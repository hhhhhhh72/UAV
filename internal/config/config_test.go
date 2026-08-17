package config_test

import (
	"os"
	"strings"
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

// TestValidateProductionHardChecks: P0 回归——生产环境必须硬校验
// DATABASE_URL / ENCRYPTION_KEY / SIGNING_SECRET，并拒绝 ADMIN_DEV_MODE，
// 任一缺失即启动失败（此前仅 WARN/静默，误配即沦陷）。
func TestValidateProductionHardChecks(t *testing.T) {
	os.Setenv("ENV", "production")
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	os.Setenv("WECHAT_APPID", "wx123")
	os.Setenv("WECHAT_APPSECRET", "sec123")
	os.Setenv("ADMIN_DEV_MODE", "true")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("ENCRYPTION_KEY")
	os.Unsetenv("SIGNING_SECRET")
	t.Cleanup(func() {
		os.Unsetenv("ENV")
		os.Unsetenv("ADMIN_DEV_MODE")
		os.Unsetenv("WECHAT_APPID")
		os.Unsetenv("WECHAT_APPSECRET")
	})

	cfg := config.Load()
	res := cfg.Validate()
	joined := strings.Join(res.Errors, " | ")
	for _, want := range []string{"DATABASE_URL", "ENCRYPTION_KEY", "SIGNING_SECRET", "ADMIN_DEV_MODE"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("production validation missing %q; errors: %v", want, res.Errors)
		}
	}
}

// TestValidateProductionOK: 生产环境变量齐全时不应报错。
func TestValidateProductionOK(t *testing.T) {
	os.Setenv("ENV", "production")
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	os.Setenv("WECHAT_APPID", "wx123")
	os.Setenv("WECHAT_APPSECRET", "sec123")
	os.Setenv("DATABASE_URL", "postgres://u:p@host/db?sslmode=require")
	os.Setenv("ENCRYPTION_KEY", "dGVzdC1rZXktMzJieXRlcy0wMTIzNDU2Nzg5MA==")
	os.Setenv("SIGNING_SECRET", "sig-secret")
	os.Unsetenv("ADMIN_DEV_MODE")
	t.Cleanup(func() {
		os.Unsetenv("ENV")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("ENCRYPTION_KEY")
		os.Unsetenv("SIGNING_SECRET")
		os.Unsetenv("WECHAT_APPID")
		os.Unsetenv("WECHAT_APPSECRET")
	})

	cfg := config.Load()
	res := cfg.Validate()
	if len(res.Errors) > 0 {
		t.Fatalf("production with complete env should pass; errors: %v", res.Errors)
	}
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

// 商业化费率（功能方案修订版 v2 第八章）：费率恒为非负，默认未启用且有说明文案。
func TestPlatformConfigMatchFee(t *testing.T) {
	pc := config.GetPlatformConfig()
	if pc.MatchFeeRate < 0 {
		t.Fatalf("MatchFeeRate should never be negative, got %v", pc.MatchFeeRate)
	}
	if pc.MatchFeeRate == 0 && pc.MatchFeeNote == "" {
		t.Fatalf("MatchFeeNote should explain the disabled state")
	}
}
