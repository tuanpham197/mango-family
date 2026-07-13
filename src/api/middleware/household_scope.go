package middleware

import (
	"errors"
	"net/http"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	householdstorage "household-finance/api/module/household/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireHousehold — tra membership, gắn household_id vào context (thay RLS — D5).
// Mọi storage query phía sau PHẢI filter theo household id này; ngoài hộ → 404.
func RequireHousehold(ac appctx.AppContext) gin.HandlerFunc {
	store := householdstorage.NewSQLStore(ac.GetDB())
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil {
			common.WriteError(c, common.NewUnauthorized(common.ErrCodeUnauthenticated, "vui lòng đăng nhập"))
			return
		}
		h, err := store.HouseholdOfUser(c.Request.Context(), u.ID)
		if err != nil {
			if errors.Is(err, householdstorage.ErrNoHousehold) {
				common.WriteError(c, common.NewAppError(http.StatusConflict, common.ErrCodeNoHousehold,
					"tài khoản chưa thuộc hộ gia đình nào — cần được thêm vào một hộ"))
				return
			}
			common.WriteError(c, common.NewInternal(err))
			return
		}
		c.Set(common.CtxKeyHouseholdID, h.ID)
		c.Next()
	}
}

// HouseholdID — lấy household id đã gắn bởi RequireHousehold.
func HouseholdID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(common.CtxKeyHouseholdID)
	id, _ := v.(uuid.UUID)
	return id
}
