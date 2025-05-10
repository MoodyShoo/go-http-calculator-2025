package agent

import (
	"os"
	"strconv"
)

type Config struct {
	OrchestratorGRPC string
	ComputingPower   int
}

func configFromEnv() *Config {
	config := Config{
		OrchestratorGRPC: "localhost:5000",
		ComputingPower:   2,
	}

	if orchAddr := os.Getenv(OrchestratorGRPCEnv); orchAddr != "" {
		config.OrchestratorGRPC = orchAddr
	}

	if val := os.Getenv(ComputingPowerEnv); val != "" {
		if power, err := strconv.Atoi(val); err == nil {
			config.ComputingPower = power
		}
	}

	return &config
}
