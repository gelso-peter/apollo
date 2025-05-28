package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"apollo/internal/contextutil"
	"apollo/repository"
)

type MeHandler struct {
	Repo *repository.MeRepository
}

// Me Handler
func (lh *MeHandler) GetMeLeagues(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextutil.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized from League", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	league, err := lh.Repo.GetMeLeagues(ctx, userID)
	if err != nil {
		http.Error(w, "could not get leagues for user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(league)
}
