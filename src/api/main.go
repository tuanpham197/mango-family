package main

import (
	"log"
	"os"

	"household-finance/api/component/appctx"
	"household-finance/api/component/pubsub"
	"household-finance/api/component/subscriber"
	"household-finance/api/component/wshub"
	"household-finance/api/middleware"

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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("không kết nối được Postgres: %v", err)
	}

	ps := pubsub.NewLocal()
	hub := wshub.NewHub()
	subscriber.Start(ps, hub)

	ac := appctx.New(db, secret, ps, hub)

	r := gin.New()
	r.Use(gin.Logger(), middleware.Recover())

	registerRoutes(r, ac)
	serveSPA(r) // prod: Go serve web/dist (một origin — design §3)

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
