package main

import (
	"github.com/gorilla/mux"
)

type APIError struct {
	Code    int
	Message string
}

type APIValidationError struct {
	Code int
	Message string
	Errors []*ValidationError
}

type ValidationError struct {
	Key string
	Value interface{}
	Message string
}

type PaginatedResult struct {
	Next     string
	Previous string
	Results  TodoList
}

func NewMux(tm *TodoManager) *mux.Router {
	h := NewHandler(tm)
	r := mux.NewRouter()
	r.HandleFunc("/", h.CreateFunc()).Methods("POST")
	r.HandleFunc("/", h.ListFunc()).Methods("GET")
	r.HandleFunc("/{id:[0-9]+}/", h.RetrieveFunc()).Methods("GET")
	r.HandleFunc("/{id:[0-9]+}/", h.UpdateFunc()).Methods("PATCH")
	r.HandleFunc("/{id:[0-9]+}/", h.DeleteFunc()).Methods("DELETE")
	return r
}
