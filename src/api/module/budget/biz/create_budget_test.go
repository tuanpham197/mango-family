package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/budget/model"
	categorymodel "household-finance/api/module/category/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCatFinder struct {
	cat *categorymodel.Category
	err error
}

func (m *mockCatFinder) FindByID(_ context.Context, _, _ uuid.UUID) (*categorymodel.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.cat == nil {
		return nil, errors.New("not found")
	}
	return m.cat, nil
}

type mockBudgetWriter struct {
	created   *model.Budget
	dup       *model.Budget
	createErr error
}

func (m *mockBudgetWriter) Create(_ context.Context, b *model.Budget) error {
	if m.createErr != nil {
		return m.createErr
	}
	b.ID = uuid.New()
	m.created = b
	return nil
}

func (m *mockBudgetWriter) FindActiveDuplicate(_ context.Context, _ uuid.UUID, _ string, _ *uuid.UUID, _ string, _ *uuid.UUID) (*model.Budget, error) {
	return m.dup, nil
}

func appErrCode(t *testing.T, err error) string {
	t.Helper()
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

func expenseCat() *categorymodel.Category {
	c := &categorymodel.Category{Type: common.TypeExpense}
	c.ID = uuid.New()
	return c
}

func categoryInput(catID uuid.UUID) BudgetFields {
	return BudgetFields{Type: model.TypeCategory, CategoryID: &catID, LimitAmount: 5000000, PeriodType: model.PeriodMonthly}
}

func newCreate(cat *categorymodel.Category, catErr error, w *mockBudgetWriter) *CreateBudgetBiz {
	return NewCreateBudgetBiz(w, &mockCatFinder{cat: cat, err: catErr})
}

func TestCreateBudget_LimitInvalid(t *testing.T) {
	cat := expenseCat()
	in := categoryInput(cat.ID)
	in.LimitAmount = 0
	_, err := newCreate(cat, nil, &mockBudgetWriter{}).Create(context.Background(), uuid.New(), uuid.New(), in)
	assert.Equal(t, common.ErrCodeLimitInvalid, appErrCode(t, err))
}

func TestCreateBudget_CategoryRequired(t *testing.T) {
	in := categoryInput(uuid.New())
	in.CategoryID = nil
	_, err := newCreate(expenseCat(), nil, &mockBudgetWriter{}).Create(context.Background(), uuid.New(), uuid.New(), in)
	assert.Equal(t, common.ErrCodeCategoryRequired, appErrCode(t, err))
}

func TestCreateBudget_CategoryTypeMismatch(t *testing.T) {
	income := &categorymodel.Category{Type: common.TypeIncome}
	income.ID = uuid.New()
	_, err := newCreate(income, nil, &mockBudgetWriter{}).Create(context.Background(), uuid.New(), uuid.New(), categoryInput(income.ID))
	assert.Equal(t, common.ErrCodeCategoryTypeMismatch, appErrCode(t, err))
}

func TestCreateBudget_CategoryHouseholdMismatch(t *testing.T) {
	// Danh mục ngoài hộ → finder lỗi → CATEGORY_HOUSEHOLD_MISMATCH (không lộ tồn tại).
	_, err := newCreate(nil, errors.New("not found"), &mockBudgetWriter{}).Create(context.Background(), uuid.New(), uuid.New(), categoryInput(uuid.New()))
	assert.Equal(t, common.ErrCodeCategoryHouseholdMismatch, appErrCode(t, err))
}

func TestCreateBudget_CategoryNotAllowedForTotal(t *testing.T) {
	catID := uuid.New()
	in := BudgetFields{Type: model.TypeTotal, CategoryID: &catID, LimitAmount: 20000000, PeriodType: model.PeriodMonthly}
	_, err := newCreate(expenseCat(), nil, &mockBudgetWriter{}).Create(context.Background(), uuid.New(), uuid.New(), in)
	assert.Equal(t, common.ErrCodeCategoryNotAllowedTotal, appErrCode(t, err))
}

func TestCreateBudget_PeriodInvalid(t *testing.T) {
	start := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	in := BudgetFields{Type: model.TypeTotal, LimitAmount: 1000, PeriodType: model.PeriodOneTime, StartDate: &start, EndDate: &end}
	_, err := newCreate(expenseCat(), nil, &mockBudgetWriter{}).Create(context.Background(), uuid.New(), uuid.New(), in)
	assert.Equal(t, common.ErrCodePeriodInvalid, appErrCode(t, err))

	// Thiếu ngày cũng PERIOD_INVALID.
	in2 := BudgetFields{Type: model.TypeTotal, LimitAmount: 1000, PeriodType: model.PeriodOneTime}
	_, err = newCreate(expenseCat(), nil, &mockBudgetWriter{}).Create(context.Background(), uuid.New(), uuid.New(), in2)
	assert.Equal(t, common.ErrCodePeriodInvalid, appErrCode(t, err))
}

func TestCreateBudget_Duplicate(t *testing.T) {
	cat := expenseCat()
	dup := &model.Budget{}
	dup.ID = uuid.New()
	_, err := newCreate(cat, nil, &mockBudgetWriter{dup: dup}).Create(context.Background(), uuid.New(), uuid.New(), categoryInput(cat.ID))
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, common.ErrCodeBudgetDuplicate, appErr.Code)
	assert.Equal(t, dup.ID, appErr.Extra["existing_budget_id"]) // chỉ tới ngân sách hiện có (D22)
}

func TestCreateBudget_SuccessSetsAuthorshipAndClearsDatesForMonthly(t *testing.T) {
	cat := expenseCat()
	w := &mockBudgetWriter{}
	hid, uid := uuid.New(), uuid.New()
	in := categoryInput(cat.ID)
	start := time.Now()
	in.StartDate = &start // MONTHLY → phải bị xóa
	bud, err := newCreate(cat, nil, w).Create(context.Background(), hid, uid, in)
	require.NoError(t, err)
	assert.Equal(t, uid, bud.CreatedBy) // created_by từ phiên (FR-011)
	assert.Equal(t, hid, bud.HouseholdID)
	assert.Equal(t, model.StatusActive, bud.Status)
	assert.Nil(t, bud.StartDate) // MONTHLY không giữ ngày
}
