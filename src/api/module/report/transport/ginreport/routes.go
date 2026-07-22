package ginreport

import (
	"time"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/middleware"
	reportbiz "household-finance/api/module/report/biz"
	reportstorage "household-finance/api/module/report/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// parseRange — đọc ?from=&to= (YYYY-MM-DD, inclusive). Thiếu/không hợp lệ → 400.
func parseRange(c *gin.Context) (from, to time.Time, ok bool) {
	f, err1 := time.ParseInLocation("2006-01-02", c.Query("from"), time.Local)
	t, err2 := time.ParseInLocation("2006-01-02", c.Query("to"), time.Local)
	if err1 != nil || err2 != nil {
		common.WriteError(c, common.NewBadRequest("thiếu hoặc sai định dạng from/to (YYYY-MM-DD)"))
		return time.Time{}, time.Time{}, false
	}
	return f, t, true
}

// Overview — GET /api/reports/overview?from=&to= (contracts §Reports).
func Overview(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		from, to, ok := parseRange(c)
		if !ok {
			return
		}
		biz := reportbiz.NewGetOverviewBiz(reportstorage.NewSQLStore(ac.GetDB()))
		res, err := biz.Get(c.Request.Context(), middleware.HouseholdID(c), from, to)
		if err != nil {
			common.WriteError(c, err)
			return
		}
		common.WriteOK(c, res)
	}
}

// Category — GET /api/reports/category/:id?from=&to= (contracts §Reports).
func Category(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			common.WriteError(c, common.NewNotFound("danh mục không tồn tại"))
			return
		}
		from, to, ok := parseRange(c)
		if !ok {
			return
		}
		biz := reportbiz.NewGetCategoryBiz(reportstorage.NewSQLStore(ac.GetDB()))
		res, err := biz.Get(c.Request.Context(), middleware.HouseholdID(c), id, from, to)
		if err != nil {
			common.WriteError(c, err)
			return
		}
		common.WriteOK(c, res)
	}
}
