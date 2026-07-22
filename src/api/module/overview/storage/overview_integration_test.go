//go:build integration

package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"household-finance/api/common"
	budgetmodel "household-finance/api/module/budget/model"
	"household-finance/api/module/overview/storage"

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
		owner, fmt.Sprintf("ov-%s@test.local", owner)).Error)
	require.NoError(t, db.Exec(`INSERT INTO households (id, name, created_by) VALUES (?, 'OV Household', ?)`, hid, owner).Error)
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

func newAccount(t *testing.T, db *gorm.DB, hid uuid.UUID, initial float64) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO accounts (id, household_id, name, type, initial_balance) VALUES (?, ?, 'Tiền mặt', 'CASH', ?)`,
		id, hid, initial).Error)
	return id
}

func insertTxn(t *testing.T, db *gorm.DB, hid, owner, catID, accID uuid.UUID, amount float64, typ string, when time.Time) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO transactions (id, household_id, created_by, amount, type, category_id, account_id, transaction_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New(), hid, owner, amount, typ, catID, accID, when).Error)
}

func TestOverviewStorage_DerivedSummaries(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	s := storage.NewSQLStore(db)
	hid, owner := newHousehold(t, db)
	other, otherOwner := newHousehold(t, db) // hộ khác — cô lập

	parent := newCategory(t, db, hid, "Ăn uống", common.TypeExpense, nil)
	child := newCategory(t, db, hid, "Cà phê", common.TypeExpense, &parent)
	move := newCategory(t, db, hid, "Di chuyển", common.TypeExpense, nil)
	salary := newCategory(t, db, hid, "Lương", common.TypeIncome, nil)
	acc := newAccount(t, db, hid, 1000000) // số dư đầu kỳ 1.000.000
	otherAcc := newAccount(t, db, other, 500000)
	otherCat := newCategory(t, db, other, "Khác hộ", common.TypeExpense, nil)

	now := time.Now()
	p := budgetmodel.ResolvePeriod(budgetmodel.PeriodMonthly, nil, nil, now)
	lastMonth := p.Start.AddDate(0, 0, -1) // trong tháng trước

	// Tháng trước: thu 100k (ảnh hưởng net worth as-of, KHÔNG vào tháng này)
	insertTxn(t, db, hid, owner, salary, acc, 100000, common.TypeIncome, lastMonth)
	// Tháng này: thu 18.000.000, chi cha 200k + con 50k (gộp vào cha), chi Di chuyển 90k
	insertTxn(t, db, hid, owner, salary, acc, 18000000, common.TypeIncome, now)
	insertTxn(t, db, hid, owner, parent, acc, 200000, common.TypeExpense, now)
	insertTxn(t, db, hid, owner, child, acc, 50000, common.TypeExpense, now)
	insertTxn(t, db, hid, owner, move, acc, 90000, common.TypeExpense, now)
	// Hộ khác: không được tính
	insertTxn(t, db, other, otherOwner, otherCat, otherAcc, 777000, common.TypeExpense, now)

	// NetWorth = 1.000.000 + (100k + 18.000.000 − 200k − 50k − 90k) = 18.760.000
	nw, err := s.NetWorth(ctx, hid)
	require.NoError(t, err)
	assert.Equal(t, float64(18760000), nw)

	// NetWorthAsOf(đầu tháng) = 1.000.000 + 100.000 (chỉ giao dịch tháng trước) = 1.100.000
	asOf, err := s.NetWorthAsOf(ctx, hid, p.Start)
	require.NoError(t, err)
	assert.Equal(t, float64(1100000), asOf)

	// Thu/Chi tháng này
	income, expense, err := s.MonthIncomeExpense(ctx, hid, p.Start, p.EndExcl)
	require.NoError(t, err)
	assert.Equal(t, float64(18000000), income)
	assert.Equal(t, float64(340000), expense) // 200k + 50k + 90k

	// Chi theo danh mục: Ăn uống 250k (gộp con) > Di chuyển 90k
	cats, err := s.CategorySpending(ctx, hid, p.Start, p.EndExcl)
	require.NoError(t, err)
	require.Len(t, cats, 2)
	assert.Equal(t, parent, cats[0].CategoryID) // sắp giảm dần
	assert.Equal(t, float64(250000), cats[0].Amount)
	assert.Equal(t, float64(90000), cats[1].Amount)

	// Recent transactions cô lập hộ: 5 mới nhất của hộ, không lẫn hộ khác
	recent, err := s.RecentTransactions(ctx, hid, 5)
	require.NoError(t, err)
	require.NotEmpty(t, recent)
	for _, r := range recent {
		assert.Equal(t, hid, r.HouseholdID)
	}
}
