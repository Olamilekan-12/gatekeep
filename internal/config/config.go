package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds everything the process needs to run. It is read once at
// startup and then treated as immutable.
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment, applying defaults.
// It fails if a value is present but malformed.
func Load() (Config, error) {
	shutdownTimeout, err := envDuration("GATEKEEP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Addr:            env("GATEKEEP_ADDR", ":8080"),
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s=%q is not a valid duration: %w", key, v, err)
	}
	return d, nil
}
