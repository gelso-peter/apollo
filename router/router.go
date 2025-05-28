package router

import (
	"apollo/db"
	"apollo/handler"
	"apollo/middleware"

	"apollo/repository"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	// Mux Router
	r := mux.NewRouter()

	// DB Connection
	pool := db.DB

	// Instantiate Repositories

	userRepository := repository.NewUserRepository(pool)
	meRepository := repository.NewMeRepository(pool)
	leagueRepository := repository.NewLeagueRepository(pool)
	userLeagueAssociationRepository := repository.NewUserLeagueAssociationRepository(pool)

	// Instantiate Handlers
	userHandler := &handler.UserHandler{Repo: userRepository}
	meHandler := &handler.MeHandler{Repo: meRepository}
	leagueHandler := &handler.LeagueHandler{League_Repo: leagueRepository, User_League_Association_Repo: userLeagueAssociationRepository}

	// Public routes
	r.HandleFunc("/signup", userHandler.SignUp).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTMiddleware)

	// me routes
	protected.HandleFunc("/me/leagues", meHandler.GetMeLeagues).Methods("GET")

	// League routes
	protected.HandleFunc("/league", leagueHandler.PostLeague).Methods("POST")

	return r
}
