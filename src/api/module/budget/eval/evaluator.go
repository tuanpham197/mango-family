// Package eval — đánh giá cảnh báo ngân sách theo sự kiện (D24). Giữ đúng chiều phụ
// thuộc: module transaction KHÔNG biết budget; budget là hạ nguồn, phản ứng qua pubsub.
package eval

import (
	"context"
	"log"
	"time"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/pubsub"
	budgetbiz "household-finance/api/module/budget/biz"
	budgetstorage "household-finance/api/module/budget/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Run — recompute cảnh báo cho một hộ trong MỘT DB transaction (D24). Idempotent:
// gọi lại cho cùng trạng thái không đổi kết quả. Dùng bởi subscriber (giao dịch đổi)
// và transport (CRUD ngân sách).
func Run(ac appctx.AppContext, householdID uuid.UUID) error {
	now := time.Now()
	return ac.GetDB().Transaction(func(tx *gorm.DB) error {
		ev := budgetbiz.NewAlertEvaluator(
			budgetstorage.NewSQLStore(tx),
			budgetstorage.NewProgressStore(tx),
			budgetstorage.NewAlertStore(tx),
		)
		return ev.EvaluateHousehold(context.Background(), householdID, now)
	})
}

// Start — subscriber pubsub: nhận transactions_changed (mọi mutation giao dịch của
// 002) → recompute → Publish(budgets_changed). KHÔNG phản ứng budgets_changed để
// tránh vòng lặp (D24). Bỏ lỡ message vẫn tự lành ở lần mutation kế (idempotent).
func Start(ac appctx.AppContext) {
	ch := ac.GetPubSub().Subscribe()
	go func() {
		for msg := range ch {
			if msg.Topic != common.TopicTransactionsChanged {
				continue
			}
			if err := Run(ac, msg.HouseholdID); err != nil {
				log.Printf("đánh giá cảnh báo ngân sách thất bại (bỏ qua): %v", err)
				continue
			}
			ac.GetPubSub().Publish(pubsub.Message{Topic: common.TopicBudgetsChanged, HouseholdID: msg.HouseholdID})
		}
	}()
}
