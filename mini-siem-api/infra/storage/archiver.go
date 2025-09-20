package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"server-analyst/mini-siem-api/domain"
)

// Archiver is used to persist artefacts to cold storage (S3/MinIO). It is optional.
type Archiver interface {
	ArchiveEvent(ctx context.Context, event *domain.Event) error
	ArchiveDetection(ctx context.Context, detection *domain.Detection) error
}

// NoopArchiver discards archive requests.
type NoopArchiver struct{}

// ArchiveEvent implements Archiver without persistence.
func (NoopArchiver) ArchiveEvent(ctx context.Context, event *domain.Event) error { return nil }

// ArchiveDetection implements Archiver without persistence.
func (NoopArchiver) ArchiveDetection(ctx context.Context, detection *domain.Detection) error {
	return nil
}

// MinIOConfig describes required settings for archiving to MinIO/S3.
type MinIOConfig struct {
	Endpoint         string
	AccessKey        string
	SecretKey        string
	UseSSL           bool
	Bucket           string
	EventsPrefix     string
	DetectionsPrefix string
}

// MinIOArchiver streams artefacts to a MinIO bucket.
type MinIOArchiver struct {
	client           *minio.Client
	bucket           string
	eventsPrefix     string
	detectionsPrefix string
}

// NewMinIOArchiver creates an archiver when configuration is complete.
func NewMinIOArchiver(cfg MinIOConfig) (Archiver, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return NoopArchiver{}, fmt.Errorf("missing credentials for minio archiver")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	eventsPrefix := withTrailingSlash(cfg.EventsPrefix, "events/")
	detectionsPrefix := withTrailingSlash(cfg.DetectionsPrefix, "detections/")
	return &MinIOArchiver{client: client, bucket: cfg.Bucket, eventsPrefix: eventsPrefix, detectionsPrefix: detectionsPrefix}, nil
}

func withTrailingSlash(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	if !strings.HasSuffix(value, "/") {
		return value + "/"
	}
	return value
}

// ArchiveEvent uploads the event as JSON into MinIO.
func (a *MinIOArchiver) ArchiveEvent(ctx context.Context, event *domain.Event) error {
	if a == nil {
		return nil
	}
	key := fmt.Sprintf("%s%d.json", a.eventsPrefix, event.ID)
	return a.putJSON(ctx, key, event)
}

// ArchiveDetection uploads the detection as JSON into MinIO.
func (a *MinIOArchiver) ArchiveDetection(ctx context.Context, detection *domain.Detection) error {
	if a == nil {
		return nil
	}
	key := fmt.Sprintf("%s%d.json", a.detectionsPrefix, detection.ID)
	return a.putJSON(ctx, key, detection)
}

func (a *MinIOArchiver) putJSON(ctx context.Context, key string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data)
	_, err = a.client.PutObject(ctx, a.bucket, key, reader, int64(len(data)), minio.PutObjectOptions{ContentType: "application/json"})
	return err
}

// HealthCheck pings the MinIO server to ensure connectivity.
func (a *MinIOArchiver) HealthCheck(ctx context.Context) error {
	if a == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := a.client.ListBuckets(ctx)
	return err
}
