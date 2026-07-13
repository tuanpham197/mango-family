package biz

import (
	"context"

	"household-finance/api/common"
	"household-finance/api/component/hasher"
	"household-finance/api/module/user/model"
)

type UserFinder interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type LoginBiz struct{ store UserFinder }

func NewLoginBiz(store UserFinder) *LoginBiz { return &LoginBiz{store: store} }

// Login xác thực email + mật khẩu; sai bất kỳ trường nào cũng trả cùng một lỗi
// INVALID_CREDENTIALS — không tiết lộ trường sai (UC-TRK-01 E1).
func (b *LoginBiz) Login(ctx context.Context, email, password string) (*model.User, error) {
	invalid := common.NewUnauthorized(common.ErrCodeInvalidCredentials, "email hoặc mật khẩu không đúng")
	if email == "" || password == "" {
		return nil, invalid
	}
	u, err := b.store.FindByEmail(ctx, email)
	if err != nil {
		return nil, invalid
	}
	if !hasher.ComparePassword(u.PasswordHash, password) {
		return nil, invalid
	}
	return u, nil
}
