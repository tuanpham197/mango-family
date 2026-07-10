# Implementation Plan: Phân Loại Giao Dịch (Transaction Categorization) — Go + Vue

**Branch**: `001-transaction-categorization` (re-plan thực hiện trên branch `003-budgeting`, 2026-07-10) | **Date**: 2026-07-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-transaction-categorization/spec.md` · [BR-001](../business-requirements/BR-001.md) · [UC-CAT-01…08](../use-cases/001-transaction-categorization/) · [Entity model](../entities/entity-model.md) · [Re-platform design](../../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md)

> ♻️ **Re-plan (2026-07-10)**: Thay thế hoàn toàn plan Flutter/Supabase cũ (xem git history). Bản Flutter từng implement & verify 53/53 task nhưng code đã gỡ bỏ — đây là kế hoạch **re-implementation** trên stack mới. Spec/BR/use case KHÔNG đổi.

## Summary

Dựng lại tính năng phân loại giao dịch trên stack mới: **hệ thống danh mục Thu/Chi dùng chung trong hộ**
(mặc định + tự tạo, danh mục con một cấp, loại bất biến, ẩn/bỏ ẩn, xóa an toàn với gán lại),
**bắt buộc chọn danh mục cùng loại khi nhập giao dịch** (luồng nhập tối thiểu để kiểm chứng),
**gợi ý danh mục** rule-based + lịch sử chung của hộ, đồng bộ giữa thành viên qua WebSocket.

Vì 001 nay đi TRƯỚC 002 trên stack mới, plan này bao gồm cả **nền tảng định danh tối thiểu**:
bảng `users` (kiêm đăng nhập — bcrypt) + `households`/`household_members` + login JWT cookie +
middleware phạm vi hộ (thay RLS cũ). Vòng đời tài khoản & quản lý hộ đầy đủ vẫn ngoài phạm vi (dev seed).

**Technical approach**: Go API (Gin + GORM) theo layout `learn_go` — mỗi aggregate một module
`model/biz/storage/transport`; bất biến nghiệp vụ thực thi ở tầng `biz` trong DB transaction,
PostgreSQL giữ CHECK/FK/UNIQUE làm hàng rào cuối; migrations SQL thuần qua **goose**; realtime =
pubsub local → **WebSocket hub theo household**; front-end Vue 3 (Vite/TS/Pinia) mobile-first.

## Technical Context

**Language/Version**: Go 1.22+ (api) · TypeScript 5.x / Node 20+ (web)

**Primary Dependencies**: API: `gin-gonic/gin`, `gorm.io/gorm` + `gorm.io/driver/postgres`, `pressly/goose/v3`, `gorilla/websocket`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `google/uuid` · Web: Vue 3, Vite, Pinia, Vue Router, native `fetch` + WebSocket client

**Storage**: PostgreSQL 16 tự quản (docker compose cho dev) — bảng `users` (kiêm credentials), `households`, `household_members`, `categories`, `transactions` (phần phân loại), `categorization_rules`; migrations goose tại `src/db/migrations/`

**Testing**: Go: `go test` + `testify` (unit tầng biz), integration storage với Postgres thật (docker), `httptest` cho transport · Web: Vitest (unit/component) · E2E: **Playwright** (Chromium, multi-context cho đa thành viên) theo quickstart

**Target Platform**: Web responsive **mobile-first** (Chromium/Firefox/Safari hiện đại); dev-test trên Chromium

**Project Type**: Web app monorepo — `src/api` (Go) + `src/web` (Vue) + `src/db` (migrations), multi-user chia sẻ theo hộ gia đình

**Performance Goals**: Tạo danh mục ≤ 3 bước & < 30s (SC-005); tìm & chọn danh mục < 10s (SC-006); thay đổi của thành viên khác hiển thị ≤ 5 giây (đồng bộ WS — nhất quán SC-006/002); danh sách danh mục mượt với hàng trăm mục

**Constraints**: Loại danh mục bất biến sau tạo (FR-005); danh mục con đúng 1 cấp, kế thừa loại (FR-010/011); không giao dịch mồ côi — xóa danh mục phải gán lại/xóa giao dịch trong MỘT transaction (FR-008/009/012, SC-007); giao dịch bắt buộc danh mục cùng loại (FR-013/014); cô lập theo hộ ở tầng API (FR-018); ngang quyền (FR-021); không ghi đè thầm lặng (`updated_at` — R13 cũ); gợi ý không AI/ML (FR-015)

**Scale/Scope**: Mỗi hộ 2–8 thành viên, ≤ vài trăm danh mục, hàng nghìn giao dịch; 8 use case UC-CAT-01…08 + nền tảng định danh tối thiểu

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` vẫn là **bản mẫu chưa phê chuẩn** (toàn placeholder) — không có nguyên tắc ràng buộc.

- **Kết luận**: PASS (không có gate ràng buộc).
- **Re-check sau Phase 1**: PASS — thiết kế theo mẫu learn_go đã chốt ở design re-platform; mỗi aggregate một module, không phát sinh độ phức tạp cần biện minh.

## Project Structure

### Documentation (this feature)

```text
specs/001-transaction-categorization/
├── plan.md              # This file
├── research.md          # Phase 0 — quyết định D1…D12 (stack mới + carry-over nghiệp vụ)
├── data-model.md        # Phase 1 — bảng/constraint + nơi thực thi bất biến (biz vs DB)
├── quickstart.md        # Phase 1 — setup Go+Vue + 17 kịch bản kiểm chứng (giữ nguyên nghiệp vụ)
├── contracts/
│   ├── db-schema.sql            # Schema tham chiếu (goose-style) — users/households/categories/…
│   └── category-api.md          # REST API + WS events (thay category-repository.md cũ)
└── tasks.md             # Phase 2 (/speckit-tasks — tái sinh, KHÔNG tạo ở bước này)
```

### Source Code (monorepo `src/` — theo layout learn_go)

```text
src/
├── api/
│   ├── main.go · main_route.go       # entry + đăng ký route các module
│   ├── .env.example · .air.toml · Dockerfile
│   ├── common/                       # app_error, app_response, paging, sql_model (base), const
│   ├── component/
│   │   ├── appctx/                   # app context: db (GORM), secret, pubsub, ws hub
│   │   ├── tokenprovider/jwt/        # phát/xác thực JWT (cookie HttpOnly)
│   │   ├── hasher/                   # bcrypt (KHÔNG dùng md5 như repo mẫu)
│   │   ├── pubsub/ + subscriber/     # local pubsub → đẩy event sang wshub
│   │   └── wshub/                    # WebSocket hub theo household
│   ├── middleware/                   # authenticate (JWT→user), household scope, recover
│   ├── module/
│   │   ├── user/                     # model/biz/storage/transport-ginuser: login, logout, me
│   │   ├── household/                # membership + helper seed danh mục mặc định khi tạo hộ
│   │   ├── category/                 # CRUD + ẩn + con 1 cấp + delete-reassign + suggest (rules)
│   │   └── transaction/              # TỐI THIỂU: tạo + list (kiểm chứng gán danh mục — FR-013/014/019/022)
│   └── cmd/seed/                     # seed dev: Alice/Bob (hộ A), Carol (hộ B) + danh mục mặc định
├── web/
│   ├── src/{views,components,stores,composables,api,router}/
│   │   # views: Login, CategoryManage, CategoryForm, TransactionEntry (tối thiểu), Ledger (tối thiểu)
│   │   # components: CategoryPicker (lọc theo loại), SuggestionChip, DeleteReassignDialog
│   └── e2e/                          # Playwright specs (17 kịch bản quickstart)
├── db/migrations/                    # goose: 00001_users … 00005_categorization_rules
└── docker-compose.yml                # Postgres 16 dev
```

**Structure Decision**: Mỗi aggregate một module theo learn_go; `categorization_rules` nằm TRONG
`module/category` (storage + biz suggest) vì chỉ phục vụ gợi ý danh mục. `module/transaction` ở 001
chỉ là luồng tối thiểu để kiểm chứng phân loại — 002 sẽ mở rộng thành vòng đời đầy đủ (accounts,
sửa/xóa, số dư). Seed danh mục mặc định là **logic app** gắn với việc tạo hộ (`module/household`),
không phải migration — dev gọi qua `cmd/seed`.

## Dependency & thứ tự nền tảng

1. **Hạ tầng chung**: docker compose Postgres + goose migrations `00001_users` → `00002_households`
   (+`household_members`) → `00003_categories` → `00004_transactions_min` → `00005_categorization_rules`.
2. **Định danh tối thiểu (TRƯỚC TIÊN — kế thừa nguyên tắc "users FIRST")**: `module/user` (login
   email+bcrypt → JWT cookie; GET /api/me) + middleware authenticate + household scope + `cmd/seed`.
3. **Danh mục**: `module/category` đầy đủ (US1→US3) + WS invalidation.
4. **Gợi ý**: rules + học từ lịch sử (US4).
5. **Nhập giao dịch tối thiểu**: `module/transaction` (kiểm chứng FR-013/014/019/022).
6. **Quản lý hộ** (tạo/mời/tham gia) vẫn là tiền đề ngoài phạm vi — dev seed thay thế.

## Tác động tới artifact khác

- `specs/entities/entity-model.md` — cấu trúc thực thể KHÔNG đổi (đã trung lập hóa 2026-07-10); không cần sửa thêm.
- `specs/use-cases/001-*` + sơ đồ — nghiệp vụ không đổi; README đã ghi trạng thái re-implementation.
- Feature 002: kế thừa toàn bộ nền tảng (users/auth/household/WS) từ plan này; re-plan 002 sau khi 001 chốt.
- `CLAUDE.md` — cập nhật con trỏ plan hiện hành (bước agent context của Phase 1).

## Complexity Tracking

> Không áp dụng — Constitution Check PASS, không có vi phạm cần biện minh.
