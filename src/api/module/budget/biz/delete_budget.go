package biz

import (
	"context"
	"time"

	"household-finance/api/common"

	"github.com/google/uuid"
)

// BudgetDeleter — xóa theo mốc lạc quan (D25); budget_alerts cascade theo FK (D23).
type BudgetDeleter interface {
	DeleteConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt *time.Time) (int64, error)
	Exists(ctx context.Context, householdID, id uuid.UUID) (bool, error)
}

type DeleteBudgetBiz struct{ store BudgetDeleter }

func NewDeleteBudgetBiz(store BudgetDeleter) *DeleteBudgetBiz {
	return &DeleteBudgetBiz{store: store}
}

// Delete — DELETE /api/budgets/:id (FR-012/013, UC-BGT-06): xóa theo id (+ mốc lạc
// quan nếu client gửi); cảnh báo liên quan xóa cascade. 0 hàng → CONCURRENCY_CONFLICT
// (nếu còn tồn tại và có mốc) hoặc RECORD_GONE. UI luôn xác nhận trước khi gọi.
func (b *DeleteBudgetBiz) Delete(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt *time.Time) error {
	rows, err := b.store.DeleteConditional(ctx, householdID, id, expectedUpdatedAt)
	if err != nil {
		return common.NewInternal(err)
	}
	if rows == 0 {
		if expectedUpdatedAt != nil {
			exists, err := b.store.Exists(ctx, householdID, id)
			if err != nil {
				return common.NewInternal(err)
			}
			if exists {
				return common.NewConflict("ngân sách vừa được thành viên khác sửa — tải lại rồi thử lại")
			}
		}
		return common.NewNotFound("ngân sách đã được xóa trước đó")
	}
	return nil
}
