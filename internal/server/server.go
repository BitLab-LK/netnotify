package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/logger"
)

type Server struct {
	srv *http.Server
	cfg config.Config
	log logger.Logger
}

func New(cfg config.Config, log logger.Logger) *Server {
	s := &Server{cfg: cfg, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	if cfg.Server.Metrics {
		mux.HandleFunc("/metrics", s.metrics)
	}
	s.srv = &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := "ok"
	if !s.cfg.Heartbeat.Enabled || s.cfg.Heartbeat.URL == "" {
		status = "disabled"
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":            status,
		"heartbeat_enabled": s.cfg.Heartbeat.Enabled,
		"heartbeat_url":     s.cfg.Heartbeat.URL,
		"interval":          s.cfg.Heartbeat.Interval.String(),
	})
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte("netnotify_up 1\nnetnotify_heartbeat_enabled 1\n"))
}
