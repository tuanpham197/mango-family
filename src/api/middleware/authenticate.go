package middleware

import (
	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/tokenprovider/jwt"
	usermodel "household-finance/api/module/user/model"
	userstorage "household-finance/api/module/user/storage"

	"github.com/gin-gonic/gin"
)

// Authenticate — đọc JWT từ cookie HttpOnly, nạp user vào context (design §2).
func Authenticate(ac appctx.AppContext) gin.HandlerFunc {
	store := userstorage.NewSQLStore(ac.GetDB())
	return func(c *gin.Context) {
		unauth := common.NewUnauthorized(common.ErrCodeUnauthenticated, "vui lòng đăng nhập")
		tokenStr, err := c.Cookie(jwt.CookieName)
		if err != nil || tokenStr == "" {
			common.WriteError(c, unauth)
			return
		}
		claims, err := jwt.Validate(ac.GetSecret(), tokenStr)
		if err != nil {
			common.WriteError(c, unauth)
			return
		}
		u, err := store.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			common.WriteError(c, unauth)
			return
		}
		c.Set(common.CtxKeyUser, u)
		c.Next()
	}
}

// CurrentUser — lấy user đã authenticate từ context.
func CurrentUser(c *gin.Context) *usermodel.User {
	v, _ := c.Get(common.CtxKeyUser)
	u, _ := v.(*usermodel.User)
	return u
}
