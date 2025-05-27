package main

import (
	"apollo/db"
	"apollo/db/migrations"
	"apollo/graph"
	"apollo/server"
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
)

const defaultPort = "8080"

// func main() {
// 	// Connect to the database

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = defaultPort
// 	}

// 	db.ConnectDB()
// 	defer db.CloseDB()
// 	migrations.RunMigrations()

// 	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

// 	srv.AddTransport(transport.Options{})
// 	srv.AddTransport(transport.GET{})
// 	srv.AddTransport(transport.POST{})

// 	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

// 	srv.Use(extension.Introspection{})
// 	srv.Use(extension.AutomaticPersistedQuery{
// 		Cache: lru.New[string](100),
// 	})

// 	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
// 	http.Handle("/query", srv)

// 	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
// 	log.Fatal(http.ListenAndServe(":"+port, nil))

// 	// Start the server
// 	err := server.StartServer()
// 	if err != nil {
// 		log.Fatalf("Error starting server: %v", err)
// 	}

// 	// Graceful shutdown on interrupt signal
// 	gracefulShutdown()
// }

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db.ConnectDB()
	defer db.CloseDB()
	migrations.RunMigrations()

	// Setup GraphQL server
	graphQLServer := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db.DB}}))
	graphQLServer.AddTransport(transport.Options{})
	graphQLServer.AddTransport(transport.GET{})
	graphQLServer.AddTransport(transport.POST{})
	graphQLServer.Use(extension.Introspection{})

	// Setup REST and GraphQL together in one mux
	mainMux := http.NewServeMux()

	// GraphQL routes
	mainMux.Handle("/query", graphQLServer)
	mainMux.Handle("/", playground.Handler("GraphQL playground", "/query"))

	// REST routes mounted under a prefix (e.g., "/api")
	restRouter := server.GetRestRouter()
	mainMux.Handle("/api/", http.StripPrefix("/api", restRouter))

	// Create HTTP server instance
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mainMux,
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
