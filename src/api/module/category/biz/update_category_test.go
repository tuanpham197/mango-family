package biz

import (
	"context"
	"testing"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/category/model"
	"household-finance/api/module/category/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateStore struct {
	rows       int64
	updated    map[string]any
	findErr    error
	findResult *model.Category
}

func (m *mockUpdateStore) FindByID(_ context.Context, _, _ uuid.UUID) (*model.Category, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.findResult != nil {
		return m.findResult, nil
	}
	return &model.Category{Name: "Hiện tại"}, nil
}

func (m *mockUpdateStore) UpdateConditional(_ context.Context, _, _ uuid.UUID, _ time.Time, changes map[string]any) (int64, error) {
	m.updated = changes
	return m.rows, nil
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func TestUpdateCategory_TypeImmutable(t *testing.T) {
	biz := NewUpdateCategoryBiz(&mockUpdateStore{})
	_, err := biz.Update(context.Background(), uuid.New(), uuid.New(),
		UpdateCategoryInput{Type: strPtr(common.TypeIncome), ExpectedUpdatedAt: time.Now()})
	assert.Equal(t, common.ErrCodeTypeImmutable, appErrCode(t, err)) // FR-005
}

func TestUpdateCategory_RequiresExpectedUpdatedAt(t *testing.T) {
	biz := NewUpdateCategoryBiz(&mockUpdateStore{})
	_, err := biz.Update(context.Background(), uuid.New(), uuid.New(),
		UpdateCategoryInput{Name: strPtr("Mới")})
	assert.Equal(t, common.ErrCodeInvalidRequest, appErrCode(t, err))
}

func TestUpdateCategory_ConflictVsGone(t *testing.T) {
	// 0 hàng + bản ghi còn → CONCURRENCY_CONFLICT (409)
	store := &mockUpdateStore{rows: 0}
	biz := NewUpdateCategoryBiz(store)
	_, err := biz.Update(context.Background(), uuid.New(), uuid.New(),
		UpdateCategoryInput{Name: strPtr("Mới"), ExpectedUpdatedAt: time.Now()})
	assert.Equal(t, common.ErrCodeConcurrencyConflict, appErrCode(t, err))

	// 0 hàng + bản ghi mất → RECORD_GONE (404)
	store = &mockUpdateStore{rows: 0, findErr: storage.ErrNotFound}
	biz = NewUpdateCategoryBiz(store)
	_, err = biz.Update(context.Background(), uuid.New(), uuid.New(),
		UpdateCategoryInput{Name: strPtr("Mới"), ExpectedUpdatedAt: time.Now()})
	assert.Equal(t, common.ErrCodeRecordGone, appErrCode(t, err))
}

func TestUpdateCategory_Success(t *testing.T) {
	store := &mockUpdateStore{rows: 1, findResult: &model.Category{Name: "Đã đổi"}}
	biz := NewUpdateCategoryBiz(store)
	cat, err := biz.Update(context.Background(), uuid.New(), uuid.New(),
		UpdateCategoryInput{Name: strPtr("  Đã đổi  "), IsHidden: boolPtr(true), ExpectedUpdatedAt: time.Now()})
	require.NoError(t, err)
	assert.Equal(t, "Đã đổi", store.updated["name"]) // trim
	assert.Equal(t, true, store.updated["is_hidden"])
	assert.Equal(t, "Đã đổi", cat.Name)
}
