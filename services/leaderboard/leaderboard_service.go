package leaderboard

import (
	"apollo/db"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LeaderboardService struct {
	gamePickRepo *db.GamePickRepository
	dbPool       *pgxpool.Pool
	ctx          context.Context
}

type LeaderboardEntry struct {
	UserID       string  `json:"user_id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        string  `json:"email"`
	TotalPoints  int32   `json:"total_points"`
	Rank         int32   `json:"rank"`
	TotalPicks   int32   `json:"total_picks"`
	WinningPicks int32   `json:"winning_picks"`
	WinRate      float64 `json:"win_rate"`
}

type SeasonLeaderboard struct {
	LeagueSeasonID string             `json:"league_season_id"`
	Entries        []LeaderboardEntry `json:"entries"`
	TotalUsers     int32              `json:"total_users"`
	GeneratedAt    string             `json:"generated_at"`
}

func NewLeaderboardService(dbPool *pgxpool.Pool) *LeaderboardService {
	return &LeaderboardService{
		gamePickRepo: db.NewGamePickRepository(dbPool),
		dbPool:       dbPool,
		ctx:          context.Background(),
	}
}

// GetSeasonLeaderboard calculates and returns the leaderboard for a specific league season
func (ls *LeaderboardService) GetSeasonLeaderboard(leagueSeasonID string) (*SeasonLeaderboard, error) {
	// Get user season points from the repository
	userPoints, err := ls.gamePickRepo.GetUserSeasonPoints(ls.ctx, leagueSeasonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user season points: %w", err)
	}

	// Get detailed pick statistics for each user
	leaderboardEntries := make([]LeaderboardEntry, 0, len(userPoints))

	for _, up := range userPoints {
		stats, err := ls.getUserPickStats(leagueSeasonID, up.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get pick stats for user %s: %w", up.UserID, err)
		}

		entry := LeaderboardEntry{
			UserID:       up.UserID,
			FirstName:    up.FirstName,
			LastName:     up.LastName,
			Email:        up.Email,
			TotalPoints:  up.Points,
			TotalPicks:   stats.TotalPicks,
			WinningPicks: stats.WinningPicks,
			WinRate:      stats.WinRate,
		}

		leaderboardEntries = append(leaderboardEntries, entry)
	}

	// Assign ranks (already sorted by points descending from repository)
	for i := range leaderboardEntries {
		leaderboardEntries[i].Rank = int32(i + 1)
	}

	return &SeasonLeaderboard{
		LeagueSeasonID: leagueSeasonID,
		Entries:        leaderboardEntries,
		TotalUsers:     int32(len(leaderboardEntries)),
		GeneratedAt:    fmt.Sprintf("%d", context.Background()),
	}, nil
}

type UserPickStats struct {
	TotalPicks   int32
	WinningPicks int32
	WinRate      float64
}

// getUserPickStats gets detailed pick statistics for a user in a season
func (ls *LeaderboardService) getUserPickStats(leagueSeasonID, userID string) (*UserPickStats, error) {
	query := `
		SELECT
			COUNT(*) as total_picks,
			COUNT(CASE WHEN spread_result > 0 THEN 1 END) as winning_picks
		FROM game_pick
		WHERE league_season_id = $1 AND user_id = $2 AND is_finalized = true
	`

	var totalPicks, winningPicks int32
	err := ls.dbPool.QueryRow(ls.ctx, query, leagueSeasonID, userID).Scan(&totalPicks, &winningPicks)
	if err != nil {
		return nil, fmt.Errorf("failed to query pick stats: %w", err)
	}

	winRate := float64(0)
	if totalPicks > 0 {
		winRate = float64(winningPicks) / float64(totalPicks)
	}

	return &UserPickStats{
		TotalPicks:   totalPicks,
		WinningPicks: winningPicks,
		WinRate:      winRate,
	}, nil
}

// GetUserRank returns the rank of a specific user in the leaderboard
func (ls *LeaderboardService) GetUserRank(leagueSeasonID, userID string) (*LeaderboardEntry, error) {
	leaderboard, err := ls.GetSeasonLeaderboard(leagueSeasonID)
	if err != nil {
		return nil, err
	}

	for _, entry := range leaderboard.Entries {
		if entry.UserID == userID {
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("user not found in leaderboard")
}

// GetTopUsers returns the top N users from the leaderboard
func (ls *LeaderboardService) GetTopUsers(leagueSeasonID string, limit int) ([]LeaderboardEntry, error) {
	leaderboard, err := ls.GetSeasonLeaderboard(leagueSeasonID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(leaderboard.Entries) {
		limit = len(leaderboard.Entries)
	}

	return leaderboard.Entries[:limit], nil
}
