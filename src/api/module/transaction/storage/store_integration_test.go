//go:build integration

package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"household-finance/api/common"
	categorymodel "household-finance/api/module/category/model"
	categorystorage "household-finance/api/module/category/storage"
	"household-finance/api/module/transaction/model"
	"household-finance/api/module/transaction/storage"

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

func newMember(t *testing.T, db *gorm.DB, householdID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, ?, 'x')`,
		userID, fmt.Sprintf("it-%s@test.local", userID), name).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, householdID, userID).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM household_members WHERE user_id = ?`, userID)
		db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	})
	return userID
}

func newHousehold(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID) {
	t.Helper()
	householdID := uuid.New()
	ownerID := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, 'Owner', 'x')`,
		ownerID, fmt.Sprintf("it-%s@test.local", ownerID)).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO households (id, name, created_by) VALUES (?, 'IT Household', ?)`, householdID, ownerID).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, householdID, ownerID).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM accounts WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM categories WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM household_members WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM households WHERE id = ?`, householdID)
		db.Exec(`DELETE FROM users WHERE id = ?`, ownerID)
	})
	return householdID, ownerID
}

// newAccount tạo một tài khoản CASH cho hộ (002: transactions.account_id NOT NULL).
func newAccount(t *testing.T, db *gorm.DB, householdID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO accounts (id, household_id, name, type) VALUES (?, ?, 'Tiền mặt', 'CASH')`, id, householdID).Error)
	return id
}

func TestTransactionListEmbedsNamesAndScopes(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	catStore := categorystorage.NewSQLStore(db)
	ctx := context.Background()

	hid, owner := newHousehold(t, db)
	alice := newMember(t, db, hid, "Alice IT")
	otherHid, otherOwner := newHousehold(t, db)

	cat := &categorymodel.Category{HouseholdID: hid, Name: "Ăn uống IT", Type: common.TypeExpense}
	require.NoError(t, catStore.Create(ctx, cat))
	otherCat := &categorymodel.Category{HouseholdID: otherHid, Name: "Hộ khác", Type: common.TypeExpense}
	require.NoError(t, catStore.Create(ctx, otherCat))
	acc := newAccount(t, db, hid)
	otherAcc := newAccount(t, db, otherHid)

	require.NoError(t, s.Create(ctx, &model.Transaction{
		HouseholdID: hid, CreatedBy: alice, Amount: 45000, Type: common.TypeExpense, CategoryID: cat.ID, AccountID: acc,
	}))
	require.NoError(t, s.Create(ctx, &model.Transaction{
		HouseholdID: hid, CreatedBy: owner, Amount: 90000, Type: common.TypeExpense, CategoryID: cat.ID, AccountID: acc,
	}))
	require.NoError(t, s.Create(ctx, &model.Transaction{
		HouseholdID: otherHid, CreatedBy: otherOwner, Amount: 11111, Type: common.TypeExpense, CategoryID: otherCat.ID, AccountID: otherAcc,
	}))

	items, total, err := s.List(ctx, hid, common.Paging{Page: 1, PageSize: 50})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total) // cô lập hộ (FR-018)
	require.Len(t, items, 2)
	names := []string{items[0].CreatedByName, items[1].CreatedByName}
	assert.Contains(t, names, "Alice IT") // sổ chung + authorship (FR-022)
	assert.Equal(t, "Ăn uống IT", items[0].CategoryName)

	// paging
	page1, total, err := s.List(ctx, hid, common.Paging{Page: 1, PageSize: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, page1, 1)
}
