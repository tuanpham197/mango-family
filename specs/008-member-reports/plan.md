# Implementation Plan: Báo Cáo Thu Chi Theo Thành Viên (Per-Member Report) — Go + Vue

**Branch**: `008-member-reports` | **Date**: 2026-08-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature spec `/specs/008-member-reports/spec.md` · BR [`BR-008.md`](../business-requirements/BR-008.md) (mã con BR-MBR-001…009) · **Nền tảng**: [plan 005](../005-reports/plan.md) (module `report` read-only — mở rộng trực tiếp) · [plan 002](../002-transaction-tracking/plan.md) (giao dịch, `created_by`, phân trang) · [plan 001](../001-transaction-categorization/plan.md).

> 🆕 Feature này **không thêm bảng/migration** và **không thêm dependency** (BE hay FE). Mở rộng module Go `report` sẵn có với tổng hợp **theo thành viên** (`GROUP BY created_by`) và màn Báo cáo (`/reports`) với một khối "theo thành viên" + drill-in danh sách giao dịch của một thành viên. Tái dùng `chart.js` (đã có từ 005) nếu cần biểu đồ cột so sánh, và `transactionmodel.ListItem` + `common.Paging` cho drill-in.

## Summary

Mở rộng màn **Báo cáo** với phần **theo thành viên**: cho khoảng thời gian chọn được (tuần/tháng/quý/năm/tùy chỉnh — tái dùng bộ giải khoảng của 005), hiển thị **tất cả thành viên hiện tại** của hộ kèm **tổng Thu / tổng Chi / ròng (Thu − Chi)**; thành viên không có giao dịch hiển thị 0/0/0; giao dịch của người **đã rời hộ** gộp vào một dòng **"Thành viên cũ / Đã rời hộ"** (chỉ hiện khi có dữ liệu) để **tổng các dòng luôn khớp tổng hộ** (SC-001, FR-010 Option A). Chọn một dòng để **drill-in** xem danh sách giao dịch (phân trang) cấu thành số liệu đó. Mọi số liệu là **giá trị suy ra**, tính bằng SQL khi đọc — không bảng, không cột spent.

**Technical approach**: Mở rộng module Go **`module/report`** (read-only) với 2 endpoint mới:
`GET /api/reports/members?from=&to=` (danh sách thành viên + tổng Thu/Chi/ròng + dòng "Thành viên cũ" + totals đối soát) và
`GET /api/reports/member/:id?from=&to=&page=&page_size=` (drill-in giao dịch của một thành viên, phân trang; `:id` = UUID thành viên hoặc sentinel `former`).
Tổng hợp bằng SQL: `SUM(CASE type) ... GROUP BY created_by` trên giao dịch của hộ trong `[from,to]`, **LEFT JOIN `household_members`** để tách thành viên hiện tại ↔ người đã rời (created_by không còn là thành viên → gộp bucket `former`); danh sách khởi từ `household_members` (LEFT JOIN sang tổng hợp) để mọi thành viên hiện tại đều xuất hiện kể cả 0 giao dịch. Drill-in tái dùng cơ chế liệt kê giao dịch phân trang của module transaction (thêm bộ lọc `created_by`). Web: mở rộng `ReportsView` + store `reports.ts` với khối `MemberBreakdown` (bảng + biểu đồ cột tùy chọn qua wrapper `chart.js` sẵn có) và panel drill-in giao dịch theo thành viên.

## Technical Context

**Language/Version**: Go 1.22+ (api) · TypeScript 5.x / Node 20+ (web) — như 001–005

**Primary Dependencies**: như 001–005 (Gin, GORM, gorilla/websocket · Vue 3, Vite, Pinia, vue-router · `chart.js` đã có từ 005). **Không thêm dependency mới** (BE/FE), **không migration**.

**Storage**: PostgreSQL 16 — **chỉ ĐỌC** `transactions` (002), `users` + `household_members` (nền tảng). Tổng hợp bằng SQL. Không thêm bảng/cột/view.

**Testing**: Go unit (tách current↔former, net = thu−chi, gộp bucket former, thành viên 0 giao dịch, sắp theo tổng Chi desc) + integration storage (aggregation `GROUP BY created_by`, cô lập hộ, đối soát tổng = tổng hộ, phân trang drill-in trên Postgres thật) + httptest (2 endpoint, validate from/to, 404 khi `:id` ngoài hộ, sentinel `former`) · Vitest (store `reports.ts` phần members, component MemberBreakdown/drill-in, format tên fallback email) · Playwright e2e (đổi khoảng cập nhật số liệu theo thành viên, thành viên 0 giao dịch = 0/0/0, drill-in + phân trang + quay lại giữ khoảng, cô lập hộ)

**Target Platform**: Web responsive mobile-first — như 001–005

**Project Type**: Web app monorepo `src/api` + `src/web` + `src/db`, chia sẻ theo hộ

**Performance Goals**: Báo cáo theo thành viên khoảng 1 tháng ≤ 2 giây (SC-002); đổi khoảng cập nhật tại chỗ (SC-004); drill-in ≤ 2 thao tác giữ khoảng (SC-004)

**Constraints**: Đối soát 100% — tổng các dòng (thành viên hiện tại + bucket former) khớp tổng hộ cùng khoảng (SC-001); mọi thành viên hiện tại luôn xuất hiện, kể cả 0 giao dịch (SC-003); cô lập hộ tuyệt đối, xác thực `:id` thuộc hộ, không nhận user ID ngoài hộ (SC-006/FR-008); tên hiển thị fallback email, không lộ mã thô (FR-009); ranh giới thời gian theo lịch nhất quán toàn hệ thống (tái dùng 005). Ngoài phạm vi: xuất file, hạn mức theo thành viên, xếp hạng/chấm điểm, phân quyền theo vai trò.

**Scale/Scope**: Mỗi hộ vài thành viên + vài nghìn giao dịch; khoảng có thể tới 1 năm; 2 endpoint mới + mở rộng 1 màn báo cáo (khối theo thành viên + drill-in).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` là **bản mẫu chưa phê chuẩn** (toàn placeholder) — không có nguyên tắc ràng buộc.

- **Kết luận**: PASS.
- **Re-check sau Phase 1**: PASS — mở rộng 1 module đọc sẵn có + 2 endpoint đọc; **không** dependency mới, **không** bảng/migration, **không** sửa dữ liệu nguồn. Giữ đúng chiều phụ thuộc (report chỉ đọc transactions/users/household_members). Không mục nào trong Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/008-member-reports/
├── plan.md              # This file
├── research.md          # Phase 0 — D42…D50 (kế thừa D1–D41 của 001–005)
├── data-model.md        # Phase 1 — DTO suy ra (không bảng mới)
├── quickstart.md        # Phase 1 — kịch bản kiểm chứng
├── contracts/
│   └── member-report-api.md     # REST GET /api/reports/members + /member/:id
├── checklists/
│   └── requirements.md          # (đã có — checklist chất lượng spec)
└── tasks.md             # Phase 2 (/speckit-tasks — KHÔNG tạo ở bước này)
```

### Source Code (mở rộng cấu trúc 001–005 — layout learn_go)

```text
src/
├── api/
│   ├── module/
│   │   ├── report/               # MỞ RỘNG (đã có từ 005): thêm tổng hợp theo thành viên
│   │   │   ├── model/            #   + MemberRow, MembersReport, MemberReport (suy ra)
│   │   │   ├── storage/          #   + SUM thu/chi GROUP BY created_by; LEFT JOIN household_members;
│   │   │   │                     #     mọi thành viên hiện tại (kể cả 0 GD); bucket former; drill-in phân trang
│   │   │   ├── biz/              #   + get_members.go, get_member.go (giải khoảng, tách current/former, sort, ráp DTO)
│   │   │   └── transport/ginreport/routes.go  # + Members, Member handlers (GET /members, /member/:id)
│   │   ├── transaction/storage/  #   + trường lọc `CreatedBy *uuid.UUID` trong ListFilter (đọc, tái dùng drill-in)
│   │   └── (users/household_members — GIỮ NGUYÊN, chỉ đọc)
│   └── main_route.go             # wire routes /api/reports/members · /api/reports/member/:id (scoped hộ)
├── web/
│   └── src/
│       ├── views/ReportsView.vue        # MỞ RỘNG: thêm khối "Theo thành viên" + drill-in panel
│       ├── components/
│       │   ├── MemberBreakdown.vue       # MỚI: bảng thành viên (Thu/Chi/ròng) + biểu đồ cột (tùy chọn)
│       │   └── MemberTransactionsPanel.vue  # MỚI: drill-in danh sách giao dịch (phân trang) của 1 thành viên
│       └── stores/reports.ts             # MỞ RỘNG: state + actions cho members + member drill-in
└── web/e2e/                              # + member-report*.spec.ts
```

**Structure Decision**: Mở rộng **module `report`** sẵn có thay vì tạo module mới — giữ đúng ngữ nghĩa "báo cáo suy ra, chỉ đọc" và tái dùng bộ giải khoảng/gom nhóm của 005. Drill-in tái dùng `transactionmodel.ListItem` + `common.Paging` (chỉ thêm một trường lọc `created_by` vào `ListFilter` của module transaction — thay đổi đọc, tương thích ngược). FE mở rộng `ReportsView`/`reports.ts` sẵn có, tái dùng `chart.js` (không thêm dependency).

## Dependency & thứ tự nền tảng

1. **Feature 001 + 002 + 005 đã triển khai**: giao dịch (loại/ngày/`created_by`), `users`+`household_members`, module `report` (giải khoảng `ResolvePeriod`, chọn đơn vị gom nhóm), màn `/reports` + phân trang giao dịch (002).
2. **Storage đọc (report, mở rộng)**: `SUM(CASE type) GROUP BY created_by` trên `[from,to]`; LEFT JOIN `household_members` để tách current↔former; danh sách khởi từ `household_members` để mọi thành viên hiện tại xuất hiện (kể cả 0 GD); bucket `former`; drill-in liệt kê giao dịch theo `created_by` (hoặc `∉ current` cho former) có phân trang.
3. **Biz + endpoints**: giải khoảng (preset FE hoặc from/to), validate end≥start (tái dùng 005), sort theo tổng Chi desc, ráp DTO, đối soát totals; `GET /api/reports/members`, `GET /api/reports/member/:id` (id = UUID hoặc `former`).
4. **UI**: `MemberBreakdown` (bảng + cột) trong `ReportsView`; `MemberTransactionsPanel` drill-in (phân trang, empty state, quay lại giữ khoảng); mở rộng store `reports.ts`.
5. **Ngoài phạm vi**: xuất PDF/Excel/CSV, hạn mức/ngân sách theo thành viên, xếp hạng/chấm điểm, phân quyền theo vai trò, xác định người thực nhận/chi (Out of Scope BR-008).

## Tác động tới artifact khác

- `src/api/module/transaction/storage` — thêm trường lọc `CreatedBy` vào `ListFilter` (đọc, tương thích ngược). Ghi chú ở research (D44).
- `specs/entities/entity-model.md` — **không đổi** (không thực thể/bảng mới; toàn giá trị suy ra; `created_by`/`household_members` đã có).
- Module transaction/users — **không sửa hành vi ghi** (report chỉ đọc; chỉ thêm bộ lọc đọc).
- `specs/005-reports` — cùng module `report` được mở rộng; báo cáo tổng quan hiện có **không đổi**.
- `CLAUDE.md` — cập nhật con trỏ plan (SPECKIT markers) → plan 008; ghi nhận feature 008 (không dep/migration mới).
- Use cases [`UC-MBR-01/02`](../use-cases/008-member-reports/) + `specs/diagrams/use-cases.puml` — đã tạo & lan truyền; giữ đồng bộ nếu FR đổi.

## Complexity Tracking

> Không có vi phạm Constitution cần biện minh. Không thêm dependency, không bảng/migration, không module mới — chỉ mở rộng module đọc sẵn có + 2 endpoint đọc.
