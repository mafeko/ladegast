// Package config loads Ladegast's runtime configuration from environment
// variables. See schema/config.md for the documented contract.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr       string
	PublicBaseURL    string
	PublicPathPrefix string
	AdminPathPrefix  string
	DataDir          string

	SessionSecret string

	EVCCBaseURL string
	EVCCAPIKey  string

	AdminUsername string
	AdminPassword string

	RateLimitMaxAttempts int
	RateLimitWindow      time.Duration
}

// Load reads and validates the configuration from the process environment.
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	var errs []string
	required := func(key string) string {
		v := getenv(key)
		if v == "" {
			errs = append(errs, key+" is required")
		}
		return v
	}

	cfg := Config{
		ListenAddr:       withDefault(getenv("LADEGAST_LISTEN_ADDR"), ":8080"),
		PublicBaseURL:    required("LADEGAST_PUBLIC_BASE_URL"),
		PublicPathPrefix: normalizePrefix(withDefault(getenv("LADEGAST_PUBLIC_PATH_PREFIX"), "/public")),
		AdminPathPrefix:  normalizePrefix(withDefault(getenv("LADEGAST_ADMIN_PATH_PREFIX"), "/admin")),
		DataDir:          withDefault(getenv("LADEGAST_DATA_DIR"), "./data"),

		SessionSecret: required("LADEGAST_SESSION_SECRET"),

		EVCCBaseURL: required("LADEGAST_EVCC_BASE_URL"),
		EVCCAPIKey:  required("LADEGAST_EVCC_API_KEY"),

		AdminUsername: getenv("LADEGAST_ADMIN_USERNAME"),
		AdminPassword: getenv("LADEGAST_ADMIN_PASSWORD"),
	}

	maxAttemptsRaw := withDefault(getenv("LADEGAST_RATE_LIMIT_MAX_ATTEMPTS"), "5")
	maxAttempts, err := strconv.Atoi(maxAttemptsRaw)
	if err != nil {
		errs = append(errs, fmt.Sprintf("LADEGAST_RATE_LIMIT_MAX_ATTEMPTS: invalid integer %q", maxAttemptsRaw))
	}
	cfg.RateLimitMaxAttempts = maxAttempts

	windowRaw := withDefault(getenv("LADEGAST_RATE_LIMIT_WINDOW"), "15m")
	window, err := time.ParseDuration(windowRaw)
	if err != nil {
		errs = append(errs, fmt.Sprintf("LADEGAST_RATE_LIMIT_WINDOW: invalid duration %q", windowRaw))
	}
	cfg.RateLimitWindow = window

	if (cfg.AdminUsername == "") != (cfg.AdminPassword == "") {
		errs = append(errs, "LADEGAST_ADMIN_USERNAME and LADEGAST_ADMIN_PASSWORD must be set together")
	}

	if len(errs) > 0 {
		return Config{}, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return cfg, nil
}

func withDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func normalizePrefix(p string) string {
	p = strings.TrimSuffix(p, "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
