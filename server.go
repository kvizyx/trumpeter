package wera

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultReadLimit int64 = 1024 // 1KB
)

var (
	ErrNoPubSub = errors.New("pub-sub implementation is not provided")
)

var expectedCloseErrors = []int{
	websocket.CloseNormalClosure,
	websocket.CloseGoingAway,
	websocket.CloseProtocolError,
	websocket.CloseUnsupportedData,
	websocket.CloseNoStatusReceived,
	websocket.CloseAbnormalClosure,
	websocket.CloseInvalidFramePayloadData,
	websocket.ClosePolicyViolation,
	websocket.CloseMessageTooBig,
	websocket.CloseMandatoryExtension,
	websocket.CloseInternalServerErr,
	websocket.CloseServiceRestart,
	websocket.CloseTryAgainLater,
	websocket.CloseTLSHandshake,
}

type Server struct {
	Upgrader websocket.Upgrader

	ps        PubSub
	psChannel string

	idOffset atomic.Uint64

	onConn       OnConnect
	onLocalMsg   OnLocalMessage
	onPubSubMsg  OnPubSubMessage
	onDisconnect OnDisconnect
	onErr        OnError

	connections       map[connID]*Connection
	connectionsLocker *sync.RWMutex

	pingInterval time.Duration
	readLimit    int64
	shutdown     chan struct{}
}

func NewServer(opts ...ServerOption) *Server {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}

	server := &Server{
		Upgrader:          upgrader,
		connections:       make(map[connID]*Connection),
		connectionsLocker: &sync.RWMutex{},
		pingInterval:      -1, // if ping interval not set, clients don't need to send pings
		readLimit:         defaultReadLimit,
		shutdown:          make(chan struct{}),
	}

	for _, opt := range opts {
		opt(server)
	}

	go server.listenPubSubBroadcasts()

	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	cid := s.getConnID()
	srvConn := newConn(cid, wsConn)

	wsConn.SetReadLimit(s.readLimit)
	wsConn.SetCloseHandler(func(code int, reason string) error {
		return s.handleConnClose(srvConn, code, reason)
	})

	s.connectionsLocker.Lock()
	s.connections[cid] = &srvConn
	s.connectionsLocker.Unlock()

	go s.serveConn(srvConn)

	if s.onConn != nil {
		s.onConn(srvConn)
	}
}

func (s *Server) Shutdown() {
	s.connectionsLocker.RLock()
	defer s.connectionsLocker.RUnlock()

	for _, conn := range s.connections {
		err := conn.Close(websocket.CloseGoingAway, "server is shutting down")
		if err != nil && s.onErr != nil {
			s.onErr(*conn, err)
		}
	}

	close(s.shutdown)
}

// BroadcastGlobal broadcasts message to pub-sub subscribers (including the current instance,
// which means that message will also be broadcast locally when pub-sub listener will get it)
//
// Works only with WithPubSubBroadcast setting.
func (s *Server) BroadcastGlobal(ctx context.Context, message []byte) error {
	if s.ps == nil {
		return ErrNoPubSub
	}

	err := s.ps.Publish(ctx, s.psChannel, message)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}

func (s *Server) BroadcastLocalFilter(message []byte, filter func(c Connection) bool) {
	closeQueue := make([]*Connection, 0)

	s.connectionsLocker.RLock()
	for _, conn := range s.connections {
		if !filter(*conn) {
			continue
		}

		if err := conn.Send(message); err != nil {
			closeQueue = append(closeQueue, conn)
		}
	}
	s.connectionsLocker.RUnlock()

	if len(closeQueue) == 0 {
		return
	}

	go func() {
		for _, conn := range closeQueue {
			if conn == nil {
				continue
			}

			err := conn.Close(websocket.CloseNormalClosure, "unreachable connection")
			if err != nil && s.onErr != nil {
				s.onErr(*conn, err)
			}
		}
	}()
}

// BroadcastLocal broadcasts message to instance scoped connections without any filter.
func (s *Server) BroadcastLocal(message []byte) {
	s.blNoFilter(message)
}

// BroadcastLocalOthers broadcasts message to instance scoped connections
// except one with provided ID.
func (s *Server) BroadcastLocalOthers(message []byte, connID connID) {
	s.BroadcastLocalFilter(message, func(c Connection) bool {
		return c.id != connID
	})
}

func (s *Server) blNoFilter(message []byte) {
	s.BroadcastLocalFilter(message, func(c Connection) bool {
		return true
	})
}

func (s *Server) listenPubSubBroadcasts() {
	if s.ps == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)

	ps := s.ps.Subscribe(ctx, s.psChannel)
	defer func() {
		cancel()
		_ = ps.Close()
	}()

	messages := ps.Messages()

	for {
		select {
		case <-s.shutdown:
			return
		case msg := <-messages:
			s.onPubSubMsg(msg.Data)
			s.BroadcastLocal(msg.Data)
		}
	}
}

func (s *Server) serveConn(conn Connection) {
	defer func() {
		if s.onDisconnect != nil {
			s.onDisconnect(conn)
		}

		s.connectionsLocker.Lock()
		delete(s.connections, conn.id)
		s.connectionsLocker.Unlock()

		if conn.isState(connStateClosed) {
			return
		}

		// force connection to close
		_ = conn.wsConn.Close()
	}()

	if s.pingInterval > 0 {
		go s.startPingChecker(conn)
	}

	for {
		mt, data, err := conn.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, expectedCloseErrors...) {
				return
			}

			if s.onErr != nil {
				s.onErr(conn, err)
			}

			return
		}

		switch mt {
		case websocket.TextMessage:
			s.onLocalMsg(conn, data)

		case websocket.PingMessage:
			err = conn.send(true, websocket.PongMessage, nil, -1)
			if err != nil {
				if s.onErr != nil {
					s.onErr(conn, err)
				}
				return
			}

		case websocket.PongMessage:
			conn.pinged.Store(true)
		}
	}
}

func (s *Server) handleConnClose(conn Connection, code int, reason string) error {
	// if closing handshake was initiated by us, then just notify internal
	// connection about client acknowledgment and return.
	if conn.isState(connStateWantClose) {
		conn.closeAck <- struct{}{}
		return nil
	}

	closeMsg := websocket.FormatCloseMessage(code, reason)
	err := conn.send(true, websocket.CloseMessage, closeMsg, -1)
	if err != nil {
		if s.onErr != nil {
			s.onErr(conn, err)
		}
		return fmt.Errorf("send close ack: %w", err)
	}

	conn.setState(connStateClosed)
	err = conn.wsConn.Close()
	if err != nil && s.onErr != nil {
		s.onErr(conn, err)
	}

	return nil
}

func (s *Server) startPingChecker(conn Connection) {
	inaccuracy := 10 * time.Millisecond
	pingTime := time.NewTicker(s.pingInterval + inaccuracy)

	for {
		select {
		case <-s.shutdown:
			return
		case <-pingTime.C:
			if conn.pinged.Swap(false) {
				continue
			}

			err := conn.Close(websocket.CloseNormalClosure, "ping not received")
			if err != nil && s.onErr != nil {
				s.onErr(conn, err)
			}
			return
		}
	}
}

// getConnID increase idOffset by one and returns new id value.
func (s *Server) getConnID() (newID uint64) {
	newID = s.idOffset.Add(1)
	return newID
}
