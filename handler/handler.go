package handler

import (
	"apollo/db"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUsersHandler handles the ger users GET request
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(r.Context(), "SELECT id, name FROM users")
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

// GetUsersHandler handles the get users GET request
func PostGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]
	message := fmt.Sprintf("Getting user by ID: %s", userId)

	response := map[string]string{"message": message}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
