package heartbeat

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/logger"
)

type Service struct {
	cfg    config.HeartbeatConfig
	client *http.Client
	log    logger.Logger
}

func New(cfg config.HeartbeatConfig, log logger.Logger) *Service {
	return &Service{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

func (s *Service) Ping(ctx context.Context, targetURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		s.log.Error("failed to create heartbeat request", err, map[string]any{"url": targetURL})
		return err
	}
	req.Header.Set("User-Agent", "netnotify-heartbeat")
	resp, err := s.client.Do(req)
	if err != nil {
		s.log.Error("heartbeat ping failed", err, map[string]any{"url": targetURL})
		return err
	}
	defer resp.Body.Close()
	s.log.Info("heartbeat ping sent", map[string]any{"url": targetURL, "status": resp.StatusCode})
	return nil
}

func (s *Service) Start(ctx context.Context) {
	if !s.cfg.Enabled || s.cfg.URL == "" {
		s.log.Info("heartbeat service disabled (no ping URL configured)", nil)
		return
	}

	interval := s.cfg.Interval
	if interval < 1*time.Second {
		interval = 1 * time.Second
	}

	s.log.Info("starting heartbeat service", map[string]any{
		"url":               s.cfg.URL,
		"interval":          interval.String(),
		"send_stop_on_exit": s.cfg.SendStopOnExit,
	})

	// Initial ping on start
	_ = s.Ping(ctx, s.cfg.URL)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if s.cfg.SendStopOnExit {
				stopURL := strings.TrimRight(s.cfg.URL, "/") + "/stop"
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_ = s.Ping(stopCtx, stopURL)
				cancel()
			}
			s.log.Info("heartbeat service stopped", nil)
			return
		case <-ticker.C:
			_ = s.Ping(ctx, s.cfg.URL)
		}
	}
}
