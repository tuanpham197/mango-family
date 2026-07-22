package ginuser

import (
	"errors"
	"net/http"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/tokenprovider/jwt"
	"household-finance/api/config"
	"household-finance/api/middleware"
	householdstorage "household-finance/api/module/household/storage"
	userbiz "household-finance/api/module/user/biz"
	userstorage "household-finance/api/module/user/storage"

	"github.com/gin-gonic/gin"
	"time"
)

// setAuthCookie ghi/xóa cookie JWT theo cấu hình môi trường (Secure/SameSite/
// Domain). Cross-site (Vercel↔Cloud Run): COOKIE_SAMESITE=none → Secure=true.
func setAuthCookie(c *gin.Context, value string, maxAge int) {
	cc := config.Cookie()
	c.SetSameSite(cc.SameSite)
	c.SetCookie(jwt.CookieName, value, maxAge, "/", cc.Domain, cc.Secure, true)
}

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
		setAuthCookie(c, token, int(jwt.TokenTTL.Seconds()))
		common.WriteData(c, http.StatusOK, gin.H{"user": u.Public()})
	}
}

// Logout — POST /api/auth/logout: xóa cookie.
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		setAuthCookie(c, "", -1)
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
