package main

import "time"

type Config struct {
	HTTPPort int
	Timeout  time.Duration
}

func defaultConfig() Config {
	return Config{HTTPPort: 8080, Timeout: 15 * time.Second}
}
