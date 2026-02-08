package telemetry

import (
	"flag"
	"log/slog"
	"os"
	"strings"
)

// Environment determines which logging handler to use
type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
	Test        Environment = "test"
)

// Config holds telemetry configuration
type Config struct {
	Environment  Environment
	LogLevel     slog.Level
	LogFile      string
	OTLPEndpoint string
}

// envOrDefault returns the env var value or a default
func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// ParseConfig reads configuration from flags and environment variables.
// Flags take precedence over environment variables.
// All env vars use the MEGAWAVE_ prefix.
func ParseConfig() Config {
	// Define flags with env var defaults
	envFlag := flag.String("env", envOrDefault("MEGAWAVE_ENV", "development"),
		"environment: production, development, test")
	logLevelFlag := flag.String("log-level", envOrDefault("MEGAWAVE_LOG_LEVEL", "info"),
		"log level: debug, info, warn, error")
	logFileFlag := flag.String("log-file", envOrDefault("MEGAWAVE_LOG_FILE", "megawave.log"),
		"log file path (development mode only)")
	otlpFlag := flag.String("otlp-endpoint", os.Getenv("MEGAWAVE_OTLP_ENDPOINT"),
		"OTLP collector endpoint (host:port, e.g., localhost:4318)")

	flag.Parse()

	return Config{
		Environment:  parseEnvironment(*envFlag),
		LogLevel:     parseLogLevel(*logLevelFlag),
		LogFile:      *logFileFlag,
		OTLPEndpoint: *otlpFlag,
	}
}

// parseEnvironment converts a string to an Environment value
func parseEnvironment(s string) Environment {
	switch strings.ToLower(s) {
	case "production", "prod":
		return Production
	case "test":
		return Test
	default:
		return Development
	}
}

// parseLogLevel converts a string to a slog.Level
func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
