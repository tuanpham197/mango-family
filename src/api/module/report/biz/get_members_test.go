package biz

import (
	"context"
	"testing"
	"time"

	"household-finance/api/module/report/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMembers struct {
	members []model.MemberInfo
	sums    []model.MemberAggRow
}

func (m *mockMembers) ListMembers(_ context.Context, _ uuid.UUID) ([]model.MemberInfo, error) {
	return m.members, nil
}
func (m *mockMembers) MemberSums(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]model.MemberAggRow, error) {
	return m.sums, nil
}

func TestGetMembers_ZeroFillSortAndReconcile(t *testing.T) {
	alice, bob, carol := uuid.New(), uuid.New(), uuid.New()
	st := &mockMembers{
		members: []model.MemberInfo{
			{ID: alice, DisplayName: "Alice"},
			{ID: bob, DisplayName: "Bob"},
			{ID: carol, DisplayName: "Carol"}, // không có giao dịch → 0/0/0
		},
		sums: []model.MemberAggRow{
			{CreatedBy: alice, Income: 12000000, Expense: 4500000},
			{CreatedBy: bob, Income: 0, Expense: 3200000},
		},
	}
	res, err := NewGetMembersBiz(st).Get(context.Background(), uuid.New(), date(2026, 8, 1), date(2026, 8, 31))
	require.NoError(t, err)
	require.Len(t, res.Members, 3)

	// Sắp theo Chi giảm dần: Alice (4.5M) → Bob (3.2M) → Carol (0)
	assert.Equal(t, "Alice", res.Members[0].DisplayName)
	assert.Equal(t, float64(7500000), res.Members[0].Net) // 12M - 4.5M
	assert.Equal(t, "Bob", res.Members[1].DisplayName)
	assert.Equal(t, float64(-3200000), res.Members[1].Net)
	assert.Equal(t, "Carol", res.Members[2].DisplayName)
	assert.Equal(t, float64(0), res.Members[2].Income)
	assert.Equal(t, float64(0), res.Members[2].Expense)
	assert.False(t, res.Members[2].IsFormer)

	// Đối soát: Σ dòng == totals (SC-001)
	var sumInc, sumExp float64
	for _, m := range res.Members {
		sumInc += m.Income
		sumExp += m.Expense
	}
	assert.Equal(t, res.Totals.Income, sumInc)
	assert.Equal(t, res.Totals.Expense, sumExp)
	assert.Equal(t, float64(12000000), res.Totals.Income)
	assert.Equal(t, float64(7700000), res.Totals.Expense)
	assert.Equal(t, res.Totals.Income-res.Totals.Expense, res.Totals.Net)
}

func TestGetMembers_FormerBucketOnlyWhenPresent(t *testing.T) {
	alice, ghost := uuid.New(), uuid.New() // ghost không thuộc members → former
	st := &mockMembers{
		members: []model.MemberInfo{{ID: alice, DisplayName: "Alice"}},
		sums: []model.MemberAggRow{
			{CreatedBy: alice, Income: 1000, Expense: 200},
			{CreatedBy: ghost, Income: 0, Expense: 800}, // người đã rời hộ
		},
	}
	res, err := NewGetMembersBiz(st).Get(context.Background(), uuid.New(), date(2026, 8, 1), date(2026, 8, 31))
	require.NoError(t, err)
	require.Len(t, res.Members, 2)

	former := res.Members[len(res.Members)-1] // former xếp cuối
	assert.True(t, former.IsFormer)
	assert.Equal(t, model.FormerMemberID, former.MemberID)
	assert.Equal(t, model.FormerMemberLabel, former.DisplayName)
	assert.Equal(t, float64(800), former.Expense)

	// Đối soát vẫn đúng khi có former (FR-005/FR-010)
	assert.Equal(t, float64(1000), res.Totals.Income)
	assert.Equal(t, float64(1000), res.Totals.Expense) // 200 + 800
}

func TestGetMembers_NoFormerRowWhenAllCurrent(t *testing.T) {
	alice := uuid.New()
	st := &mockMembers{
		members: []model.MemberInfo{{ID: alice, DisplayName: "Alice"}},
		sums:    []model.MemberAggRow{{CreatedBy: alice, Income: 500, Expense: 100}},
	}
	res, err := NewGetMembersBiz(st).Get(context.Background(), uuid.New(), date(2026, 8, 1), date(2026, 8, 31))
	require.NoError(t, err)
	require.Len(t, res.Members, 1)
	assert.False(t, res.Members[0].IsFormer)
}

func TestGetMembers_FallbackEmailAndValidateRange(t *testing.T) {
	u := uuid.New()
	st := &mockMembers{
		members: []model.MemberInfo{{ID: u, DisplayName: "", Email: "bob@example.com"}},
		sums:    []model.MemberAggRow{},
	}
	res, err := NewGetMembersBiz(st).Get(context.Background(), uuid.New(), date(2026, 8, 1), date(2026, 8, 31))
	require.NoError(t, err)
	assert.Equal(t, "bob@example.com", res.Members[0].DisplayName) // FR-009 fallback

	// to < from → 400
	_, err = NewGetMembersBiz(st).Get(context.Background(), uuid.New(), date(2026, 8, 31), date(2026, 8, 1))
	code, status := appErrCode(t, err)
	assert.Equal(t, 400, status)
	_ = code
}
