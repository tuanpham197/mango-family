package storage

import (
	"context"

	"household-finance/api/module/account/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

// ListWithBalance — tài khoản của hộ kèm số dư từ view account_balances (D14),
// một round-trip. Sắp theo created_at để tài khoản mặc định lên đầu.
func (s *SQLStore) ListWithBalance(ctx context.Context, householdID uuid.UUID) ([]model.AccountWithBalance, error) {
	items := []model.AccountWithBalance{}
	err := s.db.WithContext(ctx).
		Table("accounts a").
		Select("a.id, a.name, a.type, coalesce(b.balance, a.initial_balance) AS balance").
		Joins("LEFT JOIN account_balances b ON b.account_id = a.id").
		Where("a.household_id = ?", householdID).
		Order("a.created_at").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID — scope theo hộ; dùng để validate account_id khi tạo/sửa giao dịch.
func (s *SQLStore) FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.Account, error) {
	var a model.Account
	err := s.db.WithContext(ctx).
		Where("household_id = ? AND id = ?", householdID, id).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Count — số tài khoản của hộ (dùng khi seed idempotent).
func (s *SQLStore) Count(ctx context.Context, householdID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&model.Account{}).
		Where("household_id = ?", householdID).Count(&n).Error
	return n, err
}
