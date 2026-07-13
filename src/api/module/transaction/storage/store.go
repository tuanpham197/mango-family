package storage

import (
	"context"

	"household-finance/api/common"
	"household-finance/api/module/transaction/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

func (s *SQLStore) Create(ctx context.Context, t *model.Transaction) error {
	return s.db.WithContext(ctx).Create(t).Error
}

// List — sổ chung của hộ, mới nhất trước, embed category_name + created_by_name (FR-022).
func (s *SQLStore) List(ctx context.Context, householdID uuid.UUID, paging common.Paging) ([]model.ListItem, int64, error) {
	base := s.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("transactions.household_id = ?", householdID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := []model.ListItem{}
	err := base.
		Select("transactions.*, categories.name AS category_name, users.display_name AS created_by_name").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Joins("JOIN users ON users.id = transactions.created_by").
		Order("transactions.transaction_date DESC, transactions.id").
		Limit(paging.PageSize).Offset(paging.Offset()).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
