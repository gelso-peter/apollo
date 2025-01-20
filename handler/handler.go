package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUsersHandler handles the ger users GET request
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Getting Users!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUsersHandler handles the ger users GET request
func GetUsersByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]
	message := fmt.Sprintf("Getting user by ID: %s", userId)

	response := map[string]string{"message": message}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
