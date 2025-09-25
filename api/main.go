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

	_ = godotenv.Load()

	r := mux.NewRouter()

	routes.Setup(r)

	port := os.Getenv("PORT") // Render's port
	if port == "" {
		port = os.Getenv("APP_PORT") // Run locally
	}

	log.Fatal(http.ListenAndServe(":"+port, r))

}
