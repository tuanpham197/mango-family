package biz

import (
	"context"
	"testing"
	"time"

	"household-finance/api/module/overview/model"
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPercentOf(t *testing.T) {
	assert.Equal(t, 37, percentOf(3700000, 10000000))
	assert.Equal(t, 100, percentOf(50, 50))
	assert.Equal(t, 0, percentOf(100, 0)) // tránh chia 0
}

func TestChangePercent(t *testing.T) {
	assert.Nil(t, changePercent(100, 0)) // mẫu số 0 → null (D28)
	require.NotNil(t, changePercent(105, 100))
	assert.Equal(t, 5.0, *changePercent(105, 100))
	assert.Equal(t, -10.0, *changePercent(90, 100))
	assert.Equal(t, 5.2, *changePercent(1052, 1000)) // làm tròn 1 chữ số
}

type mockReader struct {
	netWorth   float64
	asOf       float64
	income     float64
	expense    float64
	cats       []model.CategorySpending
	recent     []transactionmodel.ListItem
	asOfBefore time.Time
}

func (m *mockReader) NetWorth(_ context.Context, _ uuid.UUID) (float64, error) {
	return m.netWorth, nil
}
func (m *mockReader) NetWorthAsOf(_ context.Context, _ uuid.UUID, before time.Time) (float64, error) {
	m.asOfBefore = before
	return m.asOf, nil
}
func (m *mockReader) MonthIncomeExpense(_ context.Context, _ uuid.UUID, _, _ time.Time) (float64, float64, error) {
	return m.income, m.expense, nil
}
func (m *mockReader) CategorySpending(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]model.CategorySpending, error) {
	return m.cats, nil
}
func (m *mockReader) RecentTransactions(_ context.Context, _ uuid.UUID, _ int) ([]transactionmodel.ListItem, error) {
	return m.recent, nil
}

func TestGetOverview_Assembles(t *testing.T) {
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	mr := &mockReader{
		netWorth: 24560000,
		asOf:     23346000, // cuối tháng trước
		income:   18200000,
		expense:  9400000,
		cats: []model.CategorySpending{
			{CategoryID: uuid.New(), CategoryName: "Ăn uống", Amount: 3500000},
			{CategoryID: uuid.New(), CategoryName: "Di chuyển", Amount: 900000},
		},
	}
	biz := NewGetOverviewBiz(mr)
	s, err := biz.Get(context.Background(), uuid.New(), now)
	require.NoError(t, err)

	assert.Equal(t, float64(24560000), s.NetWorth)
	require.NotNil(t, s.NetWorthChangePercent)
	assert.Equal(t, 5.2, *s.NetWorthChangePercent) // (24.56M-23.346M)/23.346M ≈ 5.2%
	assert.Equal(t, float64(18200000), s.Month.Income)
	assert.Equal(t, float64(9400000), s.Month.Expense)
	assert.Equal(t, float64(8800000), s.Month.Net) // income - expense

	// percent theo danh mục = round(amount/expense×100)
	assert.Equal(t, 37, s.CategorySpending[0].Percent) // 3.5M/9.4M ≈ 37%
	assert.Equal(t, 10, s.CategorySpending[1].Percent) // 0.9M/9.4M ≈ 10%

	// as-of dùng đầu tháng hiện tại (cửa sổ MONTHLY)
	assert.Equal(t, time.Date(2026, 7, 1, 0, 0, 0, 0, now.Location()), mr.asOfBefore)

	// slice rỗng-an toàn (không nil)
	assert.NotNil(t, s.RecentTransactions)
}
