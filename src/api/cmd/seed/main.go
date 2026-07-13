// cmd/seed — seed dev-only (KHÔNG phải migration — research D9):
// Alice/Bob (hộ "Gia đình A"), Carol (hộ "Gia đình B"), mật khẩu Password123!,
// bộ danh mục mặc định cho mỗi hộ. Idempotent — chạy lại không tạo trùng.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"household-finance/api/component/hasher"
	householdbiz "household-finance/api/module/household/biz"
	householdmodel "household-finance/api/module/household/model"
	householdstorage "household-finance/api/module/household/storage"
	usermodel "household-finance/api/module/user/model"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const devPassword = "Password123!"

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:app_secret@localhost:5432/finance_dev?sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		log.Fatalf("không kết nối được Postgres: %v", err)
	}
	ctx := context.Background()

	alice := upsertUser(ctx, db, "alice@dev.local", "Alice")
	bob := upsertUser(ctx, db, "bob@dev.local", "Bob")
	carol := upsertUser(ctx, db, "carol@dev.local", "Carol")

	hstore := householdstorage.NewSQLStore(db)
	houseA := upsertHousehold(ctx, db, "Gia đình A", alice.ID)
	houseB := upsertHousehold(ctx, db, "Gia đình B", carol.ID)

	must(hstore.AddMember(ctx, houseA.ID, alice.ID))
	must(hstore.AddMember(ctx, houseA.ID, bob.ID))
	must(hstore.AddMember(ctx, houseB.ID, carol.ID))

	must(householdbiz.SeedDefaultCategories(ctx, db, houseA.ID, alice.ID))
	must(householdbiz.SeedDefaultCategories(ctx, db, houseB.ID, carol.ID))

	fmt.Println("Seed xong:")
	fmt.Println("  Hộ A:", houseA.ID, "— alice@dev.local + bob@dev.local /", devPassword)
	fmt.Println("  Hộ B:", houseB.ID, "— carol@dev.local /", devPassword)
}

func upsertUser(ctx context.Context, db *gorm.DB, email, name string) *usermodel.User {
	var u usermodel.User
	err := db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err == nil {
		return &u
	}
	if err != gorm.ErrRecordNotFound {
		log.Fatal(err)
	}
	hash, err := hasher.HashPassword(devPassword)
	must(err)
	u = usermodel.User{Email: email, DisplayName: name, PasswordHash: hash}
	must(db.WithContext(ctx).Create(&u).Error)
	return &u
}

func upsertHousehold(ctx context.Context, db *gorm.DB, name string, createdBy uuid.UUID) *householdmodel.Household {
	var h householdmodel.Household
	err := db.WithContext(ctx).Where("name = ?", name).First(&h).Error
	if err == nil {
		return &h
	}
	if err != gorm.ErrRecordNotFound {
		log.Fatal(err)
	}
	h = householdmodel.Household{Name: name, CreatedBy: createdBy}
	must(db.WithContext(ctx).Create(&h).Error)
	return &h
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
