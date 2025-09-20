package actions

import (
    "context"
    "os/exec"
    "server-analyst/internal/infrastructure/observability"
)

type Adapter struct{ DryRun bool }

func NewAdapter(dry bool) *Adapter { return &Adapter{DryRun: dry} }

func (a *Adapter) ApplyBlock(ctx context.Context, ip string, reason string) error {
    if a.DryRun { observability.Logger.Info().Str("ip", ip).Msg("DRY-RUN block"); return nil }
    // Try nftables then ufw
    if err := exec.CommandContext(ctx, "nft", "add", "element", "inet", "filter", "blacklist", "{", ip, "}").Run(); err == nil { return nil }
    return exec.CommandContext(ctx, "ufw", "deny", "from", ip).Run()
}

func (a *Adapter) ApplyUnblock(ctx context.Context, ip string, reason string) error {
    if a.DryRun { observability.Logger.Info().Str("ip", ip).Msg("DRY-RUN unblock"); return nil }
    if err := exec.CommandContext(ctx, "nft", "delete", "element", "inet", "filter", "blacklist", "{", ip, "}").Run(); err == nil { return nil }
    return exec.CommandContext(ctx, "ufw", "delete", "deny", "from", ip).Run()
}
