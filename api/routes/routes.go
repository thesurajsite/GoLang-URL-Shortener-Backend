package routes

import (
	"github.com/gorilla/mux"
)

// Setup configures all routes using Gorilla Mux
func Setup(r *mux.Router) {
	// Resolve a shortened URL (GET /{url})
	r.HandleFunc("/{url}", ResolveURL).Methods("GET")

	// Shorten a new URL (POST /api/v1/)
	r.HandleFunc("/api/v1/", ShortenURL).Methods("POST")
}
