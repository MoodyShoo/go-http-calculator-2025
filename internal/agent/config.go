package agent

import (
	"os"
	"strconv"
)

type Config struct {
	Address             string
	OrchestratorAddress string
	ComputingPower      int
}

func configFromEnv() *Config {
	config := Config{
		Address:             "8081",
		OrchestratorAddress: "localhost:8080",
		ComputingPower:      2,
	}

	if addr := os.Getenv(PortEnv); addr != "" {
		config.Address = addr
	}

	if orchAddr := os.Getenv(OrchestratorAddressEnv); orchAddr != "" {
		config.OrchestratorAddress = orchAddr
	}

	if val := os.Getenv(ComputingPowerEnv); val != "" {
		if power, err := strconv.Atoi(val); err == nil {
			config.ComputingPower = power
		}
	}

	return &config
}
