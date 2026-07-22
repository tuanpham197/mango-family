package biz

import (
	"context"
	"math"
	"time"

	"household-finance/api/common"
	budgetmodel "household-finance/api/module/budget/model"
	"household-finance/api/module/overview/model"
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

const recentLimit = 5

// OverviewReader — nguồn đọc suy ra (chỉ SELECT/SUM — D27/D28/D29/D30).
type OverviewReader interface {
	NetWorth(ctx context.Context, householdID uuid.UUID) (float64, error)
	NetWorthAsOf(ctx context.Context, householdID uuid.UUID, before time.Time) (float64, error)
	MonthIncomeExpense(ctx context.Context, householdID uuid.UUID, start, endExcl time.Time) (income, expense float64, err error)
	CategorySpending(ctx context.Context, householdID uuid.UUID, start, endExcl time.Time) ([]model.CategorySpending, error)
	RecentTransactions(ctx context.Context, householdID uuid.UUID, limit int) ([]transactionmodel.ListItem, error)
}

type GetOverviewBiz struct{ store OverviewReader }

func NewGetOverviewBiz(store OverviewReader) *GetOverviewBiz {
	return &GetOverviewBiz{store: store}
}

// Get — GET /api/overview (D27): ráp OverviewSummary từ các giá trị suy ra của tháng
// hiện tại (cửa sổ tháng dùng lại budget.ResolvePeriod MONTHLY). Chỉ đọc, tính khi gọi.
func (b *GetOverviewBiz) Get(ctx context.Context, householdID uuid.UUID, now time.Time) (*model.OverviewSummary, error) {
	p := budgetmodel.ResolvePeriod(budgetmodel.PeriodMonthly, nil, nil, now) // cửa sổ tháng [Start, EndExcl)

	netWorth, err := b.store.NetWorth(ctx, householdID)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	prevEnd, err := b.store.NetWorthAsOf(ctx, householdID, p.Start) // tài sản ròng cuối tháng trước = as-of đầu tháng
	if err != nil {
		return nil, common.NewInternal(err)
	}
	income, expense, err := b.store.MonthIncomeExpense(ctx, householdID, p.Start, p.EndExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	cats, err := b.store.CategorySpending(ctx, householdID, p.Start, p.EndExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	recent, err := b.store.RecentTransactions(ctx, householdID, recentLimit)
	if err != nil {
		return nil, common.NewInternal(err)
	}

	for i := range cats {
		cats[i].Percent = percentOf(cats[i].Amount, expense)
	}
	if cats == nil {
		cats = []model.CategorySpending{}
	}
	if recent == nil {
		recent = []transactionmodel.ListItem{}
	}

	return &model.OverviewSummary{
		NetWorth:              netWorth,
		NetWorthChangePercent: changePercent(netWorth, prevEnd),
		Month:                 model.MonthCashflow{Income: income, Expense: expense, Net: income - expense},
		CategorySpending:      cats,
		RecentTransactions:    recent,
	}, nil
}

// percentOf — round(amount/total × 100); 0 khi total ≤ 0.
func percentOf(amount, total float64) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(amount / total * 100))
}

// changePercent — % thay đổi tài sản ròng so mốc trước (D28); nil khi mẫu số 0
// (không đủ dữ liệu tháng trước). Làm tròn 1 chữ số thập phân.
func changePercent(now, prev float64) *float64 {
	if prev == 0 {
		return nil
	}
	v := math.Round((now-prev)/math.Abs(prev)*1000) / 10
	return &v
}
