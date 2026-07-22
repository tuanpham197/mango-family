package biz

import (
	"context"
	"errors"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/budget/model"
	"household-finance/api/module/budget/storage"

	"github.com/google/uuid"
)

// BudgetUpdateStore — đọc hiện trạng + conditional write theo mốc lạc quan (D25).
type BudgetUpdateStore interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.BudgetView, error)
	FindActiveDuplicate(ctx context.Context, householdID uuid.UUID, typ string, categoryID *uuid.UUID, periodType string, excludeID *uuid.UUID) (*model.Budget, error)
	UpdateConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time, changes map[string]any) (int64, error)
	Exists(ctx context.Context, householdID, id uuid.UUID) (bool, error)
}

// BudgetViewer — dựng view kèm tiến độ/cảnh báo (ListBudgetsBiz) cho response + data
// đính kèm khi CONCURRENCY_CONFLICT.
type BudgetViewer interface {
	Get(ctx context.Context, householdID, id uuid.UUID, now time.Time) (*model.BudgetView, error)
}

type UpdateBudgetInput struct {
	BudgetFields
	ExpectedUpdatedAt time.Time
}

type UpdateBudgetBiz struct {
	store    BudgetUpdateStore
	catStore CategoryFinder
	viewer   BudgetViewer
}

func NewUpdateBudgetBiz(store BudgetUpdateStore, catStore CategoryFinder, viewer BudgetViewer) *UpdateBudgetBiz {
	return &UpdateBudgetBiz{store: store, catStore: catStore, viewer: viewer}
}

// Update — PATCH /api/budgets/:id (FR-012/013, D25): loại (CATEGORY/TOTAL) bất biến;
// chạy lại đầy đủ xác thực như tạo + chống trùng (loại chính nó); conditional UPDATE
// theo expected_updated_at → phân biệt CONCURRENCY_CONFLICT (409, kèm data mới) vs
// RECORD_GONE (404). Ngang quyền — không gác theo người tạo (FR-012).
func (b *UpdateBudgetBiz) Update(ctx context.Context, householdID, id uuid.UUID, in UpdateBudgetInput) (*model.BudgetView, error) {
	if in.ExpectedUpdatedAt.IsZero() {
		return nil, common.NewBadRequest("thiếu expected_updated_at").WithField("expected_updated_at")
	}
	existing, err := b.store.FindByID(ctx, householdID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, common.NewNotFound("ngân sách đã bị xóa bởi thành viên khác")
		}
		return nil, common.NewInternal(err)
	}

	// PATCH một phần (contracts): trường không gửi → giữ giá trị hiện có. `limit_amount`
	// là bắt buộc (0 → LIMIT_INVALID, backstop cho #24). Loại (CATEGORY/TOTAL) bất biến.
	fields := in.BudgetFields
	fields.Type = existing.Type
	if fields.CategoryID == nil {
		fields.CategoryID = existing.CategoryID
	}
	if fields.PeriodType == "" {
		fields.PeriodType = existing.PeriodType
	}
	if fields.StartDate == nil {
		fields.StartDate = existing.StartDate
	}
	if fields.EndDate == nil {
		fields.EndDate = existing.EndDate
	}
	fields, err = validateBudgetFields(ctx, householdID, fields, b.catStore)
	if err != nil {
		return nil, err
	}
	if dup, derr := b.store.FindActiveDuplicate(ctx, householdID, fields.Type, fields.CategoryID, fields.PeriodType, &id); derr != nil {
		return nil, common.NewInternal(derr)
	} else if dup != nil {
		return nil, duplicateErr(dup.ID)
	}

	changes := map[string]any{
		"limit_amount": fields.LimitAmount,
		"period_type":  fields.PeriodType,
		"category_id":  fields.CategoryID,
		"start_date":   fields.StartDate,
		"end_date":     fields.EndDate,
	}
	rows, err := b.store.UpdateConditional(ctx, householdID, id, in.ExpectedUpdatedAt, changes)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	if rows == 0 {
		return nil, b.conflictOrGone(ctx, householdID, id)
	}
	return b.viewer.Get(ctx, householdID, id, time.Now())
}

// conflictOrGone — 0 hàng: đã xóa (404 RECORD_GONE) hay bị sửa (409, kèm bản mới).
func (b *UpdateBudgetBiz) conflictOrGone(ctx context.Context, householdID, id uuid.UUID) error {
	exists, err := b.store.Exists(ctx, householdID, id)
	if err != nil {
		return common.NewInternal(err)
	}
	if !exists {
		return common.NewNotFound("ngân sách đã bị xóa bởi thành viên khác")
	}
	conflict := common.NewConflict("ngân sách vừa được thành viên khác sửa — xem dữ liệu mới nhất")
	if fresh, ferr := b.viewer.Get(ctx, householdID, id, time.Now()); ferr == nil {
		conflict = conflict.WithExtra(map[string]any{"data": fresh})
	}
	return conflict
}
