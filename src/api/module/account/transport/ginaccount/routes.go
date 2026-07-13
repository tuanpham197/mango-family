package ginaccount

import (
	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/middleware"
	accountbiz "household-finance/api/module/account/biz"
	accountstorage "household-finance/api/module/account/storage"

	"github.com/gin-gonic/gin"
)

// List — GET /api/accounts (contracts §Accounts): [{id, name, type, balance}].
func List(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		biz := accountbiz.NewListAccountsBiz(accountstorage.NewSQLStore(ac.GetDB()))
		items, err := biz.List(c.Request.Context(), middleware.HouseholdID(c))
		if err != nil {
			common.WriteError(c, err)
			return
		}
		common.WriteOK(c, items)
	}
}
