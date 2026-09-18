package wechatpay

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestKey 生成一把临时 RSA 私钥并落盘（PKCS#8 PEM），返回路径。
// 微信商户私钥就是这个格式，用它才能让 New() 走完整条构造路径。
func writeTestKey(t *testing.T) string {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成测试私钥: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		t.Fatalf("编码私钥: %v", err)
	}
	path := filepath.Join(t.TempDir(), "apiclient_key.pem")
	blk := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	if err := os.WriteFile(path, blk, 0o600); err != nil {
		t.Fatalf("写私钥: %v", err)
	}
	return path
}

// TestDefaultAPIBaseIsCleanURL 守住一个只在生产才会发作的坑：
// defaultAPIBase 曾写成 "`https://api.mch.weixin.qq.com`"（字符串里真的带反引号），
// 而 Config.APIBase 只有测试会覆盖 → 单测全绿、生产每笔下单都 DNS 解析失败。
func TestDefaultAPIBaseIsCleanURL(t *testing.T) {
	if strings.ContainsAny(defaultAPIBase, "`\"' ") {
		t.Fatalf("默认基址含引号/空格等非法字符：%q", defaultAPIBase)
	}
	u, err := url.Parse(defaultAPIBase)
	if err != nil {
		t.Fatalf("默认基址不是合法 URL：%v", err)
	}
	if u.Scheme != "https" {
		t.Errorf("默认基址必须是 https，当前 %q", u.Scheme)
	}
	if u.Host != "api.mch.weixin.qq.com" {
		t.Errorf("默认基址域名不对：%q", u.Host)
	}
}

// TestNewFallsBackToDefaultBase 确认 APIBase 留空时用的是默认域名。
func TestNewFallsBackToDefaultBase(t *testing.T) {
	c, err := New(Config{
		AppID: "wx-test", MchID: "1900000109",
		APIv3Key: "0123456789012345678901234567890a",
		CertSerial: "ABCDEF1234567890", PrivateKeyPath: writeTestKey(t),
		NotifyURL: "https://example.com/notify",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.baseURL != defaultAPIBase {
		t.Errorf("APIBase 留空时应回落到 %q，实际 %q", defaultAPIBase, c.baseURL)
	}
	if strings.HasSuffix(c.baseURL, "/") {
		t.Errorf("基址尾部不应有斜杠：%q", c.baseURL)
	}
}

// TestNewRejectsIncompleteConfig 五项缺一不可——缺项必须在启动时就暴露。
func TestNewRejectsIncompleteConfig(t *testing.T) {
	full := Config{
		AppID: "wx-test", MchID: "1900000109",
		APIv3Key: "0123456789012345678901234567890a",
		CertSerial: "ABCDEF1234567890", PrivateKeyPath: writeTestKey(t),
		NotifyURL: "https://example.com/notify",
	}
	cases := map[string]func(c *Config){
		"缺商户号":     func(c *Config) { c.MchID = "" },
		"缺APIv3密钥": func(c *Config) { c.APIv3Key = "" },
		"缺证书序列号":   func(c *Config) { c.CertSerial = "" },
		"缺私钥路径":    func(c *Config) { c.PrivateKeyPath = "" },
		"缺回调地址":    func(c *Config) { c.NotifyURL = "" },
		"密钥长度不足":   func(c *Config) { c.APIv3Key = "short" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := full
			mutate(&cfg)
			if _, err := New(cfg); err == nil {
				t.Fatalf("%s：应当报错但通过了", name)
			}
		})
	}
}
