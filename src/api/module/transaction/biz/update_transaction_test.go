package biz

import (
	"context"
	"testing"
	"time"

	"household-finance/api/common"
	"household-finance/api/module/transaction/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTxRW struct {
	rows    int64
	exists  bool
	current *model.ListItem
}

func (m *mockTxRW) FindByID(_ context.Context, _, id uuid.UUID) (*model.ListItem, error) {
	if m.current != nil {
		return m.current, nil
	}
	item := &model.ListItem{}
	item.ID = id
	item.Amount = 55000
	return item, nil
}

func (m *mockTxRW) UpdateConditional(_ context.Context, _, _ uuid.UUID, _ time.Time, _ map[string]any) (int64, error) {
	return m.rows, nil
}

func (m *mockTxRW) Exists(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return m.exists, nil
}

func updateInput(catID, accID uuid.UUID) UpdateTransactionInput {
	return UpdateTransactionInput{
		TxFields:          TxFields{Amount: 80000, Type: common.TypeExpense, CategoryID: &catID, AccountID: &accID},
		ExpectedUpdatedAt: time.Now(),
	}
}

func newUpdateBiz(rw *mockTxRW) *UpdateTransactionBiz {
	cat := expenseCat()
	return NewUpdateTransactionBiz(&mockCatFinder{cat: cat}, &mockAccFinder{found: true}, rw)
}

func TestUpdate_RequiresExpectedUpdatedAt(t *testing.T) {
	biz := newUpdateBiz(&mockTxRW{rows: 1})
	in := updateInput(uuid.New(), uuid.New())
	in.ExpectedUpdatedAt = time.Time{}
	_, err := biz.Update(context.Background(), uuid.New(), uuid.New(), in)
	assert.Equal(t, common.ErrCodeInvalidRequest, appErrCode(t, err))
}

func TestUpdate_Success(t *testing.T) {
	cat := expenseCat()
	rw := &mockTxRW{rows: 1}
	biz := NewUpdateTransactionBiz(&mockCatFinder{cat: cat}, &mockAccFinder{found: true}, rw)
	item, err := biz.Update(context.Background(), uuid.New(), uuid.New(), updateInput(cat.ID, uuid.New()))
	require.NoError(t, err)
	assert.NotNil(t, item)
}

func TestUpdate_ConflictCarriesFreshData(t *testing.T) {
	cat := expenseCat()
	fresh := &model.ListItem{}
	fresh.ID = uuid.New()
	fresh.Amount = 55000
	rw := &mockTxRW{rows: 0, exists: true, current: fresh}
	biz := NewUpdateTransactionBiz(&mockCatFinder{cat: cat}, &mockAccFinder{found: true}, rw)

	_, err := biz.Update(context.Background(), uuid.New(), fresh.ID, updateInput(cat.ID, uuid.New()))
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, common.ErrCodeConcurrencyConflict, appErr.Code) // D17
	assert.Contains(t, appErr.Extra, "data")                        // kèm bản mới nhất
}

func TestUpdate_RecordGoneWhenMissing(t *testing.T) {
	cat := expenseCat()
	rw := &mockTxRW{rows: 0, exists: false}
	biz := NewUpdateTransactionBiz(&mockCatFinder{cat: cat}, &mockAccFinder{found: true}, rw)
	_, err := biz.Update(context.Background(), uuid.New(), uuid.New(), updateInput(cat.ID, uuid.New()))
	assert.Equal(t, common.ErrCodeRecordGone, appErrCode(t, err)) // 404
}

func TestUpdate_RevalidatesFields(t *testing.T) {
	// future date bị chặn cả khi sửa (validate chạy lại — FR-009)
	cat := expenseCat()
	rw := &mockTxRW{rows: 1}
	biz := NewUpdateTransactionBiz(&mockCatFinder{cat: cat}, &mockAccFinder{found: true}, rw)
	in := updateInput(cat.ID, uuid.New())
	future := time.Now().Add(72 * time.Hour)
	in.TransactionDate = &future
	_, err := biz.Update(context.Background(), uuid.New(), uuid.New(), in)
	assert.Equal(t, common.ErrCodeFutureDateNotAllowed, appErrCode(t, err))
}

// --- delete biz ---

type mockTxDeleter struct {
	rows   int64
	exists bool
}

func (m *mockTxDeleter) DeleteConditional(_ context.Context, _, _ uuid.UUID, _ *time.Time) (int64, error) {
	return m.rows, nil
}
func (m *mockTxDeleter) Exists(_ context.Context, _, _ uuid.UUID) (bool, error) { return m.exists, nil }

func TestDelete_Success(t *testing.T) {
	biz := NewDeleteTransactionBiz(&mockTxDeleter{rows: 1})
	require.NoError(t, biz.Delete(context.Background(), uuid.New(), uuid.New(), nil))
}

func TestDelete_GoneWhenAlreadyDeleted(t *testing.T) {
	biz := NewDeleteTransactionBiz(&mockTxDeleter{rows: 0})
	err := biz.Delete(context.Background(), uuid.New(), uuid.New(), nil)
	assert.Equal(t, common.ErrCodeRecordGone, appErrCode(t, err))
}

func TestDelete_ConflictWhenStaleButExists(t *testing.T) {
	past := time.Unix(0, 0)
	biz := NewDeleteTransactionBiz(&mockTxDeleter{rows: 0, exists: true})
	err := biz.Delete(context.Background(), uuid.New(), uuid.New(), &past)
	assert.Equal(t, common.ErrCodeConcurrencyConflict, appErrCode(t, err))
}
