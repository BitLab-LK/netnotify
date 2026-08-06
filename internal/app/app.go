package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/logger"
	"github.com/bitlab-dev/netnotify/internal/provider/gowa"
	"github.com/bitlab-dev/netnotify/internal/queue"
	"github.com/bitlab-dev/netnotify/internal/server"
	"github.com/bitlab-dev/netnotify/pkg/provider"
)

func Run(cfg config.Config) error {
	log := logger.New(cfg.Log)
	reg := provider.NewRegistry()
	if cfg.Providers.GoWA.Enabled {
		reg.Register(gowa.New(cfg.Providers.GoWA))
	}
	q := queue.New(cfg.Queue.Size, cfg.Queue.Retries, cfg.Queue.RetryInterval, cfg.Queue.DedupTTL, cfg.Queue.RatePerSecond, reg, log)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	q.Start(ctx, cfg.Queue.Workers)
	srv := server.New(cfg, q, log)
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	log.Info("netnotify starting", map[string]any{"address": cfg.Server.Address})
	err := srv.Start()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
