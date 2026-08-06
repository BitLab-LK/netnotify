package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/heartbeat"
	"github.com/bitlab-dev/netnotify/internal/logger"
	"github.com/bitlab-dev/netnotify/internal/server"
)

func Run(cfg config.Config) error {
	log := logger.New(cfg.Log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	hb := heartbeat.New(cfg.Heartbeat, log)
	go hb.Start(ctx)

	srv := server.New(cfg, log)
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
