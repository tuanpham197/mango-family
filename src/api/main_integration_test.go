//go:build integration

// Integration test transport + middleware qua router THẬT (registerRoutes):
// 401 chưa đăng nhập, 404 ngoài hộ, 409 chưa thuộc hộ, mã lỗi contract.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"household-finance/api/component/appctx"
	"household-finance/api/component/hasher"
	"household-finance/api/component/pubsub"
	"household-finance/api/component/wshub"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testApp struct {
	engine *gin.Engine
	db     *gorm.DB
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:app_secret@localhost:5432/finance_dev?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerRoutes(r, appctx.New(db, "it-secret", pubsub.NewLocal(), wshub.NewHub()))
	return &testApp{engine: r, db: db}
}

// seedUser tạo user (tùy chọn kèm hộ riêng) và dọn dẹp sau test.
func (a *testApp) seedUser(t *testing.T, name string, withHousehold bool) (email string, householdID uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	email = fmt.Sprintf("it-%s@test.local", userID)
	hash, err := hasher.HashPassword("Password123!")
	require.NoError(t, err)
	require.NoError(t, a.db.Exec(
		`INSERT INTO users (id, email, display_name, password_hash) VALUES (?, ?, ?, ?)`,
		userID, email, name, hash).Error)
	if withHousehold {
		householdID = uuid.New()
		require.NoError(t, a.db.Exec(
			`INSERT INTO households (id, name, created_by) VALUES (?, 'IT Hộ', ?)`, householdID, userID).Error)
		require.NoError(t, a.db.Exec(
			`INSERT INTO household_members (household_id, user_id) VALUES (?, ?)`, householdID, userID).Error)
	}
	t.Cleanup(func() {
		if withHousehold {
			a.db.Exec(`DELETE FROM categorization_rules WHERE household_id = ?`, householdID)
			a.db.Exec(`DELETE FROM transactions WHERE household_id = ?`, householdID)
			a.db.Exec(`DELETE FROM categories WHERE household_id = ?`, householdID)
			a.db.Exec(`DELETE FROM household_members WHERE household_id = ?`, householdID)
			a.db.Exec(`DELETE FROM households WHERE id = ?`, householdID)
		}
		a.db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	})
	return email, householdID
}

func (a *testApp) do(method, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	a.engine.ServeHTTP(w, req)
	return w
}

func (a *testApp) login(t *testing.T, email string) []*http.Cookie {
	t.Helper()
	w := a.do(http.MethodPost, "/api/auth/login", map[string]string{"email": email, "password": "Password123!"}, nil)
	require.Equal(t, http.StatusOK, w.Code)
	res := w.Result()
	require.NotEmpty(t, res.Cookies())
	return res.Cookies()
}

func TestAuthFlowAndErrorCodes(t *testing.T) {
	app := newTestApp(t)
	email, _ := app.seedUser(t, "Alice IT", true)

	// chưa đăng nhập → 401 (quickstart #0)
	w := app.do(http.MethodGet, "/api/categories", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHENTICATED")

	// sai mật khẩu → 401 INVALID_CREDENTIALS (UC-TRK-01 E1)
	w = app.do(http.MethodPost, "/api/auth/login", map[string]string{"email": email, "password": "sai"}, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_CREDENTIALS")

	// đăng nhập đúng → /api/me trả user + household
	cookies := app.login(t, email)
	w = app.do(http.MethodGet, "/api/me", nil, cookies)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Alice IT")
}

func TestNoHousehold409(t *testing.T) {
	app := newTestApp(t)
	email, _ := app.seedUser(t, "Mồ côi", false)
	cookies := app.login(t, email)

	w := app.do(http.MethodGet, "/api/me", nil, cookies)
	assert.Equal(t, http.StatusConflict, w.Code) // UC-TRK-01 E2
	assert.Contains(t, w.Body.String(), "NO_HOUSEHOLD")

	w = app.do(http.MethodGet, "/api/categories", nil, cookies)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCrossHousehold404(t *testing.T) {
	app := newTestApp(t)
	aliceEmail, _ := app.seedUser(t, "Alice IT", true)
	carolEmail, _ := app.seedUser(t, "Carol IT", true)

	aliceCookies := app.login(t, aliceEmail)
	w := app.do(http.MethodPost, "/api/categories",
		map[string]any{"name": "Riêng hộ A", "type": "EXPENSE"}, aliceCookies)
	require.Equal(t, http.StatusCreated, w.Code)
	var created struct {
		Data struct {
			ID        string `json:"id"`
			UpdatedAt string `json:"updated_at"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	// Carol (hộ B) sửa/xóa bản ghi hộ A → 404, không lộ tồn tại (FR-018, D5)
	carolCookies := app.login(t, carolEmail)
	w = app.do(http.MethodPatch, "/api/categories/"+created.Data.ID,
		map[string]any{"name": "Hack", "expected_updated_at": created.Data.UpdatedAt}, carolCookies)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "RECORD_GONE")

	w = app.do(http.MethodDelete, "/api/categories/"+created.Data.ID,
		map[string]any{"expected_updated_at": created.Data.UpdatedAt}, carolCookies)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Carol cũng không thấy trong danh sách
	w = app.do(http.MethodGet, "/api/categories?include_hidden=true", nil, carolCookies)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "Riêng hộ A")
}

func TestContractErrorCodesOverHTTP(t *testing.T) {
	app := newTestApp(t)
	email, _ := app.seedUser(t, "Alice IT", true)
	cookies := app.login(t, email)

	// tạo danh mục thiếu type → INVALID_REQUEST (field type)
	w := app.do(http.MethodPost, "/api/categories", map[string]any{"name": "Thiếu loại"}, cookies)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// giao dịch thiếu danh mục → 422 CATEGORY_REQUIRED (FR-013)
	w = app.do(http.MethodPost, "/api/transactions", map[string]any{"amount": 1000, "type": "EXPENSE"}, cookies)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_REQUIRED")
}
