package config_test

import (
	"os"
	"testing"

	"drone-platform/internal/config"
)

// 平台配置保存回归：费率范围校验（0~100）+ 原子写（临时文件不残留）。
func TestSavePlatformConfigValidation(t *testing.T) {
	// 负数费率 → 拒绝
	if err := config.SavePlatformConfig(config.PlatformConfig{MatchFeeRate: -1}); err == nil {
		t.Fatal("negative fee rate must be rejected")
	}
	// 超 100 → 拒绝
	if err := config.SavePlatformConfig(config.PlatformConfig{MatchFeeRate: 101}); err == nil {
		t.Fatal("fee rate > 100 must be rejected")
	}
	// 合法值 → 保存成功，且无 .tmp 残留
	if err := config.SavePlatformConfig(config.PlatformConfig{MatchFeeRate: 2, MatchFeeNote: "test"}); err != nil {
		t.Fatalf("save valid config: %v", err)
	}
	if _, err := os.Stat("platform_config.json.tmp"); !os.IsNotExist(err) {
		t.Fatal("tmp file must not remain after save")
	}
	got := config.GetPlatformConfig()
	if got.MatchFeeRate != 2 {
		t.Fatalf("saved rate: %v", got.MatchFeeRate)
	}
	// 清理测试产物，避免影响其他测试
	os.Remove("platform_config.json")
	os.Remove("platform_config.json.tmp")
}
