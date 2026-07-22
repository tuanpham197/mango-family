---
description: "Task list — Thiết lập ngân sách (Budgeting, Go + Vue)"
---

# Tasks: Thiết Lập Ngân Sách (Budgeting) — Go + Vue

**Input**: Design documents from `/specs/003-budgeting/` (plan 2026-07-13)

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md) (v5), [research.md](./research.md) (D20–D26), [data-model.md](./data-model.md), [contracts/](./contracts/) (`budget-api.md` + `db-schema.sql`), [quickstart.md](./quickstart.md) (34 kịch bản) · **Nền tảng**: [feature 001](../001-transaction-categorization/tasks.md) + [feature 002](../002-transaction-tracking/tasks.md) — **đã triển khai** (users/auth/hộ/WS, danh mục, giao dịch đầy đủ, pubsub/wshub).

**Stack**: Go 1.22+ (Gin + GORM + goose + gorilla/websocket) · Vue 3 (Vite/TS/Pinia) · PostgreSQL 16 · Playwright — như 001/002, **không dependency mới**.

**Tests**: Spec không yêu cầu TDD → không sinh phase test riêng; kiểm chứng end-to-end theo 34 kịch bản [quickstart.md](./quickstart.md); test các tầng ở Polish.

**Path conventions**: API `src/api/module/budget/{model,storage,biz,transport/ginbudget,eval}/`; common `src/api/common/`; web `src/web/src/`; migrations `src/db/migrations/` (goose); e2e `src/web/e2e/`.

> 🆕 **Sinh mới 2026-07-13**: ngân sách chỉ **ĐỌC** giao dịch/danh mục để suy ra tiến độ — KHÔNG sửa module transaction/category/account/household. Tiến độ = giá trị suy ra ở biz (D21); cảnh báo qua subscriber pubsub trên `transactions_changed` (D24).
> ➕ **Cập nhật 2026-07-13 (spec v5)**: thêm **T017** (màn Tổng quan/Dashboard chứa tóm tắt ngân sách — FR-014, wireframe màn 1); các task sau T016 dời số +1 (tổng 34 task).

---

## Phase 1: Setup

**Purpose**: Mã lỗi + topic WS mới + khung module budget.

- [X] T001 Bổ sung vào `src/api/common/const.go`: mã lỗi `LIMIT_INVALID`, `CATEGORY_HOUSEHOLD_MISMATCH`, `CATEGORY_NOT_ALLOWED_FOR_TOTAL`, `PERIOD_INVALID`, `BUDGET_DUPLICATE` (tái dùng `CATEGORY_REQUIRED`/`CATEGORY_TYPE_MISMATCH` của 001/002; `CONCURRENCY_CONFLICT`/`RECORD_GONE` đã có) + hằng `TopicBudgetsChanged = "budgets_changed"` (D26)
- [X] T002 [P] Scaffold `src/api/module/budget/{model,storage,biz,transport/ginbudget,eval}/` (khung theo mẫu learn_go)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Schema budgets + budget_alerts + model + engine tính tiến độ suy ra + khung web — mọi user story phụ thuộc.

**⚠️ CRITICAL**: Không bắt đầu US nào trước khi phase này xong. Thứ tự migration: **T003 (00008) → T004 (00009)**.

- [X] T003 Goose migration `src/db/migrations/00008_budgets.sql`: bảng `budgets` (CHECK type/period_type/status, `budget_category_presence`, `budget_period_range`) + **partial unique index** `uq_budget_category_active` & `uq_budget_total_active` + `idx_budgets_household` theo `contracts/db-schema.sql` (D20, D22)
- [X] T004 Goose migration `src/db/migrations/00009_budget_alerts.sql`: bảng `budget_alerts` + `unique(budget_id, period_key, level)` + FK `on delete cascade` + `idx_budget_alerts_budget` (D23)
- [X] T005 Budget models GORM trong `src/api/module/budget/model/`: `Budget` (GORM hook làm mới `updated_at` — D25), `BudgetAlert`, `BudgetProgress` (suy ra — không map bảng); helper `PeriodKey(periodType, now)` + cửa sổ kỳ (MONTHLY `YYYY-MM` · WEEKLY ISO bắt đầu Thứ Hai · ONE_TIME `[start,end]`) (D20)
- [X] T006 Storage đọc tổng chi suy ra `src/api/module/budget/storage/progress.go`: `SUM(transactions.amount)` với `type='EXPENSE'`, cùng hộ, `transaction_date ∈ cửa sổ kỳ`, và CATEGORY → `category_id IN (id + danh mục con một cấp)` / TOTAL → toàn hộ (D21); **chỉ SELECT** transactions/categories, không sửa nguồn
- [X] T007 [P] Web khung: store `src/web/src/stores/budgets.ts` (fetch/create/update/delete + nghe `budgets_changed` & `transactions_changed` → refetch) + route `/budgets` trong `src/web/src/router/index.ts`; xác minh `src/api/component/subscriber/` forward topic mới tự động (D26)

**Checkpoint**: `goose up` (00008→00009) xong; helper `PeriodKey` + truy vấn progress có unit test cơ bản.

---

## Phase 3: User Story 1 - Tạo ngân sách theo danh mục (Priority: P1) 🎯 MVP

**Goal**: Tạo ngân sách gắn danh mục Chi với xác thực đầy đủ; xuất hiện trong danh sách của hộ kèm tiến độ từ chi tiêu hiện có. (UC-BGT-01)

**Independent Test**: quickstart #1–#6.

- [X] T008 [P] [US1] Biz `src/api/module/budget/biz/create_budget.go`: validate `limit_amount > 0` (LIMIT_INVALID); type=CATEGORY → `category_id` bắt buộc + loại EXPENSE + cùng hộ (CATEGORY_REQUIRED / CATEGORY_TYPE_MISMATCH / CATEGORY_HOUSEHOLD_MISMATCH); ONE_TIME `end ≥ start` (PERIOD_INVALID); `created_by` gán từ phiên (không nhận từ client)
- [X] T009 [US1] Storage CRUD `src/api/module/budget/storage/budget.go`: insert; list theo hộ (filter `status`); get by id; bắt vi phạm partial unique → trả `BUDGET_DUPLICATE` kèm `existing_budget_id` (D22) (phụ thuộc T003, T005)
- [X] T010 [US1] Transport `src/api/module/budget/transport/ginbudget/routes.go`: `POST /api/budgets`, `GET /api/budgets`, `GET /api/budgets/:id` theo `contracts/budget-api.md`; publish `budgets_changed`; wire route + phạm vi hộ (middleware 001) trong `src/api/main.go` (phụ thuộc T006, T008, T009)
- [X] T011 [P] [US1] Web `src/web/src/views/BudgetFormView.vue` (chế độ TẠO): chọn danh mục **Chi** (danh sách chỉ EXPENSE — tái dùng `CategoryPicker`), nhập giới hạn, chọn kỳ (MONTHLY/WEEKLY/ONE_TIME + khoảng ngày khi một lần), lỗi theo trường; `BUDGET_DUPLICATE` → chỉ tới ngân sách hiện có
- [X] T012 [US1] Web `src/web/src/views/BudgetListView.vue`: danh sách ngân sách của hộ (tên/danh mục, giới hạn, đã chi, %) + empty state hướng đi tạo; gắn route `/budgets` (phụ thuộc T007, T011)

**Checkpoint**: tạo ngân sách danh mục → thấy trong danh sách kèm tiến độ; chặn giới hạn ≤0 / thiếu danh mục / trùng / ngày sai; cô lập hộ (quickstart #1–#6).

---

## Phase 4: User Story 2 - Theo dõi tiến độ ngân sách (Priority: P1) 🎯 MVP

**Goal**: Hiển thị "đã chi/giới hạn (%)" luôn khớp tổng giao dịch Chi trong kỳ (gồm danh mục con); tính lại đúng khi giao dịch đổi kể cả quá khứ; đồng bộ realtime ≤ 5s; tóm tắt trên màn Tổng quan. (UC-BGT-03)

**Independent Test**: quickstart #7–#14, #34.

- [X] T013 [P] [US2] Biz `src/api/module/budget/biz/list_budgets.go`: với mỗi ngân sách ACTIVE, tính `spent`/`percent` kỳ hiện tại qua storage progress (T006), gắn `period_key`; embed vào response `GET /api/budgets` (D21); EXPENSE-only; danh mục con một cấp gộp vào cha
- [X] T014 [US2] Web `src/web/src/components/BudgetProgressBar.vue`: hiển thị "3.500.000/5.000.000 VNĐ (70%)" + màu theo mức (bình thường / amber ≥80% / đỏ khi vượt — BR-BGT-006/007); gắn vào `BudgetListView`
- [X] T015 [US2] Web realtime trong `src/web/src/stores/budgets.ts`: refetch `GET /api/budgets` khi nhận `transactions_changed`/`budgets_changed` (≤ 5s — SC-006); kiểm chứng tính lại đúng sau sửa/xóa/đổi danh mục giao dịch quá khứ (D21)
- [X] T016 [US2] Web `src/web/src/components/OverviewBudgetWidget.vue`: **component** tóm tắt ngân sách (tiến độ ngân sách tổng nếu có + vài danh mục nổi bật, dùng `BudgetProgressBar`) + lối "Xem tất cả ›"; tái dùng dữ liệu store `budgets.ts`
- [X] T017 [US2] Web `src/web/src/views/DashboardView.vue` (màn **Tổng quan**) + route `/` (đưa route hiện tại của 002 thành mục điều hướng): **nhúng `OverviewBudgetWidget` (T016)** hiển thị tóm tắt ngân sách kỳ hiện tại với mã màu ngưỡng, "Xem tất cả" → `BudgetListView`, thêm mục "Ngân sách" trên thanh điều hướng; theo wireframe **màn 1 · Tổng quan** (FR-014). Các thẻ khác của Dashboard (tài sản ròng / thu-chi tháng / giao dịch gần đây) chỉ là **placeholder** — thuộc feature khác (Assumptions spec); realtime cùng cơ chế T015

**Checkpoint**: MVP (US1 + US2) — tạo + theo dõi tiến độ chính xác & realtime + tóm tắt trên Tổng quan (quickstart #7–#14, #34).

---

## Phase 5: User Story 3 - Nhận cảnh báo ngưỡng & vượt ngân sách (Priority: P2)

**Goal**: Phát cảnh báo in-app 80% & vượt 100% (kèm số tiền vượt) cho mọi thành viên; mỗi mức tối đa 1 lần/kỳ, phát lại sau khi tụt dưới. (UC-BGT-04)

**Independent Test**: quickstart #15–#19.

- [X] T018 [P] [US3] Storage `src/api/module/budget/storage/alert.go`: CRUD `budget_alerts` — get active theo `(budget, period_key)`, insert (unique level), delete theo `(budget, period_key, level)` (D23)
- [X] T019 [US3] Biz máy trạng thái `src/api/module/budget/biz/evaluate_alerts.go`: với mỗi ngân sách ACTIVE tính percent → insert `THRESHOLD_80` (≥80, chưa có), insert `OVER_100` (>100, `over_amount = spent − limit`), delete khi tụt xuống dưới mức (re-arm — US3 #4); **idempotent** (phụ thuộc T006, T018)
- [X] T020 [US3] Evaluator subscriber `src/api/module/budget/eval/evaluator.go`: `Subscribe()` pubsub; nhận `transactions_changed` + sự kiện CRUD ngân sách → `evaluate_alerts` trong DB transaction → `Publish(budgets_changed)`; **KHÔNG** phản ứng `budgets_changed` (tránh vòng lặp — D24); start trong `src/api/main.go`
- [X] T021 [US3] Embed `alerts[]` (mức + over_amount + fired_at) của kỳ hiện tại vào `GET /api/budgets` — mở rộng `list_budgets.go` (T013) theo `contracts/budget-api.md`
- [X] T022 [P] [US3] Web `src/web/src/components/BudgetAlertBadge.vue`: badge/banner in-app amber "đạt 80%" / đỏ "vượt {over_amount}" (đúng số tiền vượt — SC-005); hiện trong `BudgetListView` + `OverviewBudgetWidget`; realtime cho mọi thành viên (SC-006)

**Checkpoint**: cảnh báo phát đúng ngưỡng/vượt, chống trùng + phát lại (quickstart #15–#19).

---

## Phase 6: User Story 4 - Ngân sách tổng chi tiêu (Priority: P2)

**Goal**: Ngân sách cho tổng chi của hộ, không gắn danh mục; song song & độc lập với ngân sách danh mục; dùng chung tiến độ + cảnh báo. (UC-BGT-02)

**Independent Test**: quickstart #20–#22.

- [X] T023 [P] [US4] Biz hỗ trợ `type=TOTAL` trong `create_budget.go`/`list_budgets.go`: TOTAL → `category_id` phải rỗng (CATEGORY_NOT_ALLOWED_FOR_TOTAL nếu gửi kèm); uniqueness `(household, period_type)` ACTIVE (BUDGET_DUPLICATE — D22); nhánh progress TOTAL = mọi EXPENSE của hộ (T006 đã hỗ trợ — kiểm chứng)
- [X] T024 [US4] Web `BudgetFormView`: toggle loại **theo danh mục / tổng** — ẩn ô chọn danh mục khi TOTAL; `BudgetListView` + `OverviewBudgetWidget` hiển thị nhãn "Tổng chi tiêu" cho ngân sách tổng
- [X] T025 [US4] Kiểm chứng cảnh báo + tiến độ ngân sách tổng chạy song song, độc lập ngân sách danh mục (evaluator T020 duyệt mọi ngân sách ACTIVE — không cần code mới; thêm kịch bản test)

**Checkpoint**: ngân sách tổng hoạt động song song ngân sách danh mục (quickstart #20–#22).

---

## Phase 7: User Story 5 - Quản lý ngân sách: sửa và xóa (Priority: P3)

**Goal**: Sửa (giới hạn/kỳ/danh mục) qua cùng bộ xác thực + xóa có xác nhận; ngang quyền; chống ghi đè thầm lặng; cảnh báo tính lại theo giới hạn mới. (UC-BGT-05, UC-BGT-06)

**Independent Test**: quickstart #23–#29.

- [X] T026 [P] [US5] Biz `src/api/module/budget/biz/update_budget.go`: validate như tạo (tái dùng T008/T023) + `expected_updated_at` conditional UPDATE → phân biệt `CONCURRENCY_CONFLICT` (409, kèm data mới) vs `RECORD_GONE` (404) (D25); sau sửa phát sự kiện để evaluator tính lại cảnh báo theo giới hạn mới; thêm `PATCH /api/budgets/:id` transport
- [X] T027 [US5] Biz `src/api/module/budget/biz/delete_budget.go`: DELETE (+`expected_updated_at`); cascade `budget_alerts` (DB FK — D23); 0 hàng → `RECORD_GONE`; thêm `DELETE /api/budgets/:id` transport; publish `budgets_changed`
- [X] T028 [US5] Web: `BudgetFormView` chế độ SỬA (prefill qua `GET /:id`; xử lý 409 hiện data mới / 404 thông báo + về danh sách) + `src/web/src/components/DeleteBudgetDialog.vue` (xác nhận; hủy không đổi; RECORD_GONE → toast + làm tươi); điểm vào sửa/xóa từ `BudgetListView`

**Checkpoint**: cả 6 use case UC-BGT-01…06 hoạt động độc lập (quickstart #23–#29).

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Test các tầng, e2e 34 kịch bản, đồng bộ tài liệu.

- [X] T029 [P] Unit biz (Go): `PeriodKey` (MONTHLY/WEEKLY/ONE_TIME + ranh giới kỳ), subtree danh mục con, tính %, máy trạng thái cảnh báo (chống trùng + phát lại) — `src/api/module/budget/biz/*_test.go`
- [X] T030 [P] Integration storage + evaluator (Postgres docker, tag `integration`): tiến độ khớp 100% tổng EXPENSE sau chuỗi thêm/sửa/xóa/đổi-danh-mục quá khứ (SC-003); partial unique chặn trùng; recompute idempotent (bỏ 1 message vẫn tự lành — D24) — `src/api/module/budget/**/*_integration_test.go`
- [X] T031 [P] Transport httptest: mã lỗi theo `contracts/budget-api.md` (LIMIT_INVALID, BUDGET_DUPLICATE + existing_budget_id, PERIOD_INVALID, CATEGORY_*; 409/404); 404 ngoài hộ — `src/api/module/budget/transport/ginbudget/*_test.go`
- [X] T032 [P] Vitest: `BudgetProgressBar` format %/màu ngưỡng, `OverviewBudgetWidget` (chọn danh mục nổi bật + "Xem tất cả"), `BudgetFormView` validation (giới hạn/kỳ/scope danh mục), store xử lý conflict + realtime — `src/web/src/**/__tests__/`
- [X] T033 Playwright e2e 34 kịch bản trong `src/web/e2e/budget*.spec.ts` — multi-context (#6/#11/#19/#26/#29 Alice/Bob/Carol: cảnh báo & cô lập hộ realtime); #30 danh mục ẩn; #31 kỳ một lần kết thúc; #32 sang kỳ mới; #33 tạo giữa kỳ; **#34 tóm tắt ngân sách + "Xem tất cả" trên màn Tổng quan (FR-014)**
- [X] T034 Chạy toàn bộ quickstart 003; cập nhật References/History artifact theo `CLAUDE.md` (BR-003 status; entity-model nếu lệch); chạy `/speckit-analyze` phát hiện lệch spec↔plan↔tasks

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: sau khi 001 + 002 đã triển khai (đã xong). T001 ∥ T002.
- **Foundational (P2)**: sau Setup. **BLOCKS mọi user story.** Trong phase: **T003 → T004** (migration tuần tự); T005 sau T003; T006 sau T005; T007 ∥ (T003–T006).
- **US1 (P3)**: sau Foundational; T008 → T009 → T010; T011 ∥ T008; T012 sau T011.
- **US2 (P4)**: sau Foundational + US1 (dùng list endpoint & BudgetListView); T013 → T014 → T015; T016 sau T014; **T017 sau T016 + T012** (Dashboard nhúng widget, "Xem tất cả" trỏ BudgetListView).
- **US3 (P5)**: sau US2 (cần progress); T018 → T019 → T020; T021 mở rộng T013; T022 sau T021.
- **US4 (P6)**: sau US1 + US2 (mở rộng create/list/progress cho TOTAL); T023 → T024; T025 sau T020.
- **US5 (P7)**: sau US1 (biz validate tái dùng) + US2 (BudgetListView điểm vào); T026 ∥ T027; T028 sau T026/T027.
- **Polish (P8)**: sau các story mong muốn.

### User Story Dependencies

- **US1 (P1)**: chỉ phụ thuộc Foundational — độc lập kiểm chứng (tạo + danh sách).
- **US2 (P1)**: nối tiếp US1 (hiển thị tiến độ trên danh sách + màn Tổng quan) — cùng US1 tạo MVP.
- **US3 (P2)**: phụ thuộc US2 (progress) để đánh giá cảnh báo.
- **US4 (P2)**: mở rộng US1/US2 cho loại TOTAL — độc lập US3.
- **US5 (P3)**: phụ thuộc US1/US2 — độc lập US3/US4.

### Parallel Opportunities

- Setup: T001 ∥ T002.
- Foundational: T007 (web khung) ∥ T003–T006 (API/DB).
- US1: T008 (biz) ∥ T011 (form UI); sau Foundational Dev A → US1/US2 (đọc), Dev B → US3 (cảnh báo) khi progress sẵn.
- US2: T016 (widget) chuẩn bị ∥ T014/T015; T017 (Dashboard) nối sau.
- US3: T018 (storage) ∥ T022 (badge UI).
- Polish: T029–T032 song song; T033 sau cùng; T034 khép lại.

---

## Parallel Example: User Story 1

```bash
# Sau khi Foundational xong, chạy song song trong US1:
Task: "Biz create_budget trong src/api/module/budget/biz/create_budget.go"   # T008 [P]
Task: "Web BudgetFormView (tạo) trong src/web/src/views/BudgetFormView.vue"   # T011 [P]
```

---

## Implementation Strategy

### MVP First (US1 + US2 — cả hai P1)

1. Phase 1 Setup → 2. Phase 2 Foundational (**T003→T004 migrations trước tiên**) → 3. Phase 3 US1 (tạo + danh sách) → 4. Phase 4 US2 (tiến độ + realtime + tóm tắt Tổng quan) → 5. **DỪNG & kiểm chứng** quickstart #1–#14, #34 → demo MVP (tạo ngân sách danh mục + theo dõi tiến độ + widget Tổng quan).

### Incremental Delivery

US1 (tạo) + US2 (tiến độ + Dashboard) = MVP → US3 (cảnh báo 80%/vượt) → US4 (ngân sách tổng) → US5 (sửa/xóa) → Polish (test tầng + e2e 34 kịch bản + docs). Mỗi story kiểm chứng độc lập theo nhóm kịch bản quickstart trước khi sang story sau.

### Parallel Team Strategy

Sau Foundational: Dev A cầm US1+US2 (đọc/tiến độ/Dashboard), Dev B cầm US3 (cảnh báo/evaluator) ngay khi progress (T006/T013) sẵn sàng, Dev C chuẩn bị US4/US5. Tích hợp độc lập theo topic `budgets_changed`.

---

## Notes

- [P] = khác file, không phụ thuộc nhau. Nhãn [US#] gắn task với user story.
- Tiến độ (spent/percent) **KHÔNG lưu** — tính ở biz khi đọc (D21); mọi kiểm chứng đối chiếu qua `GET /api/budgets`.
- Cảnh báo là **máy trạng thái** trong `budget_alerts` khóa theo `(budget, period_key, level)` — sang kỳ tự sạch; đánh giá bởi **subscriber pubsub** (D24), không đụng module transaction.
- Màn Tổng quan (T017): feature 003 **chỉ sở hữu phần tóm tắt ngân sách** (FR-014); các thẻ khác là placeholder thuộc BR-002/004/005 — không mở rộng phạm vi 003.
- Bất biến ở **biz trong DB transaction** (D3/001) + partial unique index là phòng tuyến cuối; KHÔNG trigger nghiệp vụ, KHÔNG AutoMigrate.
- Ngân sách chỉ **ĐỌC** giao dịch/danh mục — không sửa module transaction/category/account/household.
- Ngưỡng cấu hình được + push/email + savings goals + rollover + đa tiền tệ: **ngoài phạm vi** (Out of Scope BR-003).
- Commit sau mỗi task hoặc nhóm logic; dừng ở mỗi Checkpoint để kiểm chứng story độc lập.
