// cmd/seed — seed dev-only (KHÔNG phải migration — research D9). Idempotent.
//
// Cấu hình qua biến môi trường:
//   SEED_PASSWORD   — mật khẩu chung (mặc định "Password123!")
//   SEED_EMAIL_1/2  — nếu SET SEED_EMAIL_1: chế độ TÙY CHỈNH → 2 tài khoản CHUNG 1 HỘ
//   SEED_NAME_1/2   — tên hiển thị (mặc định = phần trước @ của email)
//   SEED_HOUSEHOLD  — tên hộ (mặc định "Gia đình")
// Không set SEED_EMAIL_1 → chế độ DEV mặc định: Alice/Bob (hộ A), Carol (hộ B)
// — dùng cho local/e2e, giữ nguyên tài khoản cũ.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// localPart trả phần trước "@" của email (tên hiển thị mặc định).
func localPart(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
}

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
	hstore := householdstorage.NewSQLStore(db)
	password := envOr("SEED_PASSWORD", "Password123!")

	// ── Chế độ TÙY CHỈNH: 2 tài khoản chung 1 hộ (dùng cho prod) ─────────────
	// ⚠️ PROD chạy nhánh TÙY CHỈNH này (SEED_EMAIL_1 set) rồi RETURN — KHÔNG bao giờ
	// chạm seed GIAO DỊCH MẪU ở nhánh DEV bên dưới (prod đã seed thật). Giao dịch mẫu
	// (feature 008) chỉ được tạo cho các hộ DEV "Gia đình A/B" do seed này tự tạo.
	if email1 := envOr("SEED_EMAIL_1", ""); email1 != "" {
		u1 := upsertUser(ctx, db, email1, envOr("SEED_NAME_1", localPart(email1)), password)
		house := upsertHousehold(ctx, db, envOr("SEED_HOUSEHOLD", "Gia đình"), u1.ID)
		must(hstore.AddMember(ctx, house.ID, u1.ID))
		members := email1
		if email2 := envOr("SEED_EMAIL_2", ""); email2 != "" {
			u2 := upsertUser(ctx, db, email2, envOr("SEED_NAME_2", localPart(email2)), password)
			must(hstore.AddMember(ctx, house.ID, u2.ID))
			members += " + " + email2
		}
		must(householdbiz.SeedDefaults(ctx, db, house.ID, u1.ID))
		fmt.Println("Seed xong (tùy chỉnh):")
		fmt.Println("  Hộ:", house.ID, "—", members, "/", password)
		return
	}

	// ── Chế độ DEV mặc định: Alice/Bob/Dave (hộ A), Carol (hộ B) ─────────────
	// CHỈ chạy khi KHÔNG set SEED_EMAIL_1 (không phải prod). Feature 008: hộ A có
	// ≥2 thành viên có giao dịch (Alice, Bob) + 1 thành viên KHÔNG có giao dịch (Dave).
	alice := upsertUser(ctx, db, "alice@dev.local", "Alice", password)
	bob := upsertUser(ctx, db, "bob@dev.local", "Bob", password)
	dave := upsertUser(ctx, db, "dave@dev.local", "Dave", password)
	carol := upsertUser(ctx, db, "carol@dev.local", "Carol", password)

	houseA := upsertHousehold(ctx, db, "Gia đình A", alice.ID)
	houseB := upsertHousehold(ctx, db, "Gia đình B", carol.ID)

	must(hstore.AddMember(ctx, houseA.ID, alice.ID))
	must(hstore.AddMember(ctx, houseA.ID, bob.ID))
	must(hstore.AddMember(ctx, houseA.ID, dave.ID)) // 008: thành viên KHÔNG có giao dịch (QS-3)
	must(hstore.AddMember(ctx, houseB.ID, carol.ID))

	must(householdbiz.SeedDefaults(ctx, db, houseA.ID, alice.ID))
	must(householdbiz.SeedDefaults(ctx, db, houseB.ID, carol.ID))

	// 008: giao dịch mẫu tháng hiện tại cho Alice & Bob (Dave để trống) — DEV-ONLY,
	// idempotent (bỏ qua nếu đã có giao dịch [seed] của người đó trong tháng này).
	seedDevTransactions(ctx, db, houseA.ID, alice.ID)
	seedDevTransactions(ctx, db, houseA.ID, bob.ID)

	fmt.Println("Seed xong (dev):")
	fmt.Println("  Hộ A:", houseA.ID, "— alice + bob (có giao dịch) + dave (trống) /", password)
	fmt.Println("  Hộ B:", houseB.ID, "— carol@dev.local /", password)
}

// seedDevTransactions — DEV-ONLY (feature 008): thêm 1 giao dịch Thu + 1 Chi trong THÁNG
// HIỆN TẠI cho một thành viên, để báo cáo theo thành viên có dữ liệu (QS-1/QS-4). Idempotent:
// bỏ qua nếu người này đã có giao dịch gắn nhãn "[seed]" trong tháng hiện tại. KHÔNG dùng cho
// prod (chỉ được gọi ở nhánh DEV; prod đã return ở nhánh tùy chỉnh phía trên).
func seedDevTransactions(ctx context.Context, db *gorm.DB, householdID, userID uuid.UUID) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)

	var existing int64
	must(db.WithContext(ctx).Table("transactions").
		Where("household_id = ? AND created_by = ? AND transaction_date >= ? AND transaction_date < ? AND description LIKE '[seed]%'",
			householdID, userID, monthStart, nextMonth).
		Count(&existing).Error)
	if existing > 0 {
		return // đã seed cho tháng này → idempotent
	}

	// Pluck vào []uuid.UUID (scan một uuid.UUID lẻ bị GORM hiểu nhầm là []uint8).
	pluckID := func(table, where string, args ...any) (uuid.UUID, bool) {
		var ids []uuid.UUID
		err := db.WithContext(ctx).Table(table).Where(where, args...).Limit(1).Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return uuid.Nil, false
		}
		return ids[0], true
	}
	accountID, okA := pluckID("accounts", "household_id = ?", householdID)
	if !okA {
		log.Printf("bỏ qua seed giao dịch (không có tài khoản cho hộ %s)", householdID)
		return
	}
	incomeCat, okI := pluckID("categories", "household_id = ? AND type = ?", householdID, "INCOME")
	expenseCat, okE := pluckID("categories", "household_id = ? AND type = ?", householdID, "EXPENSE")
	if !okI || !okE {
		log.Printf("bỏ qua seed giao dịch (thiếu danh mục Thu/Chi cho hộ %s)", householdID)
		return
	}

	// Ngày trong tháng (kẹp về giữa tháng để không vượt hiện tại / không vào tương lai).
	when := monthStart.AddDate(0, 0, 4)
	if when.After(now) {
		when = now
	}
	insert := func(amount float64, typ string, catID uuid.UUID, desc string) {
		must(db.WithContext(ctx).Exec(
			`INSERT INTO transactions (id, household_id, created_by, amount, type, category_id, account_id, description, transaction_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			uuid.New(), householdID, userID, amount, typ, catID, accountID, desc, when).Error)
	}
	insert(12000000, "INCOME", incomeCat, "[seed] Thu nhập tháng")
	insert(450000, "EXPENSE", expenseCat, "[seed] Chi tiêu")
}

func upsertUser(ctx context.Context, db *gorm.DB, email, name, password string) *usermodel.User {
	var u usermodel.User
	err := db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err == nil {
		return &u
	}
	if err != gorm.ErrRecordNotFound {
		log.Fatal(err)
	}
	hash, err := hasher.HashPassword(password)
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
