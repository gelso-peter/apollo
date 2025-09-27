package main

import (
	"apollo/db"
	"apollo/db/migrations"
	"apollo/graph"
	"apollo/internal/aws"
	"apollo/middleware"
	"apollo/router"
	"apollo/services/odds.go"
	"context"
	"fmt"
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
	// Check if running in health check mode
	if len(os.Args) > 1 && os.Args[1] == "--health-check" {
		healthCheck()
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize AWS Secrets Manager client
	secretsClient, err := aws.NewSecretsClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize AWS Secrets Manager client: %v", err)
		log.Println("Falling back to environment variables for local development")
	}

	// Get ODDS_API_KEY from AWS Secrets Manager or environment variable
	var oddsApiKey string
	if secretsClient != nil {
		oddsApiKey, err = secretsClient.GetIndividualSecret("apollo/api-keys")
		if err != nil {
			log.Printf("Failed to get ODDS_API_KEY from secrets manager: %v", err)
			log.Fatal("Missing ODDS_API_KEY and unable to retrieve from AWS Secrets Manager")
		}
	} else {
		// Fallback to environment variable
		oddsApiKey = os.Getenv("ODDS_API_KEY")
		if oddsApiKey == "" {
			log.Fatal("Missing ODDS_API_KEY and AWS Secrets Manager is not available")
		}
	}

	// Get DATABASE_URL from AWS Secrets Manager or environment variable
	var dbURL string
	if secretsClient != nil {
		dbURL, err = secretsClient.GetIndividualSecret("apollo/db-url")
		if err != nil {
			log.Printf("Failed to get DATABASE_URL from secrets manager: %v", err)
			log.Println("Falling back to DATABASE_URL environment variable")
			dbURL = os.Getenv("DATABASE_URL")
		}
	} else {
		// Fallback to environment variable
		dbURL = os.Getenv("DATABASE_URL")
	}

	if dbURL == "" {
		log.Fatal("Missing DATABASE_URL and unable to retrieve from AWS Secrets Manager")
	}

	db.ConnectDB()
	defer db.CloseDB()
	migrations.RunMigrations(dbURL)

	odds.InitOddsService(oddsApiKey)
	oddsService := odds.GetOddsService()

	// Note: Game finalization is now handled by AWS Lambda function
	log.Println("Game finalization is handled by AWS Lambda (runs Fri/Sun/Mon/Tue at 6:00 AM)")

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

	// Health check endpoint
	mainMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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
	gracefulShutdown(srv)
}

func gracefulShutdown(srv *http.Server) {
	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until signal received
	<-stop

	log.Println("Shutting down server...")

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("Server exited properly")
}

// healthCheck performs a simple health check for Docker HEALTHCHECK
func healthCheck() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
	if err != nil {
		log.Printf("Health check failed: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Health check failed with status: %d", resp.StatusCode)
		os.Exit(1)
	}

	log.Println("Health check passed")
	os.Exit(0)
}
