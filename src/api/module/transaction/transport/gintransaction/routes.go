package gintransaction

import (
	"net/http"
	"time"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/pubsub"
	"household-finance/api/middleware"
	categorystorage "household-finance/api/module/category/storage"
	transactionbiz "household-finance/api/module/transaction/biz"
	transactionstorage "household-finance/api/module/transaction/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createReq struct {
	Amount          float64    `json:"amount"`
	Type            string     `json:"type"`
	CategoryID      *uuid.UUID `json:"category_id"`
	Description     *string    `json:"description"`
	TransactionDate *time.Time `json:"transaction_date"`
}

// Create — POST /api/transactions (contracts §Transactions).
func Create(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ"))
			return
		}
		catStore := categorystorage.NewSQLStore(ac.GetDB())
		biz := transactionbiz.NewCreateTransactionBiz(
			catStore,
			transactionstorage.NewSQLStore(ac.GetDB()),
			catStore, // học rule gợi ý từ mô tả (US4, D7)
		)
		householdID := middleware.HouseholdID(c)
		t, err := biz.Create(c.Request.Context(), householdID, middleware.CurrentUser(c).ID, transactionbiz.CreateTransactionInput{
			Amount:          req.Amount,
			Type:            req.Type,
			CategoryID:      req.CategoryID,
			Description:     req.Description,
			TransactionDate: req.TransactionDate,
		})
		if err != nil {
			common.WriteError(c, err)
			return
		}
		ac.GetPubSub().Publish(pubsub.Message{Topic: common.TopicTransactionsChanged, HouseholdID: householdID})
		common.WriteData(c, http.StatusCreated, t)
	}
}

// List — GET /api/transactions?page=&page_size= (sổ chung, FR-022).
func List(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		store := transactionstorage.NewSQLStore(ac.GetDB())
		paging := common.PagingFromQuery(c)
		items, total, err := store.List(c.Request.Context(), middleware.HouseholdID(c), paging)
		if err != nil {
			common.WriteError(c, common.NewInternal(err))
			return
		}
		paging.Total = total
		c.JSON(http.StatusOK, common.PagedResponse(items, paging))
	}
}
