package biz

import (
	"context"
	"math"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/budget/model"

	"github.com/google/uuid"
)

// BudgetLister — đọc ngân sách của hộ + chuyển ONE_TIME hết kỳ sang ENDED (D20).
type BudgetLister interface {
	List(ctx context.Context, householdID uuid.UUID, status string) ([]model.BudgetView, error)
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.BudgetView, error)
	EndExpiredOneTime(ctx context.Context, householdID uuid.UUID, now time.Time) error
}

// ProgressReader — tổng chi suy ra theo kỳ (chỉ đọc transactions/categories — D21).
type ProgressReader interface {
	ChildCategoryIDs(ctx context.Context, householdID, parentID uuid.UUID) ([]uuid.UUID, error)
	SumExpenseByCategories(ctx context.Context, householdID uuid.UUID, categoryIDs []uuid.UUID, start, endExcl time.Time) (float64, error)
	SumExpenseTotal(ctx context.Context, householdID uuid.UUID, start, endExcl time.Time) (float64, error)
}

// AlertLister — cảnh báo hiện có để embed vào response (D23). nil = chưa bật (US1).
type AlertLister interface {
	ListByBudgetIDs(ctx context.Context, budgetIDs []uuid.UUID) ([]model.BudgetAlert, error)
}

type ListBudgetsBiz struct {
	store      BudgetLister
	progress   ProgressReader
	alertStore AlertLister
}

func NewListBudgetsBiz(store BudgetLister, progress ProgressReader, alertStore AlertLister) *ListBudgetsBiz {
	return &ListBudgetsBiz{store: store, progress: progress, alertStore: alertStore}
}

// List — GET /api/budgets (FR-005/006, D21): mỗi ngân sách kèm spent/percent/period_key
// kỳ hiện tại + cảnh báo đang hoạt động. Chỉ tính EXPENSE; danh mục con gộp vào cha.
func (b *ListBudgetsBiz) List(ctx context.Context, householdID uuid.UUID, status string, now time.Time) ([]model.BudgetView, error) {
	// ONE_TIME quá end_date → ENDED (best-effort, không chặn đọc).
	_ = b.store.EndExpiredOneTime(ctx, householdID, now)

	views, err := b.store.List(ctx, householdID, status)
	if err != nil {
		return nil, common.NewInternal(err)
	}

	// Cảnh báo hiện có (nếu evaluator đã bật) → gom theo budget_id.
	alertsByBudget := map[uuid.UUID][]model.BudgetAlert{}
	if b.alertStore != nil && len(views) > 0 {
		ids := make([]uuid.UUID, len(views))
		for i := range views {
			ids[i] = views[i].ID
		}
		if all, aerr := b.alertStore.ListByBudgetIDs(ctx, ids); aerr == nil {
			for _, a := range all {
				alertsByBudget[a.BudgetID] = append(alertsByBudget[a.BudgetID], a)
			}
		}
	}

	for i := range views {
		if err := b.enrich(ctx, householdID, &views[i], now, alertsByBudget[views[i].ID]); err != nil {
			return nil, common.NewInternal(err)
		}
	}
	return views, nil
}

// Get — một ngân sách của hộ kèm tiến độ/cảnh báo suy ra (GET /:id, và trả về sau
// khi tạo/sửa). Ngoài hộ / không tồn tại → store trả ErrNotFound (transport → 404).
func (b *ListBudgetsBiz) Get(ctx context.Context, householdID, id uuid.UUID, now time.Time) (*model.BudgetView, error) {
	v, err := b.store.FindByID(ctx, householdID, id)
	if err != nil {
		return nil, err
	}
	var alerts []model.BudgetAlert
	if b.alertStore != nil {
		alerts, _ = b.alertStore.ListByBudgetIDs(ctx, []uuid.UUID{id})
	}
	if err := b.enrich(ctx, householdID, v, now, alerts); err != nil {
		return nil, common.NewInternal(err)
	}
	return v, nil
}

// enrich — gắn period_key + spent + percent + alerts kỳ hiện tại cho một ngân sách.
func (b *ListBudgetsBiz) enrich(ctx context.Context, householdID uuid.UUID, v *model.BudgetView, now time.Time, alerts []model.BudgetAlert) error {
	p := model.ResolvePeriod(v.PeriodType, v.StartDate, v.EndDate, now)
	v.PeriodKey = p.Key
	spent, err := b.spentFor(ctx, householdID, v, p)
	if err != nil {
		return err
	}
	v.Spent = spent
	v.Percent = percentOf(spent, v.LimitAmount)
	v.Alerts = currentAlerts(alerts, p.Key)
	return nil
}

func (b *ListBudgetsBiz) spentFor(ctx context.Context, householdID uuid.UUID, v *model.BudgetView, p model.Period) (float64, error) {
	return computeSpent(ctx, householdID, v.Type, v.CategoryID, p, b.progress)
}

// computeSpent — tổng chi suy ra của một ngân sách trong kỳ (D21): CATEGORY gộp danh
// mục con một cấp; TOTAL cộng mọi chi của hộ. Dùng chung cho list + evaluator cảnh báo.
func computeSpent(ctx context.Context, householdID uuid.UUID, typ string, categoryID *uuid.UUID, p model.Period, progress ProgressReader) (float64, error) {
	if typ == model.TypeTotal {
		return progress.SumExpenseTotal(ctx, householdID, p.Start, p.EndExcl)
	}
	if categoryID == nil {
		return 0, nil
	}
	ids := []uuid.UUID{*categoryID}
	children, err := progress.ChildCategoryIDs(ctx, householdID, *categoryID)
	if err != nil {
		return 0, err
	}
	ids = append(ids, children...)
	return progress.SumExpenseByCategories(ctx, householdID, ids, p.Start, p.EndExcl)
}

// percentOf — round(spent/limit × 100) (D21). Frontend hiển thị chính xác từ spent/limit.
func percentOf(spent, limit float64) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Round(spent / limit * 100))
}

// currentAlerts — chỉ cảnh báo của kỳ hiện tại (period_key), map sang AlertView.
func currentAlerts(alerts []model.BudgetAlert, periodKey string) []model.AlertView {
	out := []model.AlertView{}
	for _, a := range alerts {
		if a.PeriodKey == periodKey {
			out = append(out, model.AlertView{Level: a.Level, OverAmount: a.OverAmount, FiredAt: a.FiredAt})
		}
	}
	return out
}
