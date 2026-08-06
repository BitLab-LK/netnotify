package netdata

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bitlab-dev/netnotify/internal/domain"
)

type Parser struct {
	DefaultRecipient string
	Provider         string
}

func (p Parser) Name() string { return "netdata" }
func (p Parser) Parse(in map[string]string) (domain.Notification, error) {
	n := domain.NewNotification()
	n.Source = "netdata"
	n.Provider = p.Provider
	n.Recipient = p.DefaultRecipient
	n.Title = first(in, "alarm", "ALARM", "name", "NAME")
	n.Status = strings.ToLower(first(in, "status", "STATUS"))
	n.Summary = first(in, "summary", "SUMMARY", "info", "INFO")
	n.Recipient = firstNonEmpty(first(in, "recipient", "to", "NETNOTIFY_RECIPIENT"), n.Recipient)
	chart := first(in, "chart", "CHART")
	family := first(in, "family", "FAMILY")
	host := first(in, "hostname", "HOSTNAME", "host", "HOST")
	n.Labels["chart"] = chart
	n.Labels["family"] = family
	n.Labels["host"] = host
	n.Labels["unique_id"] = first(in, "unique_id", "UNIQUE_ID")
	n.Fingerprint = strings.Join([]string{"netdata", host, chart, family, n.Title}, "|")
	if n.Title == "" {
		return n, fmt.Errorf("netdata alarm name is required")
	}
	switch n.Status {
	case "critical":
		n.Severity = domain.SeverityCritical
	case "warning":
		n.Severity = domain.SeverityWarning
	case "clear", "cleared":
		n.Severity = domain.SeverityClear
	default:
		n.Severity = domain.SeverityInfo
	}
	if v := first(in, "when", "WHEN", "date", "DATE"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			n.StartsAt = time.Unix(ts, 0).UTC()
		}
	}
	if n.Summary == "" {
		n.Summary = fmt.Sprintf("Netdata alarm %s is %s on %s", n.Title, n.Status, host)
	}
	return n, nil
}
func first(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	return ""
}
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
