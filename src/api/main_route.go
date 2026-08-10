package main

import (
	"household-finance/api/component/appctx"
	"household-finance/api/middleware"
	"household-finance/api/module/account/transport/ginaccount"
	"household-finance/api/module/budget/transport/ginbudget"
	"household-finance/api/module/category/transport/gincategory"
	"household-finance/api/module/overview/transport/ginoverview"
	"household-finance/api/module/report/transport/ginreport"
	"household-finance/api/module/transaction/transport/gintransaction"
	"household-finance/api/module/user/transport/ginuser"

	"github.com/gin-gonic/gin"
)

// registerRoutes — đăng ký route các module (mẫu learn_go main_route).
func registerRoutes(r *gin.Engine, ac appctx.AppContext) {
	// Health check (public) — xác minh app sống + DB kết nối được.
	r.GET("/healthz", healthz(ac))

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
	scoped.GET("/accounts", ginaccount.List(ac))
	scoped.POST("/transactions", gintransaction.Create(ac))
	scoped.GET("/transactions", gintransaction.List(ac))
	scoped.GET("/transactions/:id", gintransaction.Get(ac))
	scoped.PATCH("/transactions/:id", gintransaction.Update(ac))
	scoped.DELETE("/transactions/:id", gintransaction.Delete(ac))
	scoped.POST("/budgets", ginbudget.Create(ac))
	scoped.GET("/budgets", ginbudget.List(ac))
	scoped.GET("/budgets/:id", ginbudget.Get(ac))
	scoped.PATCH("/budgets/:id", ginbudget.Update(ac))
	scoped.DELETE("/budgets/:id", ginbudget.Delete(ac))
	scoped.GET("/overview", ginoverview.Get(ac))                // 004: tổng hợp giá trị suy ra màn Tổng quan
	scoped.GET("/reports/overview", ginreport.Overview(ac))     // 005: báo cáo tổng quan
	scoped.GET("/reports/category/:id", ginreport.Category(ac)) // 005: báo cáo chi tiết danh mục
	scoped.GET("/reports/members", ginreport.Members(ac))       // 008: báo cáo theo thành viên
	scoped.GET("/reports/member/:id", ginreport.Member(ac))     // 008: drill-in giao dịch một thành viên

	// WebSocket invalidation hub (D8)
	r.GET("/ws", middleware.Authenticate(ac), middleware.RequireHousehold(ac), func(c *gin.Context) {
		_ = ac.GetWSHub().ServeWS(c.Writer, c.Request, middleware.HouseholdID(c))
	})
}
