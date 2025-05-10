package main

import (
	"github.com/MoodyShoo/go-http-calculator/internal/agent"
)

func main() {
	agent := agent.New()
	agent.Run()
}
