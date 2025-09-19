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

	log.Fatal(http.ListenAndServe(":"+os.Getenv("APP_PORT"), r))
}
