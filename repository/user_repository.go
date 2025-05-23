package repository

import (
	"apollo/models"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: pool}
}

func (ur *UserRepository) CreateUser(ctx context.Context, apiUser *models.APIUser) (*models.FullUser, error) {
	user := models.FullUser{}
	id := uuid.New()

	query := `
		INSERT INTO app_user (id, first_name, last_name, email) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, first_name, last_name, email
		`
	err := ur.DB.QueryRow(ctx, query, id, apiUser.First_Name, apiUser.Last_Name, apiUser.Email).Scan(&user.ID, &user.First_Name, &user.Last_Name, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("could not app_user: %w", err)
	}

	return &user, nil
}
