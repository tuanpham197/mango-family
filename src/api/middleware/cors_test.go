package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func corsRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"https://app.vercel.app"}))
	r.GET("/api/x", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestCORS_AllowedOrigin(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Origin", "https://app.vercel.app")
	corsRouter().ServeHTTP(w, req)

	assert.Equal(t, "https://app.vercel.app", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORS_DisallowedOrigin_NoHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	corsRouter().ServeHTTP(w, req)

	// Không echo origin → trình duyệt tự chặn phản hồi cross-origin.
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_Preflight_Allowed(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/x", nil)
	req.Header.Set("Origin", "https://app.vercel.app")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	corsRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://app.vercel.app", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PATCH")
	assert.Equal(t, "content-type", w.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORS_NoOrigin_Passthrough(t *testing.T) {
	// Same-origin / curl (không có Origin) → không đụng CORS, request chạy bình thường.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	corsRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}
