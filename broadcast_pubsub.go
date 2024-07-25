package wera

import (
	"context"
	"io"
)

type PubSub interface {
	Publish(ctx context.Context, channel string, message []byte) error
	Subscribe(ctx context.Context, channel string) Subscriber
}

type Subscriber interface {
	io.Closer
	Messages() <-chan Message
}

type Message struct {
	Data []byte
}
