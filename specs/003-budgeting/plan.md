# Implementation Plan: Thiết Lập Ngân Sách (Budgeting) — Go + Vue

**Branch**: `003-budgeting` | **Date**: 2026-07-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-budgeting/spec.md` (v3) · [BR-003](../business-requirements/BR-003.md) · [UC-BGT-01…06](../use-cases/003-budgeting/) · [Entity model](../entities/entity-model.md) · [Re-platform design](../../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md) · **Nền tảng**: [plan 001 (Go+Vue)](../001-transaction-categorization/plan.md) · [plan 002 (Go+Vue)](../002-transaction-tracking/plan.md)

> 🆕 **Plan đầu tiên của 003** (không có bản Flutter cũ để thay). Kế thừa TRỰC TIẾP nền tảng đã
> triển khai của 001 (users/auth/phạm vi hộ/WS/danh mục) + 002 (giao dịch đầy đủ, tài khoản, view số dư).
> 003 chỉ **đọc** giao dịch/danh mục để suy ra tiến độ; không sửa dữ liệu nguồn.

## Summary

Cho phép mọi thành viên trong hộ đặt **giới hạn chi tiêu** theo danh mục Chi hoặc theo tổng chi
của hộ trong một kỳ (hàng tháng / hàng tuần / một lần), **theo dõi tiến độ** ("đã chi/giới hạn (%)")
suy ra realtime từ giao dịch, và **nhận cảnh báo in-app** khi chạm ngưỡng 80% hoặc vượt 100% — mỗi
mức phát tối đa một lần/kỳ, phát lại nếu tụt xuống dưới rồi vượt lại. Ngân sách dùng chung trong hộ,
ngang quyền tạo/sửa/xóa với chống ghi đè thầm lặng.

**Technical approach**: Thêm module Go **`module/budget`** (`model/storage/biz/transport`) trên nền
002; 2 migration goose mới (`00008_budgets`, `00009_budget_alerts`). **Tiến độ = giá trị suy ra tính
ở biz khi đọc** (cửa sổ kỳ từ `now()` + `period_type`; ngân sách danh mục cộng cả danh mục con một
cấp; ngân sách tổng cộng mọi giao dịch Chi của hộ; chỉ loại EXPENSE). **Cảnh báo** dùng máy trạng
thái per `(budget, period_key, level)` lưu trong `budget_alerts` (chống trùng + phát lại); được **một
subscriber pubsub** đánh giá lại (idempotent, recompute toàn phần) mỗi khi `transactions_changed` phát,
rồi publish topic mới `budgets_changed` để WS đẩy về client ≤ 5s. Sửa/xóa dùng lại mốc lạc quan
`updated_at` (D6/D17). Web: `BudgetListView` (tiến độ + badge cảnh báo amber/đỏ), `BudgetFormView`
(tạo/sửa), widget ngân sách ở màn Tổng quan; store `budgets.ts`.

## Technical Context

**Language/Version**: Go 1.22+ (api) · TypeScript 5.x / Node 20+ (web) — như 001/002

**Primary Dependencies**: như 001/002 (Gin, GORM, goose, gorilla/websocket, golang-jwt, bcrypt · Vue 3, Vite, Pinia); **không thêm dependency mới** — cảnh báo dùng lại pubsub/wshub/subscriber sẵn có

**Storage**: PostgreSQL 16 tự quản — thêm bảng `budgets` (00008) + `budget_alerts` (00009); chỉ **đọc** `transactions`/`categories`; migrations goose tại `src/db/migrations/`

**Testing**: như 001/002 — Go unit (biz: cửa sổ kỳ, subtree danh mục, máy trạng thái cảnh báo) + integration (storage + evaluator, Postgres docker) + httptest; Vitest (store/tiến độ %); **Playwright** e2e theo kịch bản quickstart 003 (multi-context đa thành viên, cảnh báo realtime)

**Target Platform**: Web responsive mobile-first — như 001/002

**Project Type**: Web app monorepo `src/api` + `src/web` + `src/db`, multi-user chia sẻ theo hộ

**Performance Goals**: Tạo ngân sách ≤ 30s / ≤ 3 bước (SC-001); 100% tiến độ khớp tổng giao dịch liên quan tại mọi thời điểm đối chiếu (SC-003); thay đổi tiến độ/cảnh báo hiển thị cho thành viên khác ≤ 5 giây (SC-006 — WS)

**Constraints**: Tiến độ = giá trị suy ra 100% khớp tổng giao dịch Chi trong kỳ, kể cả sau sửa/xóa/đổi danh mục quá khứ (SC-003 — biz, D21); mỗi mức cảnh báo phát tối đa 1 lần/kỳ, phát lại sau khi tụt dưới mức (FR-009 — máy trạng thái D23); ngưỡng cố định 80% + kênh in-app (Assumptions spec); ngân sách chỉ đọc dữ liệu nguồn; không ghi đè thầm lặng (D17); một loại tiền tệ; ngưỡng cấu hình + push/email ngoài phạm vi

**Scale/Scope**: Mỗi hộ 2–8 thành viên, vài–vài chục ngân sách; 6 use case UC-BGT-01…06; đánh giá cảnh báo chạy per mutation giao dịch (tần suất thấp theo hộ)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` vẫn là **bản mẫu chưa phê chuẩn** (toàn placeholder) — không có nguyên tắc ràng buộc.

- **Kết luận**: PASS (không có gate ràng buộc).
- **Re-check sau Phase 1**: PASS — chỉ thêm 1 module + 2 bảng + 1 subscriber trên mẫu 001/002; không phát sinh độ phức tạp cần biện minh (không dependency mới, không service mới).

## Project Structure

### Documentation (this feature)

```text
specs/003-budgeting/
├── plan.md              # This file
├── research.md          # Phase 0 — D20…D26 (kế thừa D1–D19 của 001/002)
├── data-model.md        # Phase 1 — BUDGET + BUDGET_ALERT + tiến độ suy ra
├── quickstart.md        # Phase 1 — kịch bản kiểm chứng US1–US5 + edge cases
├── contracts/
│   ├── db-schema.sql            # DELTA schema (goose 00008/00009) trên nền 001/002
│   └── budget-api.md            # REST API + WS event budgets_changed
├── checklists/
│   └── requirements.md          # (đã có — checklist chất lượng spec)
└── tasks.md             # Phase 2 (/speckit-tasks — KHÔNG tạo ở bước này)
```

### Source Code (mở rộng cấu trúc 001/002 — layout learn_go)

```text
src/
├── api/
│   ├── module/
│   │   ├── budget/               # MỚI: model/storage/biz/transport
│   │   │   ├── model/            #   Budget, BudgetAlert, BudgetProgress (suy ra), period key
│   │   │   ├── storage/          #   CRUD budgets; đọc tổng chi theo kỳ/danh mục(+con)/tổng; CRUD budget_alerts
│   │   │   ├── biz/              #   cửa sổ kỳ, subtree, %, máy trạng thái cảnh báo (80/100, phát lại), uniqueness
│   │   │   └── transport/ginbudget/  # REST /api/budgets/*; embed spent/percent/alerts
│   │   └── (user/category/household/account/transaction — GIỮ NGUYÊN, chỉ đọc)
│   ├── component/
│   │   ├── budgeteval/           # MỚI (hoặc trong module/budget): subscriber pubsub → đánh giá cảnh báo
│   │   └── (pubsub/subscriber/wshub — GIỮ NGUYÊN; thêm topic budgets_changed vào common/const.go)
│   ├── cmd/seed/                 # (tùy chọn) seed vài ngân sách mẫu cho hộ dev
│   └── main.go                   # wire budget routes + start budget evaluator subscriber
├── web/src/
│   ├── views/                    # BudgetListView (tiến độ + cảnh báo), BudgetFormView (tạo/sửa),
│   │                             #   DashboardView (màn Tổng quan chứa tóm tắt ngân sách — FR-014, wireframe màn 1)
│   ├── components/               # BudgetProgressBar, BudgetAlertBadge, DeleteBudgetDialog, OverviewBudgetWidget
│   └── stores/                   # budgets.ts (MỚI; refetch khi budgets_changed)
├── web/e2e/                      # Playwright specs 003
└── db/migrations/                # 00008_budgets.sql · 00009_budget_alerts.sql (goose)
```

**Structure Decision**: Một module mới `budget` gói trọn vòng đời ngân sách + tính tiến độ + cảnh báo.
Việc đánh giá cảnh báo theo sự kiện đặt trong một **subscriber pubsub** riêng (giữ đúng chiều phụ thuộc:
budget là hạ nguồn của transaction — transaction KHÔNG biết budget; xem D24). Không đụng vào module
transaction/category/account (chỉ đọc bảng của chúng qua storage của budget). Thêm đúng một topic WS
`budgets_changed`. UI thêm 2 view + widget Tổng quan, không sửa luồng nhập giao dịch của 002.

## Dependency & thứ tự nền tảng

1. **Feature 001 + 002 (Go+Vue) là tiền đề TRỰC TIẾP đã triển khai**: users/auth/phạm vi hộ/WS,
   danh mục (+cây một cấp), giao dịch đầy đủ (loại EXPENSE, `transaction_date`), pubsub/wshub.
2. **Budgets (00008)**: bảng `budgets` + partial unique index (uniqueness FR-010) + `updated_at`.
3. **Budget alerts (00009)**: bảng `budget_alerts` (chống trùng theo `(budget_id, period_key, level)`).
4. **Biz tiến độ**: cửa sổ kỳ từ `now()`+`period_type`; subtree danh mục một cấp; tổng theo hộ; chỉ EXPENSE.
5. **Evaluator cảnh báo**: subscriber pubsub trên `transactions_changed` + CRUD ngân sách → recompute idempotent → publish `budgets_changed`.
6. **Vòng đời UI**: form tạo/sửa → danh sách tiến độ → cảnh báo realtime → xóa có xác nhận + widget Tổng quan.
7. **Ngoài phạm vi**: ngưỡng cấu hình được, push/email, savings goals, rollover, đa tiền tệ (Out of Scope BR-003).

## Tác động tới artifact khác

- `specs/entities/entity-model.md` — **THÊM** thực thể `BUDGET` + `BUDGET_ALERT` và quan hệ (HOUSEHOLD owns BUDGET; CATEGORY tùy chọn; BUDGET has BUDGET_ALERT). Cập nhật ở Phase 1.
- `specs/003-budgeting/spec.md` — Key Entities đã mô tả Budget/Budget Alert (trung lập) → không đổi nghĩa; References mục "Dẫn xuất (design)" nay có plan/research/data-model/contracts/quickstart.
- `specs/use-cases/003-*` — nghiệp vụ không đổi (UC-BGT-01…06 vừa tạo 2026-07-13).
- `CLAUDE.md` — cập nhật con trỏ plan (SPECKIT markers) + trạng thái 003 (design generated) ở bước agent context.
- Module transaction/category/account/household — **không sửa** (budget chỉ đọc).

## Complexity Tracking

> Không áp dụng — Constitution Check PASS, không có vi phạm cần biện minh. Điểm phức tạp duy nhất
> (đánh giá cảnh báo theo sự kiện) được xử lý bằng hạ tầng pubsub/wshub sẵn có + recompute idempotent,
> không thêm dependency hay service ngoài.
