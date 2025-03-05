package orchestrator

import (
	"os"
	"strconv"
)

type Config struct {
	Address               string
	TimeAdditionMs        int
	TimeSubtractionMs     int
	TimeMultiplicationsMs int
	TimeDivisionsMs       int
}

func configFromEnv() *Config {
	config := &Config{
		Address:               "8080",
		TimeAdditionMs:        1000,
		TimeSubtractionMs:     1000,
		TimeMultiplicationsMs: 1000,
		TimeDivisionsMs:       1000,
	}

	if addr := os.Getenv(PortEnv); addr != "" {
		config.Address = addr
	}

	if val := os.Getenv(TimeAdditionMsEnv); val != "" {
		if timeMs, err := strconv.Atoi(val); err == nil {
			config.TimeAdditionMs = timeMs
		}
	}

	if val := os.Getenv(TimeSubtractionMsEnv); val != "" {
		if timeMs, err := strconv.Atoi(val); err == nil {
			config.TimeSubtractionMs = timeMs
		}
	}

	if val := os.Getenv(TimeMultiplicationsMsEnv); val != "" {
		if timeMs, err := strconv.Atoi(val); err == nil {
			config.TimeMultiplicationsMs = timeMs
		}
	}

	if val := os.Getenv(TimeDivisionsMsEnv); val != "" {
		if timeMs, err := strconv.Atoi(val); err == nil {
			config.TimeDivisionsMs = timeMs
		}
	}

	return config
}
