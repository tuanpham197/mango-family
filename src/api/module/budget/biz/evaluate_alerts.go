package biz

import (
	"context"
	"time"

	"household-finance/api/module/budget/model"

	"github.com/google/uuid"
)

// ActiveBudgetLister — ngân sách ACTIVE để đánh giá cảnh báo (evaluator — D24).
type ActiveBudgetLister interface {
	ListActive(ctx context.Context, householdID uuid.UUID) ([]model.Budget, error)
	EndExpiredOneTime(ctx context.Context, householdID uuid.UUID, now time.Time) error
}

// AlertRW — máy trạng thái cảnh báo đọc/ghi trong kỳ hiện tại (D23).
type AlertRW interface {
	GetActive(ctx context.Context, budgetID uuid.UUID, periodKey string) ([]model.BudgetAlert, error)
	Insert(ctx context.Context, a *model.BudgetAlert) error
	Delete(ctx context.Context, budgetID uuid.UUID, periodKey, level string) error
}

// AlertEvaluator — recompute idempotent tiến độ + máy trạng thái cảnh báo (D24).
type AlertEvaluator struct {
	budgets  ActiveBudgetLister
	progress ProgressReader
	alerts   AlertRW
}

func NewAlertEvaluator(budgets ActiveBudgetLister, progress ProgressReader, alerts AlertRW) *AlertEvaluator {
	return &AlertEvaluator{budgets: budgets, progress: progress, alerts: alerts}
}

// EvaluateHousehold — với mỗi ngân sách ACTIVE của hộ: tính lại tiến độ kỳ hiện tại
// rồi áp máy trạng thái cảnh báo (D23). Idempotent — chạy lại cho cùng trạng thái
// (bỏ lỡ message vẫn tự lành ở lần sau — D24). ONE_TIME hết kỳ bị loại (không cảnh báo).
func (e *AlertEvaluator) EvaluateHousehold(ctx context.Context, householdID uuid.UUID, now time.Time) error {
	if err := e.budgets.EndExpiredOneTime(ctx, householdID, now); err != nil {
		return err
	}
	budgets, err := e.budgets.ListActive(ctx, householdID)
	if err != nil {
		return err
	}
	for i := range budgets {
		if err := e.evaluateOne(ctx, householdID, budgets[i], now); err != nil {
			return err
		}
	}
	return nil
}

func (e *AlertEvaluator) evaluateOne(ctx context.Context, householdID uuid.UUID, bud model.Budget, now time.Time) error {
	p := model.ResolvePeriod(bud.PeriodType, bud.StartDate, bud.EndDate, now)
	spent, err := computeSpent(ctx, householdID, bud.Type, bud.CategoryID, p, e.progress)
	if err != nil {
		return err
	}
	existing, err := e.alerts.GetActive(ctx, bud.ID, p.Key)
	if err != nil {
		return err
	}
	has80, hasOver := false, false
	for _, a := range existing {
		switch a.Level {
		case model.LevelThreshold80:
			has80 = true
		case model.LevelOver100:
			hasOver = true
		}
	}

	d := evaluateAlertState(spent, bud.LimitAmount, has80, hasOver)
	if d.fire80 {
		if err := e.alerts.Insert(ctx, &model.BudgetAlert{BudgetID: bud.ID, PeriodKey: p.Key, Level: model.LevelThreshold80}); err != nil {
			return err
		}
	}
	if d.clear80 {
		if err := e.alerts.Delete(ctx, bud.ID, p.Key, model.LevelThreshold80); err != nil {
			return err
		}
	}
	if d.fireOver {
		over := d.overAmount
		if err := e.alerts.Insert(ctx, &model.BudgetAlert{BudgetID: bud.ID, PeriodKey: p.Key, Level: model.LevelOver100, OverAmount: &over}); err != nil {
			return err
		}
	}
	if d.clearOver {
		if err := e.alerts.Delete(ctx, bud.ID, p.Key, model.LevelOver100); err != nil {
			return err
		}
	}
	return nil
}

type alertDecision struct {
	fire80, clear80     bool
	fireOver, clearOver bool
	overAmount          float64
}

// evaluateAlertState — máy trạng thái thuần (unit-test T029): so sánh CHÍNH XÁC theo
// spent để tránh sai số làm tròn (fire 80 khi spent ≥ 80% giới hạn; over khi spent >
// giới hạn); phát lại sau khi tụt dưới mức (delete — re-arm, US3 #4). D23.
func evaluateAlertState(spent, limit float64, has80, hasOver bool) alertDecision {
	var d alertDecision
	at80 := spent >= 0.8*limit
	over := spent > limit

	if at80 && !has80 {
		d.fire80 = true
	}
	if !at80 && has80 {
		d.clear80 = true
	}
	if over && !hasOver {
		d.fireOver = true
		d.overAmount = spent - limit
	}
	if !over && hasOver {
		d.clearOver = true
	}
	return d
}
