package biz

import (
	"context"

	"household-finance/api/common"
	"household-finance/api/module/category/model"
	"household-finance/api/module/category/storage"

	"github.com/google/uuid"
)

type CategoryLister interface {
	List(ctx context.Context, householdID uuid.UUID, f storage.ListFilter) ([]model.Category, error)
}

type ListCategoriesBiz struct{ store CategoryLister }

func NewListCategoriesBiz(store CategoryLister) *ListCategoriesBiz {
	return &ListCategoriesBiz{store: store}
}

// List — GET /api/categories: lọc theo loại + ẩn, dựng cây cha/con (FR-002/014/020).
func (b *ListCategoriesBiz) List(ctx context.Context, householdID uuid.UUID, typeFilter string, includeHidden bool) ([]model.Tree, error) {
	if typeFilter != "" && typeFilter != common.TypeIncome && typeFilter != common.TypeExpense {
		return nil, common.NewBadRequest("type phải là INCOME hoặc EXPENSE").WithField("type")
	}
	cats, err := b.store.List(ctx, householdID, storage.ListFilter{Type: typeFilter, IncludeHidden: includeHidden})
	if err != nil {
		return nil, common.NewInternal(err)
	}
	return model.BuildTree(cats), nil
}
