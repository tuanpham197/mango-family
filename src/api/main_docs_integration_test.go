//go:build integration

// Integration test: health check + tài liệu API (OpenAPI/Swagger).
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	app := newTestApp(t)
	w := app.do(http.MethodGet, "/healthz", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	assert.Contains(t, w.Body.String(), `"db":"ok"`)
}

func TestDocsServed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerDocs(r)
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}

	// /openapi.yaml phục vụ spec YAML
	w := get("/openapi.yaml")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "yaml")
	assert.Contains(t, w.Body.String(), "openapi: 3.0")

	// /docs phục vụ Swagger UI
	w = get("/docs")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "SwaggerUIBundle")
}
