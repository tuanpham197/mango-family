//go:build integration

// Integration test transport 003 (budget lifecycle + mã lỗi contract) qua router THẬT.
package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (a *testApp) seedIncomeCategory(t *testing.T, householdID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, a.db.Exec(`INSERT INTO categories (id, household_id, name, type) VALUES (?, ?, 'Lương', 'INCOME')`,
		id, householdID).Error)
	return id
}

func TestBudgetLifecycleAndErrorCodes(t *testing.T) {
	app := newTestApp(t)
	email, hid := app.seedUser(t, "Alice IT", true)
	_, catID := app.seedAccountAndCategory(t, hid) // catID = EXPENSE "Ăn uống"
	incomeCat := app.seedIncomeCategory(t, hid)
	cookies := app.login(t, email)

	// giới hạn ≤ 0 → LIMIT_INVALID (422)
	w := app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "CATEGORY", "category_id": catID, "limit_amount": 0, "period_type": "MONTHLY"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "LIMIT_INVALID")

	// thiếu danh mục (CATEGORY) → CATEGORY_REQUIRED
	w = app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "CATEGORY", "limit_amount": 1000, "period_type": "MONTHLY"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_REQUIRED")

	// danh mục THU → CATEGORY_TYPE_MISMATCH
	w = app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "CATEGORY", "category_id": incomeCat, "limit_amount": 1000, "period_type": "MONTHLY"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_TYPE_MISMATCH")

	// danh mục ngoài hộ → CATEGORY_HOUSEHOLD_MISMATCH
	w = app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "CATEGORY", "category_id": uuid.New(), "limit_amount": 1000, "period_type": "MONTHLY"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_HOUSEHOLD_MISMATCH")

	// TOTAL kèm danh mục → CATEGORY_NOT_ALLOWED_FOR_TOTAL
	w = app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "TOTAL", "category_id": catID, "limit_amount": 1000, "period_type": "MONTHLY"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_NOT_ALLOWED_FOR_TOTAL")

	// ONE_TIME end < start → PERIOD_INVALID
	w = app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "TOTAL", "limit_amount": 1000, "period_type": "ONE_TIME",
			"start_date": "2026-07-10T00:00:00Z", "end_date": "2026-07-05T00:00:00Z"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "PERIOD_INVALID")

	// tạo hợp lệ → 201, embed spent/percent/period_key/alerts
	w = app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "CATEGORY", "category_id": catID, "limit_amount": 5000000, "period_type": "MONTHLY"}, cookies)
	require.Equal(t, http.StatusCreated, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "\"spent\"")
	assert.Contains(t, body, "\"percent\"")
	assert.Contains(t, body, "\"period_key\"")
	created := decodeData(t, w.Body.Bytes())
	bid := created["id"].(string)
	updatedAt := created["updated_at"].(string)

	// trùng danh mục/kỳ → BUDGET_DUPLICATE + existing_budget_id
	w = app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "CATEGORY", "category_id": catID, "limit_amount": 9000000, "period_type": "MONTHLY"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "BUDGET_DUPLICATE")
	assert.Contains(t, w.Body.String(), bid) // existing_budget_id chỉ tới ngân sách hiện có

	// PATCH mốc sai → 409 CONCURRENCY_CONFLICT
	w = app.do(http.MethodPatch, "/api/budgets/"+bid,
		map[string]any{"limit_amount": 4000000, "period_type": "MONTHLY",
			"expected_updated_at": "2020-01-01T00:00:00Z"}, cookies)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "CONCURRENCY_CONFLICT")

	// PATCH đúng mốc → 200
	w = app.do(http.MethodPatch, "/api/budgets/"+bid,
		map[string]any{"limit_amount": 4000000, "period_type": "MONTHLY", "expected_updated_at": updatedAt}, cookies)
	require.Equal(t, http.StatusOK, w.Code)

	// DELETE → 204 rồi 404
	w = app.do(http.MethodDelete, "/api/budgets/"+bid, map[string]any{}, cookies)
	require.Equal(t, http.StatusNoContent, w.Code)
	w = app.do(http.MethodDelete, "/api/budgets/"+bid, map[string]any{}, cookies)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "RECORD_GONE")
}

func TestBudgetCrossHousehold404(t *testing.T) {
	app := newTestApp(t)
	aliceEmail, aliceHid := app.seedUser(t, "Alice IT", true)
	_, catID := app.seedAccountAndCategory(t, aliceHid)
	carolEmail, _ := app.seedUser(t, "Carol IT", true)

	aliceCookies := app.login(t, aliceEmail)
	w := app.do(http.MethodPost, "/api/budgets",
		map[string]any{"type": "CATEGORY", "category_id": catID, "limit_amount": 1000000, "period_type": "MONTHLY"}, aliceCookies)
	require.Equal(t, http.StatusCreated, w.Code)
	bid := decodeData(t, w.Body.Bytes())["id"].(string)

	// Carol (hộ B) đọc/sửa/xóa ngân sách hộ A → 404 (không lộ tồn tại — D5)
	carolCookies := app.login(t, carolEmail)
	w = app.do(http.MethodGet, "/api/budgets/"+bid, nil, carolCookies)
	assert.Equal(t, http.StatusNotFound, w.Code)
	w = app.do(http.MethodPatch, "/api/budgets/"+bid,
		map[string]any{"limit_amount": 1, "period_type": "MONTHLY", "expected_updated_at": time.Now().Format(time.RFC3339)}, carolCookies)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Carol không thấy trong danh sách của mình
	w = app.do(http.MethodGet, "/api/budgets", nil, carolCookies)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), bid)
}
