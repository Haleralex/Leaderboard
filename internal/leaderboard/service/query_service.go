package service

import (
	"context"
	"fmt"

	leaderboardmodels "leaderboard-service/internal/leaderboard/models"
	authmodels "leaderboard-service/internal/auth/models"
	"leaderboard-service/internal/shared/repository"

	"github.com/google/uuid"
)

// QueryService demonstrates Specification Pattern usage
type QueryService struct {
	userRepo  repository.UserRepository
	scoreRepo repository.ScoreRepository
}

// NewQueryService creates a new query service
func NewQueryService(userRepo repository.UserRepository, scoreRepo repository.ScoreRepository) *QueryService {
	return &QueryService{
		userRepo:  userRepo,
		scoreRepo: scoreRepo,
	}
}

// SearchUsers searches for users by name or email
func (s *QueryService) SearchUsers(ctx context.Context, query string) ([]*authmodels.User, error) {
	// Build specification: search by name OR email
	spec := repository.SearchUsersSpec(query)

	return s.userRepo.FindBySpec(ctx, spec)
}

// GetActiveUsersFromDomain gets users from specific domain, ordered by creation date
func (s *QueryService) GetActiveUsersFromDomain(ctx context.Context, domain string, limit int) ([]*authmodels.User, error) {
	// Build complex specification
	spec := repository.And(
		repository.NewUserByEmailDomainSpec(domain),
		repository.NewUserOrderBySpec("created_at", true),
		repository.NewUserLimitSpec(limit),
	)

	return s.userRepo.FindBySpec(ctx, spec)
}

// GetRecentUsers gets recently created users
func (s *QueryService) GetRecentUsers(ctx context.Context, after string, limit int) ([]*authmodels.User, error) {
	spec := repository.RecentUsersSpec(after, limit)
	return s.userRepo.FindBySpec(ctx, spec)
}

// CountUsersByDomain counts users from specific domain
func (s *QueryService) CountUsersByDomain(ctx context.Context, domain string) (int64, error) {
	spec := repository.NewUserByEmailDomainSpec(domain)
	return s.userRepo.CountBySpec(ctx, spec)
}

// GetLeaderboard gets top N scores for a season
func (s *QueryService) GetLeaderboard(ctx context.Context, season string, limit int) ([]*leaderboardmodels.Score, error) {
	spec := repository.LeaderboardSpec(season, limit)
	return s.scoreRepo.FindBySpec(ctx, spec)
}

// GetHighScores gets scores above threshold
func (s *QueryService) GetHighScores(ctx context.Context, season string, minScore int64, limit int) ([]*leaderboardmodels.Score, error) {
	spec := repository.HighScoresSpec(season, minScore, limit)
	return s.scoreRepo.FindBySpec(ctx, spec)
}

// GetPaginatedLeaderboard gets leaderboard with pagination
func (s *QueryService) GetPaginatedLeaderboard(ctx context.Context, season string, page, pageSize int) ([]*leaderboardmodels.Score, int64, error) {
	spec := repository.PaginatedLeaderboardSpec(season, page, pageSize)

	scores, err := s.scoreRepo.FindBySpec(ctx, spec)
	if err != nil {
		return nil, 0, err
	}

	// Count total for pagination
	countSpec := repository.NewScoreBySeasonSpec(season)
	total, err := s.scoreRepo.CountBySpec(ctx, countSpec)
	if err != nil {
		return nil, 0, err
	}

	return scores, total, nil
}

// GetScoresInRange gets scores within specific range
func (s *QueryService) GetScoresInRange(ctx context.Context, season string, minScore, maxScore int64) ([]*leaderboardmodels.Score, error) {
	spec := repository.MidRangeScoresSpec(season, minScore, maxScore)
	return s.scoreRepo.FindBySpec(ctx, spec)
}

// GetTopScoresWithMinValue gets top scores above threshold
func (s *QueryService) GetTopScoresWithMinValue(ctx context.Context, season string, minScore int64, topN int) ([]*leaderboardmodels.Score, error) {
	// Complex specification combining multiple conditions
	spec := repository.And(
		repository.NewScoreBySeasonSpec(season),
		repository.NewScoreMinValueSpec(minScore),
		repository.NewScoreOrderBySpec("score", true),
		repository.NewScoreLimitSpec(topN),
	)

	return s.scoreRepo.FindBySpec(ctx, spec)
}

// Advanced queries demonstrating specification composition

// FindUsersNotInDomain finds users NOT from specific domain
func (s *QueryService) FindUsersNotInDomain(ctx context.Context, domain string, limit int) ([]*authmodels.User, error) {
	// Use NOT operator
	spec := repository.And(
		repository.Not(repository.NewUserByEmailDomainSpec(domain)),
		repository.NewUserLimitSpec(limit),
	)

	return s.userRepo.FindBySpec(ctx, spec)
}

// FindScoresOutsideRange finds scores outside specific range
func (s *QueryService) FindScoresOutsideRange(ctx context.Context, season string, minScore, maxScore int64, limit int) ([]*leaderboardmodels.Score, error) {
	// Use NOT and OR for complex logic
	spec := repository.And(
		repository.NewScoreBySeasonSpec(season),
		repository.Not(repository.NewScoreRangeSpec(minScore, maxScore)),
		repository.NewScoreLimitSpec(limit),
	)

	return s.scoreRepo.FindBySpec(ctx, spec)
}

// GetLeaderboardExcludingUsers gets leaderboard excluding specific users
func (s *QueryService) GetLeaderboardExcludingUsers(ctx context.Context, season string, excludeUserIDs []string, limit int) ([]*leaderboardmodels.Score, error) {
	// First get all scores for season
	spec := repository.And(
		repository.NewScoreBySeasonSpec(season),
		repository.NewScoreOrderBySpec("score", true),
		repository.NewScoreLimitSpec(limit*2), // Get more to filter
	)

	scores, err := s.scoreRepo.FindBySpec(ctx, spec)
	if err != nil {
		return nil, err
	}

	// Filter out excluded users (in-memory)
	excludeMap := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}

	filtered := make([]*leaderboardmodels.Score, 0, limit)
	for _, score := range scores {
		if !excludeMap[score.UserID.String()] {
			filtered = append(filtered, score)
			if len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

// GetMultiSeasonLeaderboard gets combined leaderboard from multiple seasons
func (s *QueryService) GetMultiSeasonLeaderboard(ctx context.Context, seasons []string, limitPerSeason int) (map[string][]*leaderboardmodels.Score, error) {
	result := make(map[string][]*leaderboardmodels.Score)

	for _, season := range seasons {
		spec := repository.LeaderboardSpec(season, limitPerSeason)
		scores, err := s.scoreRepo.FindBySpec(ctx, spec)
		if err != nil {
			return nil, fmt.Errorf("failed to get leaderboard for season %s: %w", season, err)
		}
		result[season] = scores
	}

	return result, nil
}

// CountHighScores counts scores above threshold per season
func (s *QueryService) CountHighScores(ctx context.Context, season string, minScore int64) (int64, error) {
	spec := repository.And(
		repository.NewScoreBySeasonSpec(season),
		repository.NewScoreMinValueSpec(minScore),
	)

	return s.scoreRepo.CountBySpec(ctx, spec)
}

// GetUserRankInSeason gets user's rank in season leaderboard
func (s *QueryService) GetUserRankInSeason(ctx context.Context, userID string, season string) (int, int64, error) {
	// Get user's score
	userSpec := repository.UserScoresInSeasonSpec(
		uuid.MustParse(userID),
		season,
	)
	userScores, err := s.scoreRepo.FindBySpec(ctx, userSpec)
	if err != nil || len(userScores) == 0 {
		return 0, 0, fmt.Errorf("user score not found")
	}
	userScore := userScores[0].Score

	// Count scores above user's score
	higherScoresSpec := repository.And(
		repository.NewScoreBySeasonSpec(season),
		repository.NewScoreMinValueSpec(userScore+1),
	)
	rank, err := s.scoreRepo.CountBySpec(ctx, higherScoresSpec)
	if err != nil {
		return 0, 0, err
	}

	return int(rank) + 1, userScore, nil
}
