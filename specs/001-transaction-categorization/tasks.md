---
description: "Task list — Phân loại giao dịch (Go + Vue re-implementation)"
---

# Tasks: Phân Loại Giao Dịch (Transaction Categorization) — Go + Vue

**Input**: Design documents from `/specs/001-transaction-categorization/` (re-plan 2026-07-10)

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md) (D1–D12), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Stack**: Go 1.22+ (Gin + GORM + goose + gorilla/websocket, layout `learn_go`) · Vue 3 (Vite/TS/Pinia) · PostgreSQL 16 (docker) · Playwright. Mô hình **sổ chung hộ gia đình**, ngang quyền.

**Tests**: Spec không yêu cầu TDD → không sinh phase test riêng; unit/integration/e2e tập trung ở Polish, kiểm chứng theo 18 kịch bản [quickstart.md](./quickstart.md) (#0–#17).

**Path conventions**: API `src/api/` (module-first `module/<x>/{model,biz,storage,transport}`); web `src/web/src/`; migrations `src/db/migrations/` (goose); e2e `src/web/e2e/`.

> ♻️ **Tái sinh 2026-07-10**: thay danh sách task Flutter cũ (53 task — xem git history). Vì 001 nay đi TRƯỚC 002, Phase 2 gồm cả **nền tảng định danh tối thiểu** (users + login + phạm vi hộ).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Khung monorepo Go + Vue + Postgres dev.

- [X] T001 Scaffold monorepo: `src/api` (go mod init, `main.go`, `main_route.go`, `.env.example`, `.air.toml`, `Dockerfile`), `src/web` (Vite + Vue 3 + TS + Pinia + Vue Router), `src/db/migrations/`, `src/docker-compose.yml` (Postgres 16)
- [X] T002 [P] Common package theo mẫu learn_go: `src/api/common/{app_error.go,app_response.go,paging.go,sql_model.go,const.go}` (format lỗi `{error:{code,message,field?}}` — contracts/category-api.md)
- [X] T003 [P] App context `src/api/component/appctx/app_context.go` (GORM db, secret, pubsub, wshub — interface để inject vào transport)
- [X] T004 [P] Web nền: Vite proxy `/api` + `/ws` → :8080 trong `src/web/vite.config.ts`; fetch wrapper + parse app_response trong `src/web/src/api/client.ts`; layout mobile-first + router khung trong `src/web/src/router/index.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: DB schema (goose) + định danh tối thiểu (users/login/phạm vi hộ) + realtime hub — mọi user story phụ thuộc.

**⚠️ CRITICAL**: Không bắt đầu US nào trước khi phase này xong. Migrations áp theo thứ tự 00001 → 00005.

- [X] T005 Goose migrations theo `contracts/db-schema.sql`: `src/db/migrations/00001_users.sql`, `00002_households.sql` (+`household_members`), `00003_categories.sql`, `00004_transactions_min.sql`, `00005_categorization_rules.sql` (mỗi file `-- +goose Up/Down`; CHECK/FK/UNIQUE/index như contract)
- [X] T006 [P] `src/api/component/hasher/bcrypt.go` (bcrypt — KHÔNG md5) và `src/api/component/tokenprovider/jwt/jwt.go` (HS256, hạn 7d, secret env — research D4)
- [X] T007 Module user: `src/api/module/user/{model,storage,biz,transport/ginuser}/` — `POST /api/auth/login` (401 INVALID_CREDENTIALS), `POST /api/auth/logout`, `GET /api/me` (409 NO_HOUSEHOLD khi chưa thuộc hộ) theo contracts/category-api.md §Auth (phụ thuộc T005, T006)
- [X] T008 Middleware `src/api/middleware/{authenticate.go,household_scope.go,recover.go}`: JWT cookie → user; tra membership gắn `household_id` vào context; ngoài hộ → 404 (research D5)
- [X] T009 Module household: `src/api/module/household/{model,storage,biz}/` — membership + hàm `SeedDefaultCategories(householdID)` (bộ mặc định research D9, `is_default=true`)
- [X] T010 Seed dev `src/api/cmd/seed/main.go`: Alice/Bob (hộ "Gia đình A"), Carol (hộ "Gia đình B"), mật khẩu `Password123!`, gọi SeedDefaultCategories cho mỗi hộ (KHÔNG phải migration — D9)
- [X] T011 Realtime: `src/api/component/{pubsub,subscriber,wshub}/` — local pubsub → hub WS theo household; endpoint `WS /ws` (auth cookie khi handshake) phát `categories_changed`/`transactions_changed` (D8)
- [X] T012 [P] Web auth: store `src/web/src/stores/auth.ts` + view `src/web/src/views/LoginView.vue` + router guard (chưa đăng nhập → /login; bootstrap qua GET /api/me)
- [X] T013 [P] Web WS client: composable `src/web/src/composables/useInvalidation.ts` (kết nối /ws, reconnect, nhận event → refetch store tương ứng; fallback refetch khi tab focus)

**Checkpoint**: `go run ./cmd/seed` + login Alice qua UI hoạt động (quickstart #0) — nền tảng sẵn sàng.

---

## Phase 3: User Story 1 - Bắt buộc phân loại theo đúng loại Thu/Chi (Priority: P1) 🎯 MVP

**Goal**: Nhập giao dịch phải chọn danh mục cùng loại; bộ mặc định sẵn sàng; sổ tối thiểu hiển thị người nhập. (UC-CAT-01, UC-CAT-07)

**Independent Test**: quickstart #5, #6 (lọc theo loại; chặn lưu thiếu danh mục) + #14 (authorship).

- [X] T014 [P] [US1] Module category (đọc): `src/api/module/category/{model,storage}/` — model GORM + storage list theo (household, type, include_hidden), dựng cây cha/con
- [X] T015 [US1] Biz + transport: `src/api/module/category/biz/list_categories.go` + `transport/gincategory/list.go` — `GET /api/categories` (contracts §Categories) (phụ thuộc T014)
- [X] T016 [US1] Module transaction (tối thiểu): `src/api/module/transaction/{model,storage,biz,transport}/` — `POST /api/transactions` validate amount > 0, description ≤ 255, category bắt buộc + cùng loại + cùng hộ (CATEGORY_REQUIRED / CATEGORY_TYPE_MISMATCH / AMOUNT_INVALID / DESCRIPTION_TOO_LONG); `created_by` từ phiên; publish `transactions_changed`
- [X] T017 [US1] `GET /api/transactions` (paging 50, embed `category_name` + `created_by_name`) trong `src/api/module/transaction/{storage,transport}/list.go`
- [X] T018 [P] [US1] Web: store `src/web/src/stores/categories.ts` + component `src/web/src/components/CategoryPicker.vue` (lọc theo loại, nhóm cha/con, loại trừ hidden)
- [X] T019 [US1] Web: view `src/web/src/views/TransactionEntryView.vue` (chọn loại → picker lọc; chặn lưu thiếu danh mục; hiển thị lỗi theo trường) + `src/web/src/views/LedgerView.vue` tối thiểu (mới nhất trước, tên người nhập) + store `src/web/src/stores/transactions.ts`

**Checkpoint**: US1 pass quickstart #5/#6/#14 — MVP demo được (đăng nhập → nhập giao dịch có danh mục đúng loại).

---

## Phase 4: User Story 2 - Tự tạo và quản lý danh mục (Priority: P2)

**Goal**: CRUD danh mục (mặc định đối xử như tự tạo), cảnh báo trùng tên, xóa an toàn với gán lại, ẩn/bỏ ẩn, chống ghi đè thầm lặng. (UC-CAT-02/04/05/06)

**Independent Test**: quickstart #1–#4, #8–#11, #16, #17.

- [X] T020 [P] [US2] Biz CreateCategory trong `src/api/module/category/biz/create_category.go`: type bắt buộc; cảnh báo trùng tên (NAME_DUPLICATE_WARNING + `confirm_duplicate`) — `POST /api/categories`
- [X] T021 [US2] Biz UpdateCategory trong `src/api/module/category/biz/update_category.go`: name/icon/is_hidden; TYPE_IMMUTABLE khi gửi type; mốc `expected_updated_at` → CONCURRENCY_CONFLICT (409) / RECORD_GONE (404) (D6) — `PATCH /api/categories/:id`
- [X] T022 [US2] Biz DeleteCategory trong `src/api/module/category/biz/delete_category.go`: MỘT DB transaction — kiểm đích cùng loại/cùng hộ (REASSIGN_TYPE_MISMATCH), xử lý con (FR-012), gán lại hoặc xóa giao dịch, xóa danh mục; thiếu mode khi còn giao dịch → CATEGORY_HAS_TRANSACTIONS + counts (D12) — `DELETE /api/categories/:id`; mọi mutation publish `categories_changed`
- [X] T023 [P] [US2] Web: views `src/web/src/views/{CategoryManageView,CategoryFormView}.vue` (tạo/sửa/ẩn; dialog xác nhận trùng tên; xử lý conflict → tải lại)
- [X] T024 [US2] Web: component `src/web/src/components/DeleteReassignDialog.vue` (chọn đích cùng loại; hiện số giao dịch/con bị ảnh hưởng; xử lý RECORD_GONE)

**Checkpoint**: US1 + US2 độc lập — quản lý danh mục đầy đủ, đồng bộ giữa thành viên ≤ 5s.

---

## Phase 5: User Story 3 - Danh mục con một cấp (Priority: P3)

**Goal**: Con kế thừa loại cha, đúng 1 cấp; gán được cả cha lẫn con. (UC-CAT-03)

**Independent Test**: quickstart #7.

- [X] T025 [US3] Mở rộng biz Create/Update category trong `src/api/module/category/biz/`: `parent_id` phải là danh mục gốc (NESTING_TOO_DEEP), con kế thừa type cha (PARENT_TYPE_MISMATCH nếu lệch) — FR-010/011 *(gộp trong create_category.go — đã verify NESTING_TOO_DEEP / PARENT_TYPE_MISMATCH / kế thừa loại)*
- [X] T026 [US3] Web: CategoryFormView hỗ trợ chọn cha (khóa loại theo cha); CategoryPicker/Manage hiển thị nhóm cha → con; gán giao dịch vào cha trực tiếp vẫn hợp lệ (FR-019)

**Checkpoint**: US1–US3 độc lập.

---

## Phase 6: User Story 4 - Gợi ý danh mục (Priority: P4)

**Goal**: Gợi ý rule-based + học lịch sử chung hộ; luôn ghi đè được. (UC-CAT-08)

**Independent Test**: quickstart #12.

- [ ] T027 [P] [US4] Storage rules + biz SuggestCategory trong `src/api/module/category/{storage/rule_storage.go,biz/suggest_category.go}`: normalize mô tả, match keyword cùng hộ + cùng loại, ưu tiên match_count — `GET /api/categories/suggest` (D7)
- [ ] T028 [US4] Học lịch sử: sau CreateTransaction có description, upsert rule (unique household+keyword+category, match_count++) trong `src/api/module/transaction/biz/create_transaction.go` (phụ thuộc T016, T027)
- [ ] T029 [US4] Web: component `src/web/src/components/SuggestionChip.vue` trong TransactionEntryView (chấp nhận điền sẵn / chọn khác thì ghi đè)

**Checkpoint**: Cả 4 user story hoạt động độc lập.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Test tự động các tầng, e2e Playwright, đồng bộ tài liệu.

- [ ] T030 [P] Unit biz (Go): validate category/transaction, suggest, delete-reassign, conflict — `src/api/module/category/biz/*_test.go`, `src/api/module/transaction/biz/*_test.go` (storage mock qua interface)
- [ ] T031 [P] Integration storage (Go, Postgres docker + goose up): scope hộ, cây cha/con, delete-reassign nguyên tử, conditional update — `src/api/module/*/storage/*_integration_test.go` (tag `integration`)
- [ ] T032 [P] Transport + middleware (httptest): 401 chưa đăng nhập, 404 ngoài hộ, mã lỗi contract — `src/api/middleware/*_test.go`, `src/api/module/*/transport/*_test.go`
- [ ] T033 [P] Web unit (Vitest): categories store, CategoryPicker filter, form validation — `src/web/src/**/__tests__/`
- [ ] T034 Playwright e2e 18 kịch bản (#0–#17) trong `src/web/e2e/` — multi-context cho #13–#17 (Alice/Bob/Carol); assert đồng bộ ≤ 5s (#13)
- [ ] T035 Chạy toàn bộ quickstart, cập nhật References/History các artifact theo `CLAUDE.md` (BR-001 → implemented nếu đủ; chạy `/speckit-analyze`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: bắt đầu ngay; T002–T004 song song sau T001.
- **Foundational (P2)**: sau Setup. **BLOCKS mọi user story.** Trong phase: T005 → T007/T009 (T006 song song T005); T008 sau T007; T010 sau T009; T011 sau T005; T012/T013 song song sau T004+T007.
- **US1 (P3)**: sau Foundational; T014 → T015 → T016 → T017; T018 song song T015; T019 cuối.
- **US2 (P4)** / **US3 (P5)** / **US4 (P6)**: sau US1 (US2–US4 chạm chung module/category — tuần tự theo ưu tiên hoặc chia dev theo file khác nhau).
- **Polish (P7)**: sau các story mong muốn.

### Parallel Opportunities

- Setup: T002, T003, T004 song song.
- Foundational: T006 ∥ T005; T012/T013 (web) ∥ T009–T011 (api).
- US1: T014+T018 (api model ∥ web store) trước; T016 (transaction) ∥ T015.
- Polish: T030–T033 song song; T034 sau cùng.
- Sau US1: Dev A → US2 (biz mutation), Dev B → US4 (suggest — file riêng), US3 nối sau US2.

---

## Implementation Strategy

### MVP First (Foundational + US1)

1. Phase 1 Setup → 2. Phase 2 Foundational (**T005 migrations + T007 login TRƯỚC TIÊN**) → 3. Phase 3 US1 → 4. **DỪNG & kiểm chứng** quickstart #0, #5, #6, #14 → demo MVP.

### Incremental Delivery

Foundational (định danh) → US1 (phân loại bắt buộc — MVP) → US2 (quản lý danh mục + concurrency) → US3 (danh mục con) → US4 (gợi ý) → Polish (test các tầng + e2e 18 kịch bản + docs). Mỗi story kiểm chứng độc lập theo nhóm kịch bản quickstart trước khi sang story sau.

---

## Notes

- [P] = khác file, không phụ thuộc nhau. Nhãn [US#] gắn task với user story.
- Bất biến nghiệp vụ nằm ở **biz trong DB transaction** (D3); DB giữ CHECK/FK/UNIQUE làm hàng rào cuối — KHÔNG dùng GORM AutoMigrate (D2).
- Feature 002 sẽ mở rộng module/transaction (accounts, sửa/xóa, số dư) — không làm trước ở đây.
- Commit sau mỗi task hoặc nhóm logic; dừng ở mỗi Checkpoint để kiểm chứng story độc lập.
