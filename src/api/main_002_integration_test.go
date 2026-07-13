//go:build integration

// Integration test transport 002 (accounts + transaction lifecycle) qua router THẬT.
package main

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedAccountAndCategory tạo 1 tài khoản CASH + 1 danh mục Chi cho hộ, trả về id.
func (a *testApp) seedAccountAndCategory(t *testing.T, householdID uuid.UUID) (accountID, categoryID uuid.UUID) {
	t.Helper()
	accountID, categoryID = uuid.New(), uuid.New()
	require.NoError(t, a.db.Exec(`INSERT INTO accounts (id, household_id, name, type) VALUES (?, ?, 'Tiền mặt', 'CASH')`,
		accountID, householdID).Error)
	require.NoError(t, a.db.Exec(`INSERT INTO categories (id, household_id, name, type) VALUES (?, ?, 'Ăn uống', 'EXPENSE')`,
		categoryID, householdID).Error)
	return
}

func decodeData(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var wrap struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &wrap))
	return wrap.Data
}

func TestAccountsEndpointScoped(t *testing.T) {
	app := newTestApp(t)
	email, hid := app.seedUser(t, "Alice IT", true)
	app.seedAccountAndCategory(t, hid)
	cookies := app.login(t, email)

	w := app.do(http.MethodGet, "/api/accounts", nil, cookies)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Tiền mặt")
	assert.Contains(t, w.Body.String(), "\"balance\":0")
}

func TestTransactionLifecycleAndErrors(t *testing.T) {
	app := newTestApp(t)
	email, hid := app.seedUser(t, "Alice IT", true)
	accID, catID := app.seedAccountAndCategory(t, hid)
	cookies := app.login(t, email)

	// thiếu tài khoản → ACCOUNT_REQUIRED (422)
	w := app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 50000, "type": "EXPENSE", "category_id": catID}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "ACCOUNT_REQUIRED")

	// ngày tương lai → FUTURE_DATE_NOT_ALLOWED (422)
	future := time.Now().Add(72 * time.Hour).Format(time.RFC3339)
	w = app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 1000, "type": "EXPENSE", "category_id": catID, "account_id": accID, "transaction_date": future}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "FUTURE_DATE_NOT_ALLOWED")

	// tạo hợp lệ → 201
	w = app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 50000, "type": "EXPENSE", "category_id": catID, "account_id": accID}, cookies)
	require.Equal(t, http.StatusCreated, w.Code)
	created := decodeData(t, w.Body.Bytes())
	txnID := created["id"].(string)
	updatedAt := created["updated_at"].(string)

	// số dư -50.000
	w = app.do(http.MethodGet, "/api/accounts", nil, cookies)
	assert.Contains(t, w.Body.String(), "-50000")

	// PATCH với mốc cũ sai → CONCURRENCY_CONFLICT (409)
	w = app.do(http.MethodPatch, "/api/transactions/"+txnID,
		map[string]any{"amount": 9, "type": "EXPENSE", "category_id": catID, "account_id": accID,
			"expected_updated_at": "2020-01-01T00:00:00Z"}, cookies)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "CONCURRENCY_CONFLICT")

	// PATCH đúng mốc → 200, số dư -80.000
	w = app.do(http.MethodPatch, "/api/transactions/"+txnID,
		map[string]any{"amount": 80000, "type": "EXPENSE", "category_id": catID, "account_id": accID,
			"expected_updated_at": updatedAt}, cookies)
	require.Equal(t, http.StatusOK, w.Code)
	w = app.do(http.MethodGet, "/api/accounts", nil, cookies)
	assert.Contains(t, w.Body.String(), "-80000")

	// DELETE → 204, số dư về 0
	w = app.do(http.MethodDelete, "/api/transactions/"+txnID, map[string]any{}, cookies)
	require.Equal(t, http.StatusNoContent, w.Code)
	w = app.do(http.MethodGet, "/api/accounts", nil, cookies)
	assert.Contains(t, w.Body.String(), "\"balance\":0")

	// DELETE lần nữa → RECORD_GONE (404)
	w = app.do(http.MethodDelete, "/api/transactions/"+txnID, map[string]any{}, cookies)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "RECORD_GONE")
}

func TestCrossHouseholdTransaction404(t *testing.T) {
	app := newTestApp(t)
	aliceEmail, aliceHid := app.seedUser(t, "Alice IT", true)
	accID, catID := app.seedAccountAndCategory(t, aliceHid)
	carolEmail, _ := app.seedUser(t, "Carol IT", true)

	aliceCookies := app.login(t, aliceEmail)
	w := app.do(http.MethodPost, "/api/transactions",
		map[string]any{"amount": 50000, "type": "EXPENSE", "category_id": catID, "account_id": accID}, aliceCookies)
	require.Equal(t, http.StatusCreated, w.Code)
	txnID := decodeData(t, w.Body.Bytes())["id"].(string)

	// Carol (hộ khác) đọc giao dịch hộ A → 404 (D5)
	carolCookies := app.login(t, carolEmail)
	w = app.do(http.MethodGet, "/api/transactions/"+txnID, nil, carolCookies)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
