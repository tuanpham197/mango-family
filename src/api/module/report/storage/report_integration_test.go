//go:build integration

package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"
	"household-finance/api/module/report/storage"

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
		owner, fmt.Sprintf("rpt-%s@test.local", owner)).Error)
	require.NoError(t, db.Exec(`INSERT INTO households (id, name, created_by) VALUES (?, 'RPT Household', ?)`, hid, owner).Error)
	require.NoError(t, db.Exec(`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, hid, owner).Error)
	t.Cleanup(func() {
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

func insertTxn(t *testing.T, db *gorm.DB, hid, owner, catID, accID uuid.UUID, amount float64, typ string, when time.Time) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO transactions (id, household_id, created_by, amount, type, category_id, account_id, transaction_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New(), hid, owner, amount, typ, catID, accID, when).Error)
}

func TestReportStorage_Aggregations(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	s := storage.NewSQLStore(db)
	hid, owner := newHousehold(t, db)
	other, otherOwner := newHousehold(t, db)

	parent := newCategory(t, db, hid, "Ăn uống", common.TypeExpense, nil)
	child := newCategory(t, db, hid, "Cà phê", common.TypeExpense, &parent)
	move := newCategory(t, db, hid, "Di chuyển", common.TypeExpense, nil)
	salary := newCategory(t, db, hid, "Lương", common.TypeIncome, nil)
	acc := newAccount(t, db, hid)
	otherCat := newCategory(t, db, other, "Khác hộ", common.TypeExpense, nil)
	otherAcc := newAccount(t, db, other)

	utc := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 12, 0, 0, 0, time.UTC) }
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	endExcl := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	insertTxn(t, db, hid, owner, salary, acc, 18000000, common.TypeIncome, utc(7, 2))
	insertTxn(t, db, hid, owner, parent, acc, 200000, common.TypeExpense, utc(7, 5))
	insertTxn(t, db, hid, owner, child, acc, 50000, common.TypeExpense, utc(7, 5)) // gộp vào cha
	insertTxn(t, db, hid, owner, move, acc, 90000, common.TypeExpense, utc(7, 10))
	insertTxn(t, db, hid, owner, parent, acc, 777000, common.TypeExpense, utc(6, 20))              // ngoài khoảng
	insertTxn(t, db, other, otherOwner, otherCat, otherAcc, 999000, common.TypeExpense, utc(7, 5)) // hộ khác

	// SumIncomeExpense
	income, expense, err := s.SumIncomeExpense(ctx, hid, from, endExcl)
	require.NoError(t, err)
	assert.Equal(t, float64(18000000), income)
	assert.Equal(t, float64(340000), expense) // 200k+50k+90k (không tính tháng 6, không tính hộ khác)

	// CategoryBreakdown — gộp con vào cha, sắp giảm dần
	cats, err := s.CategoryBreakdown(ctx, hid, from, endExcl)
	require.NoError(t, err)
	require.Len(t, cats, 2)
	assert.Equal(t, parent, cats[0].CategoryID)
	assert.Equal(t, float64(250000), cats[0].Amount) // 200k + 50k con
	assert.Equal(t, float64(90000), cats[1].Amount)

	// Trend theo ngày — chỉ mốc có dữ liệu (biz điền phần còn lại)
	rows, err := s.TrendIncomeExpense(ctx, hid, from, endExcl, model.GroupDay)
	require.NoError(t, err)
	byKey := map[string]model.TrendRow{}
	for _, r := range rows {
		byKey[model.BucketKey(model.GroupDay, r.Bucket)] = r
	}
	assert.Equal(t, float64(18000000), byKey["2026-07-02"].Income)
	assert.Equal(t, float64(250000), byKey["2026-07-05"].Expense) // 200k+50k cùng ngày
	assert.Equal(t, float64(90000), byKey["2026-07-10"].Expense)

	// Chi tiết danh mục cha (gộp con)
	children, err := s.ChildCategoryIDs(ctx, hid, parent)
	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{child}, children)
	ids := append([]uuid.UUID{parent}, children...)
	total, err := s.CategorySum(ctx, hid, ids, from, endExcl)
	require.NoError(t, err)
	assert.Equal(t, float64(250000), total)
	txns, err := s.CategoryTransactions(ctx, hid, ids, from, endExcl)
	require.NoError(t, err)
	assert.Len(t, txns, 2) // 200k cha + 50k con

	// FindCategoryName: trong hộ found; ngoài hộ not found
	name, found, err := s.FindCategoryName(ctx, hid, parent)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Ăn uống", name)
	_, found, err = s.FindCategoryName(ctx, hid, otherCat) // danh mục hộ khác
	require.NoError(t, err)
	assert.False(t, found)

	// Khoảng trống → 0
	emptyFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	emptyTo := time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC)
	inc, exp, err := s.SumIncomeExpense(ctx, hid, emptyFrom, emptyTo)
	require.NoError(t, err)
	assert.Zero(t, inc)
	assert.Zero(t, exp)
}
