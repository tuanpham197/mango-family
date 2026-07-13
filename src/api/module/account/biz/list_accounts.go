package biz

import (
	"context"

	"household-finance/api/common"
	"household-finance/api/module/account/model"

	"github.com/google/uuid"
)

type AccountLister interface {
	ListWithBalance(ctx context.Context, householdID uuid.UUID) ([]model.AccountWithBalance, error)
}

type ListAccountsBiz struct{ store AccountLister }

func NewListAccountsBiz(store AccountLister) *ListAccountsBiz {
	return &ListAccountsBiz{store: store}
}

// List — GET /api/accounts: tài khoản của hộ + số dư suy ra (FR-006/011, D14).
func (b *ListAccountsBiz) List(ctx context.Context, householdID uuid.UUID) ([]model.AccountWithBalance, error) {
	items, err := b.store.ListWithBalance(ctx, householdID)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	return items, nil
}
