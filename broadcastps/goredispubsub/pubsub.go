package goredispubsub

import (
	"context"

	"github.com/kvizyx/wera"
	"github.com/redis/go-redis/v9"
)

type RedisPubSub struct {
	client *redis.Client
}

var _ wera.PubSub = &RedisPubSub{}

func NewBroadcastPS(c *redis.Client) *RedisPubSub {
	return &RedisPubSub{
		client: c,
	}
}

func (rps *RedisPubSub) Publish(ctx context.Context, channel string, message []byte) error {
	cmd := rps.client.Publish(ctx, channel, message)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}

func (rps *RedisPubSub) Subscribe(ctx context.Context, channel string) wera.Subscriber {
	sub := rps.client.Subscribe(ctx, channel)
	return newSub(sub)
}
