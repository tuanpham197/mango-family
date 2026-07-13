package biz

import (
	"context"
	"errors"
	"testing"

	"household-finance/api/common"
	"household-finance/api/module/account/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLister struct {
	items []model.AccountWithBalance
	err   error
}

func (m *mockLister) ListWithBalance(_ context.Context, _ uuid.UUID) ([]model.AccountWithBalance, error) {
	return m.items, m.err
}

func TestListAccounts_ReturnsItems(t *testing.T) {
	biz := NewListAccountsBiz(&mockLister{items: []model.AccountWithBalance{
		{ID: uuid.New(), Name: "Tiền mặt", Type: "CASH", Balance: -50000},
	}})
	items, err := biz.List(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Tiền mặt", items[0].Name)
	assert.Equal(t, float64(-50000), items[0].Balance)
}

func TestListAccounts_WrapsStorageError(t *testing.T) {
	biz := NewListAccountsBiz(&mockLister{err: errors.New("db down")})
	_, err := biz.List(context.Background(), uuid.New())
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, common.ErrCodeInternal, appErr.Code)
}
