package model

import (
	"household-finance/api/common"

	"github.com/google/uuid"
)

// CategorizationRule — quy tắc gợi ý học từ lịch sử chung của hộ (FR-015, D7).
type CategorizationRule struct {
	common.SQLModel
	HouseholdID uuid.UUID `json:"household_id" gorm:"type:uuid;not null"`
	Keyword     string    `json:"keyword" gorm:"not null"` // đã chuẩn hóa lowercase (biz)
	CategoryID  uuid.UUID `json:"category_id" gorm:"type:uuid;not null"`
	MatchCount  int       `json:"match_count" gorm:"not null;default:0"`
}

func (CategorizationRule) TableName() string { return "categorization_rules" }

// RuleSuggestion — rule kèm thông tin danh mục để chấm gợi ý.
type RuleSuggestion struct {
	Keyword      string    `json:"keyword"`
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	IsHidden     bool      `json:"is_hidden"`
	MatchCount   int       `json:"match_count"`
}

// Suggestion — response GET /api/categories/suggest (contracts §Categories).
type Suggestion struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
}
