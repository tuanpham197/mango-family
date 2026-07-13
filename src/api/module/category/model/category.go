package model

import (
	"time"

	"household-finance/api/common"

	"github.com/google/uuid"
)

// Category — danh mục Thu/Chi dùng chung trong hộ (data-model 001).
type Category struct {
	common.SQLModel
	HouseholdID uuid.UUID  `json:"household_id" gorm:"type:uuid;not null;index"`
	Name        string     `json:"name" gorm:"not null"`
	Type        string     `json:"type" gorm:"not null"` // INCOME | EXPENSE — bất biến sau tạo (FR-005, biz enforce)
	Icon        *string    `json:"icon"`
	ParentID    *uuid.UUID `json:"parent_id" gorm:"type:uuid"`
	IsDefault   bool       `json:"is_default" gorm:"not null;default:false"`
	IsHidden    bool       `json:"is_hidden" gorm:"not null;default:false"`
	CreatedBy   *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"` // mốc lạc quan (D6)
}

func (Category) TableName() string { return "categories" }

// Tree — danh mục gốc kèm con (response GET /api/categories).
type Tree struct {
	Category
	Children []Category `json:"children"`
}

// BuildTree dựng cây cha → con một cấp, giữ thứ tự đầu vào.
// Con có cha không nằm trong danh sách (vd cha bị ẩn khi lọc) được nâng thành gốc
// để không biến mất khỏi response.
func BuildTree(cats []Category) []Tree {
	present := make(map[uuid.UUID]bool, len(cats))
	for _, c := range cats {
		present[c.ID] = true
	}
	children := map[uuid.UUID][]Category{}
	for _, c := range cats {
		if c.ParentID != nil && present[*c.ParentID] {
			children[*c.ParentID] = append(children[*c.ParentID], c)
		}
	}
	roots := []Tree{}
	for _, c := range cats {
		if c.ParentID == nil || !present[*c.ParentID] {
			kids := children[c.ID]
			if kids == nil {
				kids = []Category{}
			}
			roots = append(roots, Tree{Category: c, Children: kids})
		}
	}
	return roots
}
