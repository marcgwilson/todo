package main

import (
	"log"
	"net/http"
)

func main() {
	conf, err := LookupConfig()

	if err != nil {
		log.Panic(err)
	}

	db, err := OpenDB(conf.Database)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	log.Printf("Running Todo server at %s using database %s...\n", conf.Addr(), conf.Database)

	m := &TodoManager{db, conf.Limit}
	if err := http.ListenAndServe(conf.Addr(), NewMux(m)); err != nil {
		log.Fatal(err)
	}
}
