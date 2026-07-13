package biz

import (
	"context"
	"errors"
	"testing"

	"household-finance/api/common"
	"household-finance/api/module/category/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCreateStore struct {
	findByID  func(householdID, id uuid.UUID) (*model.Category, error)
	dupCount  int64
	created   *model.Category
	createErr error
}

func (m *mockCreateStore) FindByID(_ context.Context, householdID, id uuid.UUID) (*model.Category, error) {
	if m.findByID != nil {
		return m.findByID(householdID, id)
	}
	return nil, errors.New("not found")
}

func (m *mockCreateStore) CountDuplicateName(_ context.Context, _ uuid.UUID, _, _ string, _, _ *uuid.UUID) (int64, error) {
	return m.dupCount, nil
}

func (m *mockCreateStore) Create(_ context.Context, c *model.Category) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = c
	return nil
}

func appErrCode(t *testing.T, err error) string {
	t.Helper()
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

func TestCreateCategory_NameRequired(t *testing.T) {
	biz := NewCreateCategoryBiz(&mockCreateStore{})
	_, err := biz.Create(context.Background(), uuid.New(), uuid.New(), CreateCategoryInput{Name: "  ", Type: common.TypeExpense})
	assert.Equal(t, common.ErrCodeInvalidRequest, appErrCode(t, err))
}

func TestCreateCategory_TypeRequired(t *testing.T) {
	biz := NewCreateCategoryBiz(&mockCreateStore{})
	_, err := biz.Create(context.Background(), uuid.New(), uuid.New(), CreateCategoryInput{Name: "Thú cưng"})
	assert.Equal(t, common.ErrCodeInvalidRequest, appErrCode(t, err))

	_, err = biz.Create(context.Background(), uuid.New(), uuid.New(), CreateCategoryInput{Name: "Thú cưng", Type: "WRONG"})
	assert.Equal(t, common.ErrCodeInvalidRequest, appErrCode(t, err))
}

func TestCreateCategory_NestingTooDeep(t *testing.T) {
	grandParent := uuid.New()
	store := &mockCreateStore{findByID: func(_, _ uuid.UUID) (*model.Category, error) {
		return &model.Category{Type: common.TypeExpense, ParentID: &grandParent}, nil // cha đã là con
	}}
	biz := NewCreateCategoryBiz(store)
	parentID := uuid.New()
	_, err := biz.Create(context.Background(), uuid.New(), uuid.New(), CreateCategoryInput{Name: "Cấp 3", ParentID: &parentID})
	assert.Equal(t, common.ErrCodeNestingTooDeep, appErrCode(t, err))
}

func TestCreateCategory_ParentTypeMismatch(t *testing.T) {
	store := &mockCreateStore{findByID: func(_, _ uuid.UUID) (*model.Category, error) {
		return &model.Category{Type: common.TypeExpense}, nil
	}}
	biz := NewCreateCategoryBiz(store)
	parentID := uuid.New()
	_, err := biz.Create(context.Background(), uuid.New(), uuid.New(),
		CreateCategoryInput{Name: "Lệch", Type: common.TypeIncome, ParentID: &parentID})
	assert.Equal(t, common.ErrCodeParentTypeMismatch, appErrCode(t, err))
}

func TestCreateCategory_ChildInheritsParentType(t *testing.T) {
	store := &mockCreateStore{findByID: func(_, _ uuid.UUID) (*model.Category, error) {
		return &model.Category{Type: common.TypeExpense}, nil
	}}
	biz := NewCreateCategoryBiz(store)
	parentID := uuid.New()
	cat, err := biz.Create(context.Background(), uuid.New(), uuid.New(),
		CreateCategoryInput{Name: "Ăn ngoài", ParentID: &parentID}) // không gửi type
	require.NoError(t, err)
	assert.Equal(t, common.TypeExpense, cat.Type) // FR-011
}

func TestCreateCategory_DuplicateWarningThenConfirm(t *testing.T) {
	store := &mockCreateStore{dupCount: 1}
	biz := NewCreateCategoryBiz(store)
	hid, uid := uuid.New(), uuid.New()

	_, err := biz.Create(context.Background(), hid, uid, CreateCategoryInput{Name: "Ăn uống", Type: common.TypeExpense})
	assert.Equal(t, common.ErrCodeNameDuplicateWarning, appErrCode(t, err)) // FR-017 cảnh báo

	cat, err := biz.Create(context.Background(), hid, uid,
		CreateCategoryInput{Name: "Ăn uống", Type: common.TypeExpense, ConfirmDuplicate: true})
	require.NoError(t, err) // vẫn cho tạo khi xác nhận
	assert.Equal(t, "Ăn uống", cat.Name)
	assert.Equal(t, &uid, cat.CreatedBy)
}
