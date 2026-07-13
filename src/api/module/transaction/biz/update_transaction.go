package biz

import (
	"context"
	"errors"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/transaction/model"
	"household-finance/api/module/transaction/storage"

	"github.com/google/uuid"
)

type TransactionRW interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.ListItem, error)
	UpdateConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time, changes map[string]any) (int64, error)
	Exists(ctx context.Context, householdID, id uuid.UUID) (bool, error)
}

type UpdateTransactionInput struct {
	TxFields
	ExpectedUpdatedAt time.Time
}

type UpdateTransactionBiz struct {
	catStore CategoryFinder
	accStore AccountFinder
	txStore  TransactionRW
}

func NewUpdateTransactionBiz(catStore CategoryFinder, accStore AccountFinder, txStore TransactionRW) *UpdateTransactionBiz {
	return &UpdateTransactionBiz{catStore: catStore, accStore: accStore, txStore: txStore}
}

// Update — PATCH /api/transactions/:id (FR-009/014): validate như tạo mới +
// mốc lạc quan expected_updated_at → conditional UPDATE; 0 hàng phân biệt
// CONCURRENCY_CONFLICT (409, kèm data mới) vs RECORD_GONE (404) (D17).
func (b *UpdateTransactionBiz) Update(ctx context.Context, householdID, id uuid.UUID, in UpdateTransactionInput) (*model.ListItem, error) {
	if in.ExpectedUpdatedAt.IsZero() {
		return nil, common.NewBadRequest("thiếu expected_updated_at").WithField("expected_updated_at")
	}
	when, err := validateFields(ctx, householdID, in.TxFields, b.catStore, b.accStore, time.Now())
	if err != nil {
		return nil, err
	}

	changes := map[string]any{
		"amount":           in.Amount,
		"type":             in.Type,
		"category_id":      *in.CategoryID,
		"account_id":       *in.AccountID,
		"description":      in.Description,
		"transaction_date": when,
	}
	rows, err := b.txStore.UpdateConditional(ctx, householdID, id, in.ExpectedUpdatedAt, changes)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	if rows == 0 {
		// Phân biệt đã-xóa (RECORD_GONE 404) vs đã-sửa (CONCURRENCY_CONFLICT 409, kèm data mới).
		exists, err := b.txStore.Exists(ctx, householdID, id)
		if err != nil {
			return nil, common.NewInternal(err)
		}
		if !exists {
			return nil, common.NewNotFound("giao dịch đã bị xóa bởi thành viên khác")
		}
		fresh, ferr := b.txStore.FindByID(ctx, householdID, id)
		conflict := common.NewConflict("giao dịch vừa được thành viên khác sửa — xem dữ liệu mới nhất")
		if ferr == nil {
			conflict = conflict.WithExtra(map[string]any{"data": fresh})
		}
		return nil, conflict
	}

	fresh, err := b.txStore.FindByID(ctx, householdID, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, common.NewNotFound("giao dịch không còn tồn tại")
		}
		return nil, common.NewInternal(err)
	}
	return fresh, nil
}
