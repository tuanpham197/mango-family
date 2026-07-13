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

// mockDeleteStore ghi lại các thao tác trong "transaction" để kiểm chứng luồng D12.
type mockDeleteStore struct {
	category     *model.Category
	children     []model.Category
	txnCount     int64
	deleteRows   int64
	targetLookup func(id uuid.UUID) (*model.Category, error)
	reassigned   *uuid.UUID
	txnsDeleted  bool
	rulesGone    bool
	catsDeleted  []uuid.UUID
}

func (m *mockDeleteStore) FindByID(_ context.Context, _, id uuid.UUID) (*model.Category, error) {
	if m.category == nil {
		return nil, storage.ErrNotFound
	}
	if m.category.ID == id {
		return m.category, nil
	}
	// tra cứu target: trả về category khác loại/cùng loại tùy test cấu hình qua targetLookup
	if m.targetLookup != nil {
		return m.targetLookup(id)
	}
	return nil, storage.ErrNotFound
}

func (m *mockDeleteStore) ListChildren(_ context.Context, _, _ uuid.UUID) ([]model.Category, error) {
	return m.children, nil
}

func (m *mockDeleteStore) CountTransactionsByCategories(_ context.Context, _ uuid.UUID, _ []uuid.UUID) (int64, error) {
	return m.txnCount, nil
}

func (m *mockDeleteStore) ReassignTransactions(_ context.Context, _ uuid.UUID, _ []uuid.UUID, toID uuid.UUID) error {
	m.reassigned = &toID
	return nil
}

func (m *mockDeleteStore) DeleteTransactionsByCategories(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	m.txnsDeleted = true
	return nil
}

func (m *mockDeleteStore) DeleteRulesByCategories(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	m.rulesGone = true
	return nil
}

func (m *mockDeleteStore) DeleteByIDs(_ context.Context, _ uuid.UUID, ids []uuid.UUID) error {
	m.catsDeleted = append(m.catsDeleted, ids...)
	return nil
}

func (m *mockDeleteStore) DeleteConditional(_ context.Context, _, id uuid.UUID, _ time.Time) (int64, error) {
	m.catsDeleted = append(m.catsDeleted, id)
	return m.deleteRows, nil
}

func runnerFor(store *mockDeleteStore) TxRunner {
	return TxRunnerFunc(func(ctx context.Context, fn func(DeleteStore) error) error {
		return fn(store)
	})
}

func newTestCategory(typ string) *model.Category {
	c := &model.Category{Type: typ}
	c.ID = uuid.New()
	return c
}

func TestDeleteCategory_RecordGone(t *testing.T) {
	store := &mockDeleteStore{category: nil}
	biz := NewDeleteCategoryBiz(runnerFor(store))
	err := biz.Delete(context.Background(), uuid.New(), uuid.New(),
		DeleteCategoryInput{ExpectedUpdatedAt: time.Now()})
	assert.Equal(t, common.ErrCodeRecordGone, appErrCode(t, err))
}

func TestDeleteCategory_HasTransactionsRequiresMode(t *testing.T) {
	cat := newTestCategory(common.TypeExpense)
	store := &mockDeleteStore{category: cat, txnCount: 3, children: []model.Category{{}, {}}}
	biz := NewDeleteCategoryBiz(runnerFor(store))
	err := biz.Delete(context.Background(), uuid.New(), cat.ID,
		DeleteCategoryInput{ExpectedUpdatedAt: time.Now()})

	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, common.ErrCodeCategoryHasTxns, appErr.Code) // FR-008
	assert.Equal(t, int64(3), appErr.Extra["transaction_count"])
	assert.Equal(t, 2, appErr.Extra["children_count"])
}

func TestDeleteCategory_ReassignToSelfOrChildRejected(t *testing.T) {
	cat := newTestCategory(common.TypeExpense)
	child := model.Category{}
	child.ID = uuid.New()
	store := &mockDeleteStore{category: cat, txnCount: 1, children: []model.Category{child}}
	biz := NewDeleteCategoryBiz(runnerFor(store))

	for _, target := range []uuid.UUID{cat.ID, child.ID} {
		targetID := target
		err := biz.Delete(context.Background(), uuid.New(), cat.ID,
			DeleteCategoryInput{Mode: DeleteModeReassign, TargetCategoryID: &targetID, ExpectedUpdatedAt: time.Now()})
		assert.Equal(t, common.ErrCodeReassignTypeMismatch, appErrCode(t, err)) // FR-009
	}
}

func TestDeleteCategory_ReassignTypeMismatch(t *testing.T) {
	cat := newTestCategory(common.TypeExpense)
	target := newTestCategory(common.TypeIncome)
	store := &mockDeleteStore{category: cat, txnCount: 1}
	store.targetLookup = func(id uuid.UUID) (*model.Category, error) {
		if id == target.ID {
			return target, nil
		}
		return nil, storage.ErrNotFound
	}
	biz := NewDeleteCategoryBiz(runnerFor(store))
	err := biz.Delete(context.Background(), uuid.New(), cat.ID,
		DeleteCategoryInput{Mode: DeleteModeReassign, TargetCategoryID: &target.ID, ExpectedUpdatedAt: time.Now()})
	assert.Equal(t, common.ErrCodeReassignTypeMismatch, appErrCode(t, err))
}

func TestDeleteCategory_ReassignSuccess(t *testing.T) {
	cat := newTestCategory(common.TypeExpense)
	target := newTestCategory(common.TypeExpense)
	store := &mockDeleteStore{category: cat, txnCount: 2, deleteRows: 1}
	store.targetLookup = func(id uuid.UUID) (*model.Category, error) {
		if id == target.ID {
			return target, nil
		}
		return nil, storage.ErrNotFound
	}
	biz := NewDeleteCategoryBiz(runnerFor(store))
	err := biz.Delete(context.Background(), uuid.New(), cat.ID,
		DeleteCategoryInput{Mode: DeleteModeReassign, TargetCategoryID: &target.ID, ExpectedUpdatedAt: time.Now()})
	require.NoError(t, err)
	require.NotNil(t, store.reassigned)
	assert.Equal(t, target.ID, *store.reassigned) // giao dịch chuyển hết → không mồ côi (SC-007)
	assert.True(t, store.rulesGone)
	assert.Contains(t, store.catsDeleted, cat.ID)
}

func TestDeleteCategory_DeleteTransactionsMode(t *testing.T) {
	cat := newTestCategory(common.TypeExpense)
	store := &mockDeleteStore{category: cat, txnCount: 2, deleteRows: 1}
	biz := NewDeleteCategoryBiz(runnerFor(store))
	err := biz.Delete(context.Background(), uuid.New(), cat.ID,
		DeleteCategoryInput{Mode: DeleteModeDeleteAll, ExpectedUpdatedAt: time.Now()})
	require.NoError(t, err)
	assert.True(t, store.txnsDeleted)
}

func TestDeleteCategory_NoTransactionsQuickDelete(t *testing.T) {
	cat := newTestCategory(common.TypeExpense)
	store := &mockDeleteStore{category: cat, txnCount: 0, deleteRows: 1}
	biz := NewDeleteCategoryBiz(runnerFor(store))
	err := biz.Delete(context.Background(), uuid.New(), cat.ID,
		DeleteCategoryInput{ExpectedUpdatedAt: time.Now()}) // không cần mode
	require.NoError(t, err)
	assert.Nil(t, store.reassigned)
	assert.False(t, store.txnsDeleted)
}

func TestDeleteCategory_StaleTimestampConflict(t *testing.T) {
	cat := newTestCategory(common.TypeExpense)
	store := &mockDeleteStore{category: cat, txnCount: 0, deleteRows: 0} // mốc lệch → 0 hàng
	biz := NewDeleteCategoryBiz(runnerFor(store))
	err := biz.Delete(context.Background(), uuid.New(), cat.ID,
		DeleteCategoryInput{ExpectedUpdatedAt: time.Now()})
	assert.Equal(t, common.ErrCodeConcurrencyConflict, appErrCode(t, err)) // quickstart #16
}
