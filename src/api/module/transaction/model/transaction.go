package model

import (
	"time"

	"github.com/google/uuid"
)

// Transaction — bút toán Thu/Chi đầy đủ vòng đời (002 mở rộng bảng 001).
// KHÔNG embed common.SQLModel: bảng chưa có created_at (contracts/db-schema.sql).
type Transaction struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HouseholdID     uuid.UUID `json:"household_id" gorm:"type:uuid;not null;index"`
	CreatedBy       uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	Amount          float64   `json:"amount" gorm:"type:numeric(14,2);not null"`
	Type            string    `json:"type" gorm:"not null"` // INCOME | EXPENSE — luôn = type danh mục (biz)
	CategoryID      uuid.UUID `json:"category_id" gorm:"type:uuid;not null"`
	AccountID       uuid.UUID `json:"account_id" gorm:"type:uuid;not null"` // 002: FK accounts (biz kiểm cùng hộ)
	Description     *string   `json:"description"`
	TransactionDate time.Time `json:"transaction_date" gorm:"not null"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"` // mốc lạc quan (D17)
}

func (Transaction) TableName() string { return "transactions" }

// ListItem — dòng sổ giao dịch, embed tên danh mục + tài khoản + người nhập (D16).
type ListItem struct {
	Transaction
	CategoryName  string `json:"category_name"`
	AccountName   string `json:"account_name"`
	CreatedByName string `json:"created_by_name"`
}
