package handler

import (
	"apollo/db"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"apollo/repository"

	"github.com/gorilla/mux"
)

// GetUsersHandler handles the ger users GET request
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(r.Context(), "SELECT * FROM app_user")
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}
		users = append(users, map[string]interface{}{
			"id":   id,
			"name": name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUsersHandler handles the get users GET request
func GetUsersByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]
	message := fmt.Sprintf("Getting user by ID: %s", userId)

	response := map[string]string{"message": message}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type LeagueHandler struct {
	Repo *repository.LeagueRepository
}

// Create League Handler
func (lh *LeagueHandler) PostLeague(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var input struct {
		League_Nme string `json:"league_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	league, err := lh.Repo.CreateLeague(ctx, input.League_Nme)
	if err != nil {
		http.Error(w, "could not create league", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(league)
}
