package wera

import (
	"time"
)

type ServerOption func(s *Server)

// WithPubSubBroadcast enables broadcasting messages through PubSub mechanism.
// It may be helpful if you want to distribute broadcast-messages to multiple
// subscribers (e.g. different server instances).
func WithPubSubBroadcast(ps PubSub, channel string) ServerOption {
	return func(s *Server) {
		s.ps = ps
		s.psChannel = channel
	}
}

// WithPingInterval requires clients to send pings to server with given interval
// otherwise client will be disconnected.
func WithPingInterval(interval time.Duration) ServerOption {
	return func(s *Server) {
		s.pingInterval = interval
	}
}

// WithReadLimit sets a maximum size in bytes for a message to read
// from the websocket connection. If a message exceeds the limit,
// the connection will be closed.
//
// By default, read limit is 1024 bytes.
func WithReadLimit(limit int64) ServerOption {
	return func(s *Server) {
		s.readLimit = limit
	}
}
