// Package config loads runtime configuration from the environment.
package config

import "os"

// Config holds all runtime settings. Fields are added per phase.
type Config struct {
	Port              string
	DBPath            string
	AdminPasswordHash string
	SessionSecret     string
	AllowedOrigin     string
	SecureCookies     bool
}

// Load reads configuration from environment variables, applying sensible
// defaults so the server runs locally with zero setup.
func Load() Config {
	return Config{
		Port:              env("PORT", "8080"),
		DBPath:            env("DB_PATH", "bg01.db"),
		AdminPasswordHash: os.Getenv("ADMIN_PASSWORD_HASH"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
		AllowedOrigin:     os.Getenv("ALLOWED_ORIGIN"),
		// Secure cookies on by default; set COOKIE_SECURE=false for local HTTP.
		SecureCookies: env("COOKIE_SECURE", "true") != "false",
	}
}

// env returns the value of key, or def when it is unset or empty.
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
