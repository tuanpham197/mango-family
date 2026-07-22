package biz

import (
	"context"

	"household-finance/api/common"
	"household-finance/api/module/budget/model"

	"github.com/google/uuid"
)

// BudgetWriter — CRUD tối thiểu cho tạo ngân sách + chống trùng (D22).
type BudgetWriter interface {
	Create(ctx context.Context, b *model.Budget) error
	FindActiveDuplicate(ctx context.Context, householdID uuid.UUID, typ string, categoryID *uuid.UUID, periodType string, excludeID *uuid.UUID) (*model.Budget, error)
}

type CreateBudgetBiz struct {
	store    BudgetWriter
	catStore CategoryFinder
}

func NewCreateBudgetBiz(store BudgetWriter, catStore CategoryFinder) *CreateBudgetBiz {
	return &CreateBudgetBiz{store: store, catStore: catStore}
}

// Create — POST /api/budgets (FR-001…004, FR-010, FR-011): xác thực đầy đủ; chống
// trùng ngân sách ACTIVE cùng (danh mục, kỳ) / (tổng, kỳ) → BUDGET_DUPLICATE kèm id
// hiện có; created_by gán từ phiên (không nhận từ client).
func (b *CreateBudgetBiz) Create(ctx context.Context, householdID, userID uuid.UUID, in BudgetFields) (*model.Budget, error) {
	fields, err := validateBudgetFields(ctx, householdID, in, b.catStore)
	if err != nil {
		return nil, err
	}

	if dup, err := b.store.FindActiveDuplicate(ctx, householdID, fields.Type, fields.CategoryID, fields.PeriodType, nil); err != nil {
		return nil, common.NewInternal(err)
	} else if dup != nil {
		return nil, duplicateErr(dup.ID)
	}

	bud := &model.Budget{
		HouseholdID: householdID,
		Type:        fields.Type,
		CategoryID:  fields.CategoryID,
		LimitAmount: fields.LimitAmount,
		PeriodType:  fields.PeriodType,
		StartDate:   fields.StartDate,
		EndDate:     fields.EndDate,
		Status:      model.StatusActive,
		CreatedBy:   userID,
	}
	if err := b.store.Create(ctx, bud); err != nil {
		// Race với partial unique index (D22) → tra lại để trả BUDGET_DUPLICATE.
		if dup, derr := b.store.FindActiveDuplicate(ctx, householdID, fields.Type, fields.CategoryID, fields.PeriodType, nil); derr == nil && dup != nil {
			return nil, duplicateErr(dup.ID)
		}
		return nil, common.NewInternal(err)
	}
	return bud, nil
}

// duplicateErr — BUDGET_DUPLICATE kèm existing_budget_id để UI chỉ tới (D22).
func duplicateErr(existingID uuid.UUID) *common.AppError {
	return common.NewUnprocessable(common.ErrCodeBudgetDuplicate,
		"đã có ngân sách đang hoạt động cho lựa chọn này", "").
		WithExtra(map[string]any{"existing_budget_id": existingID})
}
