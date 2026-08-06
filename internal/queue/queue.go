package queue

import (
	"context"
	"sync"
	"time"

	"github.com/bitlab-dev/netnotify/internal/domain"
	"github.com/bitlab-dev/netnotify/internal/logger"
	"github.com/bitlab-dev/netnotify/pkg/provider"
)

type Queue struct {
	ch       chan domain.Notification
	dlq      []domain.Notification
	mu       sync.Mutex
	reg      *provider.Registry
	log      logger.Logger
	retries  int
	interval time.Duration
	seen     map[string]time.Time
	ttl      time.Duration
	throttle <-chan time.Time
}

func New(size, retries int, interval, ttl time.Duration, rps float64, reg *provider.Registry, log logger.Logger) *Queue {
	if rps <= 0 {
		rps = 10
	}
	return &Queue{ch: make(chan domain.Notification, size), reg: reg, log: log, retries: retries, interval: interval, seen: map[string]time.Time{}, ttl: ttl, throttle: time.Tick(time.Duration(float64(time.Second) / rps))}
}
func (q *Queue) Enqueue(n domain.Notification) bool {
	q.mu.Lock()
	key := n.DedupKey()
	if exp, ok := q.seen[key]; ok && time.Now().Before(exp) {
		q.mu.Unlock()
		return false
	}
	q.seen[key] = time.Now().Add(q.ttl)
	q.mu.Unlock()
	q.ch <- n
	return true
}
func (q *Queue) Start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go q.worker(ctx)
	}
}
func (q *Queue) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-q.ch:
			q.deliver(ctx, n)
		}
	}
}
func (q *Queue) deliver(ctx context.Context, n domain.Notification) {
	nt, ok := q.reg.Get(n.Provider)
	if !ok {
		q.dead(n)
		return
	}
	var err error
	for i := 0; i <= q.retries; i++ {
		select {
		case <-ctx.Done():
			return
		case <-q.throttle:
		}
		err = nt.Notify(ctx, n)
		if err == nil {
			return
		}
		time.Sleep(q.interval)
	}
	q.log.Error("notification failed", err, map[string]any{"provider": n.Provider})
	q.dead(n)
}
func (q *Queue) dead(n domain.Notification) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.dlq = append(q.dlq, n)
}
func (q *Queue) DeadLetters() []domain.Notification {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]domain.Notification, len(q.dlq))
	copy(out, q.dlq)
	return out
}
