package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WeChatSession is returned by code2Session.
type WeChatSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WeChatLogin calls the WeChat code2Session API.
func WeChatLogin(code, appID, appSecret string) (WeChatSession, error) {
	if appID == "" || appSecret == "" {
		return WeChatSession{}, fmt.Errorf("WeChat AppID and AppSecret are required")
	}
	u := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(appID), url.QueryEscape(appSecret), url.QueryEscape(code))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return WeChatSession{}, fmt.Errorf("wechat api: %w", err)
	}
	defer resp.Body.Close()

	var sess WeChatSession
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return WeChatSession{}, fmt.Errorf("parse wechat response: %w", err)
	}
	if sess.ErrCode != 0 {
		return sess, fmt.Errorf("wechat error %d: %s", sess.ErrCode, sess.ErrMsg)
	}
	if sess.OpenID == "" {
		return sess, fmt.Errorf("wechat returned empty openid")
	}
	return sess, nil
}

// HashToken returns the SHA-256 hash of a token string.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// GenerateRefreshToken creates a random 32-byte base64 token.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
