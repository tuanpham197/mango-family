//go:build integration

// Integration test storage với Postgres thật (docker + goose up trước khi chạy):
//
//	DATABASE_URL=... go test ./... -tags=integration
package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"household-finance/api/common"
	categorybiz "household-finance/api/module/category/biz"
	"household-finance/api/module/category/model"
	"household-finance/api/module/category/storage"

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

// newHousehold tạo hộ + user cô lập cho test; dọn sạch khi test kết thúc.
func newHousehold(t *testing.T, db *gorm.DB) (householdID, userID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	userID = uuid.New()
	householdID = uuid.New()
	email := fmt.Sprintf("it-%s@test.local", userID)
	require.NoError(t, db.WithContext(ctx).Exec(
		`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, 'IT User', 'x')`,
		userID, email).Error)
	require.NoError(t, db.WithContext(ctx).Exec(
		`INSERT INTO households (id, name, created_by) VALUES (?, 'IT Household', ?)`,
		householdID, userID).Error)
	require.NoError(t, db.WithContext(ctx).Exec(
		`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`,
		householdID, userID).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM categorization_rules WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM transactions WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM categories WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM household_members WHERE household_id = ?`, householdID)
		db.Exec(`DELETE FROM households WHERE id = ?`, householdID)
		db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	})
	return householdID, userID
}

func mkCategory(t *testing.T, s *storage.SQLStore, hid uuid.UUID, name, typ string, parentID *uuid.UUID, hidden bool) *model.Category {
	t.Helper()
	c := &model.Category{HouseholdID: hid, Name: name, Type: typ, ParentID: parentID, IsHidden: hidden}
	require.NoError(t, s.Create(context.Background(), c))
	return c
}

func mkTransaction(t *testing.T, db *gorm.DB, hid, uid, catID uuid.UUID, typ string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO transactions (id, household_id, created_by, amount, type, category_id) VALUES (?, ?, ?, 1000, ?, ?)`,
		id, hid, uid, typ, catID).Error)
	return id
}

func TestHouseholdScoping(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	hidA, _ := newHousehold(t, db)
	hidB, _ := newHousehold(t, db)
	ctx := context.Background()

	cat := mkCategory(t, s, hidA, "Chỉ hộ A", common.TypeExpense, nil, false)

	// Hộ B không thấy bản ghi hộ A (FR-018, D5)
	_, err := s.FindByID(ctx, hidB, cat.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)

	listB, err := s.List(ctx, hidB, storage.ListFilter{})
	require.NoError(t, err)
	assert.Empty(t, listB)

	found, err := s.FindByID(ctx, hidA, cat.ID)
	require.NoError(t, err)
	assert.Equal(t, "Chỉ hộ A", found.Name)
}

func TestListFiltersAndTree(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	hid, _ := newHousehold(t, db)
	ctx := context.Background()

	root := mkCategory(t, s, hid, "Ăn uống", common.TypeExpense, nil, false)
	child := mkCategory(t, s, hid, "Ăn ngoài", common.TypeExpense, &root.ID, false)
	mkCategory(t, s, hid, "Đã ẩn", common.TypeExpense, nil, true)
	mkCategory(t, s, hid, "Lương", common.TypeIncome, nil, false)

	visible, err := s.List(ctx, hid, storage.ListFilter{Type: common.TypeExpense})
	require.NoError(t, err)
	require.Len(t, visible, 2) // ẩn bị loại, INCOME bị loại

	all, err := s.List(ctx, hid, storage.ListFilter{Type: common.TypeExpense, IncludeHidden: true})
	require.NoError(t, err)
	require.Len(t, all, 3)

	trees := model.BuildTree(visible)
	require.Len(t, trees, 1)
	assert.Equal(t, root.ID, trees[0].ID)
	require.Len(t, trees[0].Children, 1)
	assert.Equal(t, child.ID, trees[0].Children[0].ID)
}

func TestUpdateConditional(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	hid, _ := newHousehold(t, db)
	ctx := context.Background()

	cat := mkCategory(t, s, hid, "Gốc", common.TypeExpense, nil, false)
	fresh, err := s.FindByID(ctx, hid, cat.ID)
	require.NoError(t, err)

	rows, err := s.UpdateConditional(ctx, hid, cat.ID, fresh.UpdatedAt, map[string]any{"name": "Đổi tên"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	// mốc cũ giờ đã lệch → 0 hàng (D6)
	rows, err = s.UpdateConditional(ctx, hid, cat.ID, fresh.UpdatedAt, map[string]any{"name": "Đổi nữa"})
	require.NoError(t, err)
	assert.Equal(t, int64(0), rows)
}

// TestDeleteReassignAtomic chạy biz THẬT trên store THẬT: gán lại nguyên tử,
// và rollback toàn bộ khi mốc lệch (SC-007, D12).
func TestDeleteReassignAtomic(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	hid, uid := newHousehold(t, db)
	ctx := context.Background()

	victim := mkCategory(t, s, hid, "Sắp xóa", common.TypeExpense, nil, false)
	child := mkCategory(t, s, hid, "Con sắp xóa", common.TypeExpense, &victim.ID, false)
	target := mkCategory(t, s, hid, "Đích", common.TypeExpense, nil, false)
	mkTransaction(t, db, hid, uid, victim.ID, common.TypeExpense)
	mkTransaction(t, db, hid, uid, child.ID, common.TypeExpense)
	require.NoError(t, s.UpsertRule(ctx, hid, "keyword xóa", victim.ID))

	runner := categorybiz.TxRunnerFunc(func(ctx context.Context, fn func(categorybiz.DeleteStore) error) error {
		return s.InTx(ctx, func(tx *storage.SQLStore) error { return fn(tx) })
	})
	biz := categorybiz.NewDeleteCategoryBiz(runner)

	freshVictim, err := s.FindByID(ctx, hid, victim.ID)
	require.NoError(t, err)

	// (a) mốc lệch → CONCURRENCY_CONFLICT và KHÔNG có gì thay đổi (rollback)
	stale := freshVictim.UpdatedAt.Add(-time.Second)
	err = biz.Delete(ctx, hid, victim.ID, categorybiz.DeleteCategoryInput{
		Mode: categorybiz.DeleteModeReassign, TargetCategoryID: &target.ID, ExpectedUpdatedAt: stale,
	})
	require.Error(t, err)
	var n int64
	db.Table("transactions").Where("category_id IN ?", []uuid.UUID{victim.ID, child.ID}).Count(&n)
	assert.Equal(t, int64(2), n, "rollback phải giữ nguyên giao dịch")
	_, err = s.FindByID(ctx, hid, child.ID)
	assert.NoError(t, err, "rollback phải giữ nguyên danh mục con")

	// (b) mốc đúng → chuyển toàn bộ giao dịch, xóa cha + con + rule
	err = biz.Delete(ctx, hid, victim.ID, categorybiz.DeleteCategoryInput{
		Mode: categorybiz.DeleteModeReassign, TargetCategoryID: &target.ID, ExpectedUpdatedAt: freshVictim.UpdatedAt,
	})
	require.NoError(t, err)
	db.Table("transactions").Where("household_id = ? AND category_id = ?", hid, target.ID).Count(&n)
	assert.Equal(t, int64(2), n, "toàn bộ giao dịch phải sang đích — không mồ côi")
	_, err = s.FindByID(ctx, hid, victim.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = s.FindByID(ctx, hid, child.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	db.Table("categorization_rules").Where("household_id = ?", hid).Count(&n)
	assert.Equal(t, int64(0), n)
}

func TestUpsertRuleAndListByType(t *testing.T) {
	db := testDB(t)
	s := storage.NewSQLStore(db)
	hid, _ := newHousehold(t, db)
	ctx := context.Background()

	eat := mkCategory(t, s, hid, "Ăn uống", common.TypeExpense, nil, false)
	salary := mkCategory(t, s, hid, "Lương", common.TypeIncome, nil, false)

	require.NoError(t, s.UpsertRule(ctx, hid, "trà sữa", eat.ID))
	require.NoError(t, s.UpsertRule(ctx, hid, "trà sữa", eat.ID)) // upsert → match_count++
	require.NoError(t, s.UpsertRule(ctx, hid, "lương tháng", salary.ID))

	rules, err := s.ListRulesByType(ctx, hid, common.TypeExpense)
	require.NoError(t, err)
	require.Len(t, rules, 1) // chỉ rule có danh mục cùng loại
	assert.Equal(t, "trà sữa", rules[0].Keyword)
	assert.Equal(t, 2, rules[0].MatchCount)
	assert.Equal(t, "Ăn uống", rules[0].CategoryName)
}
