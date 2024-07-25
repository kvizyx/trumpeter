module github.com/kvizyx/wera/_examples/broadcast_pubsub

go 1.22.5

replace (
	github.com/kvizyx/wera => ../../
	github.com/kvizyx/wera/broadcastps/goredispubsub => ../../broadcastps/goredispubsub
)

require (
	github.com/kvizyx/wera v0.0.0-20240725140842-c488c30c8654
	github.com/kvizyx/wera/broadcastps/goredispubsub v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.6.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
)
