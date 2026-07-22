package ginoverview

import (
	"time"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/middleware"
	overviewbiz "household-finance/api/module/overview/biz"
	overviewstorage "household-finance/api/module/overview/storage"

	"github.com/gin-gonic/gin"
)

// Get — GET /api/overview (contracts/overview-api.md): tổng hợp giá trị suy ra của
// màn Tổng quan (tài sản ròng + % tháng trước, Thu/Chi tháng, chi theo danh mục, giao
// dịch gần đây). Chỉ đọc; phạm vi hộ ở middleware (D5/001).
func Get(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		biz := overviewbiz.NewGetOverviewBiz(overviewstorage.NewSQLStore(ac.GetDB()))
		summary, err := biz.Get(c.Request.Context(), middleware.HouseholdID(c), time.Now())
		if err != nil {
			common.WriteError(c, err)
			return
		}
		common.WriteOK(c, summary)
	}
}
