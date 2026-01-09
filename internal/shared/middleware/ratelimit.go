package middleware

import (
	"net/http"
	"sync"
	"time"

	"leaderboard-service/internal/shared/config"

	"golang.org/x/time/rate"
)

// RateLimiter implements per-IP rate limiting
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	limit    rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.Config) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    rate.Limit(float64(cfg.RateLimit.Requests) / float64(cfg.RateLimit.WindowSeconds)),
		burst:    cfg.RateLimit.Requests,
	}

	// Cleanup old visitors every 5 minutes
	go rl.cleanupVisitors()

	return rl
}

// Limit is the middleware handler
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			respondError(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getVisitor returns the rate limiter for the given IP
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.limit, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes old visitors
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getIP extracts the real IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
