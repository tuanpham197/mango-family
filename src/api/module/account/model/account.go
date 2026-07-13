package model

import (
	"time"

	"github.com/google/uuid"
)

// Account — nguồn tiền của hộ (D13). 002 chỉ ĐỌC; CRUD đầy đủ thuộc BR-005.
// KHÔNG embed common.SQLModel: cần map created_at riêng và không dùng autoCreate.
type Account struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HouseholdID    uuid.UUID  `json:"household_id" gorm:"type:uuid;not null;index"`
	Name           string     `json:"name" gorm:"not null"`
	Type           string     `json:"type" gorm:"not null;default:CASH"`
	InitialBalance float64    `json:"initial_balance" gorm:"type:numeric(14,2);not null;default:0"`
	CreatedBy      *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (Account) TableName() string { return "accounts" }

// AccountBalance — model read-only map view `account_balances` (D14).
type AccountBalance struct {
	AccountID   uuid.UUID `json:"account_id"`
	HouseholdID uuid.UUID `json:"household_id"`
	Balance     float64   `json:"balance"`
}

func (AccountBalance) TableName() string { return "account_balances" }

// AccountWithBalance — response GET /api/accounts (contracts §Accounts).
type AccountWithBalance struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Balance float64   `json:"balance"`
}
