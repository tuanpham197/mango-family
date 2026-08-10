//go:build integration

package storage_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Feature 008 — tổng hợp theo thành viên + drill-in.
func TestReportStorage_MemberAggregations(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	s := storage.NewSQLStore(db)
	hid, alice := newHousehold(t, db) // alice là owner + thành viên
	other, otherOwner := newHousehold(t, db)

	// Bob: thành viên thứ hai của hộ.
	bob := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, 'Bob', 'x')`,
		bob, fmt.Sprintf("bob-%s@test.local", bob)).Error)
	require.NoError(t, db.Exec(`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, hid, bob).Error)
	// Dave: thành viên KHÔNG có giao dịch.
	dave := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, 'Dave', 'x')`,
		dave, fmt.Sprintf("dave-%s@test.local", dave)).Error)
	require.NoError(t, db.Exec(`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, hid, dave).Error)
	// Ghost: đã RỜI hộ — có giao dịch nhưng KHÔNG còn trong household_members.
	ghost := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, 'Ghost', 'x')`,
		ghost, fmt.Sprintf("ghost-%s@test.local", ghost)).Error)
	t.Cleanup(func() { db.Exec(`DELETE FROM users WHERE id IN (?, ?, ?)`, bob, dave, ghost) })

	salary := newCategory(t, db, hid, "Lương", common.TypeIncome, nil)
	food := newCategory(t, db, hid, "Ăn uống", common.TypeExpense, nil)
	acc := newAccount(t, db, hid)
	otherCat := newCategory(t, db, other, "Khác hộ", common.TypeExpense, nil)
	otherAcc := newAccount(t, db, other)

	utc := func(d int) time.Time { return time.Date(2026, 8, d, 12, 0, 0, 0, time.UTC) }
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endExcl := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	insertTxn(t, db, hid, alice, salary, acc, 12000000, common.TypeIncome, utc(2))
	insertTxn(t, db, hid, alice, food, acc, 4500000, common.TypeExpense, utc(5))
	insertTxn(t, db, hid, bob, food, acc, 3200000, common.TypeExpense, utc(6))
	insertTxn(t, db, hid, ghost, food, acc, 800000, common.TypeExpense, utc(7))          // former
	insertTxn(t, db, other, otherOwner, otherCat, otherAcc, 999000, common.TypeExpense, utc(8)) // hộ khác

	// ListMembers — 3 thành viên hiện tại (alice, bob, dave), KHÔNG có ghost.
	members, err := s.ListMembers(ctx, hid)
	require.NoError(t, err)
	require.Len(t, members, 3)
	ids := map[uuid.UUID]bool{}
	for _, m := range members {
		ids[m.ID] = true
	}
	assert.True(t, ids[alice] && ids[bob] && ids[dave])
	assert.False(t, ids[ghost])

	// FindMember — trong hộ true; ghost/hộ khác false (chống dò).
	_, found, err := s.FindMember(ctx, hid, bob)
	require.NoError(t, err)
	assert.True(t, found)
	_, found, err = s.FindMember(ctx, hid, ghost)
	require.NoError(t, err)
	assert.False(t, found)

	// MemberSums — GROUP BY created_by, cô lập hộ.
	sums, err := s.MemberSums(ctx, hid, from, endExcl)
	require.NoError(t, err)
	byUser := map[uuid.UUID][2]float64{}
	for _, r := range sums {
		byUser[r.CreatedBy] = [2]float64{r.Income, r.Expense}
	}
	assert.Equal(t, [2]float64{12000000, 4500000}, byUser[alice])
	assert.Equal(t, [2]float64{0, 3200000}, byUser[bob])
	assert.Equal(t, [2]float64{0, 800000}, byUser[ghost])
	_, hasDave := byUser[dave]
	assert.False(t, hasDave) // Dave không có giao dịch → không xuất hiện trong sums

	// Đối soát: Σ member sums == tổng hộ (SumIncomeExpense) — SC-001.
	hInc, hExp, err := s.SumIncomeExpense(ctx, hid, from, endExcl)
	require.NoError(t, err)
	var sInc, sExp float64
	for _, r := range sums {
		sInc += r.Income
		sExp += r.Expense
	}
	assert.Equal(t, hInc, sInc)
	assert.Equal(t, hExp, sExp)
	assert.Equal(t, float64(8500000), sExp) // 4.5M + 3.2M + 0.8M (không tính hộ khác)

	// MemberTransactions — của Alice, phân trang.
	pg := func(page, size int) common.Paging { return common.Paging{Page: page, PageSize: size} }
	txns, total, err := s.MemberTransactions(ctx, hid, alice, from, endExcl, pg(1, 20))
	require.NoError(t, err)
	assert.Equal(t, int64(2), total) // salary + food
	assert.Len(t, txns, 2)
	page2, total2, err := s.MemberTransactions(ctx, hid, alice, from, endExcl, pg(1, 1))
	require.NoError(t, err)
	assert.Equal(t, int64(2), total2)
	assert.Len(t, page2, 1) // phân trang page_size=1

	// Dave — không có giao dịch.
	_, daveTotal, err := s.MemberTransactions(ctx, hid, dave, from, endExcl, pg(1, 20))
	require.NoError(t, err)
	assert.Equal(t, int64(0), daveTotal)

	// FormerMemberTransactions — của ghost (created_by ∉ members).
	fTxns, fTotal, err := s.FormerMemberTransactions(ctx, hid, []uuid.UUID{alice, bob, dave}, from, endExcl, pg(1, 20))
	require.NoError(t, err)
	assert.Equal(t, int64(1), fTotal)
	assert.Len(t, fTxns, 1)
	assert.Equal(t, float64(800000), fTxns[0].Amount)
}
