package model

import (
	"time"

	"github.com/google/uuid"
)

// Transaction — luồng tối thiểu 001 để kiểm chứng phân loại (FR-013/014/019/022).
// Feature 002 sẽ mở rộng: account_id, updated_at, chặn ngày tương lai, sửa/xóa.
// KHÔNG embed common.SQLModel: bảng 001 chưa có created_at (contracts/db-schema.sql).
type Transaction struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HouseholdID     uuid.UUID `json:"household_id" gorm:"type:uuid;not null;index"`
	CreatedBy       uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	Amount          float64   `json:"amount" gorm:"type:numeric(14,2);not null"`
	Type            string    `json:"type" gorm:"not null"` // INCOME | EXPENSE — luôn = type danh mục (biz)
	CategoryID      uuid.UUID `json:"category_id" gorm:"type:uuid;not null"`
	Description     *string   `json:"description"`
	TransactionDate time.Time `json:"transaction_date" gorm:"not null"`
}

func (Transaction) TableName() string { return "transactions" }

// ListItem — dòng sổ giao dịch, embed tên danh mục + người nhập (FR-022).
type ListItem struct {
	Transaction
	CategoryName  string `json:"category_name"`
	CreatedByName string `json:"created_by_name"`
}
