// Package config loads Trackdate's runtime configuration from environment
// variables, so the container image stays generic across deployments.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr    string
	DBPath  string
	BaseURL string
	// CookieSecure is derived from BaseURL's scheme: only set the Secure
	// cookie attribute when the app is actually served over https.
	CookieSecure bool
	Currency     string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:     getEnv("TRACKDATE_ADDR", ":8080"),
		DBPath:   getEnv("TRACKDATE_DB_PATH", "trackdate.db"),
		BaseURL:  strings.TrimRight(getEnv("TRACKDATE_BASE_URL", "http://localhost:8080"), "/"),
		Currency: getEnv("TRACKDATE_CURRENCY", "EUR"),

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     getEnv("SMTP_FROM", "Trackdate <trackdate@localhost>"),
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return Config{}, fmt.Errorf("parsing TRACKDATE_BASE_URL: %w", err)
	}
	cfg.CookieSecure = u.Scheme == "https"

	portStr := getEnv("SMTP_PORT", "587")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Config{}, fmt.Errorf("parsing SMTP_PORT: %w", err)
	}
	cfg.SMTPPort = port

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
