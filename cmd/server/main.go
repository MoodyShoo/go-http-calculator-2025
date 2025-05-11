package main

import (
	"log"

	"github.com/MoodyShoo/go-http-calculator/internal/database"
	"github.com/MoodyShoo/go-http-calculator/internal/orchestrator"
)

func main() {
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}

	orc := orchestrator.New(db)

	if err := orc.RunServer(); err != nil {
		log.Fatalf("Start HTTP server error: %v", err)
	}
}
