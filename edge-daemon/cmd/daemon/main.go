package main

import (
	"fmt"
	"log"

	_ "github.com/caarlos0/env/v11"

	_ "rigtelemetry/edge/internal/serial"
	_ "rigtelemetry/edge/internal/storage"
	_ "rigtelemetry/edge/internal/sync"
	_ "rigtelemetry/edge/internal/wits"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("edge-daemon: starting (stub)")
	// TODO: wire serial reader → WITS parser → SQLite buffer → MQTT sync
}
