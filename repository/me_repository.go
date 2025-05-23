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

func (ur *MeRepository) GetMeLeagues(ctx context.Context) (*models.League, error) {
	dbLeague := models.League{}
	userId := "<TODO temporary hard code>"

	query := `
		SELECT l.id, l.league_name
		FROM league l
		INNER JOIN user_league_association ula ON l.id = ula.league_id
		WHERE ula.user_id = $1;
	`
	err := ur.DB.QueryRow(ctx, query, userId).Scan(&dbLeague.ID, &dbLeague.League_Name)
	if err != nil {
		return nil, fmt.Errorf("could not get leagues for user: %w", err)
	}

	return &models.League{
		ID:          dbLeague.ID,
		League_Name: dbLeague.League_Name,
	}, nil
}
