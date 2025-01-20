package server

import (
	"apollo/router"
	"log"
	"net/http"
)

// StartServer initializes and starts the server
func StartServer() error {
	// Initialize the router
	r := router.SetupRouter()

	// Define server settings
	addr := ":8080" // Set the port to listen on
	log.Printf("Starting server on %s\n", addr)

	// Start the server
	return http.ListenAndServe(addr, r)
}
