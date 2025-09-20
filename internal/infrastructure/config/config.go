package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"
)

type ServerConfig struct {
	HTTPAddr string `yaml:"http_addr"`
	APIKey   string `yaml:"api_key"`
	DryRun   bool   `yaml:"dry_run"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type MinIOConfig struct {
	Endpoint   string `yaml:"endpoint"`
	UseSSL     bool   `yaml:"use_ssl"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	Bucket     string `yaml:"bucket"`
	Prefix     string `yaml:"prefix"`
	Region     string `yaml:"region"`
	TimeoutSec int    `yaml:"timeout_sec"`
	MaxRetries int    `yaml:"max_retries"`
	SpoolDir   string `yaml:"spool_dir"`
}

type StorageConfig struct {
	DB    DBConfig    `yaml:"db"`
	MinIO MinIOConfig `yaml:"minio"`
}

type IngestSource struct {
	Type  string   `yaml:"type"`
	Units []string `yaml:"units"`
	Paths []string `yaml:"paths"`
}

type BatchConfig struct {
	Size    int `yaml:"size"`
	FlushMS int `yaml:"flush_ms"`
}

type IngestConfig struct {
	Sources []IngestSource `yaml:"sources"`
	Batch   BatchConfig    `yaml:"batch"`
}

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Value == "" {
		d.Duration = 0
		return nil
	}
	dur, err := time.ParseDuration(value.Value)
	if err != nil {
		return err
	}
	d.Duration = dur
	return nil
}

type DetectWindows struct {
	BruteForce Duration `yaml:"brute_force"`
	Flood      Duration `yaml:"flood"`
}

type DetectThresholds struct {
	SSHFailed int `yaml:"ssh_failed"`
	HTTP401   int `yaml:"http_401"`
	RPSPerIP  int `yaml:"rps_per_ip"`
}

type DetectConfig struct {
	Windows    DetectWindows    `yaml:"windows"`
	Thresholds DetectThresholds `yaml:"thresholds"`
	Enabled    []string         `yaml:"enabled"`
}

type CloudflareConfig struct {
	Enabled  bool   `yaml:"enabled"`
	APIToken string `yaml:"api_token"`
	ZoneID   string `yaml:"zone_id"`
}

type Fail2banConfig struct {
	Enabled bool `yaml:"enabled"`
}

type ActionsConfig struct {
	Default    []string         `yaml:"default"`
	Cloudflare CloudflareConfig `yaml:"cloudflare"`
	Fail2ban   Fail2banConfig   `yaml:"fail2ban"`
}

type NotifyConfig struct {
	SlackWebhook     string `yaml:"slack_webhook"`
	TelegramBotToken string `yaml:"telegram_bot_token"`
	TelegramChatID   string `yaml:"telegram_chat_id"`
}

type OpenAIConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type RulesConfig struct {
	File          string `yaml:"file"`
	JUDOLKeywords string `yaml:"judol_keywords"`
}

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Ingest  IngestConfig  `yaml:"ingest"`
	Detect  DetectConfig  `yaml:"detect"`
	Actions ActionsConfig `yaml:"actions"`
	Notify  NotifyConfig  `yaml:"notify"`
	Rules   RulesConfig   `yaml:"rules"`
	OpenAI  OpenAIConfig  `yaml:"openai"`
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		// Look for config in default places
		for _, p := range []string{
			"/etc/server-analyst/config.yaml",
			"/etc/server-analyst/config.yml",
			"./config/config.yaml",
			"./config.yaml",
		} {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}
	if path == "" {
		return cfg, nil
	}
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	// Support env var placeholders like ${ENV}
	expanded := os.ExpandEnv(string(data))
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, err
	}
	// Basic validation
	if strings.TrimSpace(cfg.Server.APIKey) == "" {
		return nil, errors.New("server.api_key must be set")
	}
	return cfg, nil
}
