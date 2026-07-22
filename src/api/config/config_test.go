package config

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCookie_SameSiteNoneForcesSecure(t *testing.T) {
	t.Setenv("COOKIE_SAMESITE", "none")
	t.Setenv("COOKIE_SECURE", "false") // dù đặt false, None vẫn buộc Secure
	cc := Cookie()
	assert.Equal(t, http.SameSiteNoneMode, cc.SameSite)
	assert.True(t, cc.Secure, "SameSite=None phải kèm Secure=true")
}

func TestCookie_DefaultLaxInsecure(t *testing.T) {
	// Không đặt env → mặc định dev: Lax, không Secure (giữ hành vi cũ).
	cc := Cookie()
	assert.Equal(t, http.SameSiteLaxMode, cc.SameSite)
	assert.False(t, cc.Secure)
	assert.Empty(t, cc.Domain)
}

func TestCookie_DomainAndStrict(t *testing.T) {
	t.Setenv("COOKIE_SAMESITE", "strict")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("COOKIE_DOMAIN", ".example.com")
	cc := Cookie()
	assert.Equal(t, http.SameSiteStrictMode, cc.SameSite)
	assert.True(t, cc.Secure)
	assert.Equal(t, ".example.com", cc.Domain)
}

func TestCORSOrigins_ParseAndTrim(t *testing.T) {
	t.Setenv("CORS_ORIGINS", " https://a.vercel.app , http://localhost:5173 ,")
	assert.Equal(t, []string{"https://a.vercel.app", "http://localhost:5173"}, CORSOrigins())
}

func TestCORSOrigins_EmptyIsNil(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")
	assert.Nil(t, CORSOrigins())
}

func TestDBPool_Defaults(t *testing.T) {
	p := DBPoolConfig()
	assert.Equal(t, 10, p.MaxOpen)
	assert.Equal(t, 5, p.MaxIdle)
	assert.False(t, p.PreferSimpleProtocol)
}
