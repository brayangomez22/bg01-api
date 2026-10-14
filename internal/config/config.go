// Package config loads runtime configuration from the environment.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// MinSessionSecretLen is the minimum SESSION_SECRET length in bytes. An HMAC
// key shorter than the hash output (32 bytes for SHA-256) weakens the session
// signature; an empty one lets anyone forge a valid token.
const MinSessionSecretLen = 32

// Config holds all runtime settings. Fields are added per phase.
type Config struct {
	Port              string
	DBPath            string
	AdminPasswordHash string
	SessionSecret     string
	AllowedOrigin     string
	SecureCookies     bool
	// Publishing: trigger the portfolio's GitHub Actions deploy on content edits.
	PublishRepo string // "owner/name", e.g. brayangomez22/bg01-station
	GithubToken string // PAT with permission to dispatch workflows on PublishRepo
}

// Load reads configuration from environment variables, applying sensible
// defaults so the server runs locally with zero setup. A local .env file is
// loaded first as a convenience for development; real environment variables
// always take precedence, so it is a no-op (and absent) in production. A .env
// that exists but cannot be read is an error.
func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	return Config{
		Port:              env("PORT", "8080"),
		DBPath:            env("DB_PATH", "bg01.db"),
		AdminPasswordHash: os.Getenv("ADMIN_PASSWORD_HASH"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
		AllowedOrigin:     os.Getenv("ALLOWED_ORIGIN"),
		// Secure cookies on by default; set COOKIE_SECURE=false for local HTTP.
		SecureCookies: env("COOKIE_SECURE", "true") != "false",
		PublishRepo:   os.Getenv("PUBLISH_REPO"),
		GithubToken:   os.Getenv("GITHUB_DISPATCH_TOKEN"),
	}, nil
}

// Validate reports configuration that would leave the API insecure. The
// server must refuse to start rather than run with forgeable sessions.
func (c Config) Validate() error {
	var errs []error
	if c.AdminPasswordHash == "" {
		errs = append(errs, errors.New("ADMIN_PASSWORD_HASH is required"))
	}
	if len(c.SessionSecret) < MinSessionSecretLen {
		errs = append(errs, fmt.Errorf("SESSION_SECRET must be at least %d bytes (got %d)", MinSessionSecretLen, len(c.SessionSecret)))
	}
	return errors.Join(errs...)
}

// env returns the value of key, or def when it is unset or empty.
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadDotEnv reads simple KEY=VALUE lines from path into the process
// environment, skipping blanks and #-comments and tolerating optional quotes
// around values. Existing environment variables are never overwritten, so the
// real environment wins over the file. A missing file is not an error.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return nil
}
