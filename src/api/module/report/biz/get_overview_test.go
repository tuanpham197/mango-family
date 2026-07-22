package biz

import (
	"context"
	"testing"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func appErrCode(t *testing.T, err error) (string, int) {
	t.Helper()
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr.Code, appErr.StatusCode
}

func TestPercentOf(t *testing.T) {
	assert.Equal(t, 79, percentOf(3500000, 4420000))
	assert.Equal(t, 0, percentOf(100, 0))
}

func TestFillTrend_FillsEmptyBuckets(t *testing.T) {
	rows := []model.TrendRow{
		{Bucket: date(2026, 7, 2), Income: 18000000},
		{Bucket: date(2026, 7, 5), Expense: 3500000},
	}
	out := fillTrend(date(2026, 7, 1), date(2026, 7, 5), model.GroupDay, rows)
	require.Len(t, out, 5) // 5 ngày liên tục
	assert.Equal(t, "2026-07-01", out[0].Bucket)
	assert.Equal(t, float64(0), out[0].Income)
	assert.Equal(t, float64(18000000), out[1].Income) // 07-02
	assert.Equal(t, float64(3500000), out[4].Expense) // 07-05
}

type mockOverview struct {
	income, expense float64
	cats            []model.CategoryBreakdown
	trend           []model.TrendRow
}

func (m *mockOverview) SumIncomeExpense(_ context.Context, _ uuid.UUID, _, _ time.Time) (float64, float64, error) {
	return m.income, m.expense, nil
}
func (m *mockOverview) CategoryBreakdown(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]model.CategoryBreakdown, error) {
	return m.cats, nil
}
func (m *mockOverview) TrendIncomeExpense(_ context.Context, _ uuid.UUID, _, _ time.Time, _ string) ([]model.TrendRow, error) {
	return m.trend, nil
}

func TestGetOverview_ValidatesRange(t *testing.T) {
	biz := NewGetOverviewBiz(&mockOverview{})
	_, err := biz.Get(context.Background(), uuid.New(), date(2026, 7, 31), date(2026, 7, 1)) // to < from
	code, status := appErrCode(t, err)
	assert.Equal(t, common.ErrCodeInvalidRequest, code)
	assert.Equal(t, 400, status)
}

func TestGetOverview_Assembles(t *testing.T) {
	mo := &mockOverview{
		income:  18000000,
		expense: 4420000,
		cats: []model.CategoryBreakdown{
			{CategoryID: uuid.New(), CategoryName: "Ăn uống", Amount: 3500000},
			{CategoryID: uuid.New(), CategoryName: "Di chuyển", Amount: 920000},
		},
	}
	res, err := NewGetOverviewBiz(mo).Get(context.Background(), uuid.New(), date(2026, 7, 1), date(2026, 7, 31))
	require.NoError(t, err)
	assert.Equal(t, model.GroupDay, res.GroupUnit)        // 31 ngày → day
	assert.Equal(t, float64(13580000), res.Net)           // income − expense
	assert.Equal(t, 79, res.CategoryBreakdown[0].Percent) // 3.5M/4.42M ≈ 79%
	assert.Equal(t, 21, res.CategoryBreakdown[1].Percent) // 0.92M/4.42M ≈ 21%
	assert.Len(t, res.Trend, 31)                          // điền đủ 31 ngày
	assert.Equal(t, "2026-07-01", res.From)
}
