package main

import (
	"apollo/db"
	"apollo/db/migrations"
	"apollo/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Connect to the database
	db.ConnectDB()
	defer db.CloseDB()
	migrations.RunMigrations()

	// Start the server
	err := server.StartServer()
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	// Graceful shutdown on interrupt signal
	gracefulShutdown()
}

// gracefulShutdown handles cleanup on termination signals
func gracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down...")
}
