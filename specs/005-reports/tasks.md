---
description: "Task list — Báo Cáo và Phân Tích Trực Quan (Reports, Go + Vue)"
---

# Tasks: Báo Cáo và Phân Tích Trực Quan (Reports) — Go + Vue

**Input**: Design documents from `/specs/005-reports/` (plan 2026-07-15)

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md) (D34–D41), [data-model.md](./data-model.md), [contracts/](./contracts/) (`report-api.md`), [quickstart.md](./quickstart.md) (16 kịch bản) · **Nền tảng**: [001](../001-transaction-categorization/tasks.md) + [002](../002-transaction-tracking/tasks.md) + [004](../004-monthly-income-expense-overview/tasks.md) — **đã triển khai** (giao dịch, danh mục, màn Tổng quan + mục nav "Báo cáo" placeholder).

**Stack**: Go 1.22+ (Gin + GORM) · Vue 3 (Vite/TS/Pinia) · PostgreSQL 16 · Playwright — như 001–004, **+ dependency FE mới `chart.js`** (D39); **không migration mới** (chỉ ĐỌC).

**Tests**: Spec không yêu cầu TDD → không sinh phase test riêng; kiểm chứng theo 16 kịch bản [quickstart.md](./quickstart.md); test các tầng ở Polish.

**Path conventions**: API `src/api/module/report/{model,storage,biz,transport/ginreport}/`; web `src/web/src/`; e2e `src/web/e2e/`.

> 🆕 **Sinh 2026-07-15**: báo cáo chỉ **ĐỌC** giao dịch/danh mục, tổng hợp bằng SQL (SUM/`date_trunc`); 2 endpoint `GET /api/reports/overview` + `/category/:id` (D34). Biểu đồ dùng **chart.js** bọc wrapper Vue mỏng (D39). Màn `/reports` **thay placeholder** của feature 004.

---

## Phase 1: Setup

**Purpose**: Khung module report + dependency biểu đồ.

- [X] T001 [P] Scaffold `src/api/module/report/{model,storage,biz,transport/ginreport}/` (khung theo mẫu learn_go)
- [X] T002 [P] Thêm `chart.js` vào `src/web/package.json` (dependency) + `npm install`; xác minh build không lỗi

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: DTO + storage helpers + 2 endpoint + hạ tầng biểu đồ + màn Báo cáo khung + đổi route — mọi user story phụ thuộc.

**⚠️ CRITICAL**: Không bắt đầu US nào trước khi phase này xong.

- [X] T003 Report model DTO trong `src/api/module/report/model/report.go`: `ReportOverview` (`from,to,group_unit,income,expense,net,category_breakdown[],trend[]`), `CategoryBreakdown`, `TrendPoint` (`bucket,income,expense`), `CategoryReport` (`total,transactions[],trend[]` với `CategoryTrendPoint{bucket,amount}`) — **suy ra, không map bảng** (D34, data-model)
- [X] T004 Report storage `src/api/module/report/storage/report.go`: struct `SQLStore` **chỉ SELECT/SUM**; helper cửa sổ khoảng + `chooseGroupUnit(from,to)` (day/week/month — D36) + bucket bằng `date_trunc`; (các truy vấn cụ thể thêm ở từng US)
- [X] T005 Report biz `src/api/module/report/biz/*.go` (`get_overview.go`, `get_category.go`): giải/validate khoảng (`to ≥ from` → 400), chọn đơn vị gom nhóm, ráp DTO + điền mốc trend trống = 0; transport `src/api/module/report/transport/ginreport/routes.go` `GET /api/reports/overview` + `GET /api/reports/category/:id`; wire route scoped hộ trong `src/api/main_route.go` (D34/D35, contracts) (phụ thuộc T003, T004)
- [X] T006 [P] Web wrapper biểu đồ `src/web/src/components/charts/`: đăng ký chart.js tree-shake (Doughnut/Pie, Line, scale/element) + `DonutChart.vue` + `LineChart.vue` (wrapper mỏng nhận `data`/`options`, hủy chart khi unmount) (D39)
- [X] T007 Web khung + đổi route: store `src/web/src/stores/reports.ts` (fetch overview/category theo from/to; nghe `transactions_changed`/`categories_changed` → refetch — D40); `TimeRangePicker.vue` (preset tuần/tháng/quý/năm + tùy chỉnh, chặn end<start); `src/web/src/views/ReportsView.vue` (khung, thay `ReportsPlaceholderView`); đổi route `/reports` → `ReportsView` trong `src/web/src/router/index.ts` (gỡ `ReportsPlaceholderView`) (phụ thuộc T006)

**Checkpoint**: `GET /api/reports/overview?from=&to=` + `/category/:id` trả cấu trúc rỗng-an toàn; `/reports` mở màn Báo cáo khung với bộ chọn khoảng.

---

## Phase 3: User Story 1 - Báo cáo tổng quan theo khoảng thời gian (Priority: P1) 🎯 MVP

**Goal**: Chọn khoảng (tuần/tháng/quý/năm/tùy chỉnh) → tổng thu/chi/số dư ròng của hộ; đổi khoảng cập nhật tại chỗ; end<start bị chặn. (quickstart #1–5)

**Independent Test**: quickstart #1–5.

- [X] T008 [US1] Storage `report.go`: `SumIncomeExpense(ctx, householdID, from, to)` → `Σ` INCOME & EXPENSE trong [from,to], cùng hộ; biz `get_overview.go` điền `income/expense/net` (FR-002, SC-001)
- [X] T009 [US1] Web `ReportsView` + `TimeRangePicker`: hiển thị tổng thu/chi/ròng từ store theo khoảng chọn; đổi preset/tùy chỉnh → refetch; báo lỗi khi end<start (FE) và xử lý 400 từ API; trạng thái trống (0) (FR-001, FR-008, FR-009, SC-003)

**Checkpoint**: tổng thu/chi/ròng đúng theo khoảng; đổi khoảng cập nhật; chặn ngày sai (quickstart #1–5).

---

## Phase 4: User Story 2 - Phân bổ chi tiêu theo danh mục (Priority: P1) 🎯 MVP

**Goal**: Biểu đồ tròn/donut phân bổ chi theo danh mục (gộp con) + số tiền/% cho khoảng đã chọn. (quickstart #6–8)

**Independent Test**: quickstart #6–8.

- [X] T010 [US2] Storage `report.go`: `CategoryBreakdown(ctx, householdID, from, to)` → `Σ` EXPENSE nhóm theo `COALESCE(parent.id,cat.id)` (gộp con — D37), kèm `category_hidden`; biz tính `percent`, sắp giảm dần, điền `category_breakdown[]` (FR-003, SC-001)
- [X] T011 [P] [US2] Web `src/web/src/components/CategoryBreakdownChart.vue`: dùng `DonutChart` + chú giải (tên · số tiền · %); empty state; nhúng vào `ReportsView` (FR-003, FR-009)

**Checkpoint**: phân bổ đúng số tiền/% (gộp con), tổng khớp tổng chi; empty state (quickstart #6–8).

---

## Phase 5: User Story 3 - Xu hướng thu nhập & chi phí theo thời gian (Priority: P2)

**Goal**: Biểu đồ đường thu & chi theo mốc thời gian; đơn vị gom nhóm hợp lý theo độ dài khoảng. (quickstart #9–10)

**Independent Test**: quickstart #9–10.

- [X] T012 [US3] Storage `report.go`: `TrendIncomeExpense(ctx, householdID, from, to, unit)` → `GROUP BY date_trunc(unit, transaction_date)` trả `{bucket, income, expense}`; biz chọn `unit` (D36) + điền mốc trống = 0, điền `trend[]` (FR-004, SC-001)
- [X] T013 [P] [US3] Web `src/web/src/components/TrendChart.vue`: dùng `LineChart` hai chuỗi (thu/chi) theo `bucket`; nhúng vào `ReportsView` (FR-004)

**Checkpoint**: hai đường thu/chi khớp tổng theo mốc; đơn vị gom nhóm đổi theo khoảng (quickstart #9–10).

---

## Phase 6: User Story 4 - Báo cáo chi tiết theo danh mục (Priority: P2)

**Goal**: Chọn một danh mục → tổng chi + danh sách mọi giao dịch (gồm con) trong khoảng; đổi khoảng cập nhật; ngoài hộ → 404. (quickstart #11–14, #16)

**Independent Test**: quickstart #11–14, #16.

- [X] T014 [US4] Storage `report.go`: `CategoryTotal` + `CategoryTransactions(ctx, householdID, categoryID+subtree, from, to)` (dùng lại `ListItem` 002, mới nhất trước); biz `get_category.go` ráp `CategoryReport` (total + transactions); transport `/category/:id` trả 404 nếu danh mục ngoài hộ (D38, FR-005, FR-007)
- [X] T015 [US4] Web `src/web/src/components/CategoryDetailPanel.vue` + drill-in: bấm một danh mục ở `CategoryBreakdownChart`/chú giải → mở chi tiết (tổng + danh sách giao dịch: số tiền/mô tả/ngày); đổi khoảng cập nhật; empty state (FR-005, SC-005)

**Checkpoint**: chi tiết danh mục đúng tổng + danh sách (gồm con); cô lập hộ (quickstart #11–14, #16).

---

## Phase 7: User Story 5 - Xu hướng chi tiêu của một danh mục (Priority: P3)

**Goal**: Biểu đồ đường chi tiêu của riêng danh mục theo thời gian trong khoảng. (quickstart #15)

**Independent Test**: quickstart #15.

- [X] T016 [US5] Storage `report.go`: `CategoryTrend(ctx, householdID, categoryID+subtree, from, to, unit)` → `{bucket, amount}` (điền mốc 0); biz điền `CategoryReport.trend` (FR-006, SC-001)
- [X] T017 [P] [US5] Web `CategoryDetailPanel`: nhúng `LineChart` xu hướng chi của danh mục theo `bucket` (FR-006)

**Checkpoint**: đường xu hướng danh mục khớp tổng chi danh mục theo mốc (quickstart #15).

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Test các tầng, e2e 16 kịch bản, đồng bộ tài liệu.

- [X] T018 [P] Unit biz (Go) `src/api/module/report/biz/*_test.go`: giải khoảng preset→[from,to] (tuần/tháng/quý/năm), `chooseGroupUnit` ranh giới (31/92 ngày), `net = income − expense`, `percent`, điền mốc trend trống = 0, validate `to<from`
- [X] T019 [P] Integration storage (Postgres docker, tag `integration`) `src/api/module/report/storage/*_integration_test.go`: SUM thu/chi theo khoảng; breakdown gộp con khớp SUM; trend bucket (day/week/month) khớp; chi tiết danh mục + subtree; cô lập hộ; khoảng trống (SC-001)
- [X] T020 [P] Transport httptest `src/api/main_005_integration_test.go`: `/overview` + `/category/:id` đủ field theo `contracts/report-api.md`; `to<from` → 400; thiếu param → 400; danh mục ngoài hộ → 404; khoảng trống → số 0/rỗng
- [X] T021 [P] Vitest `src/web/src/**/__tests__/`: `DonutChart`/`LineChart` (mock `chart.js` — dựng/hủy đúng), `TimeRangePicker` (preset + chặn end<start), store `reports.ts` (fetch overview/category + refetch), format số/%
- [X] T022 Playwright e2e 16 kịch bản trong `src/web/e2e/report*.spec.ts`: đổi khoảng (#1–5), phân bổ danh mục (#6–8), xu hướng (#9–10), drill-in chi tiết danh mục (#11–13, #15), 404 ngoài hộ (#14), cô lập hộ (#16)
- [X] T023 Chạy toàn bộ quickstart 005; cập nhật References/History artifact theo `CLAUDE.md` (spec 004 ghi chú `/reports` từ placeholder → màn thật; CLAUDE.md status 005 + ghi nhận dependency FE `chart.js` + sai lệch số hiệu BR-006/BR-004; entity-model không đổi); chạy `/speckit-analyze` phát hiện lệch spec↔plan↔tasks

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: T001 ∥ T002.
- **Foundational (P2)**: sau Setup. **BLOCKS mọi user story.** T003 → T004 → T005; T006 ∥ (T003–T005); T007 sau T006.
- **US1 (P3)**: sau Foundational; T008 → T009.
- **US2 (P4)**: sau Foundational; T010 → T011.
- **US3 (P5)**: sau Foundational; T012 → T013.
- **US4 (P6)**: sau Foundational; T014 → T015.
- **US5 (P7)**: sau US4 (nhúng vào CategoryDetailPanel); T016 → T017.
- **Polish (P8)**: sau các story mong muốn.

### User Story Dependencies

- **US1/US2 (P1)**: chỉ phụ thuộc Foundational — cùng tạo MVP (báo cáo tổng quan: số tổng + phân bổ danh mục).
- **US3 (P2)**: độc lập US2 (khác truy vấn/biểu đồ), chung file storage/biz nên phần backend tuần tự.
- **US4 (P2)**: độc lập US3; endpoint riêng `/category/:id`.
- **US5 (P3)**: mở rộng US4 (thêm trend vào panel chi tiết).

### Parallel Opportunities

- Setup: T001 ∥ T002.
- Foundational: T006 (chart wrappers) ∥ T003–T005 (API).
- Các component UI [P] khác file: T011, T013, T017 song song được (chỉ chờ store + wrapper); phần backend tương ứng (T008/T010/T012/T014/T016) tuần tự do chung `storage/report.go` + biz.
- Polish: T018–T021 song song; T022 sau cùng; T023 khép lại.

---

## Implementation Strategy

### MVP First (US1 + US2 — cả hai P1)

1. Phase 1 Setup → 2. Phase 2 Foundational (endpoint + màn khung + chart wrappers + đổi route) → 3. US1 (tổng thu/chi/ròng + bộ chọn khoảng) → 4. US2 (phân bổ danh mục + donut) → **DỪNG & kiểm chứng** quickstart #1–8 → demo MVP (báo cáo tổng quan trực quan theo khoảng).

### Incremental Delivery

MVP (US1+US2) → US3 (xu hướng thu/chi) → US4 (chi tiết theo danh mục) → US5 (xu hướng danh mục) → Polish (test tầng + e2e 16 kịch bản + docs). Mỗi story kiểm chứng độc lập theo nhóm kịch bản quickstart.

---

## Notes

- [P] = khác file, không phụ thuộc nhau. Nhãn [US#] gắn task với user story.
- Mọi số liệu **KHÔNG lưu** — tổng hợp bằng SQL khi đọc (D41); **không migration, không sửa** module transaction/category (report chỉ ĐỌC).
- Biểu đồ: **chart.js** (D39) bọc bằng wrapper Vue mỏng — dependency FE mới đầu tiên; cô lập sau `components/charts/`.
- Màn `/reports` thay **placeholder** của feature 004 (`ReportsPlaceholderView` gỡ bỏ); mục nav "Báo cáo" giữ nguyên.
- Ngoài phạm vi (BR-006): xuất PDF/Excel/CSV, forecasting, benchmark, phân tích thu chi tiết.
- Ghi nhận cần nghiệp vụ: số hiệu BR (file `BR-006.md` ↔ ID "BR-004"); baseline SC-006.
- Commit sau mỗi task hoặc nhóm logic; dừng ở mỗi Checkpoint để kiểm chứng story độc lập.
