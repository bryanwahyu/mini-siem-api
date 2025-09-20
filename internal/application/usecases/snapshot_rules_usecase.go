package usecases

import (
    "context"
    "fmt"
    "time"
    "server-analyst/internal/domain/repositories"
)

type SnapshotRulesUsecase struct { ColdStore repositories.ColdStoreRepository }

func (uc *SnapshotRulesUsecase) Snapshot(ctx context.Context, yaml []byte) error {
    object := fmt.Sprintf("rules/snapshots/%s-rules.yml", time.Now().UTC().Format(time.RFC3339))
    return uc.ColdStore.SnapshotRules(ctx, object, yaml)
}
