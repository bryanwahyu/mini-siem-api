package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Limiter provides simple keyed rate limiting with TTL eviction.
type Limiter struct {
	mu      sync.Mutex
	rate    rate.Limit
	burst   int
	ttl     time.Duration
	clients map[string]*clientLimiter
}

// New constructs a limiter for the supplied rate/burst with eviction TTL.
func New(r rate.Limit, burst int, ttl time.Duration) *Limiter {
	if burst <= 0 {
		burst = 1
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Limiter{
		rate:    r,
		burst:   burst,
		ttl:     ttl,
		clients: make(map[string]*clientLimiter),
	}
}

// Allow returns whether the action is permitted for the given key.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if key == "" {
		key = "global"
	}

	cl, ok := l.clients[key]
	if !ok {
		cl = &clientLimiter{limiter: rate.NewLimiter(l.rate, l.burst), lastSeen: now}
		l.clients[key] = cl
	}

	cl.lastSeen = now
	allowed := cl.limiter.Allow()

	l.evictExpired(now)

	return allowed
}

func (l *Limiter) evictExpired(now time.Time) {
	for key, cl := range l.clients {
		if now.Sub(cl.lastSeen) > l.ttl {
			delete(l.clients, key)
		}
	}
}
