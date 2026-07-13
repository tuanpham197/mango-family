//go:build integration

// Integration test view account_balances với Postgres thật (docker + goose up):
//
//	DATABASE_URL=... go test ./... -tags=integration
package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"household-finance/api/common"
	"household-finance/api/module/account/storage"

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

// setup tạo hộ + 1 user + 1 tài khoản "Tiền mặt" + 1 danh mục Chi; dọn sạch sau test.
func setup(t *testing.T, db *gorm.DB) (householdID, userID, accountID, categoryID uuid.UUID) {
	t.Helper()
	userID, householdID, accountID, categoryID = uuid.New(), uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, 'IT', 'x')`,
		userID, fmt.Sprintf("it-%s@test.local", userID)).Error)
	require.NoError(t, db.Exec(`INSERT INTO households (id, name, created_by) VALUES (?, 'IT Hộ', ?)`, householdID, userID).Error)
	require.NoError(t, db.Exec(`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, householdID, userID).Error)
	require.NoError(t, db.Exec(`INSERT INTO accounts (id, household_id, name, type) VALUES (?, ?, 'Tiền mặt', 'CASH')`, accountID, householdID).Error)
	require.NoError(t, db.Exec(`INSERT INTO categories (id, household_id, name, type) VALUES (?, ?, 'Ăn uống', 'EXPENSE')`, categoryID, householdID).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM accounts WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM categories WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM household_members WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM households WHERE id = ?`, householdID)
		db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	})
	return
}

func addTxn(t *testing.T, db *gorm.DB, hid, uid, accID, catID uuid.UUID, amount float64, typ string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO transactions (id, household_id, created_by, amount, type, category_id, account_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, hid, uid, amount, typ, catID, accID).Error)
	return id
}

func balanceOf(t *testing.T, s *storage.SQLStore, hid, accID uuid.UUID) float64 {
	t.Helper()
	items, err := s.ListWithBalance(context.Background(), hid)
	require.NoError(t, err)
	for _, a := range items {
		if a.ID == accID {
			return a.Balance
		}
	}
	t.Fatalf("account %s không có trong danh sách", accID)
	return 0
}

func TestAccountBalancesViewReconciles(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	hid, uid, accID, catID := setup(t, db)

	// Ban đầu 0
	assert.Equal(t, float64(0), balanceOf(t, s, hid, accID))

	// + Chi 50.000 → -50.000
	txn := addTxn(t, db, hid, uid, accID, catID, 50000, common.TypeExpense)
	assert.Equal(t, float64(-50000), balanceOf(t, s, hid, accID))

	// + Thu 200.000 → 150.000
	addTxn(t, db, hid, uid, accID, catID, 200000, common.TypeIncome)
	assert.Equal(t, float64(150000), balanceOf(t, s, hid, accID))

	// Sửa Chi 50.000 → 80.000 → balance 120.000
	require.NoError(t, db.Exec(`UPDATE transactions SET amount = 80000 WHERE id = ?`, txn).Error)
	assert.Equal(t, float64(120000), balanceOf(t, s, hid, accID))

	// Xóa giao dịch Chi → chỉ còn Thu 200.000
	require.NoError(t, db.Exec(`DELETE FROM transactions WHERE id = ?`, txn).Error)
	assert.Equal(t, float64(200000), balanceOf(t, s, hid, accID))

	// Đối chiếu tổng bút toán có dấu (SC-004)
	var expected float64
	require.NoError(t, db.Raw(
		`SELECT coalesce(sum(case type when 'INCOME' then amount else -amount end),0) FROM transactions WHERE account_id = ?`,
		accID).Scan(&expected).Error)
	assert.Equal(t, expected, balanceOf(t, s, hid, accID))
}

func TestFindByIDScopesToHousehold(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	hidA, _, accA, _ := setup(t, db)
	hidB, _, _, _ := setup(t, db)
	ctx := context.Background()

	// Tài khoản hộ A không truy được từ hộ B (D5)
	_, err := s.FindByID(ctx, hidB, accA)
	assert.Error(t, err)

	got, err := s.FindByID(ctx, hidA, accA)
	require.NoError(t, err)
	assert.Equal(t, "Tiền mặt", got.Name)
}
