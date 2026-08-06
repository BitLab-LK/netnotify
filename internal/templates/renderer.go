package templates

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/bitlab-dev/netnotify/internal/domain"
)

type Renderer struct{ dir string }

func New(dir string) *Renderer { return &Renderer{dir: dir} }
func (r *Renderer) Render(n domain.Notification) (string, error) {
	name := string(n.Severity) + ".tmpl"
	if n.Severity == "" {
		name = "test.tmpl"
	}
	t, err := template.ParseFiles(filepath.Join(r.dir, name))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var b bytes.Buffer
	if err := t.Execute(&b, n); err != nil {
		return "", err
	}
	return b.String(), nil
}
