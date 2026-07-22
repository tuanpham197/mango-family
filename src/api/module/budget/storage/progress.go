package storage

import (
	"context"
	"time"

	"household-finance/api/common"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProgressStore — CHỈ ĐỌC transactions/categories để suy ra tiến độ (D21).
// Không sửa dữ liệu nguồn (bất biến "budget chỉ đọc" — Assumptions spec).
type ProgressStore struct{ db *gorm.DB }

func NewProgressStore(db *gorm.DB) *ProgressStore { return &ProgressStore{db: db} }

// ChildCategoryIDs — id các danh mục con một cấp của một danh mục (cây 001 chỉ một
// cấp — D10/001); dùng để gộp chi danh mục con vào ngân sách cha (FR-006).
func (s *ProgressStore) ChildCategoryIDs(ctx context.Context, householdID, parentID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.db.WithContext(ctx).Table("categories").
		Where("household_id = ? AND parent_id = ?", householdID, parentID).
		Pluck("id", &ids).Error
	return ids, err
}

// SumExpenseByCategories — Σ chi (EXPENSE) trong cửa sổ kỳ, theo tập danh mục (D21).
// Cửa sổ [start, endExcl); chỉ transaction_date trong kỳ; cùng hộ. Danh sách rỗng → 0.
func (s *ProgressStore) SumExpenseByCategories(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, start, endExcl time.Time) (float64, error) {
	if len(categoryIDs) == 0 {
		return 0, nil
	}
	return s.sum(ctx, householdID, start, endExcl, func(q *gorm.DB) *gorm.DB {
		return q.Where("category_id IN ?", categoryIDs)
	})
}

// SumExpenseTotal — Σ mọi chi (EXPENSE) của hộ trong cửa sổ kỳ (ngân sách tổng — D21).
func (s *ProgressStore) SumExpenseTotal(ctx context.Context, householdID uuid.UUID, start, endExcl time.Time) (float64, error) {
	return s.sum(ctx, householdID, start, endExcl, nil)
}

func (s *ProgressStore) sum(ctx context.Context, householdID uuid.UUID, start, endExcl time.Time, scope func(*gorm.DB) *gorm.DB) (float64, error) {
	q := s.db.WithContext(ctx).Table("transactions").
		Where("household_id = ? AND type = ?", householdID, common.TypeExpense).
		Where("transaction_date >= ? AND transaction_date < ?", start, endExcl)
	if scope != nil {
		q = scope(q)
	}
	var total *float64
	if err := q.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}
