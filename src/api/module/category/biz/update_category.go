package biz

import (
	"context"
	"errors"
	"strings"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/category/model"
	"household-finance/api/module/category/storage"

	"github.com/google/uuid"
)

type UpdateStore interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.Category, error)
	UpdateConditional(ctx context.Context, householdID, id uuid.UUID, expectedUpdatedAt time.Time, changes map[string]any) (int64, error)
}

type UpdateCategoryInput struct {
	Name              *string
	Icon              *string
	IsHidden          *bool
	Type              *string // luôn bị từ chối — loại bất biến (FR-005)
	ExpectedUpdatedAt time.Time
}

type UpdateCategoryBiz struct{ store UpdateStore }

func NewUpdateCategoryBiz(store UpdateStore) *UpdateCategoryBiz {
	return &UpdateCategoryBiz{store: store}
}

// Update — PATCH /api/categories/:id: name/icon/is_hidden; mốc lạc quan
// expected_updated_at chống ghi đè thầm lặng (D6, quickstart #16).
func (b *UpdateCategoryBiz) Update(ctx context.Context, householdID, id uuid.UUID, in UpdateCategoryInput) (*model.Category, error) {
	if in.Type != nil {
		return nil, common.NewUnprocessable(common.ErrCodeTypeImmutable,
			"loại danh mục không thể thay đổi sau khi tạo — hãy tạo danh mục mới và gán lại giao dịch", "type")
	}
	if in.ExpectedUpdatedAt.IsZero() {
		return nil, common.NewBadRequest("thiếu expected_updated_at").WithField("expected_updated_at")
	}

	changes := map[string]any{}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, common.NewBadRequest("tên danh mục là bắt buộc").WithField("name")
		}
		changes["name"] = name
	}
	if in.Icon != nil {
		changes["icon"] = *in.Icon
	}
	if in.IsHidden != nil {
		changes["is_hidden"] = *in.IsHidden
	}
	if len(changes) == 0 {
		return nil, common.NewBadRequest("không có thay đổi nào để cập nhật")
	}

	rows, err := b.store.UpdateConditional(ctx, householdID, id, in.ExpectedUpdatedAt, changes)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	if rows == 0 {
		// Phân biệt: đã bị xóa (RECORD_GONE 404) vs đã bị sửa (CONCURRENCY_CONFLICT 409).
		if _, err := b.store.FindByID(ctx, householdID, id); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, common.NewNotFound("danh mục đã bị xóa bởi thành viên khác")
			}
			return nil, common.NewInternal(err)
		}
		return nil, common.NewConflict("danh mục vừa được thành viên khác sửa — tải lại rồi thử lại")
	}
	fresh, err := b.store.FindByID(ctx, householdID, id)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	return fresh, nil
}
