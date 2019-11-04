package main

import (
	"log"
	"net/http"
)

func main() {
	config, err := LookupConfig()

	if err != nil {
		log.Panic(err)
	}

	db, err := OpenDB(config.Database)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	log.Printf("Running Todo server at %s using database %s...\n", config.Addr(), config.Database)


	m := &TodoManager{db}
	h := NewHandler(m, config)

	if err := http.ListenAndServe(config.Addr(), NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
