package biz

import (
	"context"
	"time"

	"household-finance/api/common"

	"github.com/google/uuid"
)

type TransactionDeleter interface {
	DeleteConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt *time.Time) (int64, error)
	Exists(ctx context.Context, householdID, id uuid.UUID) (bool, error)
}

type DeleteTransactionBiz struct{ store TransactionDeleter }

func NewDeleteTransactionBiz(store TransactionDeleter) *DeleteTransactionBiz {
	return &DeleteTransactionBiz{store: store}
}

// Delete — DELETE /api/transactions/:id (FR-010/014): xóa theo id (+ mốc lạc quan
// nếu client gửi); 0 hàng → RECORD_GONE (đã bị xóa trước — thông báo nhẹ, D17).
// UI luôn hiện dialog xác nhận trước khi gọi (SC-005).
func (b *DeleteTransactionBiz) Delete(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt *time.Time) error {
	rows, err := b.store.DeleteConditional(ctx, householdID, id, expectedUpdatedAt)
	if err != nil {
		return common.NewInternal(err)
	}
	if rows == 0 {
		// Nếu có mốc và bản ghi vẫn còn → xung đột; ngược lại đã bị xóa trước.
		if expectedUpdatedAt != nil {
			exists, err := b.store.Exists(ctx, householdID, id)
			if err != nil {
				return common.NewInternal(err)
			}
			if exists {
				return common.NewConflict("giao dịch vừa được thành viên khác sửa — tải lại rồi thử lại")
			}
		}
		return common.NewNotFound("giao dịch đã được xóa trước đó")
	}
	return nil
}
