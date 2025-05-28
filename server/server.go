package server

import (
	"apollo/router"
	"net/http"
)

func GetRestRouter() http.Handler {
	r := router.SetupRouter()
	return r
}
