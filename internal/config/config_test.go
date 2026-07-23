package config_test

import (
	"os"
	"testing"

	"drone-platform/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("AUTH_SECRET")
	os.Unsetenv("HTTP_ADDR")
	cfg := config.Load()
	if cfg.Server.Port != ":8080" { t.Fatalf("default port: %s", cfg.Server.Port) }
	if cfg.Server.Env != "development" { t.Fatalf("default env: %s", cfg.Server.Env) }
}

func TestLoadWithEnv(t *testing.T) {
	os.Setenv("HTTP_ADDR", ":9090")
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	defer os.Unsetenv("HTTP_ADDR")
	defer os.Unsetenv("AUTH_SECRET")
	cfg := config.Load()
	if cfg.Server.Port != ":9090" { t.Fatalf("custom port: %s", cfg.Server.Port) }
}

func TestValidateMissingSecret(t *testing.T) {
	os.Unsetenv("AUTH_SECRET")
	cfg := config.Load()
	result := cfg.Validate()
	if len(result.Errors) == 0 { t.Fatal("should have errors for missing AUTH_SECRET") }
}

func TestValidateWithSecret(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	defer os.Unsetenv("AUTH_SECRET")
	cfg := config.Load()
	result := cfg.Validate()
	if len(result.Errors) > 0 { t.Fatalf("should not have errors: %v", result.Errors) }
}

func TestValidateShortSecret(t *testing.T) {
	os.Setenv("AUTH_SECRET", "short")
	defer os.Unsetenv("AUTH_SECRET")
	cfg := config.Load()
	result := cfg.Validate()
	if len(result.Errors) == 0 { t.Fatal("should have error for short AUTH_SECRET") }
}

func TestPrint(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	defer os.Unsetenv("AUTH_SECRET")
	cfg := config.Load()
	cfg.Print() // Should not panic
}

func TestPlatformConfig(t *testing.T) {
	pc := config.GetPlatformConfig()
	if len(pc.Banners) == 0 && len(pc.QuickEntries) == 0 {
		t.Log("platform config has defaults or is empty")
	}
	// Save and reload
	if err := config.SavePlatformConfig(pc); err != nil {
		t.Fatalf("save platform config: %v", err)
	}
}
