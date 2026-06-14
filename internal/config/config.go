// Package config loads runtime configuration from the environment.
package config

import "os"

// Config holds all runtime settings. Fields are added per phase; Phase 0 only
// needs the HTTP port.
type Config struct {
	Port string
}

// Load reads configuration from environment variables, applying sensible
// defaults so the server runs locally with zero setup.
func Load() Config {
	return Config{
		Port: env("PORT", "8080"),
	}
}

// env returns the value of key, or def when it is unset or empty.
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
