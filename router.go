package main

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	gob.Register(&User{})
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes { //See routes.go for list of Routes

		var handler http.Handler

		handler = route.HandlerFunc
		handler = AuthHandler(handler, route.AuthRequired)
		handler = RateLimitHandler(handler, route.RateLimited)
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
