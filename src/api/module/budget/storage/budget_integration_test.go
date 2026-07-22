//go:build integration

package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"household-finance/api/common"
	budgetbiz "household-finance/api/module/budget/biz"
	"household-finance/api/module/budget/model"
	"household-finance/api/module/budget/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:app_secret@localhost:5432/finance_dev?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	return db
}

func newHousehold(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID) {
	t.Helper()
	hid, owner := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, 'Owner', 'x')`,
		owner, fmt.Sprintf("bgt-%s@test.local", owner)).Error)
	require.NoError(t, db.Exec(`INSERT INTO households (id, name, created_by) VALUES (?, 'BGT Household', ?)`, hid, owner).Error)
	require.NoError(t, db.Exec(`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, hid, owner).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM budget_alerts WHERE budget_id IN (SELECT id FROM budgets WHERE household_id = ?)`, hid)
		db.Exec(`DELETE FROM budgets WHERE household_id = ?`, hid)
		db.Exec(`DELETE FROM transactions WHERE household_id = ?`, hid)
		db.Exec(`DELETE FROM accounts WHERE household_id = ?`, hid)
		db.Exec(`DELETE FROM categories WHERE household_id = ?`, hid)
		db.Exec(`DELETE FROM household_members WHERE household_id = ?`, hid)
		db.Exec(`DELETE FROM households WHERE id = ?`, hid)
		db.Exec(`DELETE FROM users WHERE id = ?`, owner)
	})
	return hid, owner
}

func newCategory(t *testing.T, db *gorm.DB, hid uuid.UUID, name, typ string, parent *uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO categories (id, household_id, name, type, parent_id) VALUES (?, ?, ?, ?, ?)`,
		id, hid, name, typ, parent).Error)
	return id
}

func newAccount(t *testing.T, db *gorm.DB, hid uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO accounts (id, household_id, name, type) VALUES (?, ?, 'Tiền mặt', 'CASH')`, id, hid).Error)
	return id
}

func insertTxn(t *testing.T, db *gorm.DB, hid, owner, catID, accID uuid.UUID, amount float64, typ string, when time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO transactions (id, household_id, created_by, amount, type, category_id, account_id, transaction_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, hid, owner, amount, typ, catID, accID, when).Error)
	return id
}

// SC-003: tiến độ suy ra khớp 100% tổng EXPENSE liên quan sau add/edit/delete/đổi-danh-mục.
func TestProgress_MatchesExpenseSumAcrossMutations(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	ps := storage.NewProgressStore(db)
	hid, owner := newHousehold(t, db)

	parent := newCategory(t, db, hid, "Ăn uống", common.TypeExpense, nil)
	child := newCategory(t, db, hid, "Cà phê", common.TypeExpense, &parent)
	other := newCategory(t, db, hid, "Đi lại", common.TypeExpense, nil)
	acc := newAccount(t, db, hid)
	now := time.Now()
	p := model.ResolvePeriod(model.PeriodMonthly, nil, nil, now)

	// Chi cha 100k + con 50k (gộp vào cha) + khác danh mục 30k + một khoản THU (không tính).
	insertTxn(t, db, hid, owner, parent, acc, 100000, common.TypeExpense, now)
	tChild := insertTxn(t, db, hid, owner, child, acc, 50000, common.TypeExpense, now)
	insertTxn(t, db, hid, owner, other, acc, 30000, common.TypeExpense, now)
	insertTxn(t, db, hid, owner, parent, acc, 999999, common.TypeIncome, now) // THU không vào ngân sách

	childIDs, err := ps.ChildCategoryIDs(ctx, hid, parent)
	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{child}, childIDs)

	catBudgetIDs := append([]uuid.UUID{parent}, childIDs...)
	spent, err := ps.SumExpenseByCategories(ctx, hid, catBudgetIDs, p.Start, p.EndExcl)
	require.NoError(t, err)
	assert.Equal(t, float64(150000), spent) // cha 100k + con 50k

	total, err := ps.SumExpenseTotal(ctx, hid, p.Start, p.EndExcl)
	require.NoError(t, err)
	assert.Equal(t, float64(180000), total) // mọi EXPENSE = 150k + 30k

	// Đổi danh mục con → khác danh mục: tiến độ ngân sách cha giảm 50k.
	require.NoError(t, db.Exec(`UPDATE transactions SET category_id = ? WHERE id = ?`, other, tChild).Error)
	spent, err = ps.SumExpenseByCategories(ctx, hid, catBudgetIDs, p.Start, p.EndExcl)
	require.NoError(t, err)
	assert.Equal(t, float64(100000), spent)

	// Ngoài cửa sổ kỳ (tháng trước) → không tính.
	insertTxn(t, db, hid, owner, parent, acc, 777000, common.TypeExpense, now.AddDate(0, -1, 0))
	spent, err = ps.SumExpenseByCategories(ctx, hid, catBudgetIDs, p.Start, p.EndExcl)
	require.NoError(t, err)
	assert.Equal(t, float64(100000), spent) // không đổi
}

func TestBudget_PartialUniqueBlocksActiveDuplicate(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	s := storage.NewSQLStore(db)
	hid, owner := newHousehold(t, db)
	cat := newCategory(t, db, hid, "Ăn uống", common.TypeExpense, nil)

	b1 := &model.Budget{HouseholdID: hid, Type: model.TypeCategory, CategoryID: &cat, LimitAmount: 1000, PeriodType: model.PeriodMonthly, Status: model.StatusActive, CreatedBy: owner}
	require.NoError(t, s.Create(ctx, b1))
	b2 := &model.Budget{HouseholdID: hid, Type: model.TypeCategory, CategoryID: &cat, LimitAmount: 2000, PeriodType: model.PeriodMonthly, Status: model.StatusActive, CreatedBy: owner}
	assert.Error(t, s.Create(ctx, b2), "partial unique index chặn ngân sách ACTIVE trùng (D22)")

	// Kết thúc b1 → tạo mới được (partial chỉ chặn ACTIVE).
	require.NoError(t, db.Exec(`UPDATE budgets SET status='ENDED' WHERE id = ?`, b1.ID).Error)
	b3 := &model.Budget{HouseholdID: hid, Type: model.TypeCategory, CategoryID: &cat, LimitAmount: 3000, PeriodType: model.PeriodMonthly, Status: model.StatusActive, CreatedBy: owner}
	assert.NoError(t, s.Create(ctx, b3))
}

// storage + evaluator: recompute khớp tổng + idempotent + re-arm (D23/D24, SC-004).
func TestEvaluator_RecomputeIdempotentAndRearm(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	hid, owner := newHousehold(t, db)
	cat := newCategory(t, db, hid, "Ăn uống", common.TypeExpense, nil)
	acc := newAccount(t, db, hid)
	now := time.Now()

	bstore := storage.NewSQLStore(db)
	astore := storage.NewAlertStore(db)
	ev := budgetbiz.NewAlertEvaluator(bstore, storage.NewProgressStore(db), astore)

	bud := &model.Budget{HouseholdID: hid, Type: model.TypeCategory, CategoryID: &cat, LimitAmount: 1000000, PeriodType: model.PeriodMonthly, Status: model.StatusActive, CreatedBy: owner}
	require.NoError(t, bstore.Create(ctx, bud))
	pk := model.PeriodKey(model.PeriodMonthly, now)

	// Chi 850k = 85% → phát THRESHOLD_80.
	t1 := insertTxn(t, db, hid, owner, cat, acc, 850000, common.TypeExpense, now)
	require.NoError(t, ev.EvaluateHousehold(ctx, hid, now))
	rows, _ := astore.GetActive(ctx, bud.ID, pk)
	require.Len(t, rows, 1)
	assert.Equal(t, model.LevelThreshold80, rows[0].Level)

	// Idempotent: chạy lại → vẫn 1 hàng.
	require.NoError(t, ev.EvaluateHousehold(ctx, hid, now))
	rows, _ = astore.GetActive(ctx, bud.ID, pk)
	assert.Len(t, rows, 1)

	// Thêm 300k = 115% → thêm OVER_100 với over_amount = 150k.
	insertTxn(t, db, hid, owner, cat, acc, 300000, common.TypeExpense, now)
	require.NoError(t, ev.EvaluateHousehold(ctx, hid, now))
	rows, _ = astore.GetActive(ctx, bud.ID, pk)
	require.Len(t, rows, 2)
	var over *model.BudgetAlert
	for i := range rows {
		if rows[i].Level == model.LevelOver100 {
			over = &rows[i]
		}
	}
	require.NotNil(t, over)
	require.NotNil(t, over.OverAmount)
	assert.Equal(t, float64(150000), *over.OverAmount) // spent 1.15M − limit 1M

	// Xóa 300k + 850k → 0% → gỡ cả hai (re-arm — bỏ 1 message vẫn tự lành).
	require.NoError(t, db.Exec(`DELETE FROM transactions WHERE id = ?`, t1).Error)
	require.NoError(t, db.Exec(`DELETE FROM transactions WHERE household_id = ? AND category_id = ?`, hid, cat).Error)
	require.NoError(t, ev.EvaluateHousehold(ctx, hid, now))
	rows, _ = astore.GetActive(ctx, bud.ID, pk)
	assert.Empty(t, rows)
}
