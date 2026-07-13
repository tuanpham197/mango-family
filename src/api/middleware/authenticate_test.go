package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"household-finance/api/component/appctx"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func testEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ac := appctx.New(nil, "test-secret", nil, nil)
	r.GET("/protected", Authenticate(ac), func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestAuthenticate_NoCookie401(t *testing.T) {
	r := testEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHENTICATED")
}

func TestAuthenticate_InvalidToken401(t *testing.T) {
	r := testEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "hf_token", Value: "rác-không-phải-jwt"})
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHENTICATED")
}
