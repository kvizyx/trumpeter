package wera

import (
	"context"
)

// PubSub is an interface for managing broadcasting through publish-subscribe pattern.
type PubSub interface {
	// Publish message to specified channel.
	Publish(ctx context.Context, channel string, message []byte) error

	// Subscribe for incoming messages from specified channel.
	Subscribe(ctx context.Context, channel string) Subscriber
}

// Subscriber is an PubSub subscriber.
type Subscriber interface {
	// Messages returns channel to which incoming messages from publishers
	// will be sent.
	Messages() <-chan PubSubMessage

	// Close is telling subscriber to stop receiving messages and closes channel
	// returned from Messages method.
	Close() error
}

// PubSubMessage is a message received from PubSub Subscriber.
type PubSubMessage struct {
	Data []byte
}
