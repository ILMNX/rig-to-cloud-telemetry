package sync

import (
	"context"
	"log"
	"time"

	"rigtelemetry/edge/internal/storage"
)

// Outbox is the subset of storage used by the sync worker.
type Outbox interface {
	ClaimBatch(limit int) ([]storage.OutboxRow, error)
	MarkPublished(id int64) error
	MarkFailed(id int64) error
}

// Worker polls the outbox and publishes pending rows over MQTT.
type Worker struct {
	outbox   Outbox
	pub      *Publisher
	interval time.Duration
	batch    int
}

// NewWorker creates a store-and-forward sync worker.
func NewWorker(outbox Outbox, pub *Publisher, interval time.Duration, batch int) *Worker {
	if batch <= 0 {
		batch = 50
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &Worker{outbox: outbox, pub: pub, interval: interval, batch: batch}
}

// Run loops until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.flush(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.flush(ctx)
		}
	}
}

func (w *Worker) flush(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		batch, err := w.outbox.ClaimBatch(w.batch)
		if err != nil {
			log.Printf("sync: claim failed: %v", err)
			return
		}
		if len(batch) == 0 {
			return
		}
		for _, row := range batch {
			if err := w.pub.PublishPoint(row.Point.WellID, row.Payload); err != nil {
				log.Printf("sync: publish id=%d failed: %v", row.ID, err)
				_ = w.outbox.MarkFailed(row.ID)
				return
			}
			if err := w.outbox.MarkPublished(row.ID); err != nil {
				log.Printf("sync: mark published id=%d failed: %v", row.ID, err)
			}
		}
		if len(batch) < w.batch {
			return
		}
	}
}
