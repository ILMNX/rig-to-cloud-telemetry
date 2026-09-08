package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"rigtelemetry/shared/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS origins enforced at HTTP layer for REST; WS checked below.
	},
}

type wsEnvelope struct {
	Type    string               `json:"type"`
	WellID  string               `json:"well_id,omitempty"`
	Data    *model.TelemetryPoint `json:"data,omitempty"`
	Message string               `json:"message,omitempty"`
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	wellID := r.URL.Query().Get("well_id")
	if wellID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "well_id is required")
		return
	}

	origin := r.Header.Get("Origin")
	if origin != "" {
		if _, ok := s.corsOrigins[origin]; !ok {
			writeError(w, http.StatusForbidden, "forbidden", "origin not allowed")
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("api: ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	_ = conn.WriteJSON(wsEnvelope{Type: "hello", WellID: wellID})

	ch := s.hub.Subscribe(wellID)
	defer s.hub.Unsubscribe(wellID, ch)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var env wsEnvelope
			if err := json.Unmarshal(msg, &env); err != nil {
				continue
			}
			if env.Type == "ping" {
				_ = conn.WriteJSON(wsEnvelope{Type: "pong"})
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				return
			}
		case pt, ok := <-ch:
			if !ok {
				return
			}
			point := pt
			if err := conn.WriteJSON(wsEnvelope{Type: "point", Data: &point}); err != nil {
				return
			}
		}
	}
}
