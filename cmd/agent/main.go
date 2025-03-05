package main

import (
	"log"

	"github.com/MoodyShoo/go-http-calculator/internal/agent"
)

func main() {
	agent := agent.New()
	err := agent.Run()
	if err != nil {
		log.Fatalf("Agent start error: %v", err)
	}
}
