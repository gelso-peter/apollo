package router

import (
	"apollo/handler"

	"github.com/gorilla/mux"
)

// SetupRouter sets up and returns the router
func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/api/hello", handler.HelloHandler).Methods("GET")

	return r
}
