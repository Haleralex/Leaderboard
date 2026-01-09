package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	sharedmodels "leaderboard-service/internal/shared/models"
	sharedhandlers "leaderboard-service/internal/shared/handlers"
	"leaderboard-service/internal/shared/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// LeaderboardServiceInterface defines the interface for leaderboard service
type LeaderboardServiceInterface interface {
	SubmitScore(ctx context.Context, userID uuid.UUID, req *leaderboardmodels.SubmitScoreRequest) (*leaderboardmodels.Score, error)
	GetLeaderboard(ctx context.Context, query *leaderboardmodels.LeaderboardQuery) (*leaderboardmodels.LeaderboardResponse, error)
	GetUserRank(ctx context.Context, userID uuid.UUID, season string) (*leaderboardmodels.LeaderboardEntry, error)
	BroadcastLeaderboard(ctx context.Context, season string) error
}

// LeaderboardHandler handles leaderboard endpoints
type LeaderboardHandler struct {
	leaderboardService LeaderboardServiceInterface
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler(leaderboardService LeaderboardServiceInterface) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
	}
}

// SubmitScore handles score submission
// POST /submit-score
func (h *LeaderboardHandler) SubmitScore(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		sharedhandlers.RespondError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req leaderboardmodels.SubmitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhandlers.RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate score
	if req.Score < 0 {
		sharedhandlers.RespondError(w, "score must be non-negative", http.StatusBadRequest)
		return
	}

	score, err := h.leaderboardService.SubmitScore(r.Context(), userID, &req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to submit score")
		sharedhandlers.RespondError(w, "failed to submit score", http.StatusInternalServerError)
		return
	}

	sharedhandlers.RespondJSON(w, sharedmodels.SuccessResponse{
		Success: true,
		Message: "score submitted successfully",
		Data:    score,
	}, http.StatusOK)
}

// GetLeaderboard retrieves the leaderboard with pagination
// GET /leaderboard
func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := parseLeaderboardQuery(r)

	leaderboard, err := h.leaderboardService.GetLeaderboard(r.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get leaderboard")
		sharedhandlers.RespondError(w, "failed to retrieve leaderboard", http.StatusInternalServerError)
		return
	}

	sharedhandlers.RespondJSON(w, sharedmodels.SuccessResponse{
		Success: true,
		Data:    leaderboard,
	}, http.StatusOK)
}

// GetUserRank retrieves a specific user's rank
// GET /leaderboard/user/{userID}
func (h *LeaderboardHandler) GetUserRank(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		sharedhandlers.RespondError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	season := r.URL.Query().Get("season")
	if season == "" {
		season = "global"
	}

	rank, err := h.leaderboardService.GetUserRank(r.Context(), userID, season)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user rank")
		sharedhandlers.RespondError(w, "user not found in leaderboard", http.StatusNotFound)
		return
	}

	sharedhandlers.RespondJSON(w, sharedmodels.SuccessResponse{
		Success: true,
		Data:    rank,
	}, http.StatusOK)
}

// parseLeaderboardQuery parses query parameters into LeaderboardQuery
func parseLeaderboardQuery(r *http.Request) *leaderboardmodels.LeaderboardQuery {
	params := r.URL.Query()

	// Default values
	limit := 50
	page := 0
	sortOrder := "desc"
	season := "global"

	// Parse limit
	if limitStr := params.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse page
	if pageStr := params.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}

	// Parse sort order
	if sort := params.Get("sort"); sort == "asc" || sort == "desc" {
		sortOrder = sort
	}

	// Parse season
	if s := params.Get("season"); s != "" {
		season = s
	}

	// Parse user_id filter (optional)
	var userID *uuid.UUID
	if userIDStr := params.Get("user_id"); userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			userID = &uid
		}
	}

	// Parse cursor (for cursor-based pagination)
	cursor := params.Get("cursor")

	return &leaderboardmodels.LeaderboardQuery{
		Season:    season,
		UserID:    userID,
		SortOrder: sortOrder,
		Limit:     limit,
		Page:      page,
		Cursor:    cursor,
	}
}

// TestBroadcast manually triggers a WebSocket broadcast (for testing)
// POST /test/broadcast?season=global
func (h *LeaderboardHandler) TestBroadcast(w http.ResponseWriter, r *http.Request) {
	season := r.URL.Query().Get("season")
	if season == "" {
		season = "global"
	}

	err := h.leaderboardService.BroadcastLeaderboard(r.Context(), season)
	if err != nil {
		log.Error().Err(err).Str("season", season).Msg("Failed to broadcast")
		sharedhandlers.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sharedhandlers.RespondJSON(w, map[string]string{
		"message": "Broadcast sent successfully",
		"season":  season,
	}, http.StatusOK)
}
