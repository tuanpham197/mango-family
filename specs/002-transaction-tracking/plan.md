# Implementation Plan: Ghi Chép Thu Nhập và Chi Phí (Income & Expense Tracking) — Go + Vue

**Branch**: `002-transaction-tracking` (re-plan thực hiện trên branch `003-budgeting`, 2026-07-10) | **Date**: 2026-07-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-transaction-tracking/spec.md` (v4) · [BR-002](../business-requirements/BR-002.md) · [UC-TRK-01…05](../use-cases/002-transaction-tracking/) · [Entity model](../entities/entity-model.md) · [Re-platform design](../../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md) · **Nền tảng**: [plan 001 (Go+Vue)](../001-transaction-categorization/plan.md)

> ♻️ **Re-plan (2026-07-10)**: Thay thế plan Flutter/Supabase 2026-07-06 (xem git history). Spec/BR/use case KHÔNG đổi. Khác biệt lớn so với bản cũ: nền tảng định danh (users/auth/phạm vi hộ) nay do **feature 001 (re-plan)** dựng — plan này chỉ MỞ RỘNG.

## Summary

Mở rộng nền tảng 001 thành vòng đời giao dịch đầy đủ trong sổ chung của hộ: **nhập nhanh ≤ 15s**
với xác thực đầy đủ (số tiền > 0, loại, danh mục cùng loại, **tài khoản**, mô tả ≤ 255, **không ngày
tương lai**); **sổ chung** mới-nhất-trước hiển thị tên người nhập, realtime ≤ 5s; **sửa/xóa ngang quyền**
với chống ghi đè thầm lặng; **số dư tài khoản luôn khớp tổng bút toán** qua view suy ra `account_balances`.

**Technical approach**: Thêm `module/account` (đọc + số dư) và mở rộng `module/transaction` của 001
(account_id, updated_at, PATCH/DELETE, chặn ngày tương lai); 2 migration goose mới (`00006_accounts`,
`00007_transactions_v2`); tài khoản mặc định **"Tiền mặt"** seed bằng logic app khi tạo hộ (cùng chỗ
với seed danh mục mặc định của 001); WS phát thêm `accounts_changed`. Web: form giao dịch đầy đủ
(chế độ tạo/sửa) + sổ phân trang vô hạn + chip số dư.

## Technical Context

**Language/Version**: Go 1.22+ (api) · TypeScript 5.x / Node 20+ (web) — như 001

**Primary Dependencies**: như 001 (Gin, GORM, goose, gorilla/websocket, golang-jwt, bcrypt · Vue 3, Vite, Pinia); không thêm dependency mới

**Storage**: PostgreSQL 16 tự quản — thêm bảng `accounts` + view **`account_balances`** (00006), mở rộng `transactions` với `account_id` NOT NULL + `updated_at` (00007); migrations goose tại `src/db/migrations/`

**Testing**: như 001 — Go unit (biz) + integration (storage, Postgres docker) + httptest; Vitest; **Playwright** e2e theo 23 kịch bản quickstart (multi-context đa thành viên)

**Target Platform**: Web responsive mobile-first — như 001

**Project Type**: Web app monorepo `src/api` + `src/web` + `src/db`, multi-user chia sẻ theo hộ

**Performance Goals**: Nhập giao dịch ≤ 15s (SC-001); ≥ 95% lượt lưu hợp lệ ngay lần đầu (SC-002); sổ hiển thị thay đổi của thành viên khác trong ≤ 5 giây (SC-006 — WS); danh sách mượt với hàng nghìn giao dịch (phân trang 50/trang)

**Constraints**: Số dư = giá trị suy ra 100% khớp tổng bút toán (SC-004 — view, D14); không ngày tương lai (FR-005 — biz, D15); xóa luôn có xác nhận (SC-005 — UI dialog); không ghi đè thầm lặng (SC-007 — mốc `updated_at`, D17); một loại tiền tệ; audit log chi tiết ngoài phạm vi

**Scale/Scope**: Mỗi hộ 2–8 thành viên, vài tài khoản, hàng nghìn giao dịch; 5 use case UC-TRK-01…05 (UC-TRK-01 đã được 001 dựng — ở đây chỉ kiểm chứng lại)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` vẫn là **bản mẫu chưa phê chuẩn** (toàn placeholder) — không có nguyên tắc ràng buộc.

- **Kết luận**: PASS (không có gate ràng buộc).
- **Re-check sau Phase 1**: PASS — chỉ thêm 1 module + 1 bảng + 1 view trên mẫu 001; không phát sinh độ phức tạp cần biện minh.

## Project Structure

### Documentation (this feature)

```text
specs/002-transaction-tracking/
├── plan.md              # This file
├── research.md          # Phase 0 — D13…D19 (kế thừa D1–D12 của 001)
├── data-model.md        # Phase 1 — ACCOUNT + account_balances + TRANSACTION v2
├── quickstart.md        # Phase 1 — 23 kịch bản (giữ nguyên nghiệp vụ) trên Go+Vue
├── contracts/
│   ├── db-schema.sql            # DELTA schema (goose 00006/00007) trên nền 001
│   └── transaction-api.md       # REST API + WS events (thay transaction-repository.md cũ)
└── tasks.md             # Phase 2 (/speckit-tasks — tái sinh, KHÔNG tạo ở bước này)
```

### Source Code (mở rộng cấu trúc 001 — layout learn_go)

```text
src/
├── api/
│   ├── module/
│   │   ├── account/              # MỚI: model/storage/biz/transport — list tài khoản + số dư (view)
│   │   ├── transaction/          # MỞ RỘNG 001: account_id; update/delete có mốc updated_at;
│   │   │                         #   chặn ngày tương lai; publish transactions_changed/accounts_changed
│   │   └── household/            # MỞ RỘNG: SeedDefaults thêm tài khoản mặc định "Tiền mặt" (D13)
│   └── (user/category/component/middleware — giữ nguyên từ 001)
├── web/src/
│   ├── views/                    # TransactionFormView (tạo/sửa đầy đủ), LedgerView (phân trang vô hạn)
│   ├── components/               # AccountBalanceChip, DeleteTransactionDialog (+ tái dùng CategoryPicker, SuggestionChip)
│   └── stores/                   # accounts.ts (MỚI); transactions.ts mở rộng (update/delete/conflict)
├── web/e2e/                      # Playwright specs 002 (23 kịch bản)
└── db/migrations/                # 00006_accounts.sql · 00007_transactions_v2.sql (goose)
```

**Structure Decision**: Không module mới ngoài `account`; vòng đời giao dịch nằm trọn trong
`module/transaction` (mở rộng biz/storage/transport của 001). `TransactionEntryView` tối thiểu của 001
được **thay** bằng `TransactionFormView` đầy đủ (chế độ tạo + sửa); `LedgerView` tối thiểu của 001
nâng cấp thành sổ chính thức (route `/`). Seed tài khoản mặc định đặt cạnh seed danh mục mặc định
trong `module/household` — một chỗ duy nhất cho "hộ mới có gì".

## Dependency & thứ tự nền tảng

1. **Feature 001 (Go+Vue) là tiền đề TRỰC TIẾP**: users/auth/phạm vi hộ/WS/danh mục + transactions
   tối thiểu phải xong trước (tasks 001 T001–T019 tối thiểu).
2. **Accounts (00006)**: bảng + view `account_balances` + seed mặc định "Tiền mặt" (app logic D13) + backfill.
3. **Transactions v2 (00007)**: `account_id` (backfill về tài khoản mặc định → NOT NULL) + `updated_at`.
4. **Vòng đời UI**: form đầy đủ → sổ → sửa/xóa + concurrency.
5. **Quản lý tài khoản đầy đủ** (CRUD, chuyển tiền) thuộc BR-005; **quản lý hộ** thuộc feature riêng — ngoài phạm vi.

## Tác động tới artifact khác

- `specs/entities/entity-model.md` — ACCOUNT/TRANSACTION đã mô tả đúng (trung lập); không cần sửa.
- `specs/use-cases/002-*` — nghiệp vụ không đổi.
- Feature 003 (Budgeting) — đọc transactions/categories qua cùng nền tảng; re-plan 003 sau khi 002 chốt.
- `CLAUDE.md` — cập nhật trạng thái 002 (design rewritten) ở bước agent context.

## Complexity Tracking

> Không áp dụng — Constitution Check PASS, không có vi phạm cần biện minh.
