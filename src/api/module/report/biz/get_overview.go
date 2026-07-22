package biz

import (
	"context"
	"math"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"

	"github.com/google/uuid"
)

// OverviewReader — nguồn đọc tổng hợp cho báo cáo tổng quan (chỉ SELECT/SUM — D41).
type OverviewReader interface {
	SumIncomeExpense(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) (income, expense float64, err error)
	CategoryBreakdown(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time) ([]model.CategoryBreakdown, error)
	TrendIncomeExpense(ctx context.Context, householdID uuid.UUID, from, endExcl time.Time, unit string) ([]model.TrendRow, error)
}

type GetOverviewBiz struct{ store OverviewReader }

func NewGetOverviewBiz(store OverviewReader) *GetOverviewBiz {
	return &GetOverviewBiz{store: store}
}

// Get — GET /api/reports/overview (D34): tổng thu/chi/ròng + phân bổ danh mục (gộp con,
// %) + xu hướng thu/chi (điền mốc trống). Validate to ≥ from (FR-008).
func (b *GetOverviewBiz) Get(ctx context.Context, householdID uuid.UUID, from, to time.Time) (*model.ReportOverview, error) {
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	endExcl := model.EndExclusive(to)
	unit := model.ChooseGroupUnit(from, to)

	income, expense, err := b.store.SumIncomeExpense(ctx, householdID, from, endExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	cats, err := b.store.CategoryBreakdown(ctx, householdID, from, endExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	for i := range cats {
		cats[i].Percent = percentOf(cats[i].Amount, expense)
	}
	if cats == nil {
		cats = []model.CategoryBreakdown{}
	}
	rows, err := b.store.TrendIncomeExpense(ctx, householdID, from, endExcl, unit)
	if err != nil {
		return nil, common.NewInternal(err)
	}

	return &model.ReportOverview{
		From:              dateStr(from),
		To:                dateStr(to),
		GroupUnit:         unit,
		Income:            income,
		Expense:           expense,
		Net:               income - expense,
		CategoryBreakdown: cats,
		Trend:             fillTrend(from, to, unit, rows),
	}, nil
}

// --- helpers dùng chung ---

func validateRange(from, to time.Time) error {
	if to.Before(from) {
		return common.NewBadRequest("ngày kết thúc không được trước ngày bắt đầu").WithField("to")
	}
	return nil
}

func dateStr(t time.Time) string { return t.Format("2006-01-02") }

func percentOf(amount, total float64) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(amount / total * 100))
}

// fillTrend — dựng chuỗi mốc liên tục [from,to] theo unit, điền 0 cho mốc thiếu (D36).
func fillTrend(from, to time.Time, unit string, rows []model.TrendRow) []model.TrendPoint {
	byKey := make(map[string]model.TrendRow, len(rows))
	for _, r := range rows {
		byKey[model.BucketKey(unit, r.Bucket)] = r
	}
	out := []model.TrendPoint{}
	for _, k := range model.BucketSequence(unit, from, to) {
		r := byKey[k]
		out = append(out, model.TrendPoint{Bucket: k, Income: r.Income, Expense: r.Expense})
	}
	return out
}
