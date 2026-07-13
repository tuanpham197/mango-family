package biz

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"household-finance/api/common"
	categorymodel "household-finance/api/module/category/model"
	"household-finance/api/module/transaction/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCatFinder struct{ cat *categorymodel.Category }

func (m *mockCatFinder) FindByID(_ context.Context, _, _ uuid.UUID) (*categorymodel.Category, error) {
	if m.cat == nil {
		return nil, errors.New("not found")
	}
	return m.cat, nil
}

type mockTxCreator struct{ created *model.Transaction }

func (m *mockTxCreator) Create(_ context.Context, t *model.Transaction) error {
	m.created = t
	return nil
}

type mockLearner struct {
	keyword    string
	categoryID uuid.UUID
	err        error
	calls      int
}

func (m *mockLearner) UpsertRule(_ context.Context, _ uuid.UUID, keyword string, categoryID uuid.UUID) error {
	m.calls++
	m.keyword = keyword
	m.categoryID = categoryID
	return m.err
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

func validInput(categoryID uuid.UUID) CreateTransactionInput {
	return CreateTransactionInput{Amount: 45000, Type: common.TypeExpense, CategoryID: &categoryID}
}

func TestCreateTransaction_Validation(t *testing.T) {
	cat := expenseCat()
	biz := NewCreateTransactionBiz(&mockCatFinder{cat: cat}, &mockTxCreator{}, nil)
	ctx := context.Background()
	hid, uid := uuid.New(), uuid.New()

	// type sai
	in := validInput(cat.ID)
	in.Type = "WRONG"
	_, err := biz.Create(ctx, hid, uid, in)
	assert.Equal(t, common.ErrCodeInvalidRequest, appErrCode(t, err))

	// amount <= 0 (BR-TRK-001)
	in = validInput(cat.ID)
	in.Amount = 0
	_, err = biz.Create(ctx, hid, uid, in)
	assert.Equal(t, common.ErrCodeAmountInvalid, appErrCode(t, err))

	// mô tả > 255 ký tự (BR-TRK-004)
	in = validInput(cat.ID)
	long := strings.Repeat("á", 256)
	in.Description = &long
	_, err = biz.Create(ctx, hid, uid, in)
	assert.Equal(t, common.ErrCodeDescriptionTooLong, appErrCode(t, err))

	// thiếu danh mục (FR-013)
	in = validInput(cat.ID)
	in.CategoryID = nil
	_, err = biz.Create(ctx, hid, uid, in)
	assert.Equal(t, common.ErrCodeCategoryRequired, appErrCode(t, err))
}

func TestCreateTransaction_CategoryNotFoundIs404(t *testing.T) {
	biz := NewCreateTransactionBiz(&mockCatFinder{cat: nil}, &mockTxCreator{}, nil)
	catID := uuid.New()
	_, err := biz.Create(context.Background(), uuid.New(), uuid.New(), validInput(catID))
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.StatusCode) // ngoài hộ/không tồn tại → 404 (D5)
}

func TestCreateTransaction_CategoryTypeMismatch(t *testing.T) {
	cat := expenseCat()
	biz := NewCreateTransactionBiz(&mockCatFinder{cat: cat}, &mockTxCreator{}, nil)
	in := validInput(cat.ID)
	in.Type = common.TypeIncome // giao dịch Thu nhưng danh mục Chi
	_, err := biz.Create(context.Background(), uuid.New(), uuid.New(), in)
	assert.Equal(t, common.ErrCodeCategoryTypeMismatch, appErrCode(t, err)) // FR-014
}

func TestCreateTransaction_SuccessSetsAuthorshipAndTime(t *testing.T) {
	cat := expenseCat()
	store := &mockTxCreator{}
	biz := NewCreateTransactionBiz(&mockCatFinder{cat: cat}, store, nil)
	hid, uid := uuid.New(), uuid.New()

	before := time.Now()
	tx, err := biz.Create(context.Background(), hid, uid, validInput(cat.ID))
	require.NoError(t, err)
	assert.Equal(t, uid, tx.CreatedBy) // FR-022: authorship từ phiên
	assert.Equal(t, hid, tx.HouseholdID)
	assert.False(t, tx.TransactionDate.Before(before)) // mặc định now()
	assert.Equal(t, store.created, tx)
}

func TestCreateTransaction_LearnsNormalizedKeyword(t *testing.T) {
	cat := expenseCat()
	learner := &mockLearner{}
	biz := NewCreateTransactionBiz(&mockCatFinder{cat: cat}, &mockTxCreator{}, learner)
	desc := "  Grab   Đi Làm "
	in := validInput(cat.ID)
	in.Description = &desc
	_, err := biz.Create(context.Background(), uuid.New(), uuid.New(), in)
	require.NoError(t, err)
	assert.Equal(t, 1, learner.calls)
	assert.Equal(t, "grab đi làm", learner.keyword) // chuẩn hóa (D7)
	assert.Equal(t, cat.ID, learner.categoryID)
}

func TestCreateTransaction_LearnerErrorIsBestEffort(t *testing.T) {
	cat := expenseCat()
	learner := &mockLearner{err: errors.New("db down")}
	biz := NewCreateTransactionBiz(&mockCatFinder{cat: cat}, &mockTxCreator{}, learner)
	desc := "trà sữa"
	in := validInput(cat.ID)
	in.Description = &desc
	tx, err := biz.Create(context.Background(), uuid.New(), uuid.New(), in)
	require.NoError(t, err) // lỗi học KHÔNG làm hỏng giao dịch
	assert.NotNil(t, tx)
}
