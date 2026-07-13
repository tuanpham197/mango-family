package model

import (
	"time"

	"household-finance/api/common"

	"github.com/google/uuid"
)

type Household struct {
	common.SQLModel
	Name      string    `json:"name" gorm:"not null"`
	CreatedBy uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
}

func (Household) TableName() string { return "households" }

// Member — liên kết user ↔ hộ; KHÔNG có cột vai trò (ngang quyền — FR-021/001).
type Member struct {
	HouseholdID uuid.UUID `json:"household_id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	JoinedAt    time.Time `json:"joined_at" gorm:"autoCreateTime"`
}

func (Member) TableName() string { return "household_members" }
