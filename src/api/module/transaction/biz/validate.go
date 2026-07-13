package biz

import (
	"context"
	"time"

	"household-finance/api/common"
	accountmodel "household-finance/api/module/account/model"
	categorymodel "household-finance/api/module/category/model"

	"github.com/google/uuid"
)

// futureSlack — nới lệch múi giờ khi chặn ngày tương lai (D15).
const futureSlack = 24 * time.Hour

type CategoryFinder interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*categorymodel.Category, error)
}

type AccountFinder interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*accountmodel.Account, error)
}

// TxFields — các trường chung khi tạo/sửa giao dịch.
type TxFields struct {
	Amount          float64
	Type            string
	CategoryID      *uuid.UUID
	AccountID       *uuid.UUID
	Description     *string
	TransactionDate *time.Time
}

// validateFields chạy đủ xác thực dùng chung cho POST và PATCH (FR-001…007, D15/D18).
// Trả về thời điểm giao dịch đã chuẩn hóa (mặc định now nếu thiếu).
func validateFields(ctx context.Context, householdID uuid.UUID, in TxFields, catStore CategoryFinder, accStore AccountFinder, now time.Time) (time.Time, error) {
	if in.Type != common.TypeIncome && in.Type != common.TypeExpense {
		return time.Time{}, common.NewBadRequest("type phải là INCOME hoặc EXPENSE").WithField("type")
	}
	if in.Amount <= 0 {
		return time.Time{}, common.NewUnprocessable(common.ErrCodeAmountInvalid, "số tiền phải lớn hơn 0", "amount")
	}
	if in.Description != nil && len([]rune(*in.Description)) > common.MaxDescriptionLen {
		return time.Time{}, common.NewUnprocessable(common.ErrCodeDescriptionTooLong, "mô tả tối đa 255 ký tự", "description")
	}
	if in.CategoryID == nil {
		return time.Time{}, common.NewUnprocessable(common.ErrCodeCategoryRequired, "vui lòng chọn danh mục", "category_id")
	}
	if in.AccountID == nil {
		return time.Time{}, common.NewUnprocessable(common.ErrCodeAccountRequired, "vui lòng chọn tài khoản", "account_id")
	}

	// Danh mục: cùng hộ (ngoài hộ → 404, không lộ) + cùng loại (FR-003, D18).
	cat, err := catStore.FindByID(ctx, householdID, *in.CategoryID)
	if err != nil {
		return time.Time{}, common.NewNotFound("danh mục không tồn tại")
	}
	if cat.Type != in.Type {
		return time.Time{}, common.NewUnprocessable(common.ErrCodeCategoryTypeMismatch,
			"danh mục phải cùng loại với giao dịch", "category_id")
	}

	// Tài khoản: cùng hộ (FR-006/007).
	if _, err := accStore.FindByID(ctx, householdID, *in.AccountID); err != nil {
		return time.Time{}, common.NewUnprocessable(common.ErrCodeAccountHouseholdMismatch,
			"tài khoản không thuộc hộ của bạn", "account_id")
	}

	// Ngày giao dịch: mặc định now; chặn tương lai (nới +1 ngày múi giờ — D15).
	when := now
	if in.TransactionDate != nil {
		when = *in.TransactionDate
		if when.After(now.Add(futureSlack)) {
			return time.Time{}, common.NewUnprocessable(common.ErrCodeFutureDateNotAllowed,
				"không thể ghi giao dịch ở ngày tương lai", "transaction_date")
		}
	}
	return when, nil
}
