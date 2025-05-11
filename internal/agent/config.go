package agent

import (
	"os"
	"strconv"
)

type Config struct {
	OrchestratorGRPCAddress string
	OrchestratorGRPCPort    string
	ComputingPower          int
}

func configFromEnv() *Config {
	config := Config{
		OrchestratorGRPCAddress: "localhost",
		OrchestratorGRPCPort:    "5000",
		ComputingPower:          2,
	}

	if orchAddr := os.Getenv(config.OrchestratorGRPCAddress); orchAddr != "" {
		config.OrchestratorGRPCAddress = orchAddr
	}

	if orchPort := os.Getenv(config.OrchestratorGRPCPort); orchPort != "" {
		config.OrchestratorGRPCPort = orchPort
	}

	if val := os.Getenv(ComputingPowerEnv); val != "" {
		if power, err := strconv.Atoi(val); err == nil {
			config.ComputingPower = power
		}
	}

	return &config
}
