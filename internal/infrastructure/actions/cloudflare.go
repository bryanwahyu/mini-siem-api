package actions

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Cloudflare struct {
    Enabled  bool
    APIToken string
    ZoneID   string
    DryRun   bool
}

type cfAccessRule struct {
    ID     string `json:"id,omitempty"`
    Mode   string `json:"mode"`
    Notes  string `json:"notes,omitempty"`
    Configuration struct {
        Target string `json:"target"`
        Value  string `json:"value"`
    } `json:"configuration"`
}

func (c *Cloudflare) client() *http.Client { return &http.Client{Timeout: 10 * time.Second} }

func (c *Cloudflare) apply(ctx context.Context, mode, ip, reason string) error {
    if !c.Enabled || c.APIToken == "" || c.ZoneID == "" { return nil }
    if c.DryRun { return nil }
    body := cfAccessRule{Mode: mode, Notes: reason}
    body.Configuration.Target = "ip"
    body.Configuration.Value = ip
    b, _ := json.Marshal(body)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/firewall/access_rules/rules", c.ZoneID), bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+c.APIToken)
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.client().Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 { return fmt.Errorf("cloudflare status %d", resp.StatusCode) }
    return nil
}

func (c *Cloudflare) Block(ctx context.Context, ip, reason string) error { return c.apply(ctx, "block", ip, reason) }

func (c *Cloudflare) Unblock(ctx context.Context, ip, reason string) error {
    if !c.Enabled || c.APIToken == "" || c.ZoneID == "" { return nil }
    if c.DryRun { return nil }
    // Find rules matching IP then delete
    url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/firewall/access_rules/rules?configuration_target=ip&configuration_value=%s", c.ZoneID, ip)
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    req.Header.Set("Authorization", "Bearer "+c.APIToken)
    resp, err := c.client().Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    var out struct{ Result []cfAccessRule `json:"result"` }
    _ = json.NewDecoder(resp.Body).Decode(&out)
    for _, r := range out.Result {
        delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/firewall/access_rules/rules/%s", c.ZoneID, r.ID), nil)
        delReq.Header.Set("Authorization", "Bearer "+c.APIToken)
        _, _ = c.client().Do(delReq)
    }
    return nil
}

// Implement ActionPort-style methods
func (c *Cloudflare) ApplyBlock(ctx context.Context, ip, reason string) error { return c.Block(ctx, ip, reason) }
func (c *Cloudflare) ApplyUnblock(ctx context.Context, ip, reason string) error { return c.Unblock(ctx, ip, reason) }
