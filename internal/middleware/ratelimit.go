package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-rl.window)
	
	requests := rl.requests[key]
	
	filtered := make([]time.Time, 0, len(requests))
	for _, req := range requests {
		if req.After(cutoff) {
			filtered = append(filtered, req)
		}
	}
	
	if len(filtered) >= rl.limit {
		return false
	}
	
	filtered = append(filtered, now)
	rl.requests[key] = filtered
	
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if !rl.isAllowed(key) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
