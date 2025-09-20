package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GamePickRepository struct {
	db *pgxpool.Pool
}

type GamePick struct {
	ID                string
	LeagueSeasonID    string
	SportSeasonWeekID string
	UserID            string
	SelectedTeamName  string
	OpponentTeamName  string
	SpreadSelection   int32
	SpreadResult      int32
	PointsAssigned    int32
	IsFinalized       bool
	OddsGameID        *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UserSeasonPoints struct {
	UserID    string
	FirstName string
	LastName  string
	Email     string
	Points    int32
}

func NewGamePickRepository(db *pgxpool.Pool) *GamePickRepository {
	return &GamePickRepository{db: db}
}

// GetUnfinalizedPicks retrieves all game picks that are not finalized but have an odds_game_id
func (r *GamePickRepository) GetUnfinalizedPicks(ctx context.Context) ([]GamePick, error) {
	query := `
		SELECT id, league_season_id, sport_season_week_id, user_id,
		       selected_team_name, opponent_team_name, spread_selection,
		       spread_result, points_assigned, is_finalized, odds_game_id,
		       created_at, updated_at
		FROM game_pick
		WHERE is_finalized = false AND odds_game_id IS NOT NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unfinalized picks: %w", err)
	}
	defer rows.Close()

	var picks []GamePick
	for rows.Next() {
		var pick GamePick
		err := rows.Scan(
			&pick.ID,
			&pick.LeagueSeasonID,
			&pick.SportSeasonWeekID,
			&pick.UserID,
			&pick.SelectedTeamName,
			&pick.OpponentTeamName,
			&pick.SpreadSelection,
			&pick.SpreadResult,
			&pick.PointsAssigned,
			&pick.IsFinalized,
			&pick.OddsGameID,
			&pick.CreatedAt,
			&pick.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pick: %w", err)
		}
		picks = append(picks, pick)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating picks: %w", err)
	}

	return picks, nil
}

// UpdateGamePickResult updates the spread result and finalizes the game pick
func (r *GamePickRepository) UpdateGamePickResult(ctx context.Context, pickID string, spreadResult int32, pointsAssigned int32) error {
	query := `
		UPDATE game_pick
		SET spread_result = $2, points_assigned = $3, is_finalized = true, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, pickID, spreadResult, pointsAssigned)
	if err != nil {
		return fmt.Errorf("failed to update game pick result: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no game pick found with id %s", pickID)
	}

	return nil
}

// GetUserSeasonPoints calculates the total points for all users in a given season
func (r *GamePickRepository) GetUserSeasonPoints(ctx context.Context, leagueSeasonID string) ([]UserSeasonPoints, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email,
		       COALESCE(SUM(gp.points_assigned), 0) as total_points
		FROM app_user u
		LEFT JOIN game_pick gp ON u.id = gp.user_id
		    AND gp.league_season_id = $1
		    AND gp.is_finalized = true
		    AND gp.spread_result > 0  -- Only count winning picks
		GROUP BY u.id, u.first_name, u.last_name, u.email
		ORDER BY total_points DESC
	`

	rows, err := r.db.Query(ctx, query, leagueSeasonID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user season points: %w", err)
	}
	defer rows.Close()

	var userPoints []UserSeasonPoints
	for rows.Next() {
		var up UserSeasonPoints
		err := rows.Scan(
			&up.UserID,
			&up.FirstName,
			&up.LastName,
			&up.Email,
			&up.Points,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user points: %w", err)
		}
		userPoints = append(userPoints, up)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user points: %w", err)
	}

	return userPoints, nil
}

// GetPickByOddsGameID retrieves a specific pick by its odds game ID for a user
func (r *GamePickRepository) GetPickByOddsGameID(ctx context.Context, oddsGameID string, userID string) (*GamePick, error) {
	query := `
		SELECT id, league_season_id, sport_season_week_id, user_id,
		       selected_team_name, opponent_team_name, spread_selection,
		       spread_result, points_assigned, is_finalized, odds_game_id,
		       created_at, updated_at
		FROM game_pick
		WHERE odds_game_id = $1 AND user_id = $2
	`

	var pick GamePick
	err := r.db.QueryRow(ctx, query, oddsGameID, userID).Scan(
		&pick.ID,
		&pick.LeagueSeasonID,
		&pick.SportSeasonWeekID,
		&pick.UserID,
		&pick.SelectedTeamName,
		&pick.OpponentTeamName,
		&pick.SpreadSelection,
		&pick.SpreadResult,
		&pick.PointsAssigned,
		&pick.IsFinalized,
		&pick.OddsGameID,
		&pick.CreatedAt,
		&pick.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pick by odds game ID: %w", err)
	}

	return &pick, nil
}
