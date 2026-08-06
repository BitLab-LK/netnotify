package queue

import (
	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/domain"
	"github.com/bitlab-dev/netnotify/internal/logger"
	"github.com/bitlab-dev/netnotify/pkg/provider"
	"testing"
	"time"
)

func TestDedup(t *testing.T) {
	q := New(2, 0, time.Millisecond, time.Minute, 1, provider.NewRegistry(), logger.New(config.LogConfig{}))
	n := domain.NewNotification()
	n.Title = "a"
	if !q.Enqueue(n) {
		t.Fatal("first enqueue rejected")
	}
	if q.Enqueue(n) {
		t.Fatal("duplicate accepted")
	}
}
