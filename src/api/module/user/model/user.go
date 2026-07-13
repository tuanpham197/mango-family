package model

import "household-finance/api/common"

// User — danh bạ người dùng KIÊM đăng nhập (design §2): nguồn định danh duy nhất.
type User struct {
	common.SQLModel
	Email        string `json:"email" gorm:"uniqueIndex;not null"`
	DisplayName  string `json:"display_name" gorm:"not null"`
	PasswordHash string `json:"-" gorm:"not null"` // bcrypt — không bao giờ trả về API
}

func (User) TableName() string { return "users" }

// PublicUser — payload trả về client.
type PublicUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID.String(), Email: u.Email, DisplayName: u.DisplayName}
}
