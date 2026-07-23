package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	JWT      JWTConfig
	Admin    AdminConfig
	WeChat   WeChatConfig
	Database DatabaseConfig
}

type ServerConfig struct{ Port, Env, CORSOrigin, BaseURL string }
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  int
	RefreshTokenTTL int
}
type AdminConfig struct{ SuperAdminPhone string }
type WeChatConfig struct{ AppID, AppSecret string }
type DatabaseConfig struct {
	UsePostgres bool
	DatabaseURL string
}
type ValidationResult struct {
	Warnings []string
	Errors   []string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:       envOrDefault("HTTP_ADDR", ":8080"),
			Env:        envOrDefault("ENV", "development"),
			CORSOrigin: envOrDefault("CORS_ORIGINS", "http://localhost:3000"),
			BaseURL:    envOrDefault("BASE_URL", "http://localhost:8080"),
		},
		JWT: JWTConfig{
			Secret:          os.Getenv("AUTH_SECRET"),
			AccessTokenTTL:  envOrDefaultInt("ACCESS_TOKEN_TTL", 900),
			RefreshTokenTTL: envOrDefaultInt("REFRESH_TOKEN_TTL", 604800),
		},
		Admin:    AdminConfig{SuperAdminPhone: envOrDefault("SUPER_ADMIN_PHONE", "")},
		WeChat:   WeChatConfig{AppID: os.Getenv("WECHAT_APPID"), AppSecret: os.Getenv("WECHAT_APPSECRET")},
		Database: DatabaseConfig{UsePostgres: os.Getenv("DATABASE_URL") != "", DatabaseURL: os.Getenv("DATABASE_URL")},
	}
}

func (c *Config) Validate() ValidationResult {
	var r ValidationResult
	if c.JWT.Secret == "" || len(c.JWT.Secret) < 32 {
		r.Errors = append(r.Errors, "AUTH_SECRET is required and must be at least 32 bytes")
	}
	if c.Server.Env == "production" {
		if c.WeChat.AppID == "" { r.Errors = append(r.Errors, "WECHAT_APPID is required in production") }
		if c.WeChat.AppSecret == "" { r.Errors = append(r.Errors, "WECHAT_APPSECRET is required in production") }
	}
	if c.Database.UsePostgres {
		if !strings.Contains(c.Database.DatabaseURL, "sslmode=require") && c.Server.Env == "production" {
			r.Warnings = append(r.Warnings, "PostgreSQL connection should use SSL in production")
		}
	} else {
		r.Warnings = append(r.Warnings, "DATABASE_URL not set, using in-memory storage (NOT FOR PRODUCTION)")
	}
	return r
}

func (c *Config) Print() {
	mask := func(s string) string {
		if len(s) > 8 { return s[:4] + "****" + s[len(s)-4:] }
		return "****"
	}
	fmt.Println("=== Configuration ===")
	fmt.Printf("  Server Port:     %s\n", c.Server.Port)
	fmt.Printf("  Environment:     %s\n", c.Server.Env)
	fmt.Printf("  Database:        %s\n", map[bool]string{true: "PostgreSQL", false: "JSON/Memory"}[c.Database.UsePostgres])
	if c.WeChat.AppID != "" { fmt.Printf("  WeChat AppID:    %s\n", c.WeChat.AppID) }
	if c.WeChat.AppSecret != "" { fmt.Printf("  WeChat Secret:   %s\n", mask(c.WeChat.AppSecret)) }
	fmt.Printf("  JWT Secret:      %s\n", mask(c.JWT.Secret))
	fmt.Println("======================")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" { return v }
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil { return n }
	}
	return fallback
}
