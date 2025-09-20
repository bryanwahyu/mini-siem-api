package repositories

import "context"

type ColdStoreRepository interface {
    Upload(ctx context.Context, objectPath string, contentType string, gzipped bool, data []byte) error
    SnapshotRules(ctx context.Context, objectPath string, yaml []byte) error
    FlushSpool(ctx context.Context) error
}
