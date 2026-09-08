package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rigtelemetry/cloud/internal/api"
	"rigtelemetry/cloud/internal/broker"
	"rigtelemetry/cloud/internal/config"
	"rigtelemetry/cloud/internal/hub"
	"rigtelemetry/cloud/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	h := hub.New()
	sub, err := broker.NewSubscriber(cfg.MQTTBroker, cfg.MQTTClientID, cfg.MQTTQoS, st, h)
	if err != nil {
		log.Fatalf("broker: %v", err)
	}
	defer sub.Close()

	srv := api.NewServer(st, h, sub, cfg.CORSOrigins)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("cloud-backend: listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	log.Println("cloud-backend: stopped")
}
