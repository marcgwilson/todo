package main

import (
	"github.com/gorilla/mux"
)

func NewRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", h.CreateFunc()).Methods("POST")
	r.HandleFunc("/", h.ListFunc()).Methods("GET")
	r.HandleFunc("/{id:[0-9]+}/", h.RetrieveFunc()).Methods("GET")
	r.HandleFunc("/{id:[0-9]+}/", h.UpdateFunc()).Methods("PATCH")
	r.HandleFunc("/{id:[0-9]+}/", h.DeleteFunc()).Methods("DELETE")
	return r
}
