package source

import "github.com/bitlab-dev/netnotify/internal/domain"

type Parser interface {
	Name() string
	Parse(input map[string]string) (domain.Notification, error)
}
