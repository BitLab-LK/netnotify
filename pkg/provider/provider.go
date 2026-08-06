package provider

import (
	"context"

	"github.com/bitlab-dev/netnotify/internal/domain"
)

type Notifier interface {
	Name() string
	Notify(ctx context.Context, n domain.Notification) error
	Health(ctx context.Context) error
}

type Registry struct{ providers map[string]Notifier }

func NewRegistry() *Registry                         { return &Registry{providers: map[string]Notifier{}} }
func (r *Registry) Register(n Notifier)              { r.providers[n.Name()] = n }
func (r *Registry) Get(name string) (Notifier, bool) { n, ok := r.providers[name]; return n, ok }
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.providers))
	for k := range r.providers {
		out = append(out, k)
	}
	return out
}
