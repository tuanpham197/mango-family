# Implementation Plan: Màn Tổng Quan Đầy Đủ (Dashboard) & Trang Mặc Định — Go + Vue

**Branch**: `004-monthly-income-expense-overview` | **Date**: 2026-07-14 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification `/specs/004-monthly-income-expense-overview/spec.md` (v2) · Wireframe [`specs/design/dashboard.png`](../design/dashboard.png) · **Nền tảng**: [plan 001](../001-transaction-categorization/plan.md) · [plan 002](../002-transaction-tracking/plan.md) · [plan 003](../003-budgeting/plan.md) — đều đã triển khai.

> 🆕 Feature này **không thêm bảng/migration**. Toàn bộ số liệu trên màn Tổng quan là **giá trị suy ra** đọc từ dữ liệu 001/002/003 (giao dịch, tài khoản+số dư, danh mục, ngân sách). Trọng tâm là **một endpoint tổng hợp** cho phần đọc + **dựng lại màn Tổng quan** khớp 100% wireframe + **đổi trang mặc định** sang Tổng quan.

## Summary

Biến màn Tổng quan (`/`) thành trang chủ tài chính đầy đủ theo wireframe `dashboard.png`, và đặt nó làm **trang mặc định** khi vào ứng dụng. Màn gồm 6 phần theo đúng thứ tự: lời chào + avatar → **Tổng tài sản ròng** (+% so tháng trước) → **Thu nhập / Chi phí** tháng → **Chi tiêu theo danh mục** (MỚI, trước ngân sách) → **Ngân sách** (widget 003) → **Giao dịch gần đây**; thanh điều hướng: Tổng quan · Giao dịch · ＋ · Ngân sách · Báo cáo.

**Technical approach**: Thêm module Go **`module/overview`** chỉ-đọc, phơi một endpoint tổng hợp `GET /api/overview` trả về mọi giá trị suy ra (tài sản ròng + tài sản ròng cuối tháng trước, tổng Thu/Chi tháng, chi tiêu theo danh mục + %, vài giao dịch gần đây) trong một round-trip — nhất quán triết lý "tiến độ = giá trị suy ra ở biz" (D21/003). Tóm tắt **ngân sách** dùng lại `GET /api/budgets` (widget 003, không đổi). Web: dựng lại `DashboardView` + các component mới (thẻ tài sản ròng, thẻ Thu/Chi, danh sách chi theo danh mục, danh sách giao dịch gần đây) + dời `OverviewBudgetWidget`; **đổi định tuyến**: `/` = Tổng quan (trang chủ), sổ giao dịch chuyển sang `/ledger`; store `overview.ts` refetch realtime theo các topic sẵn có.

## Technical Context

**Language/Version**: Go 1.22+ (api) · TypeScript 5.x / Node 20+ (web) — như 001/002/003

**Primary Dependencies**: như 001/002/003 (Gin, GORM, gorilla/websocket · Vue 3, Vite, Pinia) — **không thêm dependency mới**, không migration mới

**Storage**: PostgreSQL 16 — **chỉ ĐỌC** `transactions`, `accounts`, view `account_balances` (002), `categories` (001), `budgets` (003); không thêm bảng/cột

**Testing**: như 001/002/003 — Go unit (biz: cửa sổ tháng, tài sản ròng cuối tháng trước, tổng theo danh mục + %, gộp cây danh mục) + integration (storage đọc trên Postgres thật) + httptest (`GET /api/overview`, cô lập hộ) · Vitest (component thẻ/danh sách + store + format số) · Playwright e2e (bố cục màn Tổng quan, trang mặc định, realtime đa thành viên)

**Target Platform**: Web responsive mobile-first — như 001/002/003

**Project Type**: Web app monorepo `src/api` + `src/web` + `src/db`, chia sẻ theo hộ

**Performance Goals**: Màn Tổng quan hiển thị đủ tài sản ròng + Thu/Chi ngay khi mở, không thao tác thêm (SC-007); thay đổi của thành viên khác hiển thị ≤ 5s (SC-006); một round-trip cho phần suy ra (GET /api/overview) + tái dùng GET /api/budgets

**Constraints**: Mọi số liệu là **giá trị suy ra 100% khớp** dữ liệu nguồn tại mọi thời điểm (SC-001/002/003), kể cả sau sửa/xóa/đổi loại giao dịch quá khứ; chỉ đọc dữ liệu nguồn (không mutate); một loại tiền tệ; bố cục khớp **100%** `dashboard.png`

**Scale/Scope**: Mỗi hộ vài–vài chục danh mục, vài nghìn giao dịch; 6 phần màn Tổng quan + 1 endpoint tổng hợp + đổi định tuyến

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` vẫn là **bản mẫu chưa phê chuẩn** (toàn placeholder) — không có nguyên tắc ràng buộc.

- **Kết luận**: PASS (không có gate ràng buộc).
- **Re-check sau Phase 1**: PASS — chỉ thêm 1 module đọc + endpoint tổng hợp + UI; không dependency/bảng/service mới; không phát sinh độ phức tạp cần biện minh.

## Project Structure

### Documentation (this feature)

```text
specs/004-monthly-income-expense-overview/
├── plan.md              # This file
├── research.md          # Phase 0 — D27…D33 (kế thừa D1–D26 của 001/002/003)
├── data-model.md        # Phase 1 — DTO suy ra (không bảng mới)
├── quickstart.md        # Phase 1 — kịch bản kiểm chứng
├── contracts/
│   └── overview-api.md          # REST GET /api/overview (không có DB schema delta)
├── checklists/
│   └── requirements.md          # (đã có — checklist chất lượng spec)
└── tasks.md             # Phase 2 (/speckit-tasks — KHÔNG tạo ở bước này)
```

### Source Code (mở rộng cấu trúc 001/002/003 — layout learn_go)

```text
src/
├── api/
│   ├── module/
│   │   ├── overview/             # MỚI: chỉ ĐỌC, tổng hợp số liệu suy ra
│   │   │   ├── model/            #   OverviewSummary, CategorySpending (DTO suy ra — không map bảng)
│   │   │   ├── storage/          #   net worth (view account_balances) + net worth cuối tháng trước;
│   │   │   │                     #   Σ Thu/Chi tháng; Σ chi theo danh mục (gộp con); giao dịch gần đây
│   │   │   ├── biz/              #   cửa sổ tháng (dùng lại period của budget/model), % theo danh mục,
│   │   │   │                     #   % thay đổi tài sản ròng, ráp OverviewSummary
│   │   │   └── transport/ginoverview/  # GET /api/overview
│   │   └── (user/category/household/account/transaction/budget — GIỮ NGUYÊN, chỉ đọc)
│   └── main.go / main_route.go   # wire route GET /api/overview (scoped theo hộ)
├── web/src/
│   ├── views/
│   │   ├── DashboardView.vue      # DỰNG LẠI theo dashboard.png (6 phần đúng thứ tự) — nay ở "/"
│   │   └── LedgerView.vue         # GIỮ nội dung, dời route sang "/ledger"
│   ├── components/                # MỚI: NetWorthCard, IncomeExpenseCards, CategorySpendingList,
│   │                              #   RecentTransactionsList; DỜI: OverviewBudgetWidget (003)
│   ├── stores/                    # overview.ts (MỚI; refetch theo transactions/accounts/budgets/categories_changed)
│   ├── router/index.ts           # "/" = Dashboard; "/ledger" = Ledger; guard/redirect mặc định → "/"
│   └── App.vue                   # thanh nav 5 mục: Tổng quan · Giao dịch · ＋ · Ngân sách · Báo cáo
└── web/e2e/                      # spec 004 + CẬP NHẬT nav của e2e 001/002/003 (goto('/')→'/ledger', …)
```

**Structure Decision**: Một module đọc `overview` gói phần tổng hợp số liệu Tổng quan, giữ đúng chiều phụ thuộc (overview là hạ nguồn — chỉ đọc bảng của các module khác, không module nào phụ thuộc ngược overview). Không đụng logic 001/002/003; ngân sách vẫn tự phục vụ qua endpoint riêng. UI dựng lại một view + vài component thuần hiển thị.

## Dependency & thứ tự nền tảng

1. **Feature 001 + 002 + 003 đã triển khai**: giao dịch/tài khoản+số dư/danh mục/ngân sách + màn Tổng quan `/overview` (003) + WS realtime.
2. **Storage đọc (overview)**: net worth (view `account_balances`), net worth cuối tháng trước (Σ initial_balance + bút toán có dấu tới trước đầu tháng), Σ Thu/Chi tháng, Σ chi theo danh mục (gộp con), giao dịch gần đây.
3. **Biz + endpoint**: ráp `OverviewSummary`; `GET /api/overview` (scoped hộ).
4. **UI**: DashboardView theo dashboard.png (6 phần) + component mới + store + realtime.
5. **Định tuyến & mặc định**: `/` = Tổng quan; `/ledger` = sổ; đăng nhập/vào gốc → Tổng quan; nav 5 mục (Báo cáo = placeholder BR-004).
6. **Lan truyền e2e**: cập nhật điều hướng của e2e 001/002/003 theo route mới.
7. **Ngoài phạm vi**: báo cáo nâng cao (BR-004), sửa avatar/hồ sơ, đa tiền tệ.

## Tác động tới artifact khác

- `src/web/src/router/index.ts` + `App.vue` — đổi route mặc định & nav (lan truyền tới e2e 001/002/003).
- `specs/003-budgeting` — màn Tổng quan chuyển từ `/overview` sang `/` (trang chủ); widget ngân sách 003 được **dời vị trí** (sau Chi tiêu theo danh mục), logic không đổi. Ghi chú ở research (D31) — cập nhật CLAUDE.md status 003 phần route ở bước agent context.
- `specs/entities/entity-model.md` — **không đổi** (không thực thể/bảng mới; toàn giá trị suy ra).
- Module transaction/account/category/budget/household — **không sửa** (overview chỉ đọc).
- `CLAUDE.md` — cập nhật con trỏ plan (SPECKIT markers) sang plan 004.

## Complexity Tracking

> Không áp dụng — Constitution Check PASS, không vi phạm. Điểm đáng lưu ý duy nhất (đổi route mặc định làm ảnh hưởng e2e 001/002/003) được xử lý bằng cập nhật điều hướng test theo route mới (D31), không thêm dependency/độ phức tạp kiến trúc.
