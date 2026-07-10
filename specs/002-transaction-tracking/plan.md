# Implementation Plan: Ghi Chép Thu Nhập và Chi Phí (Income & Expense Tracking)

**Branch**: `002-transaction-tracking` | **Date**: 2026-07-06 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-transaction-tracking/spec.md` · [BR-002](../business-requirements/BR-002.md) · [UC-TRK-01…05](../use-cases/002-transaction-tracking/) · [Entity model](../entities/entity-model.md)

> ⚠️ **LEGACY STACK (2026-07-10)**: Plan này mô tả kiến trúc **Flutter + Supabase đã gỡ bỏ** (re-platform Go + Vue — xem [design](../../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md)). Nghiệp vụ & bất biến vẫn đúng; phần kỹ thuật chờ viết lại bằng `/speckit-plan`.

## Summary

Cho phép thành viên hộ **nhập / xem / sửa / xóa giao dịch thu-chi** trong sổ chung của hộ:
nhập nhanh ≤ 15s với xác thực đầy đủ (số tiền > 0, loại, danh mục cùng loại, tài khoản, mô tả ≤ 255,
ngày không tương lai); sổ chung hiển thị **tên người nhập**; sửa/xóa ngang quyền với chống ghi đè
thầm lặng; **số dư tài khoản luôn khớp tổng bút toán** sau mọi thao tác.

**Bước nền tảng ĐẦU TIÊN (yêu cầu 2026-07-06)**: danh sách **người dùng độc lập** (`public.users` —
định danh riêng, không FK sang hệ xác thực; phiên đối chiếu qua email) làm nguồn định danh duy nhất;
mọi FK người (membership, `created_by`) trỏ về đây. Migration `0011_users.sql` đã viết sẵn; dev
refresh qua `setup_dev.sql`.

**Technical approach**: Tái dùng toàn bộ hạ tầng feature 001 (Flutter + Riverpod + Supabase, Clean
Architecture, RLS theo membership, realtime, cập nhật lạc quan R13). Thêm module
`src/lib/features/transactions/`; DB thêm `accounts` + mở rộng `transactions`
(`account_id`, `updated_at`) + **view `account_balances`** (số dư là giá trị suy ra — luôn nhất quán).

## Technical Context

**Language/Version**: Dart 3.x trên Flutter 3.x (stable) — như feature 001

**Primary Dependencies**: `flutter_riverpod`, `supabase_flutter` (Auth + Postgres + Realtime), `go_router`, `sqflite` (cache đọc), `intl` (định dạng tiền/ngày — thêm mới); test: `flutter_test`

**Storage**: Supabase PostgreSQL — bảng `users` (0011, độc lập), `accounts` (mới), `transactions` (mở rộng `account_id` + `updated_at`), view `account_balances`; RLS theo membership qua `current_user_id()` (đối chiếu email); `created_by` do trigger DB tự gán

**Testing**: `flutter_test` (unit + widget); kiểm chứng e2e theo `quickstart.md` (API-level qua REST như feature 001 + UI Chrome)

**Target Platform**: iOS 13+ / Android 8+ / Web (Chrome để dev-test) — như 001

**Project Type**: Mobile app (Flutter) + BaaS, multi-user chia sẻ theo hộ gia đình

**Performance Goals**: Nhập giao dịch ≤ 15s (SC-001); ≥ 95% lượt lưu hợp lệ ngay lần đầu (SC-002); sổ hiển thị thay đổi của thành viên khác trong ≤ 5 giây (SC-006); danh sách mượt với hàng nghìn giao dịch (phân trang)

**Constraints**: Số dư = giá trị suy ra, 100% khớp tổng bút toán (SC-004); không ngày tương lai (FR-005); xóa luôn có xác nhận (SC-005); không ghi đè thầm lặng (SC-007 — mốc `updated_at`); một loại tiền tệ; audit log chi tiết ngoài phạm vi

**Scale/Scope**: Mỗi hộ 2–8 thành viên, vài tài khoản, hàng nghìn giao dịch; phạm vi = 5 use case UC-TRK-01…05

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` vẫn là **bản mẫu chưa phê chuẩn** (toàn placeholder) — không có nguyên tắc ràng buộc.

- **Kết luận**: PASS (không có gate ràng buộc); không có vi phạm cần ghi vào Complexity Tracking.
- **Re-check sau Phase 1**: PASS — thiết kế tái dùng mẫu đã có của 001 (RLS, R13, view thay trigger phức tạp), thêm 1 bảng + 1 view, không phát sinh độ phức tạp cần biện minh.

## Project Structure

### Documentation (this feature)

```text
specs/002-transaction-tracking/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── db-schema.sql               # users + accounts + transactions delta + view + RLS
│   └── transaction-repository.md   # Hợp đồng domain cho giao dịch/sổ/người dùng hiện tại
└── tasks.md             # Phase 2 (/speckit-tasks — KHÔNG tạo ở bước này)
```

### Source Code (dưới `src/` — Flutter project root)

```text
src/
├── lib/
│   ├── core/
│   │   ├── auth/                 # login_screen (đã có từ 001) + current_user provider (MỚI: users.id qua email)
│   │   ├── household/            # current_household (đã đổi sang tra users theo email)
│   │   ├── supabase/             # client + realtime (đã có; transactions đã được subscribe)
│   │   ├── router/               # thêm routes /ledger, /txn/new, /txn/edit
│   │   └── error/                # thêm Failure: AccountRequired, FutureDateNotAllowed, DescriptionTooLong
│   └── features/
│       ├── categorization/       # (feature 001 — giữ nguyên; TransactionEntryScreen tối thiểu sẽ được
│       │                         #  thay thế bởi luồng đầy đủ của feature này)
│       └── transactions/         # MỚI — feature-first / Clean Architecture
│           ├── domain/           # entities (TransactionEntry, Account, UserProfile), repositories, usecases
│           ├── data/             # models, datasources (transaction/account/user), repository impl
│           └── presentation/     # screens (Ledger, TransactionForm), widgets, controllers (Riverpod)
├── supabase/
│   └── migrations/               # 0011_users.sql (đã có) · 0012_accounts.sql · 0013_transactions_v2.sql (MỚI)
└── test/
    ├── unit/                     # validate rules, balance roll-up logic, use cases
    └── widget/                   # form nhập (chặn lưu), ledger hiển thị tên người nhập
```

**Structure Decision**: Thêm module `src/lib/features/transactions/` tách khỏi `categorization`
(mỗi feature một module, dùng chung `core/`). `TransactionEntryScreen` tối thiểu của 001 (dựng để
kiểm chứng phân loại) sẽ được **thay thế** bằng `TransactionFormScreen` đầy đủ của module mới;
`CategorySelectField`/`SuggestionChip` của 001 được **tái dùng** trong form. Provider "người dùng
hiện tại" (`users.id` + tên hiển thị, tra qua email) đặt ở `core/auth` vì cả hai feature cùng dùng.

## Dependency & thứ tự nền tảng

1. **Người dùng (ĐẦU TIÊN — FR-015/016)**: migration `0011_users.sql` (đã viết) + `currentUserProvider`
   + màn đăng nhập (đã có). Mọi phần sau phụ thuộc.
2. **Tài khoản (phụ thuộc BR-005, thu gọn)**: bảng `accounts` + seed **tài khoản mặc định "Tiền mặt"**
   cho mỗi hộ (trigger khi tạo hộ + backfill hộ hiện có). Quản lý tài khoản đầy đủ thuộc BR-005.
3. **Giao dịch**: mở rộng bảng + luồng UI đầy đủ.
4. **Quản lý hộ** vẫn là tiền đề ngoài phạm vi (UC-HH-* — BR riêng, chưa viết).

## Tác động tới artifact khác

- `specs/entities/entity-model.md` — ĐÃ cập nhật (ACCOUNT mới, USER độc lập, TRANSACTION mở rộng).
- `specs/use-cases/002-transaction-tracking/` + sơ đồ — ĐÃ tạo/cập nhật.
- Feature 001: `data-model.md`/`contracts/db-schema.sql` đã ghi chú FK người dùng repoint (0011);
  không cần sửa thêm.
- `CLAUDE.md` — cập nhật "Active feature" + đường dẫn plan (làm ở Phase 1, bước agent context).

## Complexity Tracking

> Không áp dụng — Constitution Check PASS, không có vi phạm cần biện minh.
