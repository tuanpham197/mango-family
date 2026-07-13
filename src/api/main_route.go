package main

import (
	"household-finance/api/component/appctx"
	"household-finance/api/middleware"
	"household-finance/api/module/category/transport/gincategory"
	"household-finance/api/module/transaction/transport/gintransaction"
	"household-finance/api/module/user/transport/ginuser"

	"github.com/gin-gonic/gin"
)

// registerRoutes — đăng ký route các module (mẫu learn_go main_route).
func registerRoutes(r *gin.Engine, ac appctx.AppContext) {
	api := r.Group("/api")

	// Public
	api.POST("/auth/login", ginuser.Login(ac))
	api.POST("/auth/logout", ginuser.Logout())

	// Authenticated
	authed := api.Group("", middleware.Authenticate(ac))
	authed.GET("/me", ginuser.Me(ac))

	// Authenticated + household scope (mọi resource theo hộ — D5)
	scoped := authed.Group("", middleware.RequireHousehold(ac))
	scoped.GET("/categories", gincategory.List(ac))
	scoped.GET("/categories/suggest", gincategory.Suggest(ac))
	scoped.POST("/categories", gincategory.Create(ac))
	scoped.PATCH("/categories/:id", gincategory.Update(ac))
	scoped.DELETE("/categories/:id", gincategory.Delete(ac))
	scoped.POST("/transactions", gintransaction.Create(ac))
	scoped.GET("/transactions", gintransaction.List(ac))

	// WebSocket invalidation hub (D8)
	r.GET("/ws", middleware.Authenticate(ac), middleware.RequireHousehold(ac), func(c *gin.Context) {
		_ = ac.GetWSHub().ServeWS(c.Writer, c.Request, middleware.HouseholdID(c))
	})
}
