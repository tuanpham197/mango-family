package storage

import (
	"context"
	"errors"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/transaction/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("transaction not found")

type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

func (s *SQLStore) Create(ctx context.Context, t *model.Transaction) error {
	return s.db.WithContext(ctx).Create(t).Error
}

// selectListItem — cột chung cho sổ: embed category_name + account_name + created_by_name (D16).
const selectListItem = "transactions.*, categories.name AS category_name, " +
	"accounts.name AS account_name, users.display_name AS created_by_name"

func (s *SQLStore) joinNames(q *gorm.DB) *gorm.DB {
	return q.
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Joins("JOIN users ON users.id = transactions.created_by")
}

// List — sổ chung của hộ, mới nhất trước, một round-trip embed đủ tên hiển thị (D16).
func (s *SQLStore) List(ctx context.Context, householdID uuid.UUID, paging common.Paging) ([]model.ListItem, int64, error) {
	base := s.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("transactions.household_id = ?", householdID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := []model.ListItem{}
	err := s.joinNames(base).
		Select(selectListItem).
		Order("transactions.transaction_date DESC, transactions.id DESC").
		Limit(paging.PageSize).Offset(paging.Offset()).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// FindByID — một giao dịch của hộ, embed tên (prefill form sửa; ngoài hộ → ErrNotFound → 404).
func (s *SQLStore) FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.ListItem, error) {
	var item model.ListItem
	err := s.joinNames(
		s.db.WithContext(ctx).Model(&model.Transaction{}).
			Where("transactions.household_id = ? AND transactions.id = ?", householdID, id),
	).Select(selectListItem).Scan(&item).Error
	if err != nil {
		return nil, err
	}
	if item.ID == uuid.Nil {
		return nil, ErrNotFound
	}
	return &item, nil
}

// UpdateConditional — mốc lạc quan (D17): UPDATE ... WHERE updated_at = expected.
// GORM tự làm mới updated_at. Trả về số hàng ảnh hưởng (0 = conflict/gone).
func (s *SQLStore) UpdateConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time, changes map[string]any) (int64, error) {
	res := s.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("household_id = ? AND id = ? AND updated_at = ?", householdID, id, expectedUpdatedAt).
		Updates(changes)
	return res.RowsAffected, res.Error
}

// DeleteConditional — DELETE ... WHERE updated_at = expected (nếu có mốc); D17.
func (s *SQLStore) DeleteConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt *time.Time) (int64, error) {
	q := s.db.WithContext(ctx).Where("household_id = ? AND id = ?", householdID, id)
	if expectedUpdatedAt != nil {
		q = q.Where("updated_at = ?", *expectedUpdatedAt)
	}
	res := q.Delete(&model.Transaction{})
	return res.RowsAffected, res.Error
}

// Exists — bản ghi còn trong hộ? (phân biệt CONCURRENCY_CONFLICT vs RECORD_GONE).
func (s *SQLStore) Exists(ctx context.Context, householdID, id uuid.UUID) (bool, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("household_id = ? AND id = ?", householdID, id).Count(&n).Error
	return n > 0, err
}
