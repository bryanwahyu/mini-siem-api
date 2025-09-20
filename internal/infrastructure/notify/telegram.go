package notify

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
)

type Telegram struct{ BotToken, ChatID string }

func (t *Telegram) Notify(ctx context.Context, title string, text string) error {
    if t.BotToken == "" || t.ChatID == "" { return nil }
    msg := fmt.Sprintf("%s\n%s", title, text)
    api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", url.PathEscape(t.BotToken))
    q := url.Values{"chat_id": {t.ChatID}, "text": {msg}}
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, api+"?"+q.Encode(), nil)
    _, err := http.DefaultClient.Do(req)
    return err
}
