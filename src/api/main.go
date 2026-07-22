package main

import (
	"log"
	"os"

	"household-finance/api/component/appctx"
	"household-finance/api/component/pubsub"
	"household-finance/api/component/subscriber"
	"household-finance/api/component/wshub"
	"household-finance/api/config"
	"household-finance/api/middleware"
	budgeteval "household-finance/api/module/budget/eval"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	_ = godotenv.Load() // .env là tùy chọn (dev)

	dsn := env("DATABASE_URL", "postgres://app:app_secret@localhost:5432/finance_dev?sslmode=disable")
	secret := env("JWT_SECRET", "dev-secret-change-me")
	port := env("PORT", "8080")

	pool := config.DBPoolConfig()
	dialector := postgres.Open(dsn)
	if pool.PreferSimpleProtocol {
		// Supabase transaction pooler (cổng 6543) không hỗ trợ prepared statement.
		dialector = postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("không kết nối được Postgres: %v", err)
	}
	// Giới hạn pool — Supabase pooler có số kết nối hữu hạn; Cloud Run stateless.
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(pool.MaxOpen)
		sqlDB.SetMaxIdleConns(pool.MaxIdle)
		sqlDB.SetConnMaxLifetime(pool.MaxLifetime)
		sqlDB.SetConnMaxIdleTime(pool.MaxIdleTime)
	}

	ps := pubsub.NewLocal()
	hub := wshub.NewHub()
	subscriber.Start(ps, hub) // pubsub → WebSocket hub

	ac := appctx.New(db, secret, ps, hub)
	budgeteval.Start(ac) // đánh giá cảnh báo ngân sách theo transactions_changed (D24)

	r := gin.New()
	r.Use(gin.Logger(), middleware.Recover(), middleware.CORS(config.CORSOrigins()))

	registerRoutes(r, ac)
	registerDocs(r) // OpenAPI spec + Swagger UI tại /docs (tắt bằng DOCS_ENABLED=false)
	serveSPA(r)     // prod: Go serve web/dist (một origin — design §3)

	log.Printf("API chạy tại :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// serveSPA phục vụ web/dist nếu tồn tại (prod build); dev dùng Vite proxy.
func serveSPA(r *gin.Engine) {
	dist := env("WEB_DIST", "../web/dist")
	if _, err := os.Stat(dist); err != nil {
		return
	}
	r.Static("/assets", dist+"/assets")
	r.StaticFile("/", dist+"/index.html")
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "không tìm thấy"}})
			return
		}
		c.File(dist + "/index.html") // SPA fallback
	})
}
