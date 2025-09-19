package routes

import (
	"github.com/gorilla/mux"
)

func Setup(r *mux.Router) {

	r.HandleFunc("/{url}", ResolveURL).Methods("GET")
	r.HandleFunc("/api/v1/", ShortenURL).Methods("POST")

}
