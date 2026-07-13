package gincategory

import (
	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/middleware"
	categorybiz "household-finance/api/module/category/biz"
	categorystorage "household-finance/api/module/category/storage"

	"github.com/gin-gonic/gin"
)

// List — GET /api/categories?type=&include_hidden= (contracts §Categories).
func List(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		biz := categorybiz.NewListCategoriesBiz(categorystorage.NewSQLStore(ac.GetDB()))
		trees, err := biz.List(
			c.Request.Context(),
			middleware.HouseholdID(c),
			c.Query("type"),
			c.Query("include_hidden") == "true",
		)
		if err != nil {
			common.WriteError(c, err)
			return
		}
		common.WriteOK(c, trees)
	}
}
