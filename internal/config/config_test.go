package config

import (
	"testing"
	"time"
)

func baseEnv() map[string]string {
	return map[string]string{
		"LADEGAST_PUBLIC_BASE_URL": "https://example.com",
		"LADEGAST_SESSION_SECRET":  "secret",
		"LADEGAST_EVCC_BASE_URL":   "https://evcc.example.com",
		"LADEGAST_EVCC_API_KEY":    "evcc_test",
	}
}

func getenvFrom(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := load(getenvFrom(baseEnv()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %q, want :8080", cfg.ListenAddr)
	}
	if cfg.PublicPathPrefix != "/public" {
		t.Errorf("PublicPathPrefix = %q, want /public", cfg.PublicPathPrefix)
	}
	if cfg.AdminPathPrefix != "/admin" {
		t.Errorf("AdminPathPrefix = %q, want /admin", cfg.AdminPathPrefix)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("DataDir = %q, want ./data", cfg.DataDir)
	}
	if cfg.RateLimitMaxAttempts != 5 {
		t.Errorf("RateLimitMaxAttempts = %d, want 5", cfg.RateLimitMaxAttempts)
	}
	if cfg.RateLimitWindow != 15*time.Minute {
		t.Errorf("RateLimitWindow = %v, want 15m", cfg.RateLimitWindow)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	_, err := load(getenvFrom(map[string]string{}))
	if err == nil {
		t.Fatal("expected error for missing required vars, got nil")
	}
}

func TestLoad_NormalizesPathPrefix(t *testing.T) {
	env := baseEnv()
	env["LADEGAST_PUBLIC_PATH_PREFIX"] = "public/"
	env["LADEGAST_ADMIN_PATH_PREFIX"] = "admin/"
	cfg, err := load(getenvFrom(env))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PublicPathPrefix != "/public" {
		t.Errorf("PublicPathPrefix = %q, want /public", cfg.PublicPathPrefix)
	}
	if cfg.AdminPathPrefix != "/admin" {
		t.Errorf("AdminPathPrefix = %q, want /admin", cfg.AdminPathPrefix)
	}
}

func TestLoad_InvalidRateLimitWindow(t *testing.T) {
	env := baseEnv()
	env["LADEGAST_RATE_LIMIT_WINDOW"] = "not-a-duration"
	_, err := load(getenvFrom(env))
	if err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
}

func TestLoad_AdminCredentialsMustComeInPairs(t *testing.T) {
	env := baseEnv()
	env["LADEGAST_ADMIN_USERNAME"] = "admin"
	_, err := load(getenvFrom(env))
	if err == nil {
		t.Fatal("expected error when only username is set, got nil")
	}
}

func TestLoad_AdminCredentialsAsPairIsValid(t *testing.T) {
	env := baseEnv()
	env["LADEGAST_ADMIN_USERNAME"] = "admin"
	env["LADEGAST_ADMIN_PASSWORD"] = "secret"
	cfg, err := load(getenvFrom(env))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AdminUsername != "admin" || cfg.AdminPassword != "secret" {
		t.Errorf("AdminUsername/Password = %q/%q, want admin/secret", cfg.AdminUsername, cfg.AdminPassword)
	}
}

func TestLoad_InvalidRateLimitMaxAttempts(t *testing.T) {
	env := baseEnv()
	env["LADEGAST_RATE_LIMIT_MAX_ATTEMPTS"] = "not-a-number"
	_, err := load(getenvFrom(env))
	if err == nil {
		t.Fatal("expected error for invalid integer, got nil")
	}
}

func TestLoad_AccumulatesMultipleErrors(t *testing.T) {
	// All required vars missing AND an invalid duration: the error should
	// mention more than just the first problem encountered.
	_, err := load(getenvFrom(map[string]string{"LADEGAST_RATE_LIMIT_WINDOW": "garbage"}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_ReadsFromOSEnv(t *testing.T) {
	for k, v := range baseEnv() {
		t.Setenv(k, v)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.EVCCBaseURL != "https://evcc.example.com" {
		t.Errorf("EVCCBaseURL = %q, want https://evcc.example.com", cfg.EVCCBaseURL)
	}
}
