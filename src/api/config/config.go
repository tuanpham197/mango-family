// Package config đọc cấu hình runtime từ biến môi trường (12-factor).
// Dùng cho deploy tách origin (web trên Vercel ↔ API trên Cloud Run) + Supabase.
package config

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func atoi(key string, def int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil && v > 0 {
		return v
	}
	return def
}

func boolEnv(key string, def bool) bool {
	switch strings.ToLower(env(key, "")) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return def
	}
}

// CORSOrigins — allow-list origin cho CORS (phân tách bằng dấu phẩy).
// Rỗng = không cho phép cross-origin (chỉ same-origin/curl). Ví dụ:
// CORS_ORIGINS="https://app.vercel.app,http://localhost:5173"
func CORSOrigins() []string {
	raw := env("CORS_ORIGINS", "")
	if raw == "" {
		return nil
	}
	var out []string
	for o := range strings.SplitSeq(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// CookieConfig — thuộc tính cookie phiên (JWT). Cross-site (Vercel↔Cloud Run)
// cần SameSite=None + Secure. Same-site (custom domain chung) dùng Lax.
type CookieConfig struct {
	Secure   bool
	SameSite http.SameSite
	Domain   string // rỗng = host-only; ".example.com" để chia sẻ giữa subdomain
}

// Cookie đọc COOKIE_SECURE / COOKIE_SAMESITE (lax|none|strict) / COOKIE_DOMAIN.
// SameSite=None luôn kèm Secure=true (yêu cầu của trình duyệt).
func Cookie() CookieConfig {
	ss := http.SameSiteLaxMode
	switch strings.ToLower(env("COOKIE_SAMESITE", "lax")) {
	case "none":
		ss = http.SameSiteNoneMode
	case "strict":
		ss = http.SameSiteStrictMode
	}
	secure := boolEnv("COOKIE_SECURE", false)
	if ss == http.SameSiteNoneMode {
		secure = true
	}
	return CookieConfig{Secure: secure, SameSite: ss, Domain: env("COOKIE_DOMAIN", "")}
}

// DBPool — giới hạn connection pool (quan trọng với Supabase pooler: số kết nối
// hữu hạn) + cờ simple-protocol cho transaction pooler (cổng 6543).
type DBPool struct {
	MaxOpen              int
	MaxIdle              int
	MaxLifetime          time.Duration
	MaxIdleTime          time.Duration
	PreferSimpleProtocol bool
}

func DBPoolConfig() DBPool {
	return DBPool{
		MaxOpen:              atoi("DB_MAX_OPEN_CONNS", 10),
		MaxIdle:              atoi("DB_MAX_IDLE_CONNS", 5),
		MaxLifetime:          time.Duration(atoi("DB_CONN_MAX_LIFETIME_SEC", 300)) * time.Second,
		MaxIdleTime:          time.Duration(atoi("DB_CONN_MAX_IDLE_SEC", 60)) * time.Second,
		PreferSimpleProtocol: boolEnv("DB_PREFER_SIMPLE_PROTOCOL", false),
	}
}
