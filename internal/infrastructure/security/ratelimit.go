package security

import (
    "net"
    "net/http"
    "sync"
    "time"
)

type tokenBucket struct {
    tokens int
    last   time.Time
}

type RateLimiter struct {
    cap   int
    refill time.Duration
    mu    sync.Mutex
    m     map[string]*tokenBucket
}

func NewRateLimiter(cap int, refill time.Duration) *RateLimiter {
    return &RateLimiter{cap: cap, refill: refill, m: make(map[string]*tokenBucket)}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        if ip == "" { ip = r.RemoteAddr }
        rl.mu.Lock()
        b := rl.m[ip]
        now := time.Now()
        if b == nil { b = &tokenBucket{tokens: rl.cap, last: now}; rl.m[ip] = b }
        // refill one token per interval
        elapsed := now.Sub(b.last)
        if elapsed > rl.refill {
            n := int(elapsed / rl.refill)
            b.tokens += n
            if b.tokens > rl.cap { b.tokens = rl.cap }
            b.last = now
        }
        if b.tokens <= 0 {
            rl.mu.Unlock()
            w.WriteHeader(http.StatusTooManyRequests)
            w.Write([]byte("rate limited"))
            return
        }
        b.tokens--
        rl.mu.Unlock()
        next.ServeHTTP(w, r)
    })
}
