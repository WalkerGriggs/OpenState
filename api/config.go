package api

import (
	"time"
)

type Config struct {
	// Address is the HTTPAdvertise address of an OpenState node
	Address string

	// WaitTime is the length of time which the API will block.
	WaitTime time.Duration
}

// DefaultConfig returns a default configuration for the client
func DefaultConfig() *Config {
	return &Config{
		Address:  "http://127.0.0.1:8080",
		WaitTime: 30 * time.Second,
	}
}
