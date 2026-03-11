package security

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	limit    rate.Limit
	burst    int
	visitors map[string]*visitor
	ttl      time.Duration
}

func NewRateLimiter(rps float64, burst int, ttl time.Duration) *RateLimiter {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &RateLimiter{
		limit:    rate.Limit(rps),
		burst:    burst,
		visitors: map[string]*visitor{},
		ttl:      ttl,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	if key == "" {
		key = "anonymous"
	}

	for visitorKey, visitor := range r.visitors {
		if now.Sub(visitor.lastSeen) > r.ttl {
			delete(r.visitors, visitorKey)
		}
	}

	current, ok := r.visitors[key]
	if !ok {
		current = &visitor{limiter: rate.NewLimiter(r.limit, r.burst), lastSeen: now}
		r.visitors[key] = current
	}
	current.lastSeen = now
	return current.limiter.Allow()
}
