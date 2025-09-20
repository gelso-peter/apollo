package main

import (
	"apollo/db"
	"apollo/db/migrations"
	"apollo/graph"
	"apollo/middleware"
	"apollo/router"
	"apollo/services/cron"
	"apollo/services/odds.go"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	oddsApiKey := os.Getenv("ODDS_API_KEY")
	if oddsApiKey == "" {
		log.Fatal("Missing ODDS_API_KEY")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Missing DATABASE_URL")
	}

	db.ConnectDB()
	defer db.CloseDB()
	migrations.RunMigrations()

	odds.InitOddsService(oddsApiKey)
	oddsService := odds.GetOddsService()

	// Setup cron job for game finalization
	gameFinalizer := cron.NewGameFinalizer(db.DB, oddsService)
	cronStop := make(chan struct{})

	// Run cron job on specific days at 6:00 AM
	go gameFinalizer.RunPeriodically(cronStop)
	log.Println("Game finalization cron job started (runs Fri/Sun/Mon/Tue at 6:00 AM)")

	// Setup GraphQL server
	graphResolvers := &graph.Resolver{
		DB:          db.DB,
		OddsService: oddsService,
	}
	graphQLServer := handler.New(
		graph.NewExecutableSchema(graph.Config{Resolvers: graphResolvers}),
	)
	graphQLServer.AddTransport(transport.Options{})
	graphQLServer.AddTransport(transport.GET{})
	graphQLServer.AddTransport(transport.POST{})
	graphQLServer.Use(extension.Introspection{})

	// Wrap the GraphQL handler with your JWT middleware
	protectedGraphQL := middleware.JWTMiddleware(graphQLServer)

	// Setup REST and GraphQL together in one mux
	mainMux := http.NewServeMux()

	// GraphQL routes
	mainMux.Handle("/query", protectedGraphQL)
	mainMux.Handle("/", playground.Handler("GraphQL playground", "/query"))

	// REST routes mounted under a prefix (e.g., "/api")
	restRouter := router.SetupRouter()

	mainMux.Handle("/api/", http.StripPrefix("/api", restRouter))

	// Create HTTP server instance
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: c.Handler(mainMux),
	}

	// Run server in a goroutine so main can listen for signals
	go func() {
		log.Printf("Server running on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for interrupt signal and gracefully shutdown the server
	gracefulShutdown(srv, cronStop)
}

func gracefulShutdown(srv *http.Server, cronStop chan struct{}) {
	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until signal received
	<-stop

	log.Println("Shutting down server...")

	// Stop the cron job first
	close(cronStop)
	log.Println("Cron job stopped")

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("Server exited properly")
}
