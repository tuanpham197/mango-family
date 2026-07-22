//go:build integration

// Integration test transport 005 (GET /api/reports/*) qua router THẬT.
package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func currentMonthRange() (from, to string) {
	now := time.Now()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	last := first.AddDate(0, 1, -1)
	return first.Format("2006-01-02"), last.Format("2006-01-02")
}

func TestReportsEndpoints(t *testing.T) {
	app := newTestApp(t)
	email, hid := app.seedUser(t, "Alice IT", true)
	accID, catID := app.seedAccountAndCategory(t, hid) // EXPENSE "Ăn uống" + "Tiền mặt"
	incomeCat := app.seedIncomeCategory(t, hid)
	from, to := currentMonthRange()

	// chưa đăng nhập → 401
	w := app.do(http.MethodGet, "/api/reports/overview?from="+from+"&to="+to, nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	cookies := app.login(t, email)

	// seed: chi 300.000 (Ăn uống) + thu 5.000.000 (Lương) trong tháng hiện tại
	require.Equal(t, http.StatusCreated, app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 300000, "type": "EXPENSE", "category_id": catID, "account_id": accID}, cookies).Code)
	require.Equal(t, http.StatusCreated, app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 5000000, "type": "INCOME", "category_id": incomeCat, "account_id": accID}, cookies).Code)

	// overview → 200, đủ field
	w = app.do(http.MethodGet, "/api/reports/overview?from="+from+"&to="+to, nil, cookies)
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	for _, k := range []string{"income", "expense", "net", "group_unit", "category_breakdown", "trend"} {
		assert.Contains(t, body, "\""+k+"\"")
	}
	data := decodeData(t, w.Body.Bytes())
	assert.Equal(t, float64(5000000), data["income"])
	assert.Equal(t, float64(300000), data["expense"])
	assert.Equal(t, float64(4700000), data["net"])
	require.Len(t, data["category_breakdown"].([]any), 1)

	// to < from → 400
	w = app.do(http.MethodGet, "/api/reports/overview?from="+to+"&to="+from, nil, cookies)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// thiếu param → 400
	w = app.do(http.MethodGet, "/api/reports/overview", nil, cookies)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// category detail → 200 với total + transactions
	w = app.do(http.MethodGet, fmt.Sprintf("/api/reports/category/%s?from=%s&to=%s", catID, from, to), nil, cookies)
	require.Equal(t, http.StatusOK, w.Code)
	cd := decodeData(t, w.Body.Bytes())
	assert.Equal(t, float64(300000), cd["total"])
	require.Len(t, cd["transactions"].([]any), 1)

	// category ngoài hộ → 404
	w = app.do(http.MethodGet, fmt.Sprintf("/api/reports/category/%s?from=%s&to=%s", uuid.New(), from, to), nil, cookies)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// cô lập hộ: Carol thấy rỗng
	carolEmail, _ := app.seedUser(t, "Carol IT", true)
	carolCookies := app.login(t, carolEmail)
	w = app.do(http.MethodGet, "/api/reports/overview?from="+from+"&to="+to, nil, carolCookies)
	require.Equal(t, http.StatusOK, w.Code)
	cdata := decodeData(t, w.Body.Bytes())
	assert.Equal(t, float64(0), cdata["income"])
	assert.Equal(t, float64(0), cdata["expense"])
	assert.Empty(t, cdata["category_breakdown"])
}
