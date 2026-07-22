//go:build integration

// Integration test transport: filter tháng/năm cho GET /api/transactions.
package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func txnPagingTotal(t *testing.T, body []byte) int64 {
	t.Helper()
	var wrap struct {
		Paging struct {
			Total int64 `json:"total"`
		} `json:"paging"`
	}
	require.NoError(t, json.Unmarshal(body, &wrap))
	return wrap.Paging.Total
}

func TestTransactionListRangeFilter(t *testing.T) {
	app := newTestApp(t)
	email, hid := app.seedUser(t, "Alice IT", true)
	accID, catID := app.seedAccountAndCategory(t, hid)
	cookies := app.login(t, email)

	mk := func(amount float64, date string) {
		w := app.do(http.MethodPost, "/api/transactions",
			map[string]any{"amount": amount, "type": "EXPENSE", "category_id": catID, "account_id": accID,
				"transaction_date": date + "T12:00:00Z"}, cookies)
		require.Equal(t, http.StatusCreated, w.Code)
	}
	// Dùng ngày quá khứ (tránh chặn ngày tương lai).
	mk(100000, "2026-06-10")
	mk(200000, "2026-06-20")
	mk(50000, "2026-05-15")

	// Tháng 6 → 2
	w := app.do(http.MethodGet, "/api/transactions?from=2026-06-01&to=2026-06-30", nil, cookies)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(2), txnPagingTotal(t, w.Body.Bytes()))

	// Tháng 5 → 1
	w = app.do(http.MethodGet, "/api/transactions?from=2026-05-01&to=2026-05-31", nil, cookies)
	assert.Equal(t, int64(1), txnPagingTotal(t, w.Body.Bytes()))

	// Không filter → 3
	w = app.do(http.MethodGet, "/api/transactions", nil, cookies)
	assert.Equal(t, int64(3), txnPagingTotal(t, w.Body.Bytes()))

	// to < from → 400
	w = app.do(http.MethodGet, "/api/transactions?from=2026-06-30&to=2026-06-01", nil, cookies)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// from/to sai định dạng → 400
	w = app.do(http.MethodGet, "/api/transactions?from=nope&to=2026-06-01", nil, cookies)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
