package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWeChatLogin_Code2SessionBranches 覆盖 code2Session 全部分支：
// 成功 / errcode 非零 / openid 为空 / 响应非法 JSON / 网络错误 / 缺凭据。
func TestWeChatLogin_Code2SessionBranches(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 断言请求参数透传（appid/secret/js_code 均 URL 编码）
			q := r.URL.Query()
			if q.Get("appid") != "wx-app" || q.Get("secret") != "s3cret" || q.Get("js_code") != "code-1" {
				t.Errorf("unexpected query: %s", r.URL.RawQuery)
			}
			json.NewEncoder(w).Encode(map[string]string{"openid": "o-openid-1", "session_key": "sk-1"})
		}))
		defer ts.Close()
		old := wechatAPIBase
		wechatAPIBase = ts.URL
		defer func() { wechatAPIBase = old }()

		sess, err := WeChatLogin("code-1", "wx-app", "s3cret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sess.OpenID != "o-openid-1" || sess.SessionKey != "sk-1" {
			t.Fatalf("session mismatch: %+v", sess)
		}
	})

	t.Run("errcode nonzero", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"errcode": 40029, "errmsg": "invalid code"})
		}))
		defer ts.Close()
		old := wechatAPIBase
		wechatAPIBase = ts.URL
		defer func() { wechatAPIBase = old }()

		_, err := WeChatLogin("bad", "wx-app", "s3cret")
		if err == nil || !strings.Contains(err.Error(), "40029") {
			t.Fatalf("want errcode error, got: %v", err)
		}
	})

	t.Run("empty openid", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"session_key": "sk"})
		}))
		defer ts.Close()
		old := wechatAPIBase
		wechatAPIBase = ts.URL
		defer func() { wechatAPIBase = old }()

		if _, err := WeChatLogin("c", "wx-app", "s3cret"); err == nil {
			t.Fatal("want empty-openid error")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not-json{"))
		}))
		defer ts.Close()
		old := wechatAPIBase
		wechatAPIBase = ts.URL
		defer func() { wechatAPIBase = old }()

		if _, err := WeChatLogin("c", "wx-app", "s3cret"); err == nil {
			t.Fatal("want parse error")
		}
	})

	t.Run("network error", func(t *testing.T) {
		// 指向一个拒绝连接的地址（127.0.0.1:1 通常立即 connection refused）
		old := wechatAPIBase
		wechatAPIBase = "http://127.0.0.1:1"
		defer func() { wechatAPIBase = old }()

		if _, err := WeChatLogin("c", "wx-app", "s3cret"); err == nil {
			t.Fatal("want network error")
		}
	})

	t.Run("missing credentials", func(t *testing.T) {
		if _, err := WeChatLogin("c", "", ""); err == nil {
			t.Fatal("want credentials error")
		}
	})
}
