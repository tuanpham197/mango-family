# Implementation Plan: Báo Cáo và Phân Tích Trực Quan (Reports) — Go + Vue

**Branch**: `005-reports` | **Date**: 2026-07-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature spec `/specs/005-reports/spec.md` · BR [`BR-006.md`](../business-requirements/BR-006.md) (ID nội bộ "BR-004", mã con BR-RPT-*) · **Nền tảng**: [plan 001](../001-transaction-categorization/plan.md) · [plan 002](../002-transaction-tracking/plan.md) · [plan 004](../004-monthly-income-expense-overview/plan.md) — đều đã triển khai.

> 🆕 Feature này **không thêm bảng/migration** (chỉ ĐỌC giao dịch/danh mục). Điểm mới về stack: **thêm một dependency FE — `chart.js`** (bọc bằng wrapper Vue mỏng tự viết) cho biểu đồ; đây là dependency mới đầu tiên kể từ re-platform (001–004 không thêm dep nào). Hiện thực hóa mục điều hướng **"Báo cáo" (`/reports`)** vốn là placeholder của feature 004.

## Summary

Cung cấp màn **Báo cáo** với: (a) **báo cáo tổng quan** theo khoảng thời gian chọn được (tuần/tháng/quý/năm/tùy chỉnh) — tổng thu/chi/số dư ròng + **biểu đồ phân bổ chi tiêu theo danh mục** (tròn/donut) + **biểu đồ xu hướng thu/chi** (đường); (b) **báo cáo chi tiết theo danh mục** — tổng chi + danh sách giao dịch + **biểu đồ xu hướng chi của danh mục**. Mọi số liệu là **giá trị suy ra** tính ở biz khi đọc (như 004), gom nhóm chuỗi thời gian bằng SQL để đạt ≤ 2s cho khoảng 1 tháng (SC-002).

**Technical approach**: Thêm module Go **`module/report`** (read-only) phơi 2 endpoint: `GET /api/reports/overview?from=&to=` và `GET /api/reports/category/:id?from=&to=`. Tổng hợp bằng SQL (`SUM`, `GROUP BY date_trunc(...)`, gộp danh mục con vào cha như D29/004). Đơn vị gom nhóm chuỗi thời gian chọn theo độ dài khoảng (ngày/tuần/tháng — D36). Web: thêm `chart.js` + **wrapper Vue mỏng** (`DonutChart`, `LineChart`) tự viết (không dùng vue-chartjs để giữ tối thiểu); `ReportsView` (thay `ReportsPlaceholderView`) + `TimeRangePicker` + drill-in chi tiết danh mục; store `reports.ts`.

## Technical Context

**Language/Version**: Go 1.22+ (api) · TypeScript 5.x / Node 20+ (web) — như 001–004

**Primary Dependencies**: như 001–004 (Gin, GORM, gorilla/websocket · Vue 3, Vite, Pinia, vue-router) **+ MỚI (FE): `chart.js`** cho biểu đồ, bọc bằng wrapper Vue tự viết (không thêm `vue-chartjs`). Không thêm dependency BE, không migration.

**Storage**: PostgreSQL 16 — **chỉ ĐỌC** `transactions` (002) + `categories` (001); tổng hợp bằng SQL. Không thêm bảng/cột/view.

**Testing**: Go unit (giải khoảng preset→[from,to], chọn đơn vị gom nhóm, gộp cây danh mục, net = thu−chi) + integration (aggregation + trend bucket trên Postgres thật, cô lập hộ) + httptest (2 endpoint, validate from/to, 404 ngoài hộ) · Vitest (wrapper chart với `chart.js` mock, `TimeRangePicker`, store, format) · Playwright e2e (đổi khoảng, phân bổ/xu hướng hiển thị, drill-in danh mục, trạng thái trống)

**Target Platform**: Web responsive mobile-first — như 001–004

**Project Type**: Web app monorepo `src/api` + `src/web` + `src/db`, chia sẻ theo hộ

**Performance Goals**: Báo cáo khoảng 1 tháng ≤ 2 giây (SC-002); đổi khoảng cập nhật tại chỗ (SC-003); drill-in ≤ 2 thao tác (SC-005)

**Constraints**: Số liệu suy ra 100% khớp giao dịch theo khoảng (SC-001); gom nhóm theo lịch nhất quán toàn hệ thống (không theo múi giờ thiết bị — Assumptions); chỉ đọc dữ liệu nguồn; khoảng tùy chỉnh end ≥ start; khoảng trống hiển thị 0/empty; ngoài phạm vi: xuất file, forecasting, benchmark, phân tích thu chi tiết

**Scale/Scope**: Mỗi hộ vài nghìn giao dịch; khoảng có thể tới 1 năm; 2 endpoint + 1 màn báo cáo (tổng quan + drill-in danh mục)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` là **bản mẫu chưa phê chuẩn** (toàn placeholder) — không có nguyên tắc ràng buộc.

- **Kết luận**: PASS.
- **Re-check sau Phase 1**: PASS — thêm 1 module đọc + 2 endpoint + 1 dependency FE (chart.js) có biện minh (BR yêu cầu biểu đồ; người dùng chọn dùng thư viện thay vì tự vẽ). Không bảng/service mới. Xem Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/005-reports/
├── plan.md              # This file
├── research.md          # Phase 0 — D34…D41 (kế thừa D1–D33)
├── data-model.md        # Phase 1 — DTO suy ra (không bảng mới)
├── quickstart.md        # Phase 1 — kịch bản kiểm chứng
├── contracts/
│   └── report-api.md            # REST GET /api/reports/overview + /category/:id
├── checklists/
│   └── requirements.md          # (đã có — checklist chất lượng spec)
└── tasks.md             # Phase 2 (/speckit-tasks — KHÔNG tạo ở bước này)
```

### Source Code (mở rộng cấu trúc 001–004 — layout learn_go)

```text
src/
├── api/
│   ├── module/
│   │   ├── report/               # MỚI: chỉ ĐỌC, tổng hợp báo cáo
│   │   │   ├── model/            #   ReportOverview, CategoryBreakdown, TrendPoint, CategoryReport (suy ra)
│   │   │   ├── storage/          #   SUM thu/chi theo khoảng; phân bổ danh mục (gộp con); trend GROUP BY date_trunc;
│   │   │   │                     #   tổng + danh sách giao dịch theo danh mục (+subtree)
│   │   │   ├── biz/              #   giải khoảng, chọn đơn vị gom nhóm, %, ráp DTO
│   │   │   └── transport/ginreport/  # GET /api/reports/overview · /api/reports/category/:id
│   │   └── (transaction/category — GIỮ NGUYÊN, chỉ đọc)
│   └── main.go / main_route.go   # wire routes /api/reports/* (scoped hộ)
├── web/
│   ├── package.json              # + chart.js
│   └── src/
│       ├── views/
│       │   └── ReportsView.vue    # THAY ReportsPlaceholderView (route /reports); drill-in danh mục
│       ├── components/            # TimeRangePicker; DonutChart, LineChart (wrapper chart.js mỏng);
│       │                          #   CategoryBreakdownChart, TrendChart, CategoryDetailPanel
│       └── stores/                # reports.ts (MỚI)
└── web/e2e/                      # report*.spec.ts
```

**Structure Decision**: Một module đọc `report` gói tổng hợp báo cáo (giữ đúng chiều phụ thuộc: report là hạ nguồn — chỉ đọc transactions/categories). Biểu đồ dùng `chart.js` bọc bằng 2–3 wrapper Vue mỏng (tự viết, không thêm `vue-chartjs`) để giới hạn footprint ở đúng một dependency. `ReportsView` thay placeholder `/reports` của feature 004.

## Dependency & thứ tự nền tảng

1. **Feature 001 + 002 + 004 đã triển khai**: giao dịch (loại/ngày/danh mục), danh mục (+cây một cấp), màn Tổng quan + mục nav "Báo cáo" (`/reports` placeholder), cửa sổ kỳ (`ResolvePeriod`).
2. **Storage đọc (report)**: SUM thu/chi theo [from,to]; phân bổ danh mục (gộp con, %); trend `GROUP BY date_trunc(unit, transaction_date)`; tổng + danh sách + trend theo một danh mục (+subtree).
3. **Biz + endpoints**: giải khoảng (preset FE hoặc from/to), validate end≥start, chọn đơn vị gom nhóm (D36), ráp DTO; `GET /api/reports/overview`, `GET /api/reports/category/:id`.
4. **FE hạ tầng biểu đồ**: thêm `chart.js`; wrapper `DonutChart`/`LineChart`.
5. **UI**: `ReportsView` + `TimeRangePicker` + phân bổ/xu hướng + drill-in danh mục (danh sách + trend); store `reports.ts`.
6. **Ngoài phạm vi**: xuất PDF/Excel/CSV, forecasting, benchmark, phân tích thu chi tiết (Out of Scope BR-006).

## Tác động tới artifact khác

- `src/web/package.json` — thêm `chart.js` (dependency FE mới đầu tiên; ghi rõ ở re-platform notes/CLAUDE.md).
- `specs/004-monthly-income-expense-overview` — mục nav "Báo cáo"/`/reports` từ **placeholder** thành màn thật; `ReportsPlaceholderView` được thay bằng `ReportsView`. Ghi chú ở research (D34) + cập nhật khi agent context.
- `specs/entities/entity-model.md` — **không đổi** (không thực thể/bảng mới; toàn giá trị suy ra).
- Module transaction/category — **không sửa** (report chỉ đọc).
- `CLAUDE.md` — cập nhật con trỏ plan (SPECKIT markers) → plan 005; ghi nhận dependency FE mới `chart.js`.

## Complexity Tracking

| Điểm phức tạp | Vì sao cần | Phương án đơn giản hơn bị loại vì |
|---------------|-----------|----------------------------------|
| Dependency FE mới `chart.js` | BR-006 yêu cầu trực quan hóa (tròn/cột/đường); người dùng chọn dùng thư viện để có biểu đồ hoàn chỉnh (trục/nhãn/tooltip) với ít mã | Tự vẽ SVG (donut/bar/line) — bị loại: nhiều mã, dễ sai trục/nhãn/tooltip, kém trau chuốt; đánh đổi không xứng cho một feature xoay quanh biểu đồ. Giới hạn rủi ro bằng **một** dep + wrapper mỏng tự viết (không thêm vue-chartjs). |
