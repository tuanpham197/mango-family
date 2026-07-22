//go:build integration

// Integration test transport 004 (GET /api/overview) qua router THẬT.
package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverviewEndpoint(t *testing.T) {
	app := newTestApp(t)
	email, hid := app.seedUser(t, "Alice IT", true)
	accID, catID := app.seedAccountAndCategory(t, hid) // EXPENSE "Ăn uống" + tài khoản "Tiền mặt"
	incomeCat := app.seedIncomeCategory(t, hid)        // (định nghĩa ở main_003_integration_test.go)

	// chưa đăng nhập → 401
	w := app.do(http.MethodGet, "/api/overview", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	cookies := app.login(t, email)

	// seed: chi 200.000 (Ăn uống) + thu 5.000.000 (Lương) trong tháng hiện tại
	require.Equal(t, http.StatusCreated, app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 200000, "type": "EXPENSE", "category_id": catID, "account_id": accID}, cookies).Code)
	require.Equal(t, http.StatusCreated, app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 5000000, "type": "INCOME", "category_id": incomeCat, "account_id": accID}, cookies).Code)

	// GET /api/overview → 200, đủ field theo contract
	w = app.do(http.MethodGet, "/api/overview", nil, cookies)
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	for _, key := range []string{"net_worth", "net_worth_change_percent", "month", "category_spending", "recent_transactions"} {
		assert.Contains(t, body, "\""+key+"\"")
	}
	data := decodeData(t, w.Body.Bytes())
	month := data["month"].(map[string]any)
	assert.Equal(t, float64(5000000), month["income"])
	assert.Equal(t, float64(200000), month["expense"])
	assert.Equal(t, float64(4800000), month["net"])
	cats := data["category_spending"].([]any)
	require.Len(t, cats, 1) // chỉ danh mục có chi
	assert.Equal(t, float64(200000), cats[0].(map[string]any)["amount"])

	// cô lập hộ: Carol (hộ khác) thấy tổng quan rỗng
	carolEmail, _ := app.seedUser(t, "Carol IT", true)
	carolCookies := app.login(t, carolEmail)
	w = app.do(http.MethodGet, "/api/overview", nil, carolCookies)
	require.Equal(t, http.StatusOK, w.Code)
	cdata := decodeData(t, w.Body.Bytes())
	cmonth := cdata["month"].(map[string]any)
	assert.Equal(t, float64(0), cmonth["income"])
	assert.Equal(t, float64(0), cmonth["expense"])
	assert.Empty(t, cdata["category_spending"])
}
