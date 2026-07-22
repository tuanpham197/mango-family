package main

import (
	_ "embed"
	"net/http"
	"os"
	"strings"

	"household-finance/api/component/appctx"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var openAPISpec []byte

// registerDocs phục vụ OpenAPI spec + Swagger UI để dễ kiểm tra/thử API sau deploy.
// Tắt bằng DOCS_ENABLED=false. Trang /docs cùng origin với API nên "Try it out"
// gửi kèm cookie (credentials: include) — đăng nhập ngay trong trang là dùng được.
func registerDocs(r *gin.Engine) {
	if strings.EqualFold(os.Getenv("DOCS_ENABLED"), "false") {
		return
	}
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
	})
	r.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})
}

// healthz — kiểm tra app sống + ping DB. Public (không cần đăng nhập) để dùng
// cho uptime check / xác minh deploy.
func healthz(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := ac.GetDB().DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "db": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "ok"})
	}
}

const swaggerHTML = `<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Household Finance API — Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/openapi.yaml',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis],
      // Gửi kèm cookie phiên khi "Try it out" (cùng origin với API)
      requestInterceptor: function (req) { req.credentials = 'include'; return req; }
    });
  </script>
</body>
</html>`
