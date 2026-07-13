package biz

import (
	"context"
	"strings"

	"household-finance/api/common"
	"household-finance/api/module/category/model"

	"github.com/google/uuid"
)

type CreateStore interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*model.Category, error)
	CountDuplicateName(ctx context.Context, householdID uuid.UUID, name, typ string, parentID, excludeID *uuid.UUID) (int64, error)
	Create(ctx context.Context, c *model.Category) error
}

type CreateCategoryInput struct {
	Name             string
	Type             string
	Icon             *string
	ParentID         *uuid.UUID
	ConfirmDuplicate bool
}

type CreateCategoryBiz struct{ store CreateStore }

func NewCreateCategoryBiz(store CreateStore) *CreateCategoryBiz {
	return &CreateCategoryBiz{store: store}
}

// Create — POST /api/categories (FR-003/004, US3: FR-010/011 danh mục con 1 cấp).
// Trùng tên trong (hộ, loại, cha) chỉ là CẢNH BÁO: client xác nhận bằng
// confirm_duplicate=true để vẫn tạo (FR-017).
func (b *CreateCategoryBiz) Create(ctx context.Context, householdID, userID uuid.UUID, in CreateCategoryInput) (*model.Category, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, common.NewBadRequest("tên danh mục là bắt buộc").WithField("name")
	}

	typ := in.Type
	if in.ParentID != nil {
		parent, err := b.store.FindByID(ctx, householdID, *in.ParentID)
		if err != nil {
			return nil, common.NewNotFound("danh mục cha không tồn tại")
		}
		if parent.ParentID != nil {
			// Chỉ một cấp con — cha phải là danh mục gốc (FR-010).
			return nil, common.NewUnprocessable(common.ErrCodeNestingTooDeep,
				"chỉ hỗ trợ một cấp danh mục con", "parent_id")
		}
		if in.Type != "" && in.Type != parent.Type {
			return nil, common.NewUnprocessable(common.ErrCodeParentTypeMismatch,
				"danh mục con phải cùng loại với danh mục cha", "type")
		}
		typ = parent.Type // con kế thừa loại cha (FR-011)
	}
	if typ != common.TypeIncome && typ != common.TypeExpense {
		// Loại là bắt buộc khi tạo danh mục gốc (FR-004, quickstart #2).
		return nil, common.NewBadRequest("vui lòng chọn loại Thu hoặc Chi").WithField("type")
	}

	dup, err := b.store.CountDuplicateName(ctx, householdID, name, typ, in.ParentID, nil)
	if err != nil {
		return nil, common.NewInternal(err)
	}
	if dup > 0 && !in.ConfirmDuplicate {
		return nil, common.NewUnprocessable(common.ErrCodeNameDuplicateWarning,
			"đã có danh mục cùng tên trong loại này — xác nhận nếu vẫn muốn tạo", "name")
	}

	cat := &model.Category{
		HouseholdID: householdID,
		Name:        name,
		Type:        typ,
		Icon:        in.Icon,
		ParentID:    in.ParentID,
		CreatedBy:   &userID,
	}
	if err := b.store.Create(ctx, cat); err != nil {
		return nil, common.NewInternal(err)
	}
	return cat, nil
}
