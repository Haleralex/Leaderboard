package handlers

import (
	"encoding/json"
	"net/http"

	"leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/auth/service"
	authservice "leaderboard-service/internal/auth/service"
	sharedhandlers "leaderboard-service/internal/shared/handlers"
	sharedmodels "leaderboard-service/internal/shared/models"

	"github.com/rs/zerolog/log"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *authservice.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhandlers.RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" || req.Email == "" || req.Password == "" {
		sharedhandlers.RespondError(w, "name, email, and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		sharedhandlers.RespondError(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register user")
		sharedhandlers.RespondError(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	sharedhandlers.RespondJSON(w, sharedmodels.SuccessResponse{
		Success: true,
		Message: "user registered successfully",
		Data:    user,
	}, http.StatusCreated)
}

// Login handles user login
// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhandlers.RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		sharedhandlers.RespondError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	loginResp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("Login failed")
		sharedhandlers.RespondError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	sharedhandlers.RespondJSON(w, sharedmodels.SuccessResponse{
		Success: true,
		Message: "login successful",
		Data:    loginResp,
	}, http.StatusOK)
}
