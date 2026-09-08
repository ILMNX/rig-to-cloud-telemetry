package pipeline

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"rigtelemetry/edge/internal/config"
	"rigtelemetry/edge/internal/serial"
	"rigtelemetry/edge/internal/storage"
	edgesync "rigtelemetry/edge/internal/sync"
	"rigtelemetry/edge/internal/wits"
)

// Pipeline wires serial → WITS → SQLite outbox → MQTT sync.
type Pipeline struct {
	cfg   config.Config
	store *storage.Store
	pub   *edgesync.Publisher
}

// New constructs a pipeline from config and opened dependencies.
func New(cfg config.Config, store *storage.Store, pub *edgesync.Publisher) *Pipeline {
	return &Pipeline{cfg: cfg, store: store, pub: pub}
}

// Run starts ingest and sync until ctx is cancelled.
func (p *Pipeline) Run(ctx context.Context) error {
	lines := make(chan []byte, 256)
	parser := wits.NewParser(p.cfg.WellID)
	reader := serial.NewReader(p.cfg.SerialPort, p.cfg.SerialBaud)
	worker := edgesync.NewWorker(p.store, p.pub, p.cfg.SyncInterval, p.cfg.SyncBatch)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := reader.Run(ctx, lines); err != nil && ctx.Err() == nil {
			log.Printf("pipeline: serial stopped: %v", err)
		}
		close(lines)
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-lines:
				if !ok {
					return
				}
				for _, pt := range parser.Feed(line) {
					if err := pt.Validate(); err != nil {
						log.Printf("pipeline: skip invalid point: %v", err)
						continue
					}
					payload, err := json.Marshal(pt)
					if err != nil {
						log.Printf("pipeline: marshal: %v", err)
						continue
					}
					if err := p.store.Enqueue(pt, payload); err != nil {
						log.Printf("pipeline: enqueue: %v", err)
						continue
					}
					log.Printf("pipeline: buffered well=%s depth=%.2f", pt.WellID, pt.BitDepth)
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		worker.Run(ctx)
	}()

	wg.Wait()
	return ctx.Err()
}
