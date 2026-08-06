package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityClear    Severity = "clear"
	SeverityInfo     Severity = "info"
)

type Notification struct {
	ID          string            `json:"id"`
	Source      string            `json:"source"`
	Provider    string            `json:"provider"`
	Severity    Severity          `json:"severity"`
	Status      string            `json:"status"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary"`
	Text        string            `json:"text"`
	Recipient   string            `json:"recipient"`
	Group       bool              `json:"group"`
	Mentions    []string          `json:"mentions"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Fingerprint string            `json:"fingerprint"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      *time.Time        `json:"ends_at,omitempty"`
	ReceivedAt  time.Time         `json:"received_at"`
}

func NewNotification() Notification {
	now := time.Now().UTC()
	return Notification{ID: newID(), ReceivedAt: now, StartsAt: now, Labels: map[string]string{}, Annotations: map[string]string{}}
}
func (n Notification) DedupKey() string {
	if n.Fingerprint != "" {
		return n.Fingerprint
	}
	h := sha256.Sum256([]byte(strings.Join([]string{n.Source, string(n.Severity), n.Status, n.Title, n.Recipient}, "|")))
	return hex.EncodeToString(h[:])
}
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" + hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" + hex.EncodeToString(b[10:])
}
