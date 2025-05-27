package server

import (
	"apollo/router"
	"net/http"
)

// StartServer initializes and starts the server
func GetRestRouter() http.Handler {
	// Initialize the router
	r := router.SetupRouter()
	return r
}
