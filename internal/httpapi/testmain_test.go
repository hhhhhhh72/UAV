package httpapi_test

import (
	"os"
	"testing"
)

// TestMain 统一测试环境：安全加固后（adminDevMode fail-closed）要求
// APP_ENV/ENV 显式声明 dev/test/development 才启用 ADMIN_DEV_MODE 路径。
// 测试即为 dev 场景，这里统一声明，避免各测试分别 Setenv 遗漏导致
// dev-only 路由（SMS/H5 JSON 路由/admin dev token）被误拦。
func TestMain(m *testing.M) {
	if os.Getenv("APP_ENV") == "" && os.Getenv("ENV") == "" {
		_ = os.Setenv("APP_ENV", "development")
	}
	os.Exit(m.Run())
}
