package router

import (
	"apollo/db"
	"apollo/handler"

	"apollo/repository"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	// Mux Router
	r := mux.NewRouter()

	// DB Connection
	pool := db.DB

	// Instantiate Repositories
	leagueRepository := repository.NewLeagueRepository(pool)
	leagueHandler := &handler.LeagueHandler{Repo: leagueRepository}

	// Instantiate routes
	r.HandleFunc("/users", handler.GetUsersHandler).Methods("GET")
	r.HandleFunc("/users/{id}", handler.GetUsersByIdHandler).Methods("GET")
	r.HandleFunc("/league", leagueHandler.PostLeague).Methods("POST")

	return r
}
