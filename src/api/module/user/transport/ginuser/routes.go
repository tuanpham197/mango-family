package ginuser

import (
	"errors"
	"net/http"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/tokenprovider/jwt"
	"household-finance/api/middleware"
	userbiz "household-finance/api/module/user/biz"
	userstorage "household-finance/api/module/user/storage"
	householdstorage "household-finance/api/module/household/storage"

	"github.com/gin-gonic/gin"
	"time"
)

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login — POST /api/auth/login: xác thực, set cookie JWT HttpOnly SameSite=Lax.
func Login(ac appctx.AppContext) gin.HandlerFunc {
	biz := userbiz.NewLoginBiz(userstorage.NewSQLStore(ac.GetDB()))
	return func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ"))
			return
		}
		u, err := biz.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			common.WriteError(c, err)
			return
		}
		token, err := jwt.Generate(ac.GetSecret(), u.ID, time.Now())
		if err != nil {
			common.WriteError(c, common.NewInternal(err))
			return
		}
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(jwt.CookieName, token, int(jwt.TokenTTL.Seconds()), "/", "", false, true)
		common.WriteData(c, http.StatusOK, gin.H{"user": u.Public()})
	}
}

// Logout — POST /api/auth/logout: xóa cookie.
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(jwt.CookieName, "", -1, "/", "", false, true)
		c.Status(http.StatusNoContent)
	}
}

// Me — GET /api/me (sau Authenticate): hồ sơ + hộ; chưa thuộc hộ → 409 NO_HOUSEHOLD.
func Me(ac appctx.AppContext) gin.HandlerFunc {
	hstore := householdstorage.NewSQLStore(ac.GetDB())
	return func(c *gin.Context) {
		u := middleware.CurrentUser(c)
		h, err := hstore.HouseholdOfUser(c.Request.Context(), u.ID)
		if err != nil {
			if errors.Is(err, householdstorage.ErrNoHousehold) {
				common.WriteError(c, common.NewAppError(http.StatusConflict, common.ErrCodeNoHousehold,
					"tài khoản chưa thuộc hộ gia đình nào — cần được thêm vào một hộ"))
				return
			}
			common.WriteError(c, common.NewInternal(err))
			return
		}
		common.WriteOK(c, gin.H{
			"user":      u.Public(),
			"household": gin.H{"id": h.ID, "name": h.Name},
		})
	}
}
