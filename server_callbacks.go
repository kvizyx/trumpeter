package wera

type (
	// OnConnect invokes when new client connects to the server.
	OnConnect = func(c Connection)

	// OnLocalMessage invokes on messages from instance-scoped clients except of control messages
	// as they're handling internally.
	OnLocalMessage = func(c Connection, msg []byte)

	// OnPubSubMessage invokes on messages from Pub-Sub.
	OnPubSubMessage = func(msg []byte)

	// OnDisconnect invokes when client disconnects from server.
	OnDisconnect = func(c Connection)

	// OnError invokes on internal server errors (including connection close errors which
	// you can detect by using ErrCloseConn error in conjunction with errors.Is).
	OnError = func(c Connection, err error)
)

func (s *Server) OnConnect(h OnConnect)             { s.onConn = h }
func (s *Server) OnLocalMessage(h OnLocalMessage)   { s.onLocalMsg = h }
func (s *Server) OnPubSubMessage(h OnPubSubMessage) { s.onPubSubMsg = h }
func (s *Server) OnDisconnect(h OnDisconnect)       { s.onDisconnect = h }
func (s *Server) OnError(h OnError)                 { s.onErr = h }
