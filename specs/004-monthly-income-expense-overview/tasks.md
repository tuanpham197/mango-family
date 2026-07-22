---
description: "Task list — Màn Tổng Quan Đầy Đủ (Dashboard) & Trang Mặc Định (Go + Vue)"
---

# Tasks: Màn Tổng Quan Đầy Đủ (Dashboard) & Trang Mặc Định — Go + Vue

**Input**: Design documents from `/specs/004-monthly-income-expense-overview/` (plan 2026-07-14)

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md) (v2), [research.md](./research.md) (D27–D33), [data-model.md](./data-model.md), [contracts/](./contracts/) (`overview-api.md`), [quickstart.md](./quickstart.md) (26 kịch bản) · **Nền tảng**: [001](../001-transaction-categorization/tasks.md) + [002](../002-transaction-tracking/tasks.md) + [003](../003-budgeting/tasks.md) — **đã triển khai** (giao dịch/tài khoản+số dư, danh mục, ngân sách, WS, màn Tổng quan `/overview`).

**Stack**: Go 1.22+ (Gin + GORM + gorilla/websocket) · Vue 3 (Vite/TS/Pinia) · PostgreSQL 16 · Playwright — như 001/002/003, **không dependency mới, không migration mới** (chỉ ĐỌC).

**Tests**: Spec không yêu cầu TDD → không sinh phase test riêng; kiểm chứng theo 26 kịch bản [quickstart.md](./quickstart.md); test các tầng ở Polish.

**Path conventions**: API `src/api/module/overview/{model,storage,biz,transport/ginoverview}/`; web `src/web/src/`; e2e `src/web/e2e/`.

> 🆕 **Sinh 2026-07-14**: màn Tổng quan chỉ **ĐỌC** dữ liệu 001/002/003 để suy ra số liệu — KHÔNG sửa module khác, KHÔNG bảng/migration. Một endpoint tổng hợp `GET /api/overview` (D27); ngân sách dùng lại `GET /api/budgets` (003). **Đổi trang mặc định**: `/` = Tổng quan, sổ giao dịch → `/ledger` (D31) → lan truyền cập nhật điều hướng e2e 001/002/003.

---

## Phase 1: Setup

**Purpose**: Khung module overview (chỉ đọc).

- [X] T001 [P] Scaffold `src/api/module/overview/{model,storage,biz,transport/ginoverview}/` (khung theo mẫu learn_go)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Endpoint tổng hợp + khung màn Tổng quan + đổi định tuyến — mọi user story phụ thuộc.

**⚠️ CRITICAL**: Không bắt đầu US nào trước khi phase này xong.

- [X] T002 Overview model DTO trong `src/api/module/overview/model/overview.go`: `OverviewSummary` (`net_worth`, `net_worth_change_percent *float64`, `month{income,expense,net}`, `category_spending[]`, `recent_transactions[]`) + `CategorySpending` (`category_id,category_name,category_hidden,amount,percent`) — **suy ra, không map bảng** (D27, data-model)
- [X] T003 Overview storage `src/api/module/overview/storage/overview.go`: struct `SQLStore` **chỉ SELECT/SUM**; khung + helper cửa sổ tháng dùng lại `budget/model.ResolvePeriod(MONTHLY, now)` (D20/003); (các phương thức đọc cụ thể thêm ở từng US)
- [X] T004 Overview biz `src/api/module/overview/biz/get_overview.go`: `GetOverviewBiz.Get(ctx, householdID, now)` ráp `OverviewSummary` qua các interface đọc (khung, điền dần theo US) + transport `src/api/module/overview/transport/ginoverview/routes.go` `GET /api/overview` + wire route scoped hộ trong `src/api/main_route.go` (D27, contracts) (phụ thuộc T002, T003)
- [X] T005 [P] Web store `src/web/src/stores/overview.ts`: fetch `GET /api/overview`; nghe `transactions_changed`/`accounts_changed`/`budgets_changed`/`categories_changed` → refetch (D33) + shell `src/web/src/views/DashboardView.vue` (khung 6 phần theo dashboard.png, chưa gắn dữ liệu; giữ `OverviewBudgetWidget` 003)
- [X] T006 Web định tuyến + nav: `src/web/src/router/index.ts` — `/` = `DashboardView`, `/ledger` = `LedgerView`, `/overview` redirect → `/`; `src/web/src/App.vue` thanh nav 5 mục **Tổng quan · Giao dịch · ＋ · Ngân sách · Báo cáo** (D31, D32) (phụ thuộc T005)

**Checkpoint**: `GET /api/overview` trả về cấu trúc rỗng-an toàn; mở `/` thấy khung màn Tổng quan; `/ledger` là sổ.

---

## Phase 3: User Story 1 - Vào ứng dụng mặc định ở màn Tổng quan (Priority: P1) 🎯 MVP

**Goal**: Vào app ở gốc / sau đăng nhập không kèm đích → màn Tổng quan; các màn khác vẫn truy cập; liên kết cụ thể được tôn trọng. (quickstart #1–4)

**Independent Test**: quickstart #1–4.

- [X] T007 [US1] Router guard `src/web/src/router/index.ts`: vào gốc/sau đăng nhập không kèm đích → Tổng quan (`/`); tôn trọng `redirect` query cho liên kết cụ thể; `src/web/src/views/LoginView.vue` điều hướng mặc định → `/` (nay là Tổng quan) (D31)
- [X] T008 [US1] Lan truyền điều hướng route mới sang e2e 001/002/003 trong `src/web/e2e/`: đổi `goto('/')` (xem sổ) → `goto('/ledger')`, khẳng định URL sau lưu giao dịch → `/ledger`, đảm bảo `helpers.login` đáp ở màn có `household-name` (DashboardView) — giữ 001/002/003 e2e xanh (D31)

**Checkpoint**: đăng nhập → Tổng quan; sổ ở `/ledger`; e2e 001/002/003 vẫn pass (quickstart #1–4).

---

## Phase 4: User Story 2 - Tóm tắt Thu/Chi tháng này (Priority: P1) 🎯 MVP

**Goal**: Hai thẻ Thu nhập / Chi phí tháng hiện tại của hộ, dưới thẻ tài sản ròng, trên Chi tiêu theo danh mục; tính lại đúng khi giao dịch đổi. (quickstart #5–8)

**Independent Test**: quickstart #5–8.

- [X] T009 [US2] Storage `src/api/module/overview/storage/overview.go`: `MonthIncomeExpense(ctx, householdID, start, endExcl)` → `Σ amount` theo `type IN (INCOME,EXPENSE)` trong cửa sổ tháng, cùng hộ; biz điền `month{income,expense,net}` vào `OverviewSummary` (`get_overview.go`) (FR-003, SC-001)
- [X] T010 [P] [US2] Web `src/web/src/components/IncomeExpenseCards.vue`: hai thẻ "Thu nhập +…" (xanh) / "Chi phí −…" (đỏ) đọc từ store `overview.ts`; nhúng vào `DashboardView` đúng vị trí (dưới tài sản ròng) (FR-003, FR-008)

**Checkpoint**: thẻ Thu/Chi khớp tổng giao dịch tháng; cập nhật khi thêm/sửa/xóa (quickstart #5–8).

---

## Phase 5: User Story 3 - Chi tiêu theo danh mục (trước Ngân sách) (Priority: P1) 🎯 MVP

**Goal**: Liệt kê chi từng danh mục (gộp con) tháng hiện tại kèm số tiền + %, sắp giảm dần, ngay trước phần Ngân sách. (quickstart #9–13)

**Independent Test**: quickstart #9–13.

- [X] T011 [US3] Storage `src/api/module/overview/storage/overview.go`: `CategorySpending(ctx, householdID, start, endExcl)` → `Σ chi EXPENSE` nhóm theo **danh mục cha** (gộp con một cấp — dùng `parent_id`), kèm `category_hidden`; biz tính `percent = round(amount/month.expense×100)`, sắp giảm dần, chỉ `amount>0`, điền `category_spending[]` (D29, FR-004, SC-002)
- [X] T012 [P] [US3] Web `src/web/src/components/CategorySpendingList.vue`: danh sách danh mục (icon/tên, số tiền, %), nhãn "đã ẩn" khi `category_hidden`; empty state; nhúng vào `DashboardView` **ngay trước** `OverviewBudgetWidget` (FR-004, FR-008)

**Checkpoint**: chi theo danh mục đúng số tiền + % (gộp con), sắp giảm dần, đứng trước Ngân sách (quickstart #9–13).

---

## Phase 6: User Story 4 - Tổng tài sản ròng ở đầu màn (Priority: P2)

**Goal**: Thẻ Tổng tài sản ròng (Σ số dư tài khoản) + % so tháng trước, là thẻ đầu tiên dưới lời chào. (quickstart #14–16)

**Independent Test**: quickstart #14–16.

- [X] T013 [US4] Storage `src/api/module/overview/storage/overview.go`: `NetWorth(ctx, householdID)` = `Σ account_balances.balance`; `NetWorthPrevMonthEnd(ctx, householdID, firstOfMonth)` = `Σ initial_balance + Σ bút toán có dấu (transaction_date < firstOfMonth)`; biz tính `net_worth_change_percent` (null khi mẫu số 0 — D28) điền `OverviewSummary` (FR-002, SC-003)
- [X] T014 [P] [US4] Web `src/web/src/components/NetWorthCard.vue`: thẻ xanh lớn "Tổng tài sản ròng {số} đ" + "▲/▼ {..}% so với tháng trước" (ẩn khi null); nhúng vào `DashboardView` **đầu tiên** dưới lời chào (FR-002, FR-008)

**Checkpoint**: tài sản ròng khớp tổng số dư; % so tháng trước hiển thị/ẩn đúng (quickstart #14–16).

---

## Phase 7: User Story 5 - Giao dịch gần đây trên màn Tổng quan (Priority: P2)

**Goal**: Vài giao dịch mới nhất của hộ (tên, danh mục · tài khoản, số tiền có dấu/màu), cuối màn. (quickstart #17–19)

**Independent Test**: quickstart #17–19.

- [X] T015 [US5] Storage `src/api/module/overview/storage/overview.go`: `RecentTransactions(ctx, householdID, limit=5)` dùng lại truy vấn `ListItem` của `module/transaction/storage` (embed tên danh mục · tài khoản · người nhập, mới nhất trước); biz điền `recent_transactions[]` (D30, FR-006)
- [X] T016 [P] [US5] Web `src/web/src/components/RecentTransactionsList.vue`: danh sách xem nhanh (icon Thu/Chi, mô tả/tên, danh mục · tài khoản, số tiền có dấu/màu — read-only); nhúng vào `DashboardView` **cuối màn** (sau Ngân sách) (FR-006, FR-008)

**Checkpoint**: giao dịch gần đây hiển thị đúng thứ tự/định dạng; giao dịch mới lên đầu (quickstart #17–19).

---

## Phase 8: User Story 6 - Bố cục & điều hướng khớp dashboard.png (Priority: P2)

**Goal**: Lời chào + avatar; thứ tự 6 phần đúng mockup; nav 5 mục với "Báo cáo" placeholder; lối phụ vào Danh mục. (quickstart #20–24)

**Independent Test**: quickstart #20–24.

- [X] T017 [US6] Web `src/web/src/views/DashboardView.vue`: lời chào theo buổi (sáng/chiều/tối) + tên + avatar placeholder; chốt **thứ tự đúng**: tài sản ròng → Thu/Chi → Chi tiêu theo danh mục → Ngân sách (`OverviewBudgetWidget` dời xuống sau Chi tiêu theo danh mục) → Giao dịch gần đây; tiêu đề "Ngân sách tháng N" + "Xem tất cả ›" giữ nguyên (FR-005, FR-007, FR-008, SC-005)
- [X] T018 [US6] Web `src/web/src/views/ReportsPlaceholderView.vue` + route `/reports` ("Báo cáo — sắp có", BR-004) gắn vào nav; giữ lối vào **Quản lý danh mục** (`/categories`) qua lối phụ (link ở đầu Tổng quan hoặc trong form tạo danh mục nhanh) (D32, FR-009)

**Checkpoint**: cả 6 phần đúng thứ tự dashboard.png; nav 5 mục; Báo cáo placeholder; Danh mục vẫn truy cập (quickstart #20–24).

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Test các tầng, e2e 26 kịch bản, đồng bộ tài liệu.

- [X] T019 [P] Unit biz (Go) `src/api/module/overview/biz/*_test.go`: cửa sổ tháng, % thay đổi tài sản ròng (kể cả mẫu số 0 → null), % theo danh mục, gộp cây danh mục con, `month.net = income − expense`
- [X] T020 [P] Integration storage (Postgres docker, tag `integration`) `src/api/module/overview/storage/*_integration_test.go`: net worth = Σ số dư; net worth cuối tháng trước; Σ Thu/Chi tháng; chi theo danh mục gộp con khớp SUM; cô lập hộ; loại trừ giao dịch ngoài tháng (SC-001/002/003)
- [X] T021 [P] Transport httptest `src/api/main_004_integration_test.go`: `GET /api/overview` trả đủ field theo `contracts/overview-api.md`; 401 chưa đăng nhập; chỉ trả dữ liệu hộ đang đăng nhập (cô lập hộ)
- [X] T022 [P] Vitest `src/web/src/**/__tests__/`: `NetWorthCard` (định dạng + ẩn % khi null), `IncomeExpenseCards`, `CategorySpendingList` (số tiền/% + nhãn ẩn), `RecentTransactionsList` (dấu/màu), store `overview.ts` (fetch + realtime refetch)
- [X] T023 Playwright e2e 26 kịch bản trong `src/web/e2e/overview*.spec.ts` (US1–US6 + đa thành viên #25 realtime / #26 cô lập hộ); chạy lại toàn bộ e2e 001/002/003 (đã cập nhật điều hướng T008) đảm bảo xanh
- [X] T024 Chạy toàn bộ quickstart 004; cập nhật References/History artifact theo `CLAUDE.md` (spec 003 ghi chú route Tổng quan chuyển `/overview`→`/`; entity-model không đổi — chỉ đọc); chạy `/speckit-analyze` phát hiện lệch spec↔plan↔tasks

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: T001.
- **Foundational (P2)**: sau Setup. **BLOCKS mọi user story.** T002 → T003 → T004; T005 ∥ (T002–T004); T006 sau T005.
- **US1 (P3)**: sau Foundational; T007 → T008.
- **US2 (P4)**: sau Foundational; T009 (storage/biz) → T010 (UI). ∥ được với US3/US4/US5 phần UI (khác file component); nhưng T009/T011/T013/T015 cùng đụng `storage/overview.go` + `get_overview.go` → tuần tự.
- **US3 (P5)**: sau Foundational; T011 → T012.
- **US4 (P6)**: sau Foundational; T013 → T014.
- **US5 (P7)**: sau Foundational; T015 → T016.
- **US6 (P8)**: sau US2–US5 (ráp mọi component vào DashboardView đúng thứ tự); T017 → T018.
- **Polish (P9)**: sau các story mong muốn.

### User Story Dependencies

- **US1 (P1)**: chỉ phụ thuộc Foundational — trang mặc định + điều hướng (độc lập kiểm chứng).
- **US2/US3 (P1)**: phụ thuộc Foundational (endpoint + shell). Cùng US1 tạo MVP (Thu/Chi + Chi tiêu theo danh mục + trang mặc định).
- **US4/US5 (P2)**: mở rộng endpoint + màn; độc lập nhau (khác component), chung file storage/biz nên tuần tự phần backend.
- **US6 (P2)**: ráp bố cục cuối cùng — sau khi các component US2–US5 tồn tại.

### Parallel Opportunities

- Setup: T001.
- Foundational: T005 (web shell/store) ∥ T002–T004 (API).
- Các component UI [P] khác file: T010, T012, T014, T016 có thể chuẩn bị song song (chỉ chờ store T005); phần backend tương ứng (T009/T011/T013/T015) tuần tự do chung `storage/overview.go`+`get_overview.go`.
- Polish: T019–T022 song song; T023 sau cùng; T024 khép lại.

---

## Implementation Strategy

### MVP First (US1 + US2 + US3 — cả ba P1)

1. Phase 1 Setup → 2. Phase 2 Foundational (endpoint + shell + đổi route) → 3. US1 (trang mặc định) → 4. US2 (Thu/Chi) → 5. US3 (Chi tiêu theo danh mục) → **DỪNG & kiểm chứng** quickstart #1–13 → demo MVP (vào thẳng Tổng quan, thấy Thu/Chi tháng + chi theo danh mục).

### Incremental Delivery

MVP (US1+US2+US3) → US4 (tài sản ròng) → US5 (giao dịch gần đây) → US6 (bố cục khớp 100% + nav Báo cáo) → Polish (test tầng + e2e 26 kịch bản + cập nhật e2e 001/002/003 + docs).

---

## Notes

- [P] = khác file, không phụ thuộc nhau. Nhãn [US#] gắn task với user story.
- Mọi số liệu **KHÔNG lưu** — suy ra ở biz khi đọc `GET /api/overview` (D27); ngân sách dùng lại `GET /api/budgets` (003).
- **Không migration, không sửa** module transaction/account/category/budget/household — overview chỉ ĐỌC.
- Đổi trang mặc định (D31) lan truyền tới e2e 001/002/003 (T008) — giữ chúng xanh là điều kiện hoàn thành US1.
- "Báo cáo" là placeholder (BR-004); Danh mục giữ lối phụ (D32) — nếu nghiệp vụ muốn bỏ hẳn Danh mục khỏi điều hướng, cần xác nhận.
- Commit sau mỗi task hoặc nhóm logic; dừng ở mỗi Checkpoint để kiểm chứng story độc lập.
