package router

import (
	"apollo/handler"

	"github.com/gorilla/mux"
)

// SetupRouter sets up and returns the router
func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/users", handler.GetUsersHandler).Methods("GET")
	r.HandleFunc("/users/{id}", handler.GetUsersByIdHandler).Methods("GET")

	return r
}
