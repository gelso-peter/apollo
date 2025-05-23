package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"apollo/models"
	"apollo/repository"
)

type UserHandler struct {
	Repo *repository.UserRepository
}

// User Handler
func (lh *UserHandler) PostCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var apiUser models.APIUser
	if err := json.NewDecoder(r.Body).Decode(&apiUser); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := lh.Repo.CreateUser(ctx, &apiUser)
	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
