package storage

import (
	"context"

	"household-finance/api/module/budget/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertStore — CRUD budget_alerts (máy trạng thái cảnh báo — D23). Khóa nghiệp vụ
// là (budget_id, period_key, level); unique index chống trùng ở tầng DB.
type AlertStore struct{ db *gorm.DB }

func NewAlertStore(db *gorm.DB) *AlertStore { return &AlertStore{db: db} }

// GetActive — các cảnh báo đang có của một ngân sách trong kỳ hiện tại (D23).
func (s *AlertStore) GetActive(ctx context.Context, budgetID uuid.UUID, periodKey string) ([]model.BudgetAlert, error) {
	var items []model.BudgetAlert
	err := s.db.WithContext(ctx).
		Where("budget_id = ? AND period_key = ?", budgetID, periodKey).
		Find(&items).Error
	return items, err
}

// ListByBudgetIDs — mọi cảnh báo của một tập ngân sách (biz lọc theo period_key hiện
// tại khi embed vào GET /api/budgets). Danh sách rỗng → không truy vấn.
func (s *AlertStore) ListByBudgetIDs(ctx context.Context, budgetIDs []uuid.UUID) ([]model.BudgetAlert, error) {
	if len(budgetIDs) == 0 {
		return nil, nil
	}
	var items []model.BudgetAlert
	err := s.db.WithContext(ctx).
		Where("budget_id IN ?", budgetIDs).
		Find(&items).Error
	return items, err
}

func (s *AlertStore) Insert(ctx context.Context, a *model.BudgetAlert) error {
	return s.db.WithContext(ctx).Create(a).Error
}

// Delete — gỡ một mức cảnh báo của kỳ (re-arm khi tiến độ tụt dưới mức — D23).
func (s *AlertStore) Delete(ctx context.Context, budgetID uuid.UUID, periodKey, level string) error {
	return s.db.WithContext(ctx).
		Where("budget_id = ? AND period_key = ? AND level = ?", budgetID, periodKey, level).
		Delete(&model.BudgetAlert{}).Error
}
