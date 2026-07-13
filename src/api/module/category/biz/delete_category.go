package biz

import (
	"context"
	"errors"
	"slices"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/category/model"
	"household-finance/api/module/category/storage"

	"github.com/google/uuid"
)

// DeleteStore — thao tác cần cho luồng delete-reassign, chạy TRONG một DB
// transaction (D12): không bao giờ để giao dịch mồ côi (SC-007).
type DeleteStore interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.Category, error)
	ListChildren(ctx context.Context, householdID, parentID uuid.UUID) ([]model.Category, error)
	CountTransactionsByCategories(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) (int64, error)
	ReassignTransactions(ctx context.Context, householdID uuid.UUID, fromIDs []uuid.UUID, toID uuid.UUID) error
	DeleteTransactionsByCategories(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) error
	DeleteRulesByCategories(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) error
	DeleteByIDs(ctx context.Context, householdID uuid.UUID, ids []uuid.UUID) error
	DeleteConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time) (int64, error)
}

// TxRunner trừu tượng hóa DB transaction để biz test được bằng mock.
type TxRunner interface {
	InTx(ctx context.Context, fn func(s DeleteStore) error) error
}

type TxRunnerFunc func(ctx context.Context, fn func(s DeleteStore) error) error

func (f TxRunnerFunc) InTx(ctx context.Context, fn func(s DeleteStore) error) error {
	return f(ctx, fn)
}

const (
	DeleteModeReassign    = "reassign"
	DeleteModeDeleteAll   = "delete_transactions"
	deleteModeUnspecified = ""
)

type DeleteCategoryInput struct {
	Mode              string
	TargetCategoryID  *uuid.UUID
	ExpectedUpdatedAt time.Time
}

type DeleteCategoryBiz struct{ runner TxRunner }

func NewDeleteCategoryBiz(runner TxRunner) *DeleteCategoryBiz {
	return &DeleteCategoryBiz{runner: runner}
}

// Delete — DELETE /api/categories/:id (FR-007/008/009/012, D12):
// MỘT DB transaction: kiểm đích cùng loại/cùng hộ → xử lý con → gán lại hoặc
// xóa giao dịch → xóa danh mục (kèm mốc lạc quan expected_updated_at — D6).
func (b *DeleteCategoryBiz) Delete(ctx context.Context, householdID, id uuid.UUID, in DeleteCategoryInput) error {
	if in.ExpectedUpdatedAt.IsZero() {
		return common.NewBadRequest("thiếu expected_updated_at").WithField("expected_updated_at")
	}
	return b.runner.InTx(ctx, func(s DeleteStore) error {
		cat, err := s.FindByID(ctx, householdID, id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return common.NewNotFound("danh mục đã bị xóa trước đó")
			}
			return common.NewInternal(err)
		}

		children, err := s.ListChildren(ctx, householdID, id)
		if err != nil {
			return common.NewInternal(err)
		}
		affectedIDs := []uuid.UUID{id}
		childIDs := make([]uuid.UUID, 0, len(children))
		for _, ch := range children {
			affectedIDs = append(affectedIDs, ch.ID)
			childIDs = append(childIDs, ch.ID)
		}

		txnCount, err := s.CountTransactionsByCategories(ctx, householdID, affectedIDs)
		if err != nil {
			return common.NewInternal(err)
		}

		switch in.Mode {
		case deleteModeUnspecified:
			if txnCount > 0 {
				// Client phải hiện dialog chọn cách xử lý (FR-008).
				return common.NewUnprocessable(common.ErrCodeCategoryHasTxns,
					"danh mục (hoặc danh mục con) còn giao dịch — chọn gán lại hoặc xóa giao dịch", "mode").
					WithExtra(map[string]any{
						"transaction_count": txnCount,
						"children_count":    len(children),
					})
			}
		case DeleteModeReassign:
			if txnCount > 0 {
				if in.TargetCategoryID == nil {
					return common.NewBadRequest("thiếu target_category_id để gán lại").WithField("target_category_id")
				}
				if slices.Contains(affectedIDs, *in.TargetCategoryID) {
					return common.NewUnprocessable(common.ErrCodeReassignTypeMismatch,
						"không thể gán lại vào chính danh mục đang xóa hoặc danh mục con của nó", "target_category_id")
				}
				target, err := s.FindByID(ctx, householdID, *in.TargetCategoryID)
				if err != nil {
					// Khác hộ / không tồn tại → cùng mã lỗi, không lộ thông tin (FR-009, D5).
					return common.NewUnprocessable(common.ErrCodeReassignTypeMismatch,
						"danh mục đích không hợp lệ", "target_category_id")
				}
				if target.Type != cat.Type {
					return common.NewUnprocessable(common.ErrCodeReassignTypeMismatch,
						"danh mục đích phải cùng loại Thu/Chi", "target_category_id")
				}
				if err := s.ReassignTransactions(ctx, householdID, affectedIDs, target.ID); err != nil {
					return common.NewInternal(err)
				}
			}
		case DeleteModeDeleteAll:
			if err := s.DeleteTransactionsByCategories(ctx, householdID, affectedIDs); err != nil {
				return common.NewInternal(err)
			}
		default:
			return common.NewBadRequest("mode phải là reassign hoặc delete_transactions").WithField("mode")
		}

		if err := s.DeleteRulesByCategories(ctx, householdID, affectedIDs); err != nil {
			return common.NewInternal(err)
		}
		if err := s.DeleteByIDs(ctx, householdID, childIDs); err != nil {
			return common.NewInternal(err)
		}
		rows, err := s.DeleteConditional(ctx, householdID, id, in.ExpectedUpdatedAt)
		if err != nil {
			return common.NewInternal(err)
		}
		if rows == 0 {
			// Bản ghi còn (FindByID ở trên thành công trong cùng tx) nhưng mốc lệch
			// → đã bị thành viên khác sửa; rollback toàn bộ (quickstart #16).
			return common.NewConflict("danh mục vừa được thành viên khác sửa — tải lại rồi thử lại")
		}
		return nil
	})
}
