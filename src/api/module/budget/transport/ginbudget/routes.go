package ginbudget

import (
	"errors"
	"log"
	"net/http"
	"time"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/pubsub"
	"household-finance/api/middleware"
	budgetbiz "household-finance/api/module/budget/biz"
	"household-finance/api/module/budget/eval"
	budgetmodel "household-finance/api/module/budget/model"
	budgetstorage "household-finance/api/module/budget/storage"
	categorystorage "household-finance/api/module/category/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// publishBudgetsChanged — WS refetch cho mọi thành viên (SC-006, D26).
func publishBudgetsChanged(ac appctx.AppContext, householdID uuid.UUID) {
	ac.GetPubSub().Publish(pubsub.Message{Topic: common.TopicBudgetsChanged, HouseholdID: householdID})
}

// evaluate — recompute cảnh báo ngay sau CRUD ngân sách (giới hạn/kỳ đổi → cảnh báo
// tính lại; best-effort). Sự kiện giao dịch do subscriber eval.Start xử lý (D24).
func evaluate(ac appctx.AppContext, householdID uuid.UUID) {
	if err := eval.Run(ac, householdID); err != nil {
		log.Printf("đánh giá cảnh báo ngân sách thất bại (bỏ qua): %v", err)
	}
}

func paramID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.WriteError(c, common.NewNotFound("ngân sách không tồn tại"))
		return uuid.Nil, false
	}
	return id, true
}

type writeReq struct {
	Type              string     `json:"type"`
	CategoryID        *uuid.UUID `json:"category_id"`
	LimitAmount       float64    `json:"limit_amount"`
	PeriodType        string     `json:"period_type"`
	StartDate         *time.Time `json:"start_date"`
	EndDate           *time.Time `json:"end_date"`
	ExpectedUpdatedAt *time.Time `json:"expected_updated_at"`
}

func (r writeReq) fields() budgetbiz.BudgetFields {
	return budgetbiz.BudgetFields{
		Type:        r.Type,
		CategoryID:  r.CategoryID,
		LimitAmount: r.LimitAmount,
		PeriodType:  r.PeriodType,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
	}
}

func listBiz(ac appctx.AppContext) *budgetbiz.ListBudgetsBiz {
	db := ac.GetDB()
	return budgetbiz.NewListBudgetsBiz(
		budgetstorage.NewSQLStore(db),
		budgetstorage.NewProgressStore(db),
		budgetstorage.NewAlertStore(db),
	)
}

// Create — POST /api/budgets (contracts §Budgets).
func Create(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req writeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ"))
			return
		}
		db := ac.GetDB()
		biz := budgetbiz.NewCreateBudgetBiz(budgetstorage.NewSQLStore(db), categorystorage.NewSQLStore(db))
		householdID := middleware.HouseholdID(c)
		bud, err := biz.Create(c.Request.Context(), householdID, middleware.CurrentUser(c).ID, req.fields())
		if err != nil {
			common.WriteError(c, err)
			return
		}
		// Đánh giá cảnh báo ngay (ngân sách tạo giữa kỳ có thể đã trên ngưỡng — #33) + WS.
		evaluate(ac, householdID)
		publishBudgetsChanged(ac, householdID)
		// Trả về view kèm tiến độ suy ra (spent gồm chi từ đầu kỳ).
		view, gerr := listBiz(ac).Get(c.Request.Context(), householdID, bud.ID, time.Now())
		if gerr != nil {
			common.WriteData(c, http.StatusCreated, bud)
			return
		}
		common.WriteData(c, http.StatusCreated, view)
	}
}

// List — GET /api/budgets?status= (contracts §Budgets); mặc định ACTIVE.
func List(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", budgetmodel.StatusActive)
		items, err := listBiz(ac).List(c.Request.Context(), middleware.HouseholdID(c), status, time.Now())
		if err != nil {
			common.WriteError(c, err)
			return
		}
		common.WriteOK(c, items)
	}
}

// Get — GET /api/budgets/:id (prefill form sửa; ngoài hộ → 404).
func Get(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := paramID(c)
		if !ok {
			return
		}
		view, err := listBiz(ac).Get(c.Request.Context(), middleware.HouseholdID(c), id, time.Now())
		if err != nil {
			if errors.Is(err, budgetstorage.ErrNotFound) {
				common.WriteError(c, common.NewNotFound("ngân sách không tồn tại"))
				return
			}
			common.WriteError(c, err)
			return
		}
		common.WriteOK(c, view)
	}
}

// Update — PATCH /api/budgets/:id (contracts §Budgets; FR-012/013, D25).
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
		db := ac.GetDB()
		biz := budgetbiz.NewUpdateBudgetBiz(budgetstorage.NewSQLStore(db), categorystorage.NewSQLStore(db), listBiz(ac))
		householdID := middleware.HouseholdID(c)
		view, err := biz.Update(c.Request.Context(), householdID, id, budgetbiz.UpdateBudgetInput{
			BudgetFields:      req.fields(),
			ExpectedUpdatedAt: expected,
		})
		if err != nil {
			common.WriteError(c, err)
			return
		}
		// Giới hạn/kỳ đổi → tính lại cảnh báo theo giá trị mới (UC-BGT-05 AC-1) + WS.
		evaluate(ac, householdID)
		publishBudgetsChanged(ac, householdID)
		// Dựng lại view SAU evaluate để alerts phản ánh giới hạn mới (UC-BGT-05 AC-1).
		if fresh, gerr := listBiz(ac).Get(c.Request.Context(), householdID, id, time.Now()); gerr == nil {
			view = fresh
		}
		common.WriteOK(c, view)
	}
}

// Delete — DELETE /api/budgets/:id (contracts §Budgets; UI đã xác nhận trước — UC-BGT-06).
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
		biz := budgetbiz.NewDeleteBudgetBiz(budgetstorage.NewSQLStore(ac.GetDB()))
		householdID := middleware.HouseholdID(c)
		if err := biz.Delete(c.Request.Context(), householdID, id, req.ExpectedUpdatedAt); err != nil {
			common.WriteError(c, err)
			return
		}
		publishBudgetsChanged(ac, householdID) // cảnh báo liên quan đã cascade
		c.Status(http.StatusNoContent)
	}
}
