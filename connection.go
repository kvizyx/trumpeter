package wera

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type (
	connState int32

	connID = uint32
)

const (
	connStateActive connState = iota
	connStateWantClose
	connStateClosed
)

var (
	defaultControlTimeout = 5 * time.Second
	closeAckTimeout       = 5 * time.Second

	ErrInactiveConn = errors.New("connection is inactive")
)

// Connection is a wrapper around raw Websocket connection.
type Connection struct {
	id connID

	// writeLocker ensures that no more than one go-routine
	// can write to Websocket connection.
	writeLocker *sync.Mutex
	wsConn      *websocket.Conn

	state    *atomic.Int32
	pinged   *atomic.Bool
	closeAck chan struct{}
}

func newConn(id connID, wsConn *websocket.Conn) Connection {
	conn := Connection{
		id:          id,
		writeLocker: &sync.Mutex{},
		wsConn:      wsConn,
		state:       &atomic.Int32{},
		pinged:      &atomic.Bool{},
		closeAck:    make(chan struct{}),
	}

	conn.setState(connStateActive)

	return conn
}

// ID returns identifier of connection on server.
func (c *Connection) ID() uint32 {
	return c.id
}

// Send sends text message to client with provided data.
func (c *Connection) Send(data []byte) error {
	return c.send(false, websocket.TextMessage, data, -1)
}

// Close closes connection with closing handshake and then delete it from server permanently.
// If closing handshake cannot be initiated, connection will be forced to close.
func (c *Connection) Close(closeCode int, reason string) error {
	if !c.isState(connStateActive) {
		return ErrInactiveConn
	}

	closeMsg := websocket.FormatCloseMessage(closeCode, reason)

	err := c.send(true, websocket.CloseMessage, closeMsg, -1)
	if err != nil {
		c.setState(connStateClosed)
		_ = c.wsConn.Close()

		return fmt.Errorf("send close message: %w", err)
	}

	c.setState(connStateWantClose)

	go func() {
		closeAckTimer := time.NewTimer(closeAckTimeout)

		defer func() {
			c.setState(connStateClosed)
			_ = c.wsConn.Close()
		}()

		for {
			select {
			case <-closeAckTimer.C:
				return
			case <-c.closeAck:
				return
			}
		}
	}()

	return nil
}

func (c *Connection) send(control bool, msgType int, data []byte, timeout time.Duration) error {
	if !c.isState(connStateActive) {
		return ErrInactiveConn
	}

	c.writeLocker.Lock()
	defer c.writeLocker.Unlock()

	if control {
		if timeout < 0 {
			timeout = defaultControlTimeout
		}
		return c.wsConn.WriteControl(msgType, data, time.Now().Add(timeout))
	}

	return c.wsConn.WriteMessage(msgType, data)
}

func (c *Connection) setState(state connState) {
	c.state.Swap(int32(state))
}

func (c *Connection) isState(state connState) bool {
	return c.state.Load() == int32(state)
}
