package main

import (
	"github.com/gorilla/mux"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIValidationError struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Errors  []*ValidationError `json:"errors"`
}

type ValidationError struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

type PaginatedResult struct {
	Next     string   `json:"next"`
	Previous string   `json:"previous"`
	Results  TodoList `json:"results"`
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
