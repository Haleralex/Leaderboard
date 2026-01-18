package handlers

import (
	"net/http"

	"leaderboard-service/internal/shared/database"
	"leaderboard-service/internal/shared/models"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db    *database.PostgresDB
	redis *database.RedisClient
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.PostgresDB, redis *database.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Health performs a health check
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database
	dbHealthy := true
	if err := h.db.Health(ctx); err != nil {
		dbHealthy = false
	}

	// Check Redis (optional)
	redisHealthy := true
	if h.redis != nil {
		if err := h.redis.Health(ctx); err != nil {
			redisHealthy = false
		}
	} else {
		redisHealthy = false // Redis not configured
	}

	status := http.StatusOK
	if !dbHealthy {
		status = http.StatusServiceUnavailable
	}

	respondJSON(w, map[string]interface{}{
		"status": "ok",
		"services": map[string]bool{
			"database": dbHealthy,
			"redis":    redisHealthy,
		},
	}, status)
}

// Readiness checks if the service is ready to accept traffic
// GET /ready
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check critical dependencies
	if err := h.db.Health(ctx); err != nil {
		respondError(w, "database not ready", http.StatusServiceUnavailable)
		return
	}

	respondJSON(w, models.SuccessResponse{
		Success: true,
		Message: "service is ready",
	}, http.StatusOK)
}

// Liveness checks if the service is alive
// GET /live
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, models.SuccessResponse{
		Success: true,
		Message: "service is alive",
	}, http.StatusOK)
}
