package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kvizyx/wera"
)

func main() {
	server := wera.NewServer(
		wera.WithPingInterval(10 * time.Second),
	)
	defer server.Shutdown()

	server.OnConnect(func(c wera.Connection) {
		log.Printf("connected: client-%d", c.ID())
	})

	server.OnError(func(c wera.Connection, err error) {
		log.Printf("error from client-%d: %s", c.ID(), err)
	})

	server.OnLocalMessage(func(c wera.Connection, msg []byte) {
		log.Printf("message from client-%d: %s", c.ID(), string(msg))

		server.BroadcastLocalOthers(msg, c.ID())
	})

	server.OnDisconnect(func(c wera.Connection) {
		log.Printf("disconnected: client-%d", c.ID())
	})

	http.HandleFunc("/ws", server.ServeHTTP)

	log.Println("server started")
	panic(http.ListenAndServe("localhost:8080", nil))
}
