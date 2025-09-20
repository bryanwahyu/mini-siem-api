package usecases

import (
    "context"
)

type ActionPort interface {
    ApplyBlock(ctx context.Context, ip string, reason string) error
    ApplyUnblock(ctx context.Context, ip string, reason string) error
}

type ApplyActionUsecase struct { Port ActionPort }

func (uc *ApplyActionUsecase) Block(ctx context.Context, ip, reason string) error {
    return uc.Port.ApplyBlock(ctx, ip, reason)
}

func (uc *ApplyActionUsecase) Unblock(ctx context.Context, ip, reason string) error {
    return uc.Port.ApplyUnblock(ctx, ip, reason)
}
