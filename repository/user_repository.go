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

func (ur *UserRepository) CreateUser(ctx context.Context, signupRequest *models.SignupRequest, passwordHash []byte) (*models.FullUser, error) {
	user := models.FullUser{}
	id := uuid.New()
	query := `
		INSERT INTO app_user (id, first_name, last_name, email, password_hash) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, first_name, last_name, email
		`
	err := ur.DB.QueryRow(ctx, query, id, signupRequest.First_Name, signupRequest.Last_Name, signupRequest.Email, passwordHash).Scan(&user.ID, &user.First_Name, &user.Last_Name, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("could not app_user: %w", err)
	}

	return &user, nil
}

func (ur *UserRepository) GetUserIdAndPasswordHashByEmail(ctx context.Context, email string) (models.UserIdPasswordHash, error) {
	var result models.UserIdPasswordHash
	query := `
		SELECT id, password_hash
		FROM app_user
		WHERE email = $1;
		`
	err := ur.DB.QueryRow(ctx, query, email).Scan(&result.User_Id, &result.Password_Hash)
	if err != nil {
	}

	return result, nil
}
