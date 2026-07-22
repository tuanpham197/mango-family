package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS cho phép các origin trong allow-list gọi API kèm cookie (credentials).
// Cần cho kiến trúc tách origin: web (Vercel) ↔ API (Cloud Run). Origin không
// nằm trong danh sách sẽ không nhận header CORS → trình duyệt tự chặn.
//
// Lưu ý: có credentials nên KHÔNG dùng "*" — phải echo đúng origin được phép.
func CORS(origins []string) gin.HandlerFunc {
	allow := make(map[string]bool, len(origins))
	for _, o := range origins {
		allow[o] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := origin != "" && allow[origin]

		if allowed {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Add("Vary", "Origin")
			if c.Request.Method == http.MethodOptions {
				reqHeaders := c.GetHeader("Access-Control-Request-Headers")
				if reqHeaders == "" {
					reqHeaders = "Content-Type"
				}
				h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", reqHeaders)
				h.Set("Access-Control-Max-Age", "3600")
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		} else if c.Request.Method == http.MethodOptions {
			// Preflight từ origin không được phép → 204 trơn (không có header cho phép).
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
