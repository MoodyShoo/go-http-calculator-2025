package main

import (
	"log"

	"github.com/MoodyShoo/go-http-calculator/internal/orchestrator"
)

func main() {
	orch := orchestrator.New()
	err := orch.RunServer()
	if err != nil {
		log.Fatalf("Start server error: %v", err)
	}
}
