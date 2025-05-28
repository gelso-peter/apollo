package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"apollo/internal/contextutil"
	"apollo/models"
	"apollo/repository"
)

type LeagueHandler struct {
	League_Repo                  *repository.LeagueRepository
	User_League_Association_Repo *repository.UserLeagueAssociationRepository
}

// Create League Handler
func (lh *LeagueHandler) PostLeague(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextutil.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized from League", http.StatusUnauthorized)
		return
	}

	var input struct {
		League_Name string `json:"league_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Create the timeout context for DB operations
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Create the League
	league, err := lh.League_Repo.CreateLeague(ctx, input.League_Name)
	if err != nil {
		http.Error(w, "could not create league", http.StatusInternalServerError)
		return
	}

	associationToCreate := models.UserLeagueAssociation{
		User_Id:   userID,
		League_id: league.ID,
	}

	// Create the Association b/w the user and the league
	_, err = lh.User_League_Association_Repo.CreateUserLeagueAssociation(ctx, &associationToCreate)
	if err != nil {
		http.Error(w, "could not create user association with the league", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(league)
}
