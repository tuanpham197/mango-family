package gintransaction

import (
	"errors"
	"net/http"
	"time"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/pubsub"
	"household-finance/api/middleware"
	accountstorage "household-finance/api/module/account/storage"
	categorystorage "household-finance/api/module/category/storage"
	transactionbiz "household-finance/api/module/transaction/biz"
	transactionstorage "household-finance/api/module/transaction/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// publishTxnChanged — mọi mutation giao dịch làm số dư đổi → phát cả hai event (D14, contracts §WS).
func publishTxnChanged(ac appctx.AppContext, householdID uuid.UUID) {
	ac.GetPubSub().Publish(pubsub.Message{Topic: common.TopicTransactionsChanged, HouseholdID: householdID})
	ac.GetPubSub().Publish(pubsub.Message{Topic: common.TopicAccountsChanged, HouseholdID: householdID})
}

func paramID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.WriteError(c, common.NewNotFound("giao dịch không tồn tại"))
		return uuid.Nil, false
	}
	return id, true
}

type writeReq struct {
	Amount            float64    `json:"amount"`
	Type              string     `json:"type"`
	CategoryID        *uuid.UUID `json:"category_id"`
	AccountID         *uuid.UUID `json:"account_id"`
	Description       *string    `json:"description"`
	TransactionDate   *time.Time `json:"transaction_date"`
	ExpectedUpdatedAt *time.Time `json:"expected_updated_at"`
}

func (r writeReq) fields() transactionbiz.TxFields {
	return transactionbiz.TxFields{
		Amount:          r.Amount,
		Type:            r.Type,
		CategoryID:      r.CategoryID,
		AccountID:       r.AccountID,
		Description:     r.Description,
		TransactionDate: r.TransactionDate,
	}
}

// Create — POST /api/transactions (contracts §Transactions).
func Create(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req writeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ"))
			return
		}
		catStore := categorystorage.NewSQLStore(ac.GetDB())
		biz := transactionbiz.NewCreateTransactionBiz(
			catStore,
			accountstorage.NewSQLStore(ac.GetDB()),
			transactionstorage.NewSQLStore(ac.GetDB()),
			catStore, // học rule gợi ý từ mô tả (US4, D7)
		)
		householdID := middleware.HouseholdID(c)
		t, err := biz.Create(c.Request.Context(), householdID, middleware.CurrentUser(c).ID, req.fields())
		if err != nil {
			common.WriteError(c, err)
			return
		}
		publishTxnChanged(ac, householdID)
		common.WriteData(c, http.StatusCreated, t)
	}
}

// List — GET /api/transactions?page=&page_size=&from=&to= (sổ chung, D16).
// from/to (YYYY-MM-DD, inclusive) lọc theo tháng/năm — filter màn Giao dịch.
func List(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter, ok := listFilterFromQuery(c)
		if !ok {
			return
		}
		store := transactionstorage.NewSQLStore(ac.GetDB())
		paging := common.PagingFromQuery(c)
		items, total, err := store.List(c.Request.Context(), middleware.HouseholdID(c), paging, filter)
		if err != nil {
			common.WriteError(c, common.NewInternal(err))
			return
		}
		paging.Total = total
		c.JSON(http.StatusOK, common.PagedResponse(items, paging))
	}
}

// listFilterFromQuery — đọc ?from=&to= (YYYY-MM-DD). Cả hai bỏ trống → không lọc.
// to → EndExcl = đầu ngày sau `to` (bao trọn ngày cuối). Sai định dạng / to<from → 400.
func listFilterFromQuery(c *gin.Context) (transactionstorage.ListFilter, bool) {
	var f transactionstorage.ListFilter
	fromStr, toStr := c.Query("from"), c.Query("to")
	if fromStr == "" && toStr == "" {
		return f, true
	}
	from, err1 := time.ParseInLocation("2006-01-02", fromStr, time.Local)
	to, err2 := time.ParseInLocation("2006-01-02", toStr, time.Local)
	if err1 != nil || err2 != nil {
		common.WriteError(c, common.NewBadRequest("from/to phải theo định dạng YYYY-MM-DD"))
		return f, false
	}
	if to.Before(from) {
		common.WriteError(c, common.NewBadRequest("ngày kết thúc không được trước ngày bắt đầu").WithField("to"))
		return f, false
	}
	endExcl := to.AddDate(0, 0, 1)
	f.From, f.EndExcl = &from, &endExcl
	return f, true
}

// Get — GET /api/transactions/:id (prefill form sửa; ngoài hộ → 404).
func Get(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := paramID(c)
		if !ok {
			return
		}
		store := transactionstorage.NewSQLStore(ac.GetDB())
		item, err := store.FindByID(c.Request.Context(), middleware.HouseholdID(c), id)
		if err != nil {
			if errors.Is(err, transactionstorage.ErrNotFound) {
				common.WriteError(c, common.NewNotFound("giao dịch không tồn tại"))
				return
			}
			common.WriteError(c, common.NewInternal(err))
			return
		}
		common.WriteOK(c, item)
	}
}

// Update — PATCH /api/transactions/:id (FR-009/014, D17/D18).
func Update(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := paramID(c)
		if !ok {
			return
		}
		var req writeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ"))
			return
		}
		var expected time.Time
		if req.ExpectedUpdatedAt != nil {
			expected = *req.ExpectedUpdatedAt
		}
		biz := transactionbiz.NewUpdateTransactionBiz(
			categorystorage.NewSQLStore(ac.GetDB()),
			accountstorage.NewSQLStore(ac.GetDB()),
			transactionstorage.NewSQLStore(ac.GetDB()),
		)
		householdID := middleware.HouseholdID(c)
		item, err := biz.Update(c.Request.Context(), householdID, id, transactionbiz.UpdateTransactionInput{
			TxFields:          req.fields(),
			ExpectedUpdatedAt: expected,
		})
		if err != nil {
			common.WriteError(c, err)
			return
		}
		publishTxnChanged(ac, householdID)
		common.WriteOK(c, item)
	}
}

// Delete — DELETE /api/transactions/:id (FR-010/014; UI đã xác nhận trước — SC-005).
func Delete(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := paramID(c)
		if !ok {
			return
		}
		var req struct {
			ExpectedUpdatedAt *time.Time `json:"expected_updated_at"`
		}
		_ = c.ShouldBindJSON(&req) // body tùy chọn
		biz := transactionbiz.NewDeleteTransactionBiz(transactionstorage.NewSQLStore(ac.GetDB()))
		householdID := middleware.HouseholdID(c)
		if err := biz.Delete(c.Request.Context(), householdID, id, req.ExpectedUpdatedAt); err != nil {
			common.WriteError(c, err)
			return
		}
		publishTxnChanged(ac, householdID)
		c.Status(http.StatusNoContent)
	}
}
