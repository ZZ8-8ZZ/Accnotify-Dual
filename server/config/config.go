package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	// Server settings
	Host string
	Port int

	// Database
	DBPath string

	// WebSocket settings
	WSPingInterval int // seconds
	WSPongTimeout  int // seconds

	// Security
	EnableHTTPS bool
	CertFile    string
	KeyFile     string

	// APNs settings
	APNsEnabled     bool
	APNsCertFile    string
	APNsKeyFile     string
	APNsKeyID       string
	APNsTeamID      string
	APNsTopic       string
	APNsDevelopment bool
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:           "0.0.0.0",
		Port:           8080,
		DBPath:         "./data/accnotify.db",
		WSPingInterval: 30,
		WSPongTimeout:  60,
		EnableHTTPS:    false,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	if host := os.Getenv("ACCNOTIFY_HOST"); host != "" {
		cfg.Host = host
	}

	if port := os.Getenv("ACCNOTIFY_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if dbPath := os.Getenv("ACCNOTIFY_DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	if pingInterval := os.Getenv("ACCNOTIFY_WS_PING_INTERVAL"); pingInterval != "" {
		if p, err := strconv.Atoi(pingInterval); err == nil {
			cfg.WSPingInterval = p
		}
	}

	if os.Getenv("ACCNOTIFY_ENABLE_HTTPS") == "true" {
		cfg.EnableHTTPS = true
		cfg.CertFile = os.Getenv("ACCNOTIFY_CERT_FILE")
		cfg.KeyFile = os.Getenv("ACCNOTIFY_KEY_FILE")
	}

	// APNs settings
	if os.Getenv("ACCNOTIFY_APNS_ENABLED") == "true" {
		cfg.APNsEnabled = true
		cfg.APNsCertFile = os.Getenv("ACCNOTIFY_APNS_CERT_FILE")
		cfg.APNsKeyFile = os.Getenv("ACCNOTIFY_APNS_KEY_FILE")
		cfg.APNsKeyID = os.Getenv("ACCNOTIFY_APNS_KEY_ID")
		cfg.APNsTeamID = os.Getenv("ACCNOTIFY_APNS_TEAM_ID")
		cfg.APNsTopic = os.Getenv("ACCNOTIFY_APNS_TOPIC")
		cfg.APNsDevelopment = os.Getenv("ACCNOTIFY_APNS_DEVELOPMENT") == "true"
	}

	return cfg
}
