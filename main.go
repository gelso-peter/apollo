package main

import (
	"apollo/server"
	"log"
)

func main() {
	// Start the server
	err := server.StartServer()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
