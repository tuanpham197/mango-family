package middleware

import (
	"log"
	"net/http"

	"household-finance/api/common"

	"github.com/gin-gonic/gin"
)

// Recover — bắt panic, trả 500 theo format app_response.
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v", r)
				common.WriteError(c, common.NewAppError(http.StatusInternalServerError, common.ErrCodeInternal, "đã có lỗi xảy ra, vui lòng thử lại"))
			}
		}()
		c.Next()
	}
}
