package repository

import (
	"apollo/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserLeagueAssociationRepository struct {
	DB *pgxpool.Pool
}

func NewUserLeagueAssociationRepository(pool *pgxpool.Pool) *UserLeagueAssociationRepository {
	return &UserLeagueAssociationRepository{DB: pool}
}

func (ula *UserLeagueAssociationRepository) CreateUserLeagueAssociation(ctx context.Context, association *models.UserLeagueAssociation) (*models.UserLeagueAssociation, error) {

	userLeagueAssociation := models.UserLeagueAssociation{}
	query := `
		INSERT INTO user_league_association (user_id, league_id) 
		VALUES ($1, $2) 
		RETURNING user_id, league_id
		`
	err := ula.DB.QueryRow(ctx, query, association.User_Id, association.League_id).Scan(&userLeagueAssociation.User_Id, &userLeagueAssociation.League_id)
	if err != nil {
		return nil, fmt.Errorf("could not attached user league association: %w", err)
	}

	return association, nil
}
