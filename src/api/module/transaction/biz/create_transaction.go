package biz

import (
	"context"
	"time"

	"household-finance/api/common"
	categorymodel "household-finance/api/module/category/model"
	"household-finance/api/module/transaction/model"

	"github.com/google/uuid"
)

type CategoryFinder interface {
	FindByID(ctx context.Context, householdID, id uuid.UUID) (*categorymodel.Category, error)
}

type TransactionCreator interface {
	Create(ctx context.Context, t *model.Transaction) error
}

type CreateTransactionInput struct {
	Amount          float64
	Type            string
	CategoryID      *uuid.UUID
	Description     *string
	TransactionDate *time.Time
}

type CreateTransactionBiz struct {
	catStore CategoryFinder
	txStore  TransactionCreator
}

func NewCreateTransactionBiz(catStore CategoryFinder, txStore TransactionCreator) *CreateTransactionBiz {
	return &CreateTransactionBiz{catStore: catStore, txStore: txStore}
}

// Create — POST /api/transactions: bắt buộc danh mục cùng loại + cùng hộ;
// created_by gán từ phiên đăng nhập (FR-013/014/019/022).
func (b *CreateTransactionBiz) Create(ctx context.Context, householdID, userID uuid.UUID, in CreateTransactionInput) (*model.Transaction, error) {
	if in.Type != common.TypeIncome && in.Type != common.TypeExpense {
		return nil, common.NewBadRequest("type phải là INCOME hoặc EXPENSE").WithField("type")
	}
	if in.Amount <= 0 {
		return nil, common.NewUnprocessable(common.ErrCodeAmountInvalid, "số tiền phải lớn hơn 0", "amount")
	}
	if in.Description != nil && len([]rune(*in.Description)) > common.MaxDescriptionLen {
		return nil, common.NewUnprocessable(common.ErrCodeDescriptionTooLong, "mô tả tối đa 255 ký tự", "description")
	}
	if in.CategoryID == nil {
		return nil, common.NewUnprocessable(common.ErrCodeCategoryRequired, "vui lòng chọn danh mục", "category_id")
	}

	cat, err := b.catStore.FindByID(ctx, householdID, *in.CategoryID)
	if err != nil {
		// Ngoài hộ / không tồn tại → 404, không lộ thông tin (research D5).
		return nil, common.NewNotFound("danh mục không tồn tại")
	}
	if cat.Type != in.Type {
		return nil, common.NewUnprocessable(common.ErrCodeCategoryTypeMismatch,
			"danh mục phải cùng loại với giao dịch", "category_id")
	}

	when := time.Now()
	if in.TransactionDate != nil {
		when = *in.TransactionDate
	}
	t := &model.Transaction{
		HouseholdID:     householdID,
		CreatedBy:       userID,
		Amount:          in.Amount,
		Type:            in.Type,
		CategoryID:      *in.CategoryID,
		Description:     in.Description,
		TransactionDate: when,
	}
	if err := b.txStore.Create(ctx, t); err != nil {
		return nil, common.NewInternal(err)
	}
	return t, nil
}
