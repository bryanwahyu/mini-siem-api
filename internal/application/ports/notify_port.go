package ports

import "context"

type Notifier interface {
    Notify(ctx context.Context, title string, text string) error
}
