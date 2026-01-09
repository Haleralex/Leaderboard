package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"leaderboard-service/internal/shared/config"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Requests:      5,
			WindowSeconds: 1,
		},
	}

	rateLimiter := NewRateLimiter(cfg)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rateLimiter.Limit(nextHandler)

	// Make 5 requests (should all succeed)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimiter_BlocksExcessiveRequests(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			Requests:      2,
			WindowSeconds: 1,
		},
	}

	rateLimiter := NewRateLimiter(cfg)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rateLimiter.Limit(nextHandler)

	// Make 3 requests (3rd should be rate limited)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, rr.Code, "Request %d should succeed", i+1)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Request %d should be rate limited", i+1)
		}
	}
}
