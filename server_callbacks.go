package wera

type (
	// OnConnect is invoking on new client connection.
	OnConnect = func(c Connection)

	// OnLocalMessage is invoking on messages from local-connected clients except of control messages
	// as they're handling internally.
	OnLocalMessage = func(c Connection, msg []byte)

	// OnPubSubMessage is invoking on messages from Pub-Sub.
	OnPubSubMessage = func(msg []byte)

	// OnDisconnect is invoking on disconnection of client from server initiated by either peer.
	OnDisconnect = func(c Connection)

	// OnError is invoking on server errors.
	OnError = func(c Connection, err error)
)

func (s *Server) OnConnect(h OnConnect)             { s.onConn = h }
func (s *Server) OnLocalMessage(h OnLocalMessage)   { s.onLocalMsg = h }
func (s *Server) OnPubSubMessage(h OnPubSubMessage) { s.onPubSubMsg = h }
func (s *Server) OnDisconnect(h OnDisconnect)       { s.onDisconnect = h }
func (s *Server) OnError(h OnError)                 { s.onErr = h }
