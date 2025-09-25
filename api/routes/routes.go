package routes

import (
	"github.com/gorilla/mux"
)

func Setup(r *mux.Router) {

	r.HandleFunc("/api/v1/", ShortenURL).Methods("POST")
	r.HandleFunc("/{id}", ResolveURL).Methods("GET")
	r.HandleFunc("/count/", UrlCount).Methods("GET")

}
