package logger

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
)

type Logger struct {
	l     *log.Logger
	level string
}

func New(c config.LogConfig) Logger {
	var w io.Writer = os.Stdout
	if c.File != "" {
		f, err := os.OpenFile(c.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			w = f
		}
	}
	return Logger{l: log.New(w, "", 0), level: strings.ToLower(c.Level)}
}
func (l Logger) Info(msg string, fields map[string]any) { l.write("info", msg, fields) }
func (l Logger) Error(msg string, err error, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.write("error", msg, fields)
}
func (l Logger) write(level, msg string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["level"] = level
	fields["message"] = msg
	fields["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	b, _ := json.Marshal(fields)
	l.l.Println(string(b))
}
