package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	leaderboardhandler "leaderboard-service/internal/leaderboard/handler"
	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	"leaderboard-service/internal/shared/middleware"
	"leaderboard-service/internal/shared/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLeaderboardService is a mock for LeaderboardService
type MockLeaderboardService struct {
	mock.Mock
}

func (m *MockLeaderboardService) SubmitScore(ctx context.Context, userID uuid.UUID, req *leaderboardmodels.SubmitScoreRequest) (*leaderboardmodels.Score, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*leaderboardmodels.Score), args.Error(1)
}

func (m *MockLeaderboardService) GetLeaderboard(ctx context.Context, query *leaderboardmodels.LeaderboardQuery) (*leaderboardmodels.LeaderboardResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*leaderboardmodels.LeaderboardResponse), args.Error(1)
}

func (m *MockLeaderboardService) GetUserRank(ctx context.Context, userID uuid.UUID, season string) (*leaderboardmodels.LeaderboardEntry, error) {
	args := m.Called(ctx, userID, season)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*leaderboardmodels.LeaderboardEntry), args.Error(1)
}

func (m *MockLeaderboardService) BroadcastLeaderboard(ctx context.Context, season string) error {
	args := m.Called(ctx, season)
	return args.Error(0)
}

// TestSubmitScore_Success tests successful score submission
func TestSubmitScore_Success(t *testing.T) {
	mockService := new(MockLeaderboardService)
	handler := leaderboardhandler.NewLeaderboardHandler(mockService)

	userID := uuid.New()
	scoreReq := &leaderboardmodels.SubmitScoreRequest{
		Score:  1000,
		Season: "global",
	}

	expectedScore := &leaderboardmodels.Score{
		ID:     uuid.New(),
		UserID: userID,
		Score:  1000,
		Season: "global",
	}

	mockService.On("SubmitScore", mock.Anything, userID, scoreReq).Return(expectedScore, nil)

	// Create request
	body, _ := json.Marshal(scoreReq)
	req := httptest.NewRequest(http.MethodPost, "/submit-score", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add user ID to context (simulating JWT middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Call handler
	handler.SubmitScore(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.SuccessResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

// TestGetLeaderboard_Success tests successful leaderboard retrieval
func TestGetLeaderboard_Success(t *testing.T) {
	mockService := new(MockLeaderboardService)
	handler := leaderboardhandler.NewLeaderboardHandler(mockService)

	expectedResponse := &leaderboardmodels.LeaderboardResponse{
		Entries: []leaderboardmodels.LeaderboardEntry{
			{Rank: 1, UserID: uuid.New(), UserName: "Player1", Score: 1000, Season: "global"},
			{Rank: 2, UserID: uuid.New(), UserName: "Player2", Score: 800, Season: "global"},
		},
		TotalCount: 2,
		Page:       0,
		Limit:      50,
		HasNext:    false,
	}

	mockService.On("GetLeaderboard", mock.Anything, mock.AnythingOfType("*leaderboardmodels.LeaderboardQuery")).
		Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard?season=global&limit=50", nil)
	rr := httptest.NewRecorder()

	handler.GetLeaderboard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.SuccessResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

// TestGetUserRank_Success tests successful user rank retrieval
func TestGetUserRank_Success(t *testing.T) {
	mockService := new(MockLeaderboardService)
	handler := leaderboardhandler.NewLeaderboardHandler(mockService)

	userID := uuid.New()
	expectedEntry := &leaderboardmodels.LeaderboardEntry{
		Rank:     5,
		UserID:   userID,
		UserName: "TestPlayer",
		Score:    750,
		Season:   "global",
	}

	mockService.On("GetUserRank", mock.Anything, userID, "global").Return(expectedEntry, nil)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard/user/"+userID.String()+"?season=global", nil)
	rr := httptest.NewRecorder()

	// Setup Chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetUserRank(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.SuccessResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}
