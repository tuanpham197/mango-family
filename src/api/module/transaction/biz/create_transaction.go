package biz

import (
	"context"
	"log"
	"time"

	"household-finance/api/common"
	categorybiz "household-finance/api/module/category/biz"
	"household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

type TransactionCreator interface {
	Create(ctx context.Context, t *model.Transaction) error
}

// RuleLearner — upsert quy tắc gợi ý từ mô tả giao dịch (US4, D7). Best-effort:
// lỗi học KHÔNG làm hỏng giao dịch đã lưu.
type RuleLearner interface {
	UpsertRule(ctx context.Context, householdID uuid.UUID, keyword string, categoryID uuid.UUID) error
}

type CreateTransactionInput = TxFields

type CreateTransactionBiz struct {
	catStore CategoryFinder
	accStore AccountFinder
	txStore  TransactionCreator
	learner  RuleLearner // nil = không học (tùy chọn)
}

func NewCreateTransactionBiz(catStore CategoryFinder, accStore AccountFinder, txStore TransactionCreator, learner RuleLearner) *CreateTransactionBiz {
	return &CreateTransactionBiz{catStore: catStore, accStore: accStore, txStore: txStore, learner: learner}
}

// Create — POST /api/transactions (FR-001…007): số tiền > 0, danh mục cùng loại +
// cùng hộ, tài khoản cùng hộ, mô tả ≤ 255, không ngày tương lai; created_by gán
// từ phiên (FR-013). Học rule gợi ý từ mô tả (D7).
func (b *CreateTransactionBiz) Create(ctx context.Context, householdID, userID uuid.UUID, in CreateTransactionInput) (*model.Transaction, error) {
	when, err := validateFields(ctx, householdID, in, b.catStore, b.accStore, time.Now())
	if err != nil {
		return nil, err
	}

	t := &model.Transaction{
		HouseholdID:     householdID,
		CreatedBy:       userID,
		Amount:          in.Amount,
		Type:            in.Type,
		CategoryID:      *in.CategoryID,
		AccountID:       *in.AccountID,
		Description:     in.Description,
		TransactionDate: when,
	}
	if err := b.txStore.Create(ctx, t); err != nil {
		return nil, common.NewInternal(err)
	}
	if b.learner != nil && in.Description != nil {
		if keyword := categorybiz.NormalizeKeyword(*in.Description); keyword != "" {
			if err := b.learner.UpsertRule(ctx, householdID, keyword, t.CategoryID); err != nil {
				log.Printf("upsert categorization rule thất bại (bỏ qua): %v", err)
			}
		}
	}
	return t, nil
}
