package config

import (
	"os"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"server-analyst/mini-siem-api/infra/database"
	"server-analyst/mini-siem-api/infra/ratelimit"
	"server-analyst/mini-siem-api/infra/storage"
)

// Config holds runtime configuration.
type Config struct {
	Server struct {
		Address     string
		AdminAPIKey string
		IngestRate  float64
		IngestBurst int
		RateTTL     time.Duration
	}
	Database database.Config
	MinIO    storage.MinIOConfig
	Logging  struct {
		Level string
	}
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Server.Address = firstNonEmpty(os.Getenv("MINI_SIEM_ADDRESS"), ":8085")
	cfg.Server.AdminAPIKey = os.Getenv("MINI_SIEM_ADMIN_API_KEY")
	cfg.Server.IngestRate = parseFloat(os.Getenv("MINI_SIEM_INGEST_RATE"), 5.0)
	cfg.Server.IngestBurst = parseInt(os.Getenv("MINI_SIEM_INGEST_BURST"), 10)
	cfg.Server.RateTTL = parseDuration(os.Getenv("MINI_SIEM_RATE_TTL"), 5*time.Minute)

	cfg.Database = database.Config{
		Driver:          firstNonEmpty(os.Getenv("MINI_SIEM_DB_DRIVER"), "sqlite"),
		DSN:             os.Getenv("MINI_SIEM_DB_DSN"),
		MaxOpenConns:    parseInt(os.Getenv("MINI_SIEM_DB_MAX_OPEN"), 10),
		MaxIdleConns:    parseInt(os.Getenv("MINI_SIEM_DB_MAX_IDLE"), 5),
		ConnMaxLifetime: parseDuration(os.Getenv("MINI_SIEM_DB_CONN_TTL"), time.Hour),
	}

	cfg.MinIO = storage.MinIOConfig{
		Endpoint:         os.Getenv("MINI_SIEM_MINIO_ENDPOINT"),
		AccessKey:        os.Getenv("MINI_SIEM_MINIO_ACCESS_KEY"),
		SecretKey:        os.Getenv("MINI_SIEM_MINIO_SECRET_KEY"),
		UseSSL:           parseBool(os.Getenv("MINI_SIEM_MINIO_SSL"), false),
		Bucket:           os.Getenv("MINI_SIEM_MINIO_BUCKET"),
		EventsPrefix:     os.Getenv("MINI_SIEM_MINIO_EVENTS_PREFIX"),
		DetectionsPrefix: os.Getenv("MINI_SIEM_MINIO_DETECTIONS_PREFIX"),
	}

	cfg.Logging.Level = firstNonEmpty(os.Getenv("MINI_SIEM_LOG_LEVEL"), "info")

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseFloat(value string, fallback float64) float64 {
	if value == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return f
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}

func parseBool(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return b
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

// BuildLimiter constructs a limiter instance from configuration.
func (c *Config) BuildLimiter() *ratelimit.Limiter {
	rateValue := c.Server.IngestRate
	if rateValue <= 0 {
		rateValue = 5.0
	}
	if c.Server.IngestBurst <= 0 {
		c.Server.IngestBurst = 10
	}
	if c.Server.RateTTL <= 0 {
		c.Server.RateTTL = 5 * time.Minute
	}
	return ratelimit.New(rate.Limit(rateValue), c.Server.IngestBurst, c.Server.RateTTL)
}

// MinIOArchiver attempts to build a MinIO archiver if configuration is complete.
func (c *Config) MinIOArchiver() storage.Archiver {
	archiver, err := storage.NewMinIOArchiver(c.MinIO)
	if err != nil {
		return storage.NoopArchiver{}
	}
	return archiver
}
