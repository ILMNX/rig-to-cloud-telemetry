package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"rigtelemetry/cloud/internal/broker"
	"rigtelemetry/cloud/internal/hub"
	"rigtelemetry/cloud/internal/store"
	"rigtelemetry/shared/model"
)

// Server serves REST and WebSocket APIs.
type Server struct {
	store      *store.Store
	hub        *hub.Hub
	broker     *broker.Subscriber
	corsOrigins map[string]struct{}
}

// NewServer constructs the HTTP API server.
func NewServer(st *store.Store, h *hub.Hub, sub *broker.Subscriber, corsOrigins []string) *Server {
	origins := make(map[string]struct{}, len(corsOrigins))
	for _, o := range corsOrigins {
		origins[o] = struct{}{}
	}
	return &Server{store: st, hub: h, broker: sub, corsOrigins: origins}
}

// Router returns the Chi router.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(s.cors)

	r.Get("/api/v1/health", s.handleHealth)
	r.Get("/api/v1/wells", s.handleWells)
	r.Get("/api/v1/wells/{wellID}/latest", s.handleLatest)
	r.Get("/api/v1/wells/{wellID}/telemetry", s.handleTelemetry)
	r.Get("/api/v1/ws/telemetry", s.handleWS)
	return r
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := s.corsOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	dbStatus := "up"
	if err := s.store.Ping(r.Context()); err != nil {
		dbStatus = "down"
	}
	mqttStatus := "down"
	if s.broker != nil && s.broker.Connected() {
		mqttStatus = "up"
	}
	status := "ok"
	if dbStatus != "up" {
		status = "degraded"
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": status,
		"db":     dbStatus,
		"mqtt":   mqttStatus,
	})
}

func (s *Server) handleWells(w http.ResponseWriter, r *http.Request) {
	wells, err := s.store.ListWells(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wells": wells})
}

func (s *Server) handleLatest(w http.ResponseWriter, r *http.Request) {
	wellID := chi.URLParam(r, "wellID")
	pt, err := s.store.Latest(r.Context(), wellID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if pt == nil {
		writeError(w, http.StatusNotFound, "not_found", "no telemetry for well")
		return
	}
	writeJSON(w, http.StatusOK, pt)
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	wellID := chi.URLParam(r, "wellID")
	now := time.Now().UTC()
	from := now.Add(-1 * time.Hour)
	to := now
	limit := 5000

	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid from (RFC3339)")
			return
		}
		from = t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid to (RFC3339)")
			return
		}
		to = t
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid limit")
			return
		}
		limit = n
	}

	points, err := s.store.Range(r.Context(), wellID, from, to, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if points == nil {
		points = []model.TelemetryPoint{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"well_id": wellID,
		"from":    from,
		"to":      to,
		"points":  points,
	})
}
