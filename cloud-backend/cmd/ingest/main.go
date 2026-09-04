package main

import (
	"fmt"
	"log"

	_ "rigtelemetry/cloud/internal/api"
	_ "rigtelemetry/cloud/internal/broker"
	_ "rigtelemetry/cloud/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("cloud-backend ingest: starting (stub)")
	// TODO: wire MQTT broker → TimescaleDB store → HTTP/WebSocket API
}
