package graph

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetCurrentWeekNumber(ctx context.Context, db *pgxpool.Pool, seasonID string) (int, error) {
	var weekNumber int
	query := `
        SELECT week_number
        FROM sport_season_week
        WHERE sport_season_id = $1
          AND start_date <= CURRENT_DATE
          AND end_date >= CURRENT_DATE
        LIMIT 1;
    `
	err := db.QueryRow(ctx, query, seasonID).Scan(&weekNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no week found for current date in season %s", seasonID)
		}
		return 0, err
	}
	return weekNumber, nil
}

func GetSportSeasonId(ctx context.Context, db *pgxpool.Pool, sport string, yearStart int, yearEnd int) (string, error) {
	var sportSeasonId string
	query := `
        SELECT id
        FROM sport_season
        WHERE 
			sport = $1
			year_start = $2
			year_end = $3
        LIMIT 1;
    `
	err := db.QueryRow(ctx, query, sport, yearStart, yearEnd).Scan(&sportSeasonId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no sport season ID found for sport %s in years %d - %d", sport, yearStart, yearEnd)
		}
		return "", err
	}
	return sportSeasonId, nil
}
func CreateSportSeason(ctx context.Context, db *pgxpool.Pool, sport string, yearStart int, yearEnd int) (string, error) {
	id := uuid.New()

	_, err := db.Exec(ctx, `
		INSERT INTO sport_season (
			id, sport, year_start, year_end, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, now(), now()
		)
	`, id, sport, yearStart, yearEnd)

	if err != nil {
		return "", fmt.Errorf("failed to create season: %w", err)
	}

	return id.String(), nil
}
