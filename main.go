package main

import (
	"log"
	"os"

	"github.com/BestPrice/backend/http"
	"github.com/BestPrice/backend/sql"
)

func main() {

	// open database
	c := &sql.Client{
		Path: os.Getenv("DATABASE_URL"),
	}
	if err := c.Open(); err != nil {
		log.Fatal(err)
	}

	// create server on PORT with handler
	s := http.Server{
		Port:    ":" + os.Getenv("PORT"),
		Handler: http.NewHandler(c.Connect().Service()),
	}

	// Run backend server
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
