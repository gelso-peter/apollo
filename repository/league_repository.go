package repository

import (
	"apollo/models"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LeagueRepository struct {
	DB *pgxpool.Pool
}

func NewLeagueRepository(pool *pgxpool.Pool) *LeagueRepository {
	return &LeagueRepository{DB: pool}
}

func (ur *LeagueRepository) CreateLeague(ctx context.Context, league_name string) (*models.League, error) {
	dbLeague := models.League{}
	id := uuid.New()

	query := `INSERT INTO league (id, league_name) VALUES ($1, $2) RETURNING id, league_name`
	err := ur.DB.QueryRow(ctx, query, id, league_name).Scan(&dbLeague.ID, &dbLeague.League_Name)
	if err != nil {
		return nil, fmt.Errorf("could not insert league: %w", err)
	}

	return &models.League{
		ID:          dbLeague.ID,
		League_Name: dbLeague.League_Name,
	}, nil
}

// func (ur *LeagueRepository) GetLeagueByID(id int) (*models.DBLeague, error) {
// 	var league models.DBLeague
// 	query := `SELECT id, name FROM league WHERE id = $1`
// 	err := ur.DB.QueryRow(query, id).Scan(&league.ID)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not fetch league: %w", err)
// 	}
// 	return &league, nil
// }
