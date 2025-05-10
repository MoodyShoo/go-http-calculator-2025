package main

import (
	"log"

	"github.com/MoodyShoo/go-http-calculator/internal/orchestrator"
)

func main() {
	orc := orchestrator.New()

	if err := orc.RunServer(); err != nil {
		log.Fatalf("Start HTTP server error: %v", err)
	}
}
