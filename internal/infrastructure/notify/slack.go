package notify

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
)

type Slack struct{ Webhook string }

func (s *Slack) Notify(ctx context.Context, title string, text string) error {
    if s.Webhook == "" { return nil }
    payload := map[string]any{"text": "["+title+"]\n"+text}
    b, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.Webhook, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    _, err := http.DefaultClient.Do(req)
    return err
}
