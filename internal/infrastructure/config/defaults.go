package config

import "time"

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPAddr: ":8080",
			APIKey:   "CHANGE_ME",
			DryRun:   true,
		},
		Storage: StorageConfig{
			DB: DBConfig{
				Driver: "sqlite",
				DSN:    "file:/var/lib/server-analyst/app.db?_busy_timeout=5000",
			},
			MinIO: MinIOConfig{
				Endpoint:   "minio.local:9000",
				UseSSL:     false,
				AccessKey:  "MINIO_ACCESS_KEY",
				SecretKey:  "MINIO_SECRET_KEY",
				Bucket:     "server-analyst",
				Prefix:     "prod",
				Region:     "us-east-1",
				TimeoutSec: 10,
				MaxRetries: 5,
				SpoolDir:   "/var/lib/server-analyst/spool",
			},
		},
		Ingest: IngestConfig{
			Sources: []IngestSource{
				{Type: "journald", Units: []string{"ssh", "sshd", "nginx", "apache2"}},
				{Type: "file", Paths: []string{"/var/log/nginx/access.log", "/var/log/nginx/error.log", "/var/log/auth.log"}},
			},
			Batch: BatchConfig{Size: 2000, FlushMS: 800},
		},
		Detect: DetectConfig{
			Windows: DetectWindows{BruteForce: Duration{Duration: 15 * time.Minute}, Flood: Duration{Duration: time.Minute}},
			Thresholds: DetectThresholds{
				SSHFailed: 8,
				HTTP401:   20,
				RPSPerIP:  120,
			},
			Enabled: []string{"judol", "sqli", "xss", "traversal", "scanner", "flood", "brute"},
		},
		Actions: ActionsConfig{
			Default:    []string{"nftables", "nginx_blocklist"},
			Cloudflare: CloudflareConfig{Enabled: false, APIToken: "CF_TOKEN", ZoneID: "CF_ZONE"},
			Fail2ban:   Fail2banConfig{Enabled: true},
		},
		Notify: NotifyConfig{},
		Rules:  RulesConfig{File: "/etc/server-analyst/rules.yml", JUDOLKeywords: "/etc/server-analyst/keywords_judol.yml"},
		OpenAI: OpenAIConfig{},
	}
}
