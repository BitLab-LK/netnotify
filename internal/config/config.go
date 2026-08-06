package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Log       LogConfig
	Queue     QueueConfig
	Providers ProvidersConfig
	Sources   SourcesConfig
	Templates TemplateConfig
}
type ServerConfig struct {
	Address   string
	Metrics   bool
	BasicAuth BasicAuthConfig
}
type BasicAuthConfig struct {
	Enabled            bool
	Username, Password string
}
type LogConfig struct {
	Level, Format, File               string
	MaxSizeMB, MaxBackups, MaxAgeDays int
}
type QueueConfig struct {
	Workers, Size, Retries  int
	RetryInterval, DedupTTL time.Duration
	RatePerSecond           float64
}
type ProvidersConfig struct {
	Default string
	GoWA    GoWAConfig
}
type GoWAConfig struct {
	Enabled                                          bool
	BaseURL, Username, Password, DeviceID, Recipient string
	Group                                            bool
	Timeout                                          time.Duration
	TLSSkipVerify                                    bool
}
type SourcesConfig struct{ Netdata NetdataConfig }
type NetdataConfig struct{ DefaultRecipient, Provider string }
type TemplateConfig struct{ Directory string }

func Load(path string) (Config, error) {
	c := defaults()
	if path != "" {
		if err := loadFile(&c, path); err != nil {
			return c, err
		}
	}
	env(&c)
	if c.Providers.GoWA.Enabled && c.Providers.GoWA.BaseURL == "" {
		return c, fmt.Errorf("providers.gowa.base_url is required when gowa is enabled")
	}
	return c, nil
}
func defaults() Config {
	return Config{Server: ServerConfig{Address: ":8080", Metrics: true}, Log: LogConfig{Level: "info", Format: "json", MaxSizeMB: 50, MaxBackups: 5, MaxAgeDays: 30}, Queue: QueueConfig{Workers: 4, Size: 1000, Retries: 3, RetryInterval: 5 * time.Second, DedupTTL: 5 * time.Minute, RatePerSecond: 10}, Providers: ProvidersConfig{Default: "gowa", GoWA: GoWAConfig{Timeout: 10 * time.Second}}, Sources: SourcesConfig{Netdata: NetdataConfig{Provider: "gowa"}}, Templates: TemplateConfig{Directory: "templates"}}
}
func env(c *Config) {
	setS(&c.Server.Address, "NETNOTIFY_SERVER_ADDRESS")
	setS(&c.Providers.GoWA.BaseURL, "NETNOTIFY_PROVIDERS_GOWA_BASE_URL")
	setS(&c.Providers.GoWA.Username, "NETNOTIFY_PROVIDERS_GOWA_USERNAME")
	setS(&c.Providers.GoWA.Password, "NETNOTIFY_PROVIDERS_GOWA_PASSWORD")
	setS(&c.Providers.GoWA.DeviceID, "NETNOTIFY_PROVIDERS_GOWA_DEVICE_ID")
	setS(&c.Providers.GoWA.Recipient, "NETNOTIFY_PROVIDERS_GOWA_RECIPIENT")
	setS(&c.Sources.Netdata.DefaultRecipient, "NETNOTIFY_SOURCES_NETDATA_DEFAULT_RECIPIENT")
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
	case "log..level":
		c.Log.Level = v
	case "providers..default":
		c.Providers.Default = v
	case "providers.gowa.enabled":
		c.Providers.GoWA.Enabled = parseB(v)
	case "providers.gowa.base_url":
		c.Providers.GoWA.BaseURL = v
	case "providers.gowa.username":
		c.Providers.GoWA.Username = v
	case "providers.gowa.password":
		c.Providers.GoWA.Password = v
	case "providers.gowa.device_id":
		c.Providers.GoWA.DeviceID = v
	case "providers.gowa.recipient":
		c.Providers.GoWA.Recipient = v
	case "providers.gowa.group":
		c.Providers.GoWA.Group = parseB(v)
	case "providers.gowa.timeout":
		c.Providers.GoWA.Timeout = parseD(v, c.Providers.GoWA.Timeout)
	case "providers.gowa.tls_skip_verify":
		c.Providers.GoWA.TLSSkipVerify = parseB(v)
	case "sources.netdata.provider":
		c.Sources.Netdata.Provider = v
	case "sources.netdata.default_recipient":
		c.Sources.Netdata.DefaultRecipient = v
	case "templates..directory":
		c.Templates.Directory = v
	case "queue..workers":
		c.Queue.Workers = parseI(v, c.Queue.Workers)
	case "queue..size":
		c.Queue.Size = parseI(v, c.Queue.Size)
	case "queue..retries":
		c.Queue.Retries = parseI(v, c.Queue.Retries)
	case "queue..retry_interval":
		c.Queue.RetryInterval = parseD(v, c.Queue.RetryInterval)
	case "queue..dedup_ttl":
		c.Queue.DedupTTL = parseD(v, c.Queue.DedupTTL)
	}
}
func parseB(v string) bool { b, _ := strconv.ParseBool(v); return b }
func parseI(v string, d int) int {
	i, e := strconv.Atoi(v)
	if e != nil {
		return d
	}
	return i
}
func parseD(v string, d time.Duration) time.Duration {
	x, e := time.ParseDuration(v)
	if e != nil {
		return d
	}
	return x
}
