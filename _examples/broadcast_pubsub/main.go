package main

import (
	"context"
	"log"
	"net/http"

	"github.com/kvizyx/wera"
	"github.com/kvizyx/wera/broadcastps/goredispubsub"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	redisBroadcast := goredispubsub.NewBroadcastPS(redisClient)

	server := wera.NewServer(
		wera.WithPubSubBroadcast(redisBroadcast, "messsages"),
	)
	defer server.Shutdown()

	server.OnConnect(func(c wera.Connection) {
		log.Printf("connected: %d", c.ID())
	})

	server.OnError(func(c wera.Connection, err error) {
		log.Printf("error from %d: %s", c.ID(), err)
	})

	server.OnPubSubMessage(func(msg []byte) {
		log.Printf("pubsub message: %s", string(msg))
	})

	server.OnLocalMessage(func(c wera.Connection, msg []byte) {
		log.Printf("local message from %d: %s", c.ID(), string(msg))

		if err := server.BroadcastGlobal(context.TODO(), msg); err != nil {
			log.Printf("broadcast error: %s", err)
		}
	})

	server.OnDisconnect(func(c wera.Connection) {
		log.Printf("disconnected: %d", c.ID())
	})

	http.HandleFunc("/ws", server.ServeHTTP)

	log.Println("server started")
	panic(http.ListenAndServe("localhost:8081", nil))
}
