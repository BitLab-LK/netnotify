package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/domain"
	"github.com/bitlab-dev/netnotify/internal/logger"
	"github.com/bitlab-dev/netnotify/internal/queue"
	"github.com/bitlab-dev/netnotify/internal/source/netdata"
)

type Server struct {
	srv *http.Server
	q   *queue.Queue
	cfg config.Config
	log logger.Logger
}

func New(cfg config.Config, q *queue.Queue, log logger.Logger) *Server {
	s := &Server{q: q, cfg: cfg, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/v1/notify", s.notify)
	mux.HandleFunc("/v1/sources/netdata", s.netdata)
	if cfg.Server.Metrics {
		mux.HandleFunc("/metrics", s.metrics)
	}
	handler := s.auth(mux)
	s.srv = &http.Server{Addr: cfg.Server.Address, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	return s
}
func (s *Server) Start() error                       { return s.srv.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }
func (s *Server) auth(next http.Handler) http.Handler {
	if !s.cfg.Server.BasicAuth.Enabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != s.cfg.Server.BasicAuth.Username || p != s.cfg.Server.BasicAuth.Password {
			w.Header().Set("WWW-Authenticate", "Basic")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte("netnotify_up 1\n"))
}
func (s *Server) notify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var n domain.Notification
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if n.ID == "" {
		base := domain.NewNotification()
		n.ID = base.ID
		n.ReceivedAt = base.ReceivedAt
	}
	if n.Provider == "" {
		n.Provider = s.cfg.Providers.Default
	}
	accepted := s.q.Enqueue(n)
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]bool{"accepted": accepted})
}
func (s *Server) netdata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	_ = r.ParseForm()
	vals := map[string]string{}
	for k, v := range r.Form {
		if len(v) > 0 {
			vals[k] = v[0]
		}
	}
	p := netdata.Parser{DefaultRecipient: s.cfg.Sources.Netdata.DefaultRecipient, Provider: s.cfg.Sources.Netdata.Provider}
	n, err := p.Parse(vals)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if n.Provider == "" {
		n.Provider = s.cfg.Providers.Default
	}
	accepted := s.q.Enqueue(n)
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]bool{"accepted": accepted})
}
