package biz

import (
	"context"

	"household-finance/api/common"
	accountmodel "household-finance/api/module/account/model"
	categorymodel "household-finance/api/module/category/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type defaultCategory struct {
	Name string
	Icon string
}

// Bộ danh mục mặc định theo hộ (research D9 + wireframe specs/design/design.html;
// danh sách cuối cùng chờ nghiệp vụ chốt — Open Question BR-001).
var defaultExpense = []defaultCategory{
	{"Ăn uống", "🍜"}, {"Di chuyển", "🚕"}, {"Hóa đơn", "⚡"}, {"Mua sắm", "🛍️"},
	{"Giải trí", "🎬"}, {"Sức khỏe", "🏥"}, {"Học tập", "🎓"}, {"Nhà cửa", "🏠"}, {"Khác", "📦"},
}
var defaultIncome = []defaultCategory{
	{"Lương", "💰"}, {"Thưởng", "🎁"}, {"Khác", "📦"},
}

// SeedDefaultCategories tạo bộ danh mục mặc định cho một hộ nếu hộ CHƯA có danh mục
// nào (idempotent). Là logic app gắn với việc tạo hộ — KHÔNG phải migration (D9).
func SeedDefaultCategories(ctx context.Context, db *gorm.DB, householdID, createdBy uuid.UUID) error {
	var count int64
	if err := db.WithContext(ctx).Model(&categorymodel.Category{}).
		Where("household_id = ?", householdID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	var cats []categorymodel.Category
	add := func(list []defaultCategory, typ string) {
		for _, d := range list {
			icon := d.Icon
			cats = append(cats, categorymodel.Category{
				HouseholdID: householdID, Name: d.Name, Type: typ,
				Icon: &icon, IsDefault: true, CreatedBy: &createdBy,
			})
		}
	}
	add(defaultExpense, common.TypeExpense)
	add(defaultIncome, common.TypeIncome)
	return db.WithContext(ctx).Create(&cats).Error
}

// SeedDefaultAccount tạo tài khoản mặc định "Tiền mặt" (CASH) cho hộ nếu hộ CHƯA
// có tài khoản nào (idempotent). Hộ mới luôn có ≥ 1 tài khoản (FR-006, D13).
func SeedDefaultAccount(ctx context.Context, db *gorm.DB, householdID, createdBy uuid.UUID) error {
	var count int64
	if err := db.WithContext(ctx).Model(&accountmodel.Account{}).
		Where("household_id = ?", householdID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&accountmodel.Account{
		HouseholdID: householdID,
		Name:        "Tiền mặt",
		Type:        common.AccountTypeCash,
		CreatedBy:   &createdBy,
	}).Error
}

// SeedDefaults — một chỗ duy nhất định nghĩa "hộ mới có gì": danh mục + tài khoản
// mặc định (D9 + D13). Gọi khi tạo hộ (và trong cmd/seed cho hộ dev).
func SeedDefaults(ctx context.Context, db *gorm.DB, householdID, createdBy uuid.UUID) error {
	if err := SeedDefaultCategories(ctx, db, householdID, createdBy); err != nil {
		return err
	}
	return SeedDefaultAccount(ctx, db, householdID, createdBy)
}
