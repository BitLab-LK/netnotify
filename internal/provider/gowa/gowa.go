package gowa

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bitlab-dev/netnotify/internal/config"
	"github.com/bitlab-dev/netnotify/internal/domain"
)

type Client struct {
	cfg  config.GoWAConfig
	http *http.Client
}
type sendMessageRequest struct {
	Phone          string `json:"phone"`
	Message        string `json:"message"`
	ReplyMessageID string `json:"reply_message_id,omitempty"`
	IsForwarded    bool   `json:"is_forwarded,omitempty"`
}

func New(cfg config.GoWAConfig) *Client {
	t := cfg.Timeout
	if t == 0 {
		t = 10 * time.Second
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: t, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.TLSSkipVerify}}}}
}
func (c *Client) Name() string { return "gowa" }
func (c *Client) Notify(ctx context.Context, n domain.Notification) error {
	to := n.Recipient
	if to == "" {
		to = c.cfg.Recipient
	}
	if to == "" {
		return fmt.Errorf("gowa recipient is required")
	}
	msg := n.Text
	if msg == "" {
		msg = n.Summary
	}
	if msg == "" {
		msg = n.Title
	}
	for _, m := range n.Mentions {
		if !strings.Contains(msg, "@"+m) {
			msg += " @" + m
		}
	}
	body, _ := json.Marshal(sendMessageRequest{Phone: to, Message: msg})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.BaseURL, "/")+"/send/message", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.DeviceID != "" {
		req.Header.Set("X-Device-Id", c.cfg.DeviceID)
	}
	if c.cfg.Username != "" || c.cfg.Password != "" {
		req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("gowa send failed: status=%d body=%s", resp.StatusCode, string(b))
	}
	return nil
}
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.cfg.BaseURL, "/")+"/app/devices", nil)
	if err != nil {
		return err
	}
	if c.cfg.Username != "" || c.cfg.Password != "" {
		req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("gowa health failed: status=%d", resp.StatusCode)
	}
	return nil
}
