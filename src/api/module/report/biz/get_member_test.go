package biz

import (
	"context"
	"testing"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMember struct {
	members  []model.MemberInfo
	sums     []model.MemberAggRow
	txns     []transactionmodel.ListItem
	total    int64
	former   []transactionmodel.ListItem
	formerN  int64
	lastUser uuid.UUID // created_by nhận ở MemberTransactions (kiểm passthrough)
}

func (m *mockMember) ListMembers(_ context.Context, _ uuid.UUID) ([]model.MemberInfo, error) {
	return m.members, nil
}
func (m *mockMember) FindMember(_ context.Context, _, userID uuid.UUID) (model.MemberInfo, bool, error) {
	for _, mi := range m.members {
		if mi.ID == userID {
			return mi, true, nil
		}
	}
	return model.MemberInfo{}, false, nil
}
func (m *mockMember) MemberSums(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]model.MemberAggRow, error) {
	return m.sums, nil
}
func (m *mockMember) MemberTransactions(_ context.Context, _, userID uuid.UUID, _, _ time.Time, _ common.Paging) ([]transactionmodel.ListItem, int64, error) {
	m.lastUser = userID
	return m.txns, m.total, nil
}
func (m *mockMember) FormerMemberTransactions(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _, _ time.Time, _ common.Paging) ([]transactionmodel.ListItem, int64, error) {
	return m.former, m.formerN, nil
}

func paging() common.Paging { return common.Paging{Page: 1, PageSize: 20} }

func TestGetMember_CurrentMember(t *testing.T) {
	alice := uuid.New()
	st := &mockMember{
		members: []model.MemberInfo{{ID: alice, DisplayName: "Alice"}},
		sums:    []model.MemberAggRow{{CreatedBy: alice, Income: 1000, Expense: 300}},
		txns:    []transactionmodel.ListItem{{}, {}},
		total:   2,
	}
	res, err := NewGetMemberBiz(st).Get(context.Background(), uuid.New(), alice.String(), date(2026, 8, 1), date(2026, 8, 31), paging())
	require.NoError(t, err)
	assert.Equal(t, alice.String(), res.MemberID)
	assert.Equal(t, "Alice", res.DisplayName)
	assert.False(t, res.IsFormer)
	assert.Equal(t, float64(700), res.Net) // 1000 - 300
	assert.Equal(t, int64(2), res.Total)
	assert.Equal(t, 1, res.Page)
	assert.Equal(t, alice, st.lastUser) // created_by passthrough
}

func TestGetMember_FormerBucket(t *testing.T) {
	alice, ghost := uuid.New(), uuid.New()
	st := &mockMember{
		members: []model.MemberInfo{{ID: alice, DisplayName: "Alice"}},
		sums: []model.MemberAggRow{
			{CreatedBy: alice, Income: 1000, Expense: 300},
			{CreatedBy: ghost, Income: 0, Expense: 800},
		},
		former:  []transactionmodel.ListItem{{}},
		formerN: 1,
	}
	res, err := NewGetMemberBiz(st).Get(context.Background(), uuid.New(), model.FormerMemberID, date(2026, 8, 1), date(2026, 8, 31), paging())
	require.NoError(t, err)
	assert.True(t, res.IsFormer)
	assert.Equal(t, model.FormerMemberID, res.MemberID)
	assert.Equal(t, model.FormerMemberLabel, res.DisplayName)
	assert.Equal(t, float64(800), res.Expense) // chỉ ghost
	assert.Equal(t, float64(-800), res.Net)
	assert.Equal(t, int64(1), res.Total)
}

func TestGetMember_NotAMemberReturns404(t *testing.T) {
	st := &mockMember{members: []model.MemberInfo{{ID: uuid.New(), DisplayName: "Alice"}}}
	_, err := NewGetMemberBiz(st).Get(context.Background(), uuid.New(), uuid.New().String(), date(2026, 8, 1), date(2026, 8, 31), paging())
	_, status := appErrCode(t, err)
	assert.Equal(t, 404, status)
}

func TestGetMember_BadUUIDReturns404(t *testing.T) {
	st := &mockMember{}
	_, err := NewGetMemberBiz(st).Get(context.Background(), uuid.New(), "not-a-uuid", date(2026, 8, 1), date(2026, 8, 31), paging())
	_, status := appErrCode(t, err)
	assert.Equal(t, 404, status)
}

func TestGetMember_ValidateRange(t *testing.T) {
	st := &mockMember{}
	_, err := NewGetMemberBiz(st).Get(context.Background(), uuid.New(), model.FormerMemberID, date(2026, 8, 31), date(2026, 8, 1), paging())
	_, status := appErrCode(t, err)
	assert.Equal(t, 400, status)
}
