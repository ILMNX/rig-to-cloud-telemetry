package hub

import (
	"sync"

	"rigtelemetry/shared/model"
)

// Hub fans out telemetry points to WebSocket subscribers by well ID.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan model.TelemetryPoint]struct{}
}

// New creates an empty hub.
func New() *Hub {
	return &Hub{subs: make(map[string]map[chan model.TelemetryPoint]struct{})}
}

// Subscribe registers a buffered channel for wellID. Caller must Unsubscribe.
func (h *Hub) Subscribe(wellID string) chan model.TelemetryPoint {
	ch := make(chan model.TelemetryPoint, 64)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[wellID] == nil {
		h.subs[wellID] = make(map[chan model.TelemetryPoint]struct{})
	}
	h.subs[wellID][ch] = struct{}{}
	return ch
}

// Unsubscribe removes ch and closes it.
func (h *Hub) Unsubscribe(wellID string, ch chan model.TelemetryPoint) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.subs[wellID]; ok {
		if _, exists := set[ch]; exists {
			delete(set, ch)
			close(ch)
		}
		if len(set) == 0 {
			delete(h.subs, wellID)
		}
	}
}

// Publish sends point to all subscribers of its well (non-blocking per subscriber).
func (h *Hub) Publish(point model.TelemetryPoint) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[point.WellID] {
		select {
		case ch <- point:
		default:
			// Drop if subscriber is slow to avoid blocking ingest.
		}
	}
}
