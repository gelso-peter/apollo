package repository

import (
	"apollo/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MeRepository struct {
	DB *pgxpool.Pool
}

func NewMeRepository(pool *pgxpool.Pool) *MeRepository {
	return &MeRepository{DB: pool}
}

func (ur *MeRepository) GetMeLeagues(ctx context.Context) ([]*models.League, error) {
	query := `
		SELECT l.id, l.league_name
		FROM league l
		INNER JOIN user_league_association ula ON l.id = ula.league_id
		WHERE ula.user_id = $1;
	`

	rows, err := ur.DB.Query(ctx, query, "b8a38c51-872e-4441-b4a1-417cf5861663")
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var leagues []*models.League

	for rows.Next() {
		var league models.League
		if err := rows.Scan(&league.ID, &league.League_Name); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		leagues = append(leagues, &league)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return leagues, nil
}
