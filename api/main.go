package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/suraj/url-shortener/routes"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Create a new Gorilla Mux router
	r := mux.NewRouter()

	// Register routes
	// Resolve short URL → GET /{url}
	r.HandleFunc("/{url}", routes.ResolveURL).Methods("GET")

	// Shorten URL → POST /api/v1/
	r.HandleFunc("/api/v1/", routes.ShortenURL).Methods("POST")

	// Start server
	log.Fatal(http.ListenAndServe(":"+os.Getenv("APP_PORT"), r))
}
