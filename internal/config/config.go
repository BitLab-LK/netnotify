package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Log       LogConfig
	Heartbeat HeartbeatConfig
}

type ServerConfig struct {
	Address string
	Metrics bool
}

type LogConfig struct {
	Level, Format, File               string
	MaxSizeMB, MaxBackups, MaxAgeDays int
}

type HeartbeatConfig struct {
	Enabled        bool
	URL            string
	Interval       time.Duration
	SendStopOnExit bool
}

func Load(path string) (Config, error) {
	c := defaults()
	if path != "" {
		if err := loadFile(&c, path); err != nil {
			return c, err
		}
	}
	env(&c)
	if c.Heartbeat.Enabled && c.Heartbeat.URL == "" {
		// If URL is empty, heartbeat is disabled automatically
		c.Heartbeat.Enabled = false
	}
	return c, nil
}

func defaults() Config {
	return Config{
		Server: ServerConfig{
			Address: ":8080",
			Metrics: true,
		},
		Log: LogConfig{
			Level:      "info",
			Format:     "json",
			MaxSizeMB:  50,
			MaxBackups: 5,
			MaxAgeDays: 30,
		},
		Heartbeat: HeartbeatConfig{
			Enabled:        true,
			Interval:       1 * time.Minute,
			SendStopOnExit: true,
		},
	}
}

func env(c *Config) {
	setS(&c.Server.Address, "NETNOTIFY_SERVER_ADDRESS")
	setS(&c.Heartbeat.URL, "NETNOTIFY_HEARTBEAT_URL")
	if v := os.Getenv("NETNOTIFY_HEARTBEAT_INTERVAL"); v != "" {
		c.Heartbeat.Interval = parseD(v, c.Heartbeat.Interval)
	}
	if v := os.Getenv("NETNOTIFY_HEARTBEAT_ENABLED"); v != "" {
		c.Heartbeat.Enabled = parseB(v)
	}
	if v := os.Getenv("NETNOTIFY_HEARTBEAT_SEND_STOP_ON_EXIT"); v != "" {
		c.Heartbeat.SendStopOnExit = parseB(v)
	}
}

func setS(p *string, k string) {
	if v := os.Getenv(k); v != "" {
		*p = v
	}
}

func loadFile(c *Config, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var section, sub string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		indent := len(sc.Text()) - len(strings.TrimLeft(sc.Text(), " "))
		kv := strings.SplitN(line, ":", 2)
		if len(kv) < 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
		val = os.ExpandEnv(val)
		if val == "" {
			if indent == 0 {
				section = key
				sub = ""
			} else if indent == 2 {
				sub = key
			}
			continue
		}
		assign(c, section, sub, key, val)
	}
	return sc.Err()
}

func assign(c *Config, s, sub, k, v string) {
	switch s + "." + sub + "." + k {
	case "server..address":
		c.Server.Address = v
	case "server..metrics":
		c.Server.Metrics = parseB(v)
	case "log..level":
		c.Log.Level = v
	case "log..format":
		c.Log.Format = v
	case "log..file":
		c.Log.File = v
	case "heartbeat..enabled":
		c.Heartbeat.Enabled = parseB(v)
	case "heartbeat..url":
		c.Heartbeat.URL = v
	case "heartbeat..interval":
		c.Heartbeat.Interval = parseD(v, c.Heartbeat.Interval)
	case "heartbeat..send_stop_on_exit":
		c.Heartbeat.SendStopOnExit = parseB(v)
	}
}

func parseB(v string) bool {
	b, _ := strconv.ParseBool(v)
	return b
}

func parseD(v string, d time.Duration) time.Duration {
	x, err := time.ParseDuration(v)
	if err != nil {
		return d
	}
	return x
}
