package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"apollo/repository"
)

type MeHandler struct {
	Repo *repository.MeRepository
}

// Me Handler
func (lh *MeHandler) GetMeLeagues(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	league, err := lh.Repo.GetMeLeagues(ctx)
	if err != nil {
		http.Error(w, "could not get leagues for user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(league)
}
