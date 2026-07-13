package gincategory

import (
	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/middleware"
	categorybiz "household-finance/api/module/category/biz"
	categorystorage "household-finance/api/module/category/storage"

	"github.com/gin-gonic/gin"
)

// Suggest — GET /api/categories/suggest?description=&type= (FR-015, D7).
// Không match → {data: null}.
func Suggest(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		biz := categorybiz.NewSuggestCategoryBiz(categorystorage.NewSQLStore(ac.GetDB()))
		suggestion, err := biz.Suggest(
			c.Request.Context(),
			middleware.HouseholdID(c),
			c.Query("description"),
			c.Query("type"),
		)
		if err != nil {
			common.WriteError(c, err)
			return
		}
		if suggestion == nil {
			common.WriteOK(c, nil)
			return
		}
		common.WriteOK(c, suggestion)
	}
}
