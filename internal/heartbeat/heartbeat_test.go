package heartbeat_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/heartbeat"
	"github.com/bitlab-dev/netnotify/internal/logger"
)

func TestHeartbeatPingsAndStop(t *testing.T) {
	var mu sync.Mutex
	var receivedPaths []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedPaths = append(receivedPaths, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	log := logger.New(config.LogConfig{Level: "error"})
	cfg := config.HeartbeatConfig{
		Enabled:        true,
		URL:            ts.URL + "/ping",
		Interval:       50 * time.Millisecond,
		SendStopOnExit: true,
	}

	svc := heartbeat.New(cfg, log)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.Start(ctx)
		close(done)
	}()

	time.Sleep(120 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("heartbeat service did not stop within timeout")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(receivedPaths) < 2 {
		t.Fatalf("expected at least 2 requests (initial ping + stop), got %d: %v", len(receivedPaths), receivedPaths)
	}

	hasInitial := false
	hasStop := false
	for _, p := range receivedPaths {
		if p == "/ping" {
			hasInitial = true
		}
		if p == "/ping/stop" {
			hasStop = true
		}
	}

	if !hasInitial {
		t.Errorf("expected initial ping to /ping, received: %v", receivedPaths)
	}
	if !hasStop {
		t.Errorf("expected graceful stop ping to /ping/stop, received: %v", receivedPaths)
	}
}
