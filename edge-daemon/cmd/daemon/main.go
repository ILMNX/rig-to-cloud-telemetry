package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"rigtelemetry/edge/internal/config"
	"rigtelemetry/edge/internal/pipeline"
	"rigtelemetry/edge/internal/storage"
	edgesync "rigtelemetry/edge/internal/sync"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	store, err := storage.Open(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	defer store.Close()

	pub, err := edgesync.NewPublisher(cfg.MQTTBroker, cfg.MQTTClientID, cfg.MQTTQoS)
	if err != nil {
		log.Fatalf("mqtt: %v", err)
	}
	defer pub.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("edge-daemon: well=%s port=%s broker=%s", cfg.WellID, cfg.SerialPort, cfg.MQTTBroker)
	pl := pipeline.New(cfg, store, pub)
	if err := pl.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("pipeline: %v", err)
	}
	log.Println("edge-daemon: stopped")
}
