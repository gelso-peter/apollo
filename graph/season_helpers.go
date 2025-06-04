package graph

import (
	"apollo/graph/model"
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetCurrentWeekData(ctx context.Context, db *pgxpool.Pool, sport string, yearStart int, yearEnd int) (*model.SportSeasonWeekData, error) {
	var sportSeasonWeekData *model.SportSeasonWeekData
	query := `
        SELECT ssw.id, ssw.start_date, ssw.end_date
		FROM sport_season_week ssw
			JOIN sport_season ss ON ss.id = ssw.sport_season_id
			WHERE ss.sport = $1
  			AND CURRENT_DATE BETWEEN ssw.start_date AND ssw.end_date
			AND ss.year_start = $2
			AND ss.year_end = $3
    	`
	err := db.QueryRow(ctx, query, sport, yearStart, yearEnd).Scan(&sportSeasonWeekData.SportSeasonWeekID, &sportSeasonWeekData.WeekStart, &sportSeasonWeekData.WeekEnd)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no week data found for current sport %s for years %d - %d", sport, yearStart, yearEnd)
		}
		return nil, err
	}
	return sportSeasonWeekData, nil
}

func GetSportSeasonId(ctx context.Context, db *pgxpool.Pool, sport string, yearStart int, yearEnd int) (string, error) {
	var sportSeasonId string
	query := `
        SELECT id
        FROM sport_season
        WHERE 
			sport = $1 AND
			year_start = $2 AND
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
