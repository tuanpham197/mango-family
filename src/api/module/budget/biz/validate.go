package biz

import (
	"context"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/budget/model"
	categorymodel "household-finance/api/module/category/model"

	"github.com/google/uuid"
)

// CategoryFinder — đọc danh mục để xác thực ngân sách danh mục (chỉ đọc — D21).
type CategoryFinder interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*categorymodel.Category, error)
}

// BudgetFields — trường chung khi tạo/sửa ngân sách (validate dùng chung — FR-012).
type BudgetFields struct {
	Type        string
	CategoryID  *uuid.UUID
	LimitAmount float64
	PeriodType  string
	StartDate   *time.Time
	EndDate     *time.Time
}

// validateBudgetFields — bộ xác thực dùng chung cho POST và PATCH (FR-001…004, D21/D22).
// Trả về các trường đã chuẩn hóa (start/end xóa khi không phải ONE_TIME).
func validateBudgetFields(ctx context.Context, householdID uuid.UUID, in BudgetFields, catStore CategoryFinder) (BudgetFields, error) {
	if in.Type != model.TypeCategory && in.Type != model.TypeTotal {
		return in, common.NewBadRequest("type phải là CATEGORY hoặc TOTAL").WithField("type")
	}
	if in.LimitAmount <= 0 {
		return in, common.NewUnprocessable(common.ErrCodeLimitInvalid, "giới hạn phải lớn hơn 0", "limit_amount")
	}
	if in.PeriodType != model.PeriodMonthly && in.PeriodType != model.PeriodWeekly && in.PeriodType != model.PeriodOneTime {
		return in, common.NewBadRequest("kỳ áp dụng không hợp lệ").WithField("period_type")
	}

	switch in.Type {
	case model.TypeCategory:
		if in.CategoryID == nil {
			return in, common.NewUnprocessable(common.ErrCodeCategoryRequired, "vui lòng chọn danh mục Chi", "category_id")
		}
		cat, err := catStore.FindByID(ctx, householdID, *in.CategoryID)
		if err != nil {
			// Ngoài hộ / không tồn tại → không lộ, coi như khác hộ (FR-001, FR-011).
			return in, common.NewUnprocessable(common.ErrCodeCategoryHouseholdMismatch,
				"danh mục không thuộc hộ của bạn", "category_id")
		}
		if cat.Type != common.TypeExpense {
			return in, common.NewUnprocessable(common.ErrCodeCategoryTypeMismatch,
				"ngân sách chỉ áp dụng cho danh mục Chi", "category_id")
		}
	case model.TypeTotal:
		if in.CategoryID != nil {
			return in, common.NewUnprocessable(common.ErrCodeCategoryNotAllowedTotal,
				"ngân sách tổng không gắn danh mục", "category_id")
		}
	}

	// Kỳ áp dụng: ONE_TIME cần đủ ngày, end ≥ start; MONTHLY/WEEKLY bỏ ngày (tự lặp).
	if in.PeriodType == model.PeriodOneTime {
		if in.StartDate == nil || in.EndDate == nil {
			return in, common.NewUnprocessable(common.ErrCodePeriodInvalid,
				"kỳ một lần cần ngày bắt đầu và kết thúc", "end_date")
		}
		if in.EndDate.Before(*in.StartDate) {
			return in, common.NewUnprocessable(common.ErrCodePeriodInvalid,
				"ngày kết thúc không được trước ngày bắt đầu", "end_date")
		}
	} else {
		in.StartDate, in.EndDate = nil, nil
	}
	return in, nil
}
