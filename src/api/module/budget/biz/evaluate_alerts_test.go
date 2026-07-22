package biz

import (
	"context"
	"testing"
	"time"

	"household-finance/api/module/budget/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateAlertState(t *testing.T) {
	cases := []struct {
		name                                 string
		spent, limit                         float64
		has80, hasOver                       bool
		fire80, clear80, fireOver, clearOver bool
		overAmount                           float64
	}{
		{"dưới 80% không làm gì", 700, 1000, false, false, false, false, false, false, 0},
		{"đạt 80% → phát", 800, 1000, false, false, true, false, false, false, 0},
		{"giữ trên 80% không phát lại", 850, 1000, true, false, false, false, false, false, 0},
		{"vượt 100% → phát kèm over_amount", 1050, 1000, true, false, false, false, true, false, 50},
		{"đã có cả hai → không phát lại", 1050, 1000, true, true, false, false, false, false, 0},
		{"tụt dưới 80% → re-arm cả hai", 750, 1000, true, true, false, true, false, true, 0},
		{"tụt dưới 100% vẫn trên 80% → chỉ gỡ over", 900, 1000, true, true, false, false, false, true, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := evaluateAlertState(c.spent, c.limit, c.has80, c.hasOver)
			assert.Equal(t, c.fire80, d.fire80, "fire80")
			assert.Equal(t, c.clear80, d.clear80, "clear80")
			assert.Equal(t, c.fireOver, d.fireOver, "fireOver")
			assert.Equal(t, c.clearOver, d.clearOver, "clearOver")
			if c.fireOver {
				assert.Equal(t, c.overAmount, d.overAmount)
			}
		})
	}
}

func TestPercentOf(t *testing.T) {
	assert.Equal(t, 70, percentOf(3500000, 5000000))
	assert.Equal(t, 103, percentOf(5137035, 5000000)) // round
	assert.Equal(t, 0, percentOf(100, 0))             // tránh chia 0
}

// --- computeSpent: gộp danh mục con một cấp (D21) ---

type mockProgress struct {
	spent       float64
	children    []uuid.UUID
	lastCatIDs  []uuid.UUID
	totalCalled bool
}

func (m *mockProgress) ChildCategoryIDs(_ context.Context, _, _ uuid.UUID) ([]uuid.UUID, error) {
	return m.children, nil
}
func (m *mockProgress) SumExpenseByCategories(_ context.Context, _ uuid.UUID, ids []uuid.UUID, _, _ time.Time) (float64, error) {
	m.lastCatIDs = ids
	return m.spent, nil
}
func (m *mockProgress) SumExpenseTotal(_ context.Context, _ uuid.UUID, _, _ time.Time) (float64, error) {
	m.totalCalled = true
	return m.spent, nil
}

func TestComputeSpent_CategoryIncludesChildren(t *testing.T) {
	parent := uuid.New()
	c1, c2 := uuid.New(), uuid.New()
	mp := &mockProgress{spent: 1234, children: []uuid.UUID{c1, c2}}
	p := model.Period{}
	got, err := computeSpent(context.Background(), uuid.New(), model.TypeCategory, &parent, p, mp)
	require.NoError(t, err)
	assert.Equal(t, float64(1234), got)
	assert.Equal(t, []uuid.UUID{parent, c1, c2}, mp.lastCatIDs) // cha + con một cấp
}

func TestComputeSpent_TotalSumsAll(t *testing.T) {
	mp := &mockProgress{spent: 9999}
	got, err := computeSpent(context.Background(), uuid.New(), model.TypeTotal, nil, model.Period{}, mp)
	require.NoError(t, err)
	assert.Equal(t, float64(9999), got)
	assert.True(t, mp.totalCalled)
}

// --- EvaluateHousehold end-to-end: dedup + re-arm (idempotent) ---

type mockActiveBudgets struct{ budgets []model.Budget }

func (m *mockActiveBudgets) ListActive(_ context.Context, _ uuid.UUID) ([]model.Budget, error) {
	return m.budgets, nil
}
func (m *mockActiveBudgets) EndExpiredOneTime(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}

type mockAlertRW struct{ rows []model.BudgetAlert }

func (m *mockAlertRW) GetActive(_ context.Context, budgetID uuid.UUID, periodKey string) ([]model.BudgetAlert, error) {
	var out []model.BudgetAlert
	for _, a := range m.rows {
		if a.BudgetID == budgetID && a.PeriodKey == periodKey {
			out = append(out, a)
		}
	}
	return out, nil
}
func (m *mockAlertRW) Insert(_ context.Context, a *model.BudgetAlert) error {
	m.rows = append(m.rows, *a)
	return nil
}
func (m *mockAlertRW) Delete(_ context.Context, budgetID uuid.UUID, periodKey, level string) error {
	kept := m.rows[:0]
	for _, a := range m.rows {
		if a.BudgetID == budgetID && a.PeriodKey == periodKey && a.Level == level {
			continue
		}
		kept = append(kept, a)
	}
	m.rows = kept
	return nil
}

func (m *mockAlertRW) levels(budgetID uuid.UUID) map[string]bool {
	out := map[string]bool{}
	for _, a := range m.rows {
		if a.BudgetID == budgetID {
			out[a.Level] = true
		}
	}
	return out
}

func TestEvaluateHousehold_FireDedupRearm(t *testing.T) {
	bud := model.Budget{LimitAmount: 1000, Type: model.TypeCategory, PeriodType: model.PeriodMonthly}
	bud.ID = uuid.New()
	catID := uuid.New()
	bud.CategoryID = &catID
	hid := uuid.New()
	now := time.Now()

	mp := &mockProgress{}
	alerts := &mockAlertRW{}
	ev := NewAlertEvaluator(&mockActiveBudgets{budgets: []model.Budget{bud}}, mp, alerts)

	run := func() { require.NoError(t, ev.EvaluateHousehold(context.Background(), hid, now)) }

	// 1) 85% → phát THRESHOLD_80
	mp.spent = 850
	run()
	assert.True(t, alerts.levels(bud.ID)[model.LevelThreshold80])
	assert.False(t, alerts.levels(bud.ID)[model.LevelOver100])

	// 2) 105% → thêm OVER_100 (over_amount = 50)
	mp.spent = 1050
	run()
	assert.True(t, alerts.levels(bud.ID)[model.LevelOver100])
	assert.Len(t, alerts.rows, 2)

	// 3) idempotent: chạy lại không thêm hàng mới (chống trùng — SC-004)
	run()
	assert.Len(t, alerts.rows, 2)

	// 4) tụt về 70% → gỡ cả hai (re-arm)
	mp.spent = 700
	run()
	assert.Empty(t, alerts.rows)

	// 5) vượt 80% lần nữa → phát lại (US3 #4)
	mp.spent = 900
	run()
	assert.True(t, alerts.levels(bud.ID)[model.LevelThreshold80])
}
