package storage

import (
	"context"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SQLStore — CHỈ ĐỌC transactions/categories, tổng hợp bằng SQL (D41). Không mutate.
type SQLStore struct{ db *gorm.DB }

func NewSQLStore(db *gorm.DB) *SQLStore { return &SQLStore{db: db} }

func (s *SQLStore) inRange(householdID uuid.UUID, from, endExcl time.Time) *gorm.DB {
	return s.db.Table("transactions").
		Where("household_id = ? AND transaction_date >= ? AND transaction_date < ?", householdID, from, endExcl)
}

// SumIncomeExpense — Σ thu / Σ chi của hộ trong [from, endExcl) (FR-002).
func (s *SQLStore) SumIncomeExpense(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) (income, expense float64, err error) {
	type row struct {
		Type  string
		Total float64
	}
	var rows []row
	err = s.inRange(householdID, from, endExcl).WithContext(ctx).
		Where("type IN ?", []string{common.TypeIncome, common.TypeExpense}).
		Select("type, COALESCE(SUM(amount), 0) AS total").Group("type").Scan(&rows).Error
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

// CategoryBreakdown — Σ chi (EXPENSE) gộp danh mục con một cấp vào cha (D37), sắp giảm dần.
func (s *SQLStore) CategoryBreakdown(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) ([]model.CategoryBreakdown, error) {
	items := []model.CategoryBreakdown{}
	err := s.db.WithContext(ctx).Table("transactions t").
		Select("COALESCE(parent.id, cat.id) AS category_id, "+
			"COALESCE(parent.name, cat.name) AS category_name, "+
			"COALESCE(parent.is_hidden, cat.is_hidden) AS category_hidden, "+
			"COALESCE(SUM(t.amount), 0) AS amount").
		Joins("JOIN categories cat ON cat.id = t.category_id").
		Joins("LEFT JOIN categories parent ON parent.id = cat.parent_id").
		Where("t.household_id = ? AND t.type = ? AND t.transaction_date >= ? AND t.transaction_date < ?",
			householdID, common.TypeExpense, from, endExcl).
		Group("COALESCE(parent.id, cat.id), COALESCE(parent.name, cat.name), COALESCE(parent.is_hidden, cat.is_hidden)").
		Order("amount DESC").Scan(&items).Error
	return items, err
}

// TrendIncomeExpense — Σ thu/chi theo mốc date_trunc(unit) (D36). Chỉ mốc có dữ liệu.
func (s *SQLStore) TrendIncomeExpense(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time, unit string) ([]model.TrendRow, error) {
	pg := model.PgTruncUnit(unit)
	rows := []model.TrendRow{}
	err := s.inRange(householdID, from, endExcl).WithContext(ctx).
		Select("date_trunc('"+pg+"', transaction_date) AS bucket, "+
			"COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) AS income, "+
			"COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) AS expense",
			common.TypeIncome, common.TypeExpense).
		Group("bucket").Order("bucket").Scan(&rows).Error
	return rows, err
}

// ChildCategoryIDs — id danh mục con một cấp (gộp subtree cho báo cáo danh mục).
func (s *SQLStore) ChildCategoryIDs(ctx context.Context, householdID, parentID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.db.WithContext(ctx).Table("categories").
		Where("household_id = ? AND parent_id = ?", householdID, parentID).
		Pluck("id", &ids).Error
	return ids, err
}

// FindCategoryName — kiểm tra danh mục thuộc hộ + lấy tên. found=false khi ngoài hộ /
// không tồn tại (biz → 404, không lộ tồn tại — D5/001).
func (s *SQLStore) FindCategoryName(ctx context.Context, householdID, id uuid.UUID) (name string, found bool, err error) {
	err = s.db.WithContext(ctx).Table("categories").
		Where("household_id = ? AND id = ?", householdID, id).
		Limit(1).Pluck("name", &name).Error
	if err != nil {
		return "", false, err
	}
	return name, name != "", nil
}

// CategorySum — Σ chi (EXPENSE) của tập danh mục (cha + con) trong khoảng (D38).
func (s *SQLStore) CategorySum(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, from, endExcl time.Time) (float64, error) {
	if len(categoryIDs) == 0 {
		return 0, nil
	}
	var total *float64
	err := s.inRange(householdID, from, endExcl).WithContext(ctx).
		Where("type = ? AND category_id IN ?", common.TypeExpense, categoryIDs).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	if err != nil || total == nil {
		return 0, err
	}
	return *total, nil
}

// CategoryTransactions — mọi giao dịch (EXPENSE) của tập danh mục trong khoảng, mới nhất
// trước, embed tên danh mục · tài khoản · người nhập (dùng lại ListItem 002 — D38).
func (s *SQLStore) CategoryTransactions(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, from, endExcl time.Time) ([]transactionmodel.ListItem, error) {
	items := []transactionmodel.ListItem{}
	if len(categoryIDs) == 0 {
		return items, nil
	}
	err := s.db.WithContext(ctx).Table("transactions").
		Select("transactions.*, categories.name AS category_name, accounts.name AS account_name, users.display_name AS created_by_name").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Joins("JOIN users ON users.id = transactions.created_by").
		Where("transactions.household_id = ? AND transactions.type = ? AND transactions.category_id IN ? AND transactions.transaction_date >= ? AND transactions.transaction_date < ?",
			householdID, common.TypeExpense, categoryIDs, from, endExcl).
		Order("transactions.transaction_date DESC, transactions.id DESC").
		Scan(&items).Error
	return items, err
}

// ─── Feature 008: báo cáo THEO THÀNH VIÊN (chỉ đọc) ─────────────────────────

// ListMembers — thành viên HIỆN TẠI của hộ (household_members JOIN users), kèm email
// để fallback tên hiển thị (D43/D46). Sắp theo tên cho ổn định.
func (s *SQLStore) ListMembers(ctx context.Context, householdID uuid.UUID) ([]model.MemberInfo, error) {
	items := []model.MemberInfo{}
	err := s.db.WithContext(ctx).Table("household_members hm").
		Select("u.id AS id, u.display_name AS display_name, u.email AS email").
		Joins("JOIN users u ON u.id = hm.user_id").
		Where("hm.household_id = ?", householdID).
		Order("u.display_name, u.email").
		Scan(&items).Error
	return items, err
}

// FindMember — thành viên hiện tại của hộ theo id. found=false khi không thuộc hộ /
// không tồn tại (biz → 404, chống dò — D48/FR-008).
func (s *SQLStore) FindMember(ctx context.Context, householdID, userID uuid.UUID) (model.MemberInfo, bool, error) {
	var m model.MemberInfo
	err := s.db.WithContext(ctx).Table("household_members hm").
		Select("u.id AS id, u.display_name AS display_name, u.email AS email").
		Joins("JOIN users u ON u.id = hm.user_id").
		Where("hm.household_id = ? AND hm.user_id = ?", householdID, userID).
		Limit(1).Scan(&m).Error
	if err != nil {
		return model.MemberInfo{}, false, err
	}
	return m, m.ID != uuid.Nil, nil
}

// MemberSums — Σ thu / Σ chi của hộ trong [from, endExcl) GROUP BY created_by (D42).
// Trả CHỈ những `created_by` có giao dịch; biz zero-fill thành viên hiện tại còn lại (D43).
func (s *SQLStore) MemberSums(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) ([]model.MemberAggRow, error) {
	rows := []model.MemberAggRow{}
	err := s.inRange(householdID, from, endExcl).WithContext(ctx).
		Where("type IN ?", []string{common.TypeIncome, common.TypeExpense}).
		Select("created_by, "+
			"COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) AS income, "+
			"COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) AS expense",
			common.TypeIncome, common.TypeExpense).
		Group("created_by").Scan(&rows).Error
	return rows, err
}

// memberTxnQuery — cơ sở liệt kê giao dịch của hộ trong khoảng, embed tên danh mục ·
// tài khoản · người nhập (dùng lại ListItem 002 — D47). Lọc created_by áp ở caller.
func (s *SQLStore) memberTxnQuery(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) *gorm.DB {
	return s.db.WithContext(ctx).Table("transactions").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Joins("JOIN users ON users.id = transactions.created_by").
		Where("transactions.household_id = ? AND transactions.transaction_date >= ? AND transactions.transaction_date < ?",
			householdID, from, endExcl)
}

const memberTxnSelect = "transactions.*, categories.name AS category_name, " +
	"accounts.name AS account_name, users.display_name AS created_by_name"

// MemberTransactions — giao dịch của MỘT thành viên (created_by = userID) trong khoảng,
// mới nhất trước, phân trang (D47/FR-006/FR-007).
func (s *SQLStore) MemberTransactions(ctx context.Context, householdID, userID uuid.UUID, from, endExcl time.Time, paging common.Paging) ([]transactionmodel.ListItem, int64, error) {
	base := s.memberTxnQuery(ctx, householdID, from, endExcl).Where("transactions.created_by = ?", userID)
	return scanTxnPage(base, paging)
}

// FormerMemberTransactions — giao dịch của người ĐÃ RỜI hộ (created_by ∉ thành viên hiện
// tại) trong khoảng, phân trang (D44). currentIDs rỗng → không loại trừ ai.
func (s *SQLStore) FormerMemberTransactions(ctx context.Context, householdID uuid.UUID, currentIDs []uuid.UUID, from, endExcl time.Time, paging common.Paging) ([]transactionmodel.ListItem, int64, error) {
	base := s.memberTxnQuery(ctx, householdID, from, endExcl)
	if len(currentIDs) > 0 {
		base = base.Where("transactions.created_by NOT IN ?", currentIDs)
	}
	return scanTxnPage(base, paging)
}

// scanTxnPage — đếm tổng + lấy trang giao dịch (mới nhất trước) dùng chung member/former.
func scanTxnPage(base *gorm.DB, paging common.Paging) ([]transactionmodel.ListItem, int64, error) {
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []transactionmodel.ListItem{}
	err := base.Select(memberTxnSelect).
		Order("transactions.transaction_date DESC, transactions.id DESC").
		Limit(paging.PageSize).Offset(paging.Offset()).Scan(&items).Error
	return items, total, err
}

// CategoryTrend — Σ chi của tập danh mục theo mốc date_trunc(unit) (D36/D38).
func (s *SQLStore) CategoryTrend(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, from, endExcl time.Time, unit string) ([]model.CategoryTrendRow, error) {
	rows := []model.CategoryTrendRow{}
	if len(categoryIDs) == 0 {
		return rows, nil
	}
	pg := model.PgTruncUnit(unit)
	err := s.inRange(householdID, from, endExcl).WithContext(ctx).
		Where("type = ? AND category_id IN ?", common.TypeExpense, categoryIDs).
		Select("date_trunc('" + pg + "', transaction_date) AS bucket, COALESCE(SUM(amount), 0) AS amount").
		Group("bucket").Order("bucket").Scan(&rows).Error
	return rows, err
}
