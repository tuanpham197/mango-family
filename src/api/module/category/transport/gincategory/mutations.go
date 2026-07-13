package gincategory

import (
	"context"
	"net/http"
	"time"

	"household-finance/api/common"
	"household-finance/api/component/appctx"
	"household-finance/api/component/pubsub"
	"household-finance/api/middleware"
	categorybiz "household-finance/api/module/category/biz"
	categorystorage "household-finance/api/module/category/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func publishChanged(ac appctx.AppContext, householdID uuid.UUID) {
	ac.GetPubSub().Publish(pubsub.Message{Topic: common.TopicCategoriesChanged, HouseholdID: householdID})
}

func paramID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.WriteError(c, common.NewNotFound("danh mục không tồn tại"))
		return uuid.Nil, false
	}
	return id, true
}

type createReq struct {
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	Icon             *string    `json:"icon"`
	ParentID         *uuid.UUID `json:"parent_id"`
	ConfirmDuplicate bool       `json:"confirm_duplicate"`
}

// Create — POST /api/categories.
func Create(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ"))
			return
		}
		biz := categorybiz.NewCreateCategoryBiz(categorystorage.NewSQLStore(ac.GetDB()))
		householdID := middleware.HouseholdID(c)
		cat, err := biz.Create(c.Request.Context(), householdID, middleware.CurrentUser(c).ID, categorybiz.CreateCategoryInput{
			Name:             req.Name,
			Type:             req.Type,
			Icon:             req.Icon,
			ParentID:         req.ParentID,
			ConfirmDuplicate: req.ConfirmDuplicate,
		})
		if err != nil {
			common.WriteError(c, err)
			return
		}
		publishChanged(ac, householdID)
		common.WriteData(c, http.StatusCreated, cat)
	}
}

type updateReq struct {
	Name              *string   `json:"name"`
	Icon              *string   `json:"icon"`
	IsHidden          *bool     `json:"is_hidden"`
	Type              *string   `json:"type"`
	ExpectedUpdatedAt time.Time `json:"expected_updated_at"`
}

// Update — PATCH /api/categories/:id.
func Update(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := paramID(c)
		if !ok {
			return
		}
		var req updateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ"))
			return
		}
		biz := categorybiz.NewUpdateCategoryBiz(categorystorage.NewSQLStore(ac.GetDB()))
		householdID := middleware.HouseholdID(c)
		cat, err := biz.Update(c.Request.Context(), householdID, id, categorybiz.UpdateCategoryInput{
			Name:              req.Name,
			Icon:              req.Icon,
			IsHidden:          req.IsHidden,
			Type:              req.Type,
			ExpectedUpdatedAt: req.ExpectedUpdatedAt,
		})
		if err != nil {
			common.WriteError(c, err)
			return
		}
		publishChanged(ac, householdID)
		common.WriteOK(c, cat)
	}
}

type deleteReq struct {
	Mode              string     `json:"mode"`
	TargetCategoryID  *uuid.UUID `json:"target_category_id"`
	ExpectedUpdatedAt time.Time  `json:"expected_updated_at"`
}

// Delete — DELETE /api/categories/:id (delete-reassign nguyên tử — D12).
func Delete(ac appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := paramID(c)
		if !ok {
			return
		}
		var req deleteReq
		if err := c.ShouldBindJSON(&req); err != nil {
			common.WriteError(c, common.NewBadRequest("body không hợp lệ (cần expected_updated_at)"))
			return
		}
		store := categorystorage.NewSQLStore(ac.GetDB())
		runner := categorybiz.TxRunnerFunc(func(ctx context.Context, fn func(categorybiz.DeleteStore) error) error {
			return store.InTx(ctx, func(tx *categorystorage.SQLStore) error { return fn(tx) })
		})
		biz := categorybiz.NewDeleteCategoryBiz(runner)
		householdID := middleware.HouseholdID(c)
		err := biz.Delete(c.Request.Context(), householdID, id, categorybiz.DeleteCategoryInput{
			Mode:              req.Mode,
			TargetCategoryID:  req.TargetCategoryID,
			ExpectedUpdatedAt: req.ExpectedUpdatedAt,
		})
		if err != nil {
			common.WriteError(c, err)
			return
		}
		publishChanged(ac, householdID)
		c.Status(http.StatusNoContent)
	}
}
