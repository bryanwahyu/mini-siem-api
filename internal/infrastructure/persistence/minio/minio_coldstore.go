package minio

import (
    "bytes"
    "compress/gzip"
    "context"
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"

    minio "github.com/minio/minio-go/v7"
    "github.com/rs/zerolog/log"

    "server-analyst/internal/domain/repositories"
    "server-analyst/internal/domain/entities"
    "server-analyst/internal/infrastructure/observability"
    "server-analyst/internal/infrastructure/spool"
)

type ColdStore struct {
    Client   *Client
    Bucket   string
    Prefix   string
    Region   string
    Spool    *spool.DiskSpool
    UploaderInterval time.Duration
    SpoolMeta repositories.SpoolRepository
}

var _ repositories.ColdStoreRepository = (*ColdStore)(nil)

func NewColdStore(c *Client, bucket, prefix, region, spoolDir string) *ColdStore {
    return &ColdStore{Client: c, Bucket: bucket, Prefix: strings.TrimSuffix(prefix, "/"), Region: region, Spool: spool.New(spoolDir), UploaderInterval: 30 * time.Second}
}

func (cs *ColdStore) StartBackgroundUploader(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(cs.UploaderInterval)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                if err := cs.FlushSpool(ctx); err != nil {
                    log.Warn().Err(err).Msg("spool flush error")
                }
            }
        }
    }()
}

func (cs *ColdStore) Upload(ctx context.Context, objectPath string, contentType string, gzipped bool, data []byte) error {
	// Normalize objectPath to avoid repeated "spooled_" prefixes and overly long names
	const maxObjectPathLen = 255
	if len(objectPath) > maxObjectPathLen {
		log.Warn().Str("objectPath", objectPath).Msg("objectPath too long, truncating")
		objectPath = objectPath[:maxObjectPathLen]
	}
	// Remove repeated "spooled_" prefixes if any
	for strings.HasPrefix(objectPath, "spooled_spooled_") {
		objectPath = strings.TrimPrefix(objectPath, "spooled_")
	}

	// Retry logic parameters
	const maxRetries = 3
	const baseBackoff = 500 * time.Millisecond

	// Ensure bucket with retry
	var err error
	for i := 0; i < maxRetries; i++ {
		err = cs.Client.EnsureBucket(ctx, cs.Bucket, cs.Region)
		if err == nil {
			break
		}
		log.Warn().Err(err).Int("attempt", i+1).Msg("ensure bucket failed; retrying")
		time.Sleep(baseBackoff * time.Duration(1<<i)) // exponential backoff
	}
    if err != nil {
        log.Warn().Err(err).Msg("ensure bucket failed after retries; spooling")
        observability.UploadsFailed.Inc()
        fn, e2 := cs.Spool.Enqueue("", data)
        if e2 != nil { return e2 }
        cs.saveSpoolMeta(ctx, fn, "upload", objectPath, contentType, int64(len(data)), err)
        return nil
    }

	var reader io.Reader
	var size int64
	if gzipped {
		reader = bytes.NewReader(data)
		size = int64(len(data))
	} else {
		// gzip on the fly
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write(data)
		gz.Close()
		reader = bytes.NewReader(buf.Bytes())
		size = int64(buf.Len())
		if contentType == "application/x-ndjson" {
			contentType = "application/gzip"
		}
	}
	opt := minio.PutObjectOptions{ContentType: contentType}

	// Try upload with retry
	for i := 0; i < maxRetries; i++ {
		_, err = cs.Client.MC.PutObject(ctx, cs.Bucket, filepath.Join(cs.Prefix, objectPath), reader, size, opt)
		if err == nil {
			break
		}
		log.Warn().Err(err).Str("object", objectPath).Int("attempt", i+1).Msg("upload failed; retrying")
		time.Sleep(baseBackoff * time.Duration(1<<i)) // exponential backoff
		// Reset reader for retry
		if seeker, ok := reader.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		} else {
			// recreate reader if not seekable
			if gzipped {
				reader = bytes.NewReader(data)
			} else {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				gz.Write(data)
				gz.Close()
				reader = bytes.NewReader(buf.Bytes())
			}
		}
	}
    if err != nil {
        observability.UploadsFailed.Inc()
        log.Warn().Err(err).Str("object", objectPath).Msg("upload failed after retries; spooling")
        fn, e2 := cs.Spool.Enqueue("", data)
        if e2 != nil { return e2 }
        cs.saveSpoolMeta(ctx, fn, "upload", objectPath, contentType, int64(len(data)), err)
        return nil
    }

	observability.UploadsTotal.Inc()
	return nil
}

func (cs *ColdStore) SnapshotRules(ctx context.Context, objectPath string, yaml []byte) error {
	// Normalize objectPath to avoid repeated "spooled_" prefixes and overly long names
	const maxObjectPathLen = 255
	if len(objectPath) > maxObjectPathLen {
		log.Warn().Str("objectPath", objectPath).Msg("SnapshotRules objectPath too long, truncating")
		objectPath = objectPath[:maxObjectPathLen]
	}
	for strings.HasPrefix(objectPath, "spooled_spooled_") {
		objectPath = strings.TrimPrefix(objectPath, "spooled_")
	}
	return cs.Upload(ctx, objectPath, "text/yaml", false, yaml)
}

func (cs *ColdStore) FlushSpool(ctx context.Context) error {
	files, err := cs.Spool.List()
	if err != nil {
		return err
	}
	observability.SpoolQueueSize.Set(float64(len(files)))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		// Use the spool file's basename without assuming original object name
		name := filepath.Base(f)
		// Check for suspiciously long or repetitive names before upload
		if len(name) > 255 {
			log.Warn().Str("filename", name).Msg("spool filename too long, truncating")
			name = name[:255]
		}
		for strings.HasPrefix(name, "spooled_spooled_") {
			name = strings.TrimPrefix(name, "spooled_")
		}
		if err := cs.Upload(ctx, filepath.Join("spooled", name), "application/octet-stream", true, data); err != nil {
			return err
		}
        _ = cs.Spool.Remove(f)
        if cs.SpoolMeta != nil {
            if id := extractSpoolID(name); id != "" { _ = cs.SpoolMeta.DeleteByID(ctx, id) }
        }
    }
    return nil
}

// Note: spool filenames are randomized; no need to sanitize object path for spooling.
func (cs *ColdStore) saveSpoolMeta(ctx context.Context, fn string, kind string, orig string, ctype string, size int64, cause error) {
    if cs.SpoolMeta == nil { return }
    base := filepath.Base(fn)
    id := extractSpoolID(base)
    it := &entities.SpoolItem{
        ID:           id,
        CreatedAt:    time.Now().UTC(),
        FilePath:     fn,
        Kind:         kind,
        OriginalPath: orig,
        ContentType:  ctype,
        Size:         size,
        RetryCount:   0,
    }
    if cause != nil { it.LastError = cause.Error() }
    if err := cs.SpoolMeta.Save(ctx, it); err != nil {
        log.Warn().Err(err).Msg("save spool meta failed")
    }
}

func extractSpoolID(name string) string {
    // name pattern: YYYYMMDD-HHMMSS-<hex>.part
    base := strings.TrimSuffix(name, filepath.Ext(name))
    i := strings.LastIndexByte(base, '-')
    if i < 0 || i+1 >= len(base) { return "" }
    return base[i+1:]
}
