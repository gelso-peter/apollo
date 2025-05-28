package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"apollo/models"
	"apollo/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Repo *repository.UserRepository
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// User Handler
func (lh *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req models.SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user, err := lh.Repo.CreateUser(ctx, &req, hashedPassword)
	if err != nil {
		http.Error(w, "could not create app_user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created"))
}

func (lh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	userIdAndPasswordHash, err := lh.Repo.GetUserIdAndPasswordHashByEmail(ctx, req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// ✅ Compare the incoming password to the stored hashed password
	err = bcrypt.CompareHashAndPassword([]byte(userIdAndPasswordHash.Password_Hash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 🔐 Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userIdAndPasswordHash.User_Id,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	// sign the token with your secret
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Could not sign token", http.StatusInternalServerError)
		return
	}

	// send the token back
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
}
