---
description: "Task list — Ghi chép thu nhập & chi phí (Go + Vue re-implementation)"
---

# Tasks: Ghi Chép Thu Nhập và Chi Phí (Income & Expense Tracking) — Go + Vue

**Input**: Design documents from `/specs/002-transaction-tracking/` (re-plan 2026-07-10)

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md) (v4), [research.md](./research.md) (D13–D19), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md) · **Nền tảng**: [feature 001 (Go+Vue)](../001-transaction-categorization/tasks.md) — tối thiểu T001–T019 của 001 phải xong

**Stack**: Go 1.22+ (Gin + GORM + goose + gorilla/websocket) · Vue 3 (Vite/TS/Pinia) · PostgreSQL 16 · Playwright — như 001, không dependency mới.

**Tests**: Spec không yêu cầu TDD → không sinh phase test riêng; kiểm chứng end-to-end theo 23 kịch bản [quickstart.md](./quickstart.md); test các tầng ở Polish.

**Path conventions**: như 001 — API `src/api/module/<x>/{model,biz,storage,transport}`; web `src/web/src/`; migrations `src/db/migrations/` (goose); e2e `src/web/e2e/`.

> ♻️ **Tái sinh 2026-07-10**: thay danh sách task Flutter cũ (T001–T032 — xem git history). US5 (đăng nhập/định danh) đã được **feature 001 dựng** (D19) — ở đây chỉ kiểm chứng lại.

---

## Phase 1: Setup

**Purpose**: Mã lỗi mới + khung module account.

- [X] T001 Bổ sung mã lỗi 002 vào `src/api/common/const.go`: `ACCOUNT_REQUIRED`, `ACCOUNT_HOUSEHOLD_MISMATCH`, `FUTURE_DATE_NOT_ALLOWED` (bộ lỗi chung CONCURRENCY_CONFLICT/RECORD_GONE đã có từ 001)
- [X] T002 [P] Scaffold `src/api/module/account/{model,storage,biz,transport/ginaccount}/` (khung theo mẫu learn_go)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Schema accounts + view số dư + transactions v2 + seed mặc định — mọi user story phụ thuộc.

**⚠️ CRITICAL**: Không bắt đầu US nào trước khi phase này xong. Thứ tự migration: **T003 (00006) → T004 (00007)**.

- [X] T003 Goose migration `src/db/migrations/00006_accounts.sql`: bảng `accounts` (CHECK type, index household) + view **`account_balances`** (bọc `-- +goose StatementBegin/End`) theo `contracts/db-schema.sql` (D13, D14)
- [X] T004 Goose migration `src/db/migrations/00007_transactions_v2.sql`: `account_id` (nullable → backfill về tài khoản mặc định của hộ → NOT NULL, FK restrict) + `updated_at` + `idx_transactions_account` (D17)
- [X] T005 Mở rộng `SeedDefaults` trong `src/api/module/household/biz/seed_defaults.go`: tạo tài khoản mặc định **"Tiền mặt" (CASH)** cùng chỗ seed danh mục (D13); `src/api/cmd/seed/main.go` backfill cho hộ dev
- [X] T006 Module account: model GORM (+ model read-only `AccountBalance` map view) trong `src/api/module/account/model/`; storage list theo hộ join balance; biz `ListAccounts`; transport `GET /api/accounts` (contracts §Accounts)
- [X] T007 Mở rộng transaction model `src/api/module/transaction/model/transaction.go`: `account_id`, `updated_at` (GORM hook làm mới khi update — D17); embed thêm `account_name` trong response
- [X] T008 [P] Web: store `src/web/src/stores/accounts.ts` + component `src/web/src/components/AccountBalanceChip.vue` (hiển thị số dư từ GET /api/accounts)
- [X] T009 Subscriber `src/api/component/subscriber/`: mọi mutation transaction publish thêm `accounts_changed` (số dư đổi theo — contracts §WebSocket); web `useInvalidation` refetch accounts store

**Checkpoint**: goose up + seed xong — mỗi hộ có "Tiền mặt", `GET /api/accounts` trả balance 0.

---

## Phase 3: User Story 5 - Đăng nhập và định danh người dùng (Priority: P1 — nền tảng, đã dựng ở 001) 🎯

**Goal**: Xác nhận nền tảng định danh của 001 phủ đúng US5/FR-015/FR-016 của spec 002. (UC-TRK-01, D19)

**Independent Test**: quickstart #1–#3.

- [X] T010 [US5] Kiểm chứng nền tảng định danh trên luồng 002: Playwright specs `src/web/e2e/auth.spec.ts` (đăng nhập đúng hộ + đúng tên; guard chưa đăng nhập; sai mật khẩu an toàn — quickstart #1–#3); sửa các phát sinh nhỏ nếu có trong `src/api/module/user/` / `src/web/src/stores/auth.ts`

**Checkpoint**: US5 pass #1–#3 — không cần code mới nếu 001 đã chuẩn.

---

## Phase 4: User Story 1 - Nhập giao dịch thu/chi nhanh và hợp lệ (Priority: P1) 🎯 MVP

**Goal**: Nhập ≤ 15s với xác thực đầy đủ (thêm tài khoản + chặn ngày tương lai so với 001); số dư cập nhật đúng. (UC-TRK-02)

**Independent Test**: quickstart #4–#10, #23.

- [X] T011 [P] [US1] Mở rộng biz `src/api/module/transaction/biz/create_transaction.go`: `account_id` bắt buộc + cùng hộ (ACCOUNT_REQUIRED/ACCOUNT_HOUSEHOLD_MISMATCH); chặn `transaction_date > now()+1d` (FUTURE_DATE_NOT_ALLOWED — D15); giữ nguyên validate 001 (amount/category/description)
- [X] T012 [US1] Transport POST cập nhật body theo `contracts/transaction-api.md`; publish `transactions_changed` + `accounts_changed` (phụ thuộc T009, T011)
- [X] T013 [US1] Web: `src/web/src/views/TransactionFormView.vue` (chế độ TẠO) — số tiền, loại Thu/Chi, `CategoryPicker` + `SuggestionChip` tái dùng từ 001, chọn tài khoản (chọn sẵn khi hộ chỉ có 1 — FR-006), date picker `max = hôm nay`, mô tả `maxlength=255`, lỗi theo trường; **thay** TransactionEntryView tối thiểu của 001 (route cũ trỏ về form mới)
- [X] T014 [US1] Store `src/web/src/stores/transactions.ts`: submit + refetch sổ/balances sau lưu; **chống double-submit** (disable khi pending; thử lại sau lỗi mạng không tạo trùng — quickstart #23)

**Checkpoint**: US5 + US1 = MVP (đăng nhập → nhập giao dịch hợp lệ → số dư đúng).

---

## Phase 5: User Story 2 - Xem sổ giao dịch chung của hộ (Priority: P2)

**Goal**: Sổ mới-nhất-trước hiển thị TÊN người nhập; realtime ≤ 5s; điểm vào sửa/xóa. (UC-TRK-03)

**Independent Test**: quickstart #11–#13.

- [X] T015 [P] [US2] Mở rộng storage/biz list `src/api/module/transaction/storage/list.go`: embed `users(display_name)` + `categories(name,icon)` + `accounts(name)` một round-trip; `ORDER BY transaction_date DESC, id DESC`; paging offset 50 + total (D16)
- [X] T016 [US2] Web: nâng cấp `src/web/src/views/LedgerView.vue` thành sổ chính thức — infinite scroll, mỗi dòng số tiền/loại/danh mục/ngày/**tên người nhập** (email nếu thiếu tên), empty state hướng đi nhập, realtime refetch qua `useInvalidation` (≤ 5s — SC-006)
- [X] T017 [US2] Router `src/web/src/router/index.ts`: Ledger làm màn hình chính `/`; app bar điều hướng Quản lý danh mục + Nhập giao dịch + badge người dùng/đăng xuất

**Checkpoint**: US1 + US2 độc lập; sổ chung minh bạch.

---

## Phase 6: User Story 3 - Chỉnh sửa giao dịch (Priority: P3)

**Goal**: Sửa mọi trường với xác thực như nhập mới; đổi loại buộc chọn lại danh mục; số dư tính lại; không ghi đè thầm lặng. (UC-TRK-04)

**Independent Test**: quickstart #14–#18.

- [X] T018 [P] [US3] Biz `src/api/module/transaction/biz/update_transaction.go`: validate như tạo mới + mốc `expected_updated_at` — conditional UPDATE, phân biệt CONCURRENCY_CONFLICT (409, kèm data mới nhất) vs RECORD_GONE (404) (D17); thêm `GET /api/transactions/:id` + `PATCH` trong transport
- [X] T019 [US3] Web: `TransactionFormView` chế độ SỬA — prefill qua GET /:id; đổi loại → `CategoryPicker` coi danh mục cũ là "chưa chọn" (D18); xử lý CONCURRENCY_CONFLICT (hiện dữ liệu mới, cho sửa tiếp) và RECORD_GONE (thông báo + về sổ) trong `src/web/src/views/TransactionFormView.vue`
- [X] T020 [US3] Điểm vào sửa từ sổ (tap dòng giao dịch → form sửa) trong `src/web/src/views/LedgerView.vue`

**Checkpoint**: US1 + US2 + US3 độc lập.

---

## Phase 7: User Story 4 - Xóa giao dịch có xác nhận (Priority: P4)

**Goal**: Xóa luôn qua cảnh báo/xác nhận; số dư hoàn tác đúng; xử lý đã-bị-xóa-trước. (UC-TRK-05)

**Independent Test**: quickstart #19–#21.

- [X] T021 [P] [US4] Biz `src/api/module/transaction/biz/delete_transaction.go`: DELETE theo id (+`expected_updated_at` nếu có); 0 hàng → RECORD_GONE (thông báo nhẹ); publish 2 event; transport `DELETE /api/transactions/:id`
- [X] T022 [US4] Web: `src/web/src/components/DeleteTransactionDialog.vue` — cảnh báo nêu rõ xóa vĩnh viễn (SC-005), hủy không đổi gì; RECORD_GONE → toast + làm tươi sổ; gắn vào LedgerView + TransactionFormView

**Checkpoint**: Cả 5 user story hoạt động độc lập.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Test các tầng, e2e 23 kịch bản, đồng bộ tài liệu.

- [X] T023 [P] Unit biz (Go): validate account/ngày tương lai, phân biệt conflict/gone, đổi loại — `src/api/module/transaction/biz/*_test.go`, `src/api/module/account/biz/*_test.go`
- [X] T024 [P] Integration storage (Postgres docker, tag `integration`): view `account_balances` khớp tổng sau chuỗi thêm/sửa/đổi-tài-khoản/xóa (SC-004); conditional update — `src/api/module/*/storage/*_integration_test.go`
- [X] T025 [P] Transport httptest: mã lỗi theo `contracts/transaction-api.md`; 404 ngoài hộ cho accounts/transactions — `src/api/module/*/transport/*_test.go`
- [X] T026 [P] Vitest: form validation (thiếu tài khoản/danh mục, ngày tương lai), store xử lý conflict — `src/web/src/**/__tests__/`
- [X] T027 Playwright e2e 23 kịch bản trong `src/web/e2e/` — multi-context #11/#13/#16–#18/#21 (Alice/Bob/Carol); #22 đối chiếu số dư qua `GET /api/accounts`; #23 offline (context.setOffline) → thử lại không trùng
- [ ] T028 Chạy toàn bộ quickstart 001+002; cập nhật References/History các artifact theo `CLAUDE.md` (BR-002 → implemented nếu đủ; entity-model/data-model nếu lệch; chạy `/speckit-analyze`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: sau khi 001 xong Foundational + US1 (T001–T019 của 001).
- **Foundational (P2)**: sau Setup. **BLOCKS mọi user story.** Trong phase: T003 → T004 (migration tuần tự); T005 sau T003; T006 sau T003; T007 sau T004; T008 ∥ T006; T009 sau T006.
- **US5 (P3)**: ngay sau Foundational — chỉ kiểm chứng (nền tảng có từ 001).
- **US1 (P4)**: sau Foundational; T011 → T012 → T013 → T014.
- **US2 (P5)**: sau Foundational (độc lập US1 về API; UI dùng chung route); T015 → T016 → T017.
- **US3 (P6)** / **US4 (P7)**: sau US1 (form) + US2 (điểm vào sổ).
- **Polish (P8)**: sau các story mong muốn.

### Parallel Opportunities

- Setup: T001 ∥ T002.
- Foundational: T005/T006 ∥ T008; T009 nối sau.
- US1: T011 (biz) ∥ chuẩn bị T013 (form UI); Dev A → US1 (form), Dev B → US2 (ledger) sau Foundational.
- Polish: T023–T026 song song; T027 sau cùng.

---

## Implementation Strategy

### MVP First (US5 + US1)

1. Phase 1 Setup → 2. Phase 2 Foundational (**T003→T004 migrations trước tiên**) → 3. Phase 3 US5 (kiểm chứng #1–#3) → 4. Phase 4 US1 → 5. **DỪNG & kiểm chứng** quickstart #4–#10, #23 → demo MVP.

### Incremental Delivery

US5 (kiểm chứng) → US1 (nhập — MVP) → US2 (sổ + realtime) → US3 (sửa + concurrency) → US4 (xóa) → Polish (test + e2e 23 kịch bản + docs). Mỗi story kiểm chứng độc lập theo nhóm kịch bản quickstart trước khi sang story sau.

---

## Notes

- [P] = khác file, không phụ thuộc nhau. Nhãn [US#] gắn task với user story (US5 = nền tảng, đã dựng ở 001).
- Số dư KHÔNG lưu cột riêng — view `account_balances` (D14); mọi kiểm chứng số dư đối chiếu qua view/GET /api/accounts.
- Bất biến ở **biz trong DB transaction** (D3/001); KHÔNG trigger nghiệp vụ, KHÔNG AutoMigrate.
- Quản lý tài khoản đầy đủ (CRUD, chuyển tiền) thuộc BR-005; quản lý hộ thuộc feature riêng — ở đây chỉ seed mặc định.
- Commit sau mỗi task hoặc nhóm logic; dừng ở mỗi Checkpoint để kiểm chứng story độc lập.
