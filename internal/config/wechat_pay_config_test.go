package config_test

import (
	"os"
	"strings"
	"testing"

	"drone-platform/internal/config"
)

// 微信支付配置（真实资金）的校验回归。
//
// 半配置是最危险的形态：有商户号却没私钥，会在真实收款时才失败——那时用户的钱可能已经
// 进了商户号但平台账上没入账。因此规则是"五项齐全或一项都不填"，启动即拒绝。
func TestWeChatPayConfigPartialIsRejected(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	os.Setenv("WECHAT_PAY_MCHID", "1900000109")
	os.Unsetenv("WECHAT_PAY_API_V3_KEY")
	os.Unsetenv("WECHAT_PAY_CERT_SERIAL")
	os.Unsetenv("WECHAT_PAY_PRIVATE_KEY_PATH")
	os.Unsetenv("WECHAT_PAY_NOTIFY_URL")
	t.Cleanup(func() {
		os.Unsetenv("AUTH_SECRET")
		os.Unsetenv("WECHAT_PAY_MCHID")
	})

	cfg := config.Load()
	if cfg.WeChatPay.Configured() {
		t.Fatal("只填商户号不应视为已配置")
	}
	res := cfg.Validate()
	joined := strings.Join(res.Errors, " | ")
	if !strings.Contains(joined, "WECHAT_PAY_MCHID") || !strings.Contains(joined, "不完整") {
		t.Fatalf("半配置应启动失败并指出已设置的项，实际 errors=%v", res.Errors)
	}
}

// 一项都不填＝真实支付未开通：不报错（资金继续走 internal_* 内部记账）。
func TestWeChatPayConfigAbsentIsOK(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	for _, k := range []string{"WECHAT_PAY_MCHID", "WECHAT_PAY_API_V3_KEY", "WECHAT_PAY_CERT_SERIAL", "WECHAT_PAY_PRIVATE_KEY_PATH", "WECHAT_PAY_NOTIFY_URL"} {
		os.Unsetenv(k)
	}
	t.Cleanup(func() { os.Unsetenv("AUTH_SECRET") })

	cfg := config.Load()
	if cfg.WeChatPay.Configured() {
		t.Fatal("未配置时 Configured() 应为 false")
	}
	for _, e := range cfg.Validate().Errors {
		if strings.Contains(e, "微信支付") || strings.Contains(e, "WECHAT_PAY") {
			t.Fatalf("未配置微信支付不应报错，实际: %s", e)
		}
	}
}

// 五项齐全 → 视为已配置；生产环境下回调地址必须是 https。
func TestWeChatPayConfigComplete(t *testing.T) {
	os.Setenv("AUTH_SECRET", "test-secret-key-32bytes-0123456789")
	os.Setenv("WECHAT_PAY_MCHID", "1900000109")
	os.Setenv("WECHAT_PAY_API_V3_KEY", "0123456789012345678901234567890a")
	os.Setenv("WECHAT_PAY_CERT_SERIAL", "ABCDEF1234567890")
	os.Setenv("WECHAT_PAY_PRIVATE_KEY_PATH", "/certs/apiclient_key.pem")
	os.Setenv("WECHAT_PAY_NOTIFY_URL", "https://api.cqnarc.cn/api/v1/payments/wechat/notify")
	t.Cleanup(func() {
		for _, k := range []string{"AUTH_SECRET", "WECHAT_PAY_MCHID", "WECHAT_PAY_API_V3_KEY", "WECHAT_PAY_CERT_SERIAL", "WECHAT_PAY_PRIVATE_KEY_PATH", "WECHAT_PAY_NOTIFY_URL"} {
			os.Unsetenv(k)
		}
	})

	cfg := config.Load()
	if !cfg.WeChatPay.Configured() {
		t.Fatal("五项齐全应视为已配置")
	}
	for _, e := range cfg.Validate().Errors {
		if strings.Contains(e, "WECHAT_PAY") || strings.Contains(e, "微信支付") {
			t.Fatalf("完整配置不应报错，实际: %s", e)
		}
	}

	// 生产环境：http 回调地址必须被拒绝（微信支付只回调 https）
	os.Setenv("ENV", "production")
	os.Setenv("WECHAT_PAY_NOTIFY_URL", "http://api.cqnarc.cn/api/v1/payments/wechat/notify")
	os.Setenv("DATABASE_URL", "postgres://u:p@db:5432/x?sslmode=require")
	os.Setenv("ENCRYPTION_KEY", "01234567890123456789012345678901")
	os.Setenv("SIGNING_SECRET", "01234567890123456789012345678901")
	os.Setenv("WECHAT_APPID", "wx123")
	os.Setenv("WECHAT_APPSECRET", "sec123")
	t.Cleanup(func() {
		for _, k := range []string{"ENV", "DATABASE_URL", "ENCRYPTION_KEY", "SIGNING_SECRET", "WECHAT_APPID", "WECHAT_APPSECRET"} {
			os.Unsetenv(k)
		}
	})
	res := config.Load().Validate()
	joined := strings.Join(res.Errors, " | ")
	if !strings.Contains(joined, "WECHAT_PAY_NOTIFY_URL must be https") {
		t.Fatalf("生产环境 http 回调地址应被拒绝，实际 errors=%v", res.Errors)
	}
}
