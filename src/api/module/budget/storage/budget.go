package storage

import (
	"context"
	"errors"
	"time"

	"household-finance/api/module/budget/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("budget not found")

type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

// selectBudgetView — embed tên danh mục (+ trạng thái ẩn) và tên người tạo (contracts).
const selectBudgetView = "budgets.*, categories.name AS category_name, " +
	"categories.is_hidden AS category_hidden, users.display_name AS created_by_name"

func (s *SQLStore) joinNames(q *gorm.DB) *gorm.DB {
	return q.
		Joins("LEFT JOIN categories ON categories.id = budgets.category_id"). // TOTAL: category_id NULL
		Joins("JOIN users ON users.id = budgets.created_by")
}

func (s *SQLStore) Create(ctx context.Context, b *model.Budget) error {
	return s.db.WithContext(ctx).Create(b).Error
}

// List — ngân sách của hộ (status: "ACTIVE" mặc định · "" / "all" cho cả ENDED).
// spent/percent/alerts để rỗng — biz suy ra (D21).
func (s *SQLStore) List(ctx context.Context, householdID uuid.UUID, status string) ([]model.BudgetView, error) {
	q := s.db.WithContext(ctx).Model(&model.Budget{}).Where("budgets.household_id = ?", householdID)
	if status != "" && status != "all" {
		q = q.Where("budgets.status = ?", status)
	}
	items := []model.BudgetView{}
	err := s.joinNames(q).Select(selectBudgetView).
		Order("budgets.type, budgets.created_at").
		Scan(&items).Error
	return items, err
}

// FindByID — một ngân sách của hộ + tên hiển thị (prefill form sửa; ngoài hộ → 404).
func (s *SQLStore) FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.BudgetView, error) {
	var v model.BudgetView
	err := s.joinNames(
		s.db.WithContext(ctx).Model(&model.Budget{}).
			Where("budgets.household_id = ? AND budgets.id = ?", householdID, id),
	).Select(selectBudgetView).Scan(&v).Error
	if err != nil {
		return nil, err
	}
	if v.ID == uuid.Nil {
		return nil, ErrNotFound
	}
	return &v, nil
}

// ListActive — ngân sách ACTIVE thô (không join) cho evaluator cảnh báo (D24).
func (s *SQLStore) ListActive(ctx context.Context, householdID uuid.UUID) ([]model.Budget, error) {
	var items []model.Budget
	err := s.db.WithContext(ctx).
		Where("household_id = ? AND status = ?", householdID, model.StatusActive).
		Find(&items).Error
	return items, err
}

// FindActiveDuplicate — ngân sách ACTIVE trùng (danh mục, kỳ) hoặc (tổng, kỳ) đang tồn
// tại (FR-010, D22). excludeID để bỏ chính bản ghi khi sửa. Không có → (nil, nil).
func (s *SQLStore) FindActiveDuplicate(ctx context.Context, householdID uuid.UUID, typ string, categoryID *uuid.UUID, periodType string, excludeID *uuid.UUID) (*model.Budget, error) {
	q := s.db.WithContext(ctx).
		Where("household_id = ? AND status = ? AND type = ? AND period_type = ?",
			householdID, model.StatusActive, typ, periodType)
	if typ == model.TypeCategory {
		q = q.Where("category_id = ?", categoryID)
	}
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var b model.Budget
	err := q.First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// EndExpiredOneTime — chuyển ngân sách ONE_TIME đã qua end_date sang ENDED (D20).
// So sánh theo ngày: end_date < đầu ngày hôm nay → cả kỳ đã trôi qua.
func (s *SQLStore) EndExpiredOneTime(ctx context.Context, householdID uuid.UUID, now time.Time) error {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return s.db.WithContext(ctx).Model(&model.Budget{}).
		Where("household_id = ? AND period_type = ? AND status = ? AND end_date < ?",
			householdID, model.PeriodOneTime, model.StatusActive, today).
		Update("status", model.StatusEnded).Error
}

// UpdateConditional — mốc lạc quan (D25): UPDATE ... WHERE updated_at = expected.
// GORM tự làm mới updated_at (autoUpdateTime). 0 hàng = conflict/gone (biz phân biệt).
func (s *SQLStore) UpdateConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time, changes map[string]any) (int64, error) {
	res := s.db.WithContext(ctx).Model(&model.Budget{}).
		Where("household_id = ? AND id = ? AND updated_at = ?", householdID, id, expectedUpdatedAt).
		Updates(changes)
	return res.RowsAffected, res.Error
}

// DeleteConditional — DELETE ... WHERE updated_at = expected (nếu có mốc); D25.
// budget_alerts xóa cascade theo FK (00009).
func (s *SQLStore) DeleteConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt *time.Time) (int64, error) {
	q := s.db.WithContext(ctx).Where("household_id = ? AND id = ?", householdID, id)
	if expectedUpdatedAt != nil {
		q = q.Where("updated_at = ?", *expectedUpdatedAt)
	}
	res := q.Delete(&model.Budget{})
	return res.RowsAffected, res.Error
}

// Exists — bản ghi còn trong hộ? (phân biệt CONCURRENCY_CONFLICT vs RECORD_GONE).
func (s *SQLStore) Exists(ctx context.Context, householdID, id uuid.UUID) (bool, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&model.Budget{}).
		Where("household_id = ? AND id = ?", householdID, id).Count(&n).Error
	return n > 0, err
}
