package biz

import (
	"context"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/report/model"
	transactionmodel "household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

// CategoryReader — nguồn đọc cho báo cáo chi tiết theo danh mục (chỉ đọc — D38).
type CategoryReader interface {
	FindCategoryName(ctx context.Context, householdID, id uuid.UUID) (name string, found bool, err error)
	ChildCategoryIDs(ctx context.Context, householdID, parentID uuid.UUID) ([]uuid.UUID, error)
	CategorySum(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, from, endExcl time.Time) (float64, error)
	CategoryTransactions(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, from, endExcl time.Time) ([]transactionmodel.ListItem, error)
	CategoryTrend(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, from, endExcl time.Time, unit string) ([]model.CategoryTrendRow, error)
}

type GetCategoryBiz struct{ store CategoryReader }

func NewGetCategoryBiz(store CategoryReader) *GetCategoryBiz {
	return &GetCategoryBiz{store: store}
}

// Get — GET /api/reports/category/:id (D38): tổng chi + danh sách giao dịch + xu hướng
// của một danh mục (gộp danh mục con một cấp) trong khoảng. Ngoài hộ → 404 (FR-007).
func (b *GetCategoryBiz) Get(ctx context.Context, householdID, categoryID uuid.UUID, from, to time.Time) (*model.CategoryReport, error) {
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	name, found, err := b.store.FindCategoryName(ctx, householdID, categoryID)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	if !found {
		return nil, common.NewNotFound("danh mục không tồn tại")
	}

	children, err := b.store.ChildCategoryIDs(ctx, householdID, categoryID)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	ids := append([]uuid.UUID{categoryID}, children...) // danh mục + con một cấp

	endExcl := model.EndExclusive(to)
	unit := model.ChooseGroupUnit(from, to)

	total, err := b.store.CategorySum(ctx, householdID, ids, from, endExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	txns, err := b.store.CategoryTransactions(ctx, householdID, ids, from, endExcl)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	rows, err := b.store.CategoryTrend(ctx, householdID, ids, from, endExcl, unit)
	if err != nil {
		return nil, common.NewInternal(err)
	}

	return &model.CategoryReport{
		CategoryID:   categoryID,
		CategoryName: name,
		From:         dateStr(from),
		To:           dateStr(to),
		GroupUnit:    unit,
		Total:        total,
		Transactions: txns,
		Trend:        fillCategoryTrend(from, to, unit, rows),
	}, nil
}

// fillCategoryTrend — như fillTrend nhưng cho một chuỗi amount (chi của danh mục).
func fillCategoryTrend(from, to time.Time, unit string, rows []model.CategoryTrendRow) []model.CategoryTrendPoint {
	byKey := make(map[string]float64, len(rows))
	for _, r := range rows {
		byKey[model.BucketKey(unit, r.Bucket)] = r.Amount
	}
	out := []model.CategoryTrendPoint{}
	for _, k := range model.BucketSequence(unit, from, to) {
		out = append(out, model.CategoryTrendPoint{Bucket: k, Amount: byKey[k]})
	}
	return out
}
