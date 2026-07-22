package storage

import (
	"context"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/overview/model"
	transactionmodel "household-finance/api/module/transaction/model"
	transactionstorage "household-finance/api/module/transaction/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SQLStore — CHỈ ĐỌC transactions/accounts/categories để suy ra số liệu Tổng quan (D27).
// Không mutate dữ liệu nguồn (bất biến "overview chỉ đọc").
type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

func scanFloat(q *gorm.DB) (float64, error) {
	var v *float64
	if err := q.Scan(&v).Error; err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	return *v, nil
}

// NetWorth — Σ số dư mọi tài khoản của hộ (view account_balances — 002/D14).
func (s *SQLStore) NetWorth(ctx context.Context, householdID uuid.UUID) (float64, error) {
	return scanFloat(s.db.WithContext(ctx).Table("account_balances").
		Where("household_id = ?", householdID).
		Select("COALESCE(SUM(balance), 0)"))
}

// NetWorthAsOf — tài sản ròng tại thời điểm `before` (đầu tháng hiện tại): Σ số dư đầu
// kỳ tài khoản + Σ bút toán có dấu (INCOME +, EXPENSE −) có transaction_date < before (D28).
func (s *SQLStore) NetWorthAsOf(ctx context.Context, householdID uuid.UUID, before time.Time) (float64, error) {
	initial, err := scanFloat(s.db.WithContext(ctx).Table("accounts").
		Where("household_id = ?", householdID).
		Select("COALESCE(SUM(initial_balance), 0)"))
	if err != nil {
		return 0, err
	}
	flow, err := scanFloat(s.db.WithContext(ctx).Table("transactions").
		Where("household_id = ? AND transaction_date < ?", householdID, before).
		Select("COALESCE(SUM(CASE type WHEN 'INCOME' THEN amount ELSE -amount END), 0)"))
	if err != nil {
		return 0, err
	}
	return initial + flow, nil
}

// MonthIncomeExpense — Σ Thu / Σ Chi của hộ trong cửa sổ [start, endExcl).
func (s *SQLStore) MonthIncomeExpense(ctx context.Context, householdID uuid.UUID, start, endExcl time.Time) (income, expense float64, err error) {
	type row struct {
		Type  string
		Total float64
	}
	var rows []row
	err = s.db.WithContext(ctx).Table("transactions").
		Where("household_id = ? AND transaction_date >= ? AND transaction_date < ?", householdID, start, endExcl).
		Where("type IN ?", []string{common.TypeIncome, common.TypeExpense}).
		Select("type, COALESCE(SUM(amount), 0) AS total").
		Group("type").Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}
	for _, r := range rows {
		switch r.Type {
		case common.TypeIncome:
			income = r.Total
		case common.TypeExpense:
			expense = r.Total
		}
	}
	return income, expense, nil
}

// CategorySpending — Σ chi (EXPENSE) trong kỳ, gộp danh mục con một cấp vào danh mục
// cha (D29): quy về COALESCE(parent.id, cat.id); sắp giảm dần. Percent tính ở biz.
func (s *SQLStore) CategorySpending(ctx context.Context, householdID uuid.UUID, start, endExcl time.Time) ([]model.CategorySpending, error) {
	items := []model.CategorySpending{}
	err := s.db.WithContext(ctx).Table("transactions t").
		Select("COALESCE(parent.id, cat.id) AS category_id, "+
			"COALESCE(parent.name, cat.name) AS category_name, "+
			"COALESCE(parent.is_hidden, cat.is_hidden) AS category_hidden, "+
			"COALESCE(SUM(t.amount), 0) AS amount").
		Joins("JOIN categories cat ON cat.id = t.category_id").
		Joins("LEFT JOIN categories parent ON parent.id = cat.parent_id").
		Where("t.household_id = ? AND t.type = ? AND t.transaction_date >= ? AND t.transaction_date < ?",
			householdID, common.TypeExpense, start, endExcl).
		Group("COALESCE(parent.id, cat.id), COALESCE(parent.name, cat.name), COALESCE(parent.is_hidden, cat.is_hidden)").
		Order("amount DESC").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// RecentTransactions — dùng lại sổ chung 002 (embed tên danh mục · tài khoản · người nhập),
// mới nhất trước (D30).
func (s *SQLStore) RecentTransactions(ctx context.Context, householdID uuid.UUID, limit int) ([]transactionmodel.ListItem, error) {
	items, _, err := transactionstorage.NewSQLStore(s.db).
		List(ctx, householdID, common.Paging{Page: 1, PageSize: limit}, transactionstorage.ListFilter{})
	return items, err
}
