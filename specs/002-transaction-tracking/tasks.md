---
description: "Task list — Ghi chép thu nhập & chi phí (sổ chung hộ gia đình)"
---

# Tasks: Ghi Chép Thu Nhập và Chi Phí (Income & Expense Tracking)

> ⚠️ **LEGACY STACK 2026-07-10 — Re-platform Go + Vue**: Các task bên dưới được sinh cho stack Flutter/Supabase đã gỡ bỏ (chưa task nào thực hiện). Phasing theo user story và nội dung nghiệp vụ vẫn đúng, nhưng đường dẫn/công nghệ sẽ được **tái sinh** khi re-plan theo stack mới — xem [design re-platform](../../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md).

**Input**: Design documents from `/specs/002-transaction-tracking/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md), [UC-TRK-01…05](../use-cases/002-transaction-tracking/)

**Stack**: Flutter (Dart 3.x) · Riverpod · Supabase (PostgreSQL + Auth + RLS + Realtime). Tái dùng hạ tầng feature 001 (✅ implemented).

**Tests**: Spec không yêu cầu TDD → không sinh phase test riêng. Kiểm chứng end-to-end theo 23 kịch bản trong [quickstart.md](./quickstart.md); unit/widget test bổ sung ở Polish.

**Path conventions**: module mới `src/lib/features/transactions/{domain,data,presentation}`; hạ tầng chung `src/lib/core/`; migration `src/supabase/migrations/`; script dev `src/supabase/setup_dev.sql`.

> **Thứ tự nền tảng (yêu cầu 2026-07-06)**: **NGƯỜI DÙNG ĐẦU TIÊN** — bảng `users` độc lập (migration `0011_users.sql` đã viết) phải áp & kiểm trước, rồi mới đến tài khoản (0012) và giao dịch (0013). US5 (đăng nhập/định danh) là story nền tảng, thực hiện trước US1–US4.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Khung module transactions + tiện ích chung.

- [ ] T001 Tạo cấu trúc module theo plan.md: `src/lib/features/transactions/{domain/{entities,repositories,usecases},data/{models,datasources,repositories},presentation/{screens,widgets,controllers}}/` và thêm dependency `intl` vào `src/pubspec.yaml`
- [ ] T002 [P] Bổ sung Failure mới (`NotAuthenticated`, `AccountRequired`, `FutureDateNotAllowed`, `DescriptionTooLong`, `RecordGone`, `AmountInvalid`) trong `src/lib/core/error/failures.dart`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Users độc lập (ĐẦU TIÊN) → accounts + view số dư → transactions v2 → hạ tầng domain/data dùng chung.

**⚠️ CRITICAL**: Không bắt đầu US nào trước khi phase này xong. Thứ tự migration bắt buộc: **T003 (users/0011) → T005 (0012) → T006 (0013)**.

- [ ] T003 **Áp & kiểm migration `0011_users.sql` (users ĐỘC LẬP — đã viết)**: dán `src/supabase/setup_dev.sql` (refresh) vào Supabase SQL Editor; verify qua API: `public.users` = 3, `users.id ≠` id đăng nhập, FK `household_members.user_id`/`*.created_by` → `users(id)`, `current_user_id()` hoạt động *(tiền đề bắt buộc — mọi task sau phụ thuộc)*
- [ ] T004 Provider "người dùng hiện tại" (`users.id` + `displayName`, đối chiếu email; tự tạo hồ sơ nếu thiếu — UC-TRK-01 5a) trong `src/lib/core/auth/current_user.dart` (contracts §CurrentUser)
- [ ] T005 Migration: bảng `accounts` + RLS `member_accounts` + trigger `set_created_by` + trigger/backfill **tài khoản mặc định "Tiền mặt"** + view **`account_balances`** (security_invoker) trong `src/supabase/migrations/0012_accounts.sql` (data-model §ACCOUNT, R16, R17)
- [ ] T006 Migration: `transactions` v2 — thêm `account_id` (backfill về tài khoản mặc định rồi NOT NULL) + `updated_at` + trigger touch + `enforce_txn_rules` mở rộng (account cùng hộ, **chặn ngày tương lai**) trong `src/supabase/migrations/0013_transactions_v2.sql` (data-model §TRANSACTION, R18, R20)
- [ ] T007 Gộp 0012 + 0013 vào `src/supabase/setup_dev.sql` (idempotent) và dán lại SQL Editor; verify: mỗi hộ có tài khoản "Tiền mặt", `account_balances` trả số dư 0
- [ ] T008 [P] Domain entities `TransactionEntry`, `Account`, `UserProfile` trong `src/lib/features/transactions/domain/entities/`
- [ ] T009 Interface `TransactionRepository` + `AccountRepository` (theo `contracts/transaction-repository.md`) trong `src/lib/features/transactions/domain/repositories/`
- [ ] T010 [P] DTO models + mappers (embed `users(display_name)`, `categories(name,icon)`, `accounts(name)`) trong `src/lib/features/transactions/data/models/transaction_model.dart`
- [ ] T011 Datasources Supabase: `transaction_remote_datasource.dart` (list embed + insert + update/delete có điều kiện mốc `updated_at` — R20) và `account_remote_datasource.dart` (accounts + `account_balances`) trong `src/lib/features/transactions/data/datasources/` (phụ thuộc T004)
- [ ] T012 `TransactionRepositoryImpl` + `AccountRepositoryImpl` + providers (ghép datasource + hộ hiện tại + người dùng hiện tại; phân biệt `ConcurrencyConflict` vs `RecordGone`) trong `src/lib/features/transactions/data/repositories/transaction_repository_impl.dart` (phụ thuộc T009, T011)
- [ ] T013 [P] Router: thêm routes `/ledger`, `/txn/new`, `/txn/edit` trong `src/lib/core/router/app_router.dart`

**Checkpoint**: users/accounts/transactions v2 sẵn sàng trên DB; repo + provider ghép xong → bắt đầu các story.

---

## Phase 3: User Story 5 - Đăng nhập và định danh người dùng (Priority: P1 — nền tảng, ĐẦU TIÊN) 🎯

**Goal**: Đăng nhập bằng tài khoản của mình; hệ thống biết "tôi là ai" (tên hiển thị) và "tôi thuộc hộ nào"; bản ghi mới ghi đúng người tạo. (UC-TRK-01)

**Independent Test**: quickstart kịch bản #1–#3 (đăng nhập đúng hộ + đúng tên; chặn khi chưa đăng nhập; sai mật khẩu bị từ chối an toàn).

> Màn đăng nhập + guard router đã có từ 001; story này hoàn thiện phần **định danh hồ sơ**.

- [ ] T014 [US5] Hiển thị tên người dùng hiện tại (từ `currentUserProvider`) + nút đăng xuất trên app bar các màn chính trong `src/lib/features/transactions/presentation/widgets/current_user_badge.dart` (dùng ở ledger/manage)
- [ ] T015 [US5] Kiểm chứng luồng định danh: đăng nhập lần đầu tự tạo hồ sơ nếu thiếu (UC-TRK-01 5a); chưa thuộc hộ → thông báo rõ (E2) — hoàn thiện trong `src/lib/core/auth/current_user.dart` + `src/lib/core/household/current_household.dart`

**Checkpoint**: US5 pass quickstart #1–#3 — nền tảng định danh sẵn sàng cho mọi story.

---

## Phase 4: User Story 1 - Nhập giao dịch thu/chi nhanh và hợp lệ (Priority: P1) 🎯 MVP

**Goal**: Nhập giao dịch ≤ 15s với xác thực đầy đủ; số dư tài khoản cập nhật đúng. (UC-TRK-02)

**Independent Test**: quickstart kịch bản #4–#10 (nhập hợp lệ + số dư; chặn số tiền/danh mục/mô tả/ngày tương lai; ngày mặc định; lọc danh mục theo loại).

- [ ] T016 [P] [US1] Use case `AddTransaction` (validate: amount > 0, mô tả ≤ 255, ngày ≤ hiện tại, category/account bắt buộc) trong `src/lib/features/transactions/domain/usecases/add_transaction.dart`
- [ ] T017 [US1] Màn `TransactionFormScreen` (chế độ tạo): số tiền, loại Thu/Chi, `CategorySelectField` + `SuggestionChip` tái dùng từ 001, chọn tài khoản (chọn sẵn khi hộ chỉ có 1 — FR-006), date picker giới hạn `lastDate = hôm nay`, mô tả `maxLength: 255` trong `src/lib/features/transactions/presentation/screens/transaction_form_screen.dart`
- [ ] T018 [US1] Controller form (submit qua `AddTransaction`, hiển thị lỗi theo trường, làm tươi ledger/số dư sau lưu; chống double-submit — thử lại sau mất kết nối không tạo bản ghi trùng, quickstart #23) trong `src/lib/features/transactions/presentation/controllers/transaction_form_controller.dart`
- [ ] T019 [US1] Widget hiển thị số dư tài khoản (từ `account_balances`) trong `src/lib/features/transactions/presentation/widgets/account_balance_chip.dart`
- [ ] T020 [US1] Thay màn nhập tối thiểu của 001: xóa `src/lib/features/categorization/presentation/screens/transaction_entry_screen.dart`, route `/txn` cũ trỏ về `/txn/new` mới trong `src/lib/core/router/app_router.dart`

**Checkpoint**: US5 + US1 = MVP demo được (đăng nhập → nhập giao dịch hợp lệ → số dư đúng).

---

## Phase 5: User Story 2 - Xem sổ giao dịch chung của hộ (Priority: P2)

**Goal**: Sổ chung mới-nhất-trước, hiển thị **tên** người nhập; realtime giữa thành viên; điểm vào sửa/xóa. (UC-TRK-03)

**Independent Test**: quickstart kịch bản #11–#13 (Bob thấy giao dịch Alice kèm tên; thứ tự & phân trang; Carol cô lập).

- [ ] T021 [P] [US2] Use case `ListTransactions` (phân trang 50/trang) trong `src/lib/features/transactions/domain/usecases/list_transactions.dart`
- [ ] T022 [US2] Controller sổ (paging + realtime invalidation qua `subscribeHouseholdChanges` đã có) trong `src/lib/features/transactions/presentation/controllers/ledger_controller.dart`
- [ ] T023 [US2] Màn `LedgerScreen`: danh sách mới nhất trước, mỗi dòng số tiền/loại/danh mục/ngày/**tên người nhập** (email nếu thiếu tên — UC-TRK-03 3a), infinite scroll, empty state hướng đi nhập trong `src/lib/features/transactions/presentation/screens/ledger_screen.dart`
- [ ] T024 [US2] Đặt Ledger làm màn hình chính `/`: điều hướng sang Quản lý danh mục + Nhập giao dịch từ app bar trong `src/lib/core/router/app_router.dart`

**Checkpoint**: US1 + US2 hoạt động độc lập; sổ chung minh bạch.

---

## Phase 6: User Story 3 - Chỉnh sửa giao dịch (Priority: P3)

**Goal**: Sửa mọi trường với xác thực như nhập mới; đổi loại buộc chọn lại danh mục; số dư tính lại; không ghi đè thầm lặng. (UC-TRK-04)

**Independent Test**: quickstart kịch bản #14–#18 (số dư chênh lệch đúng; đổi loại; ngang quyền; xung đột đồng thời; bị xóa trong lúc sửa).

- [ ] T025 [P] [US3] Use case `UpdateTransaction` (validate như nhập mới + mốc `expectedUpdatedAt`) trong `src/lib/features/transactions/domain/usecases/update_transaction.dart`
- [ ] T026 [US3] Chế độ sửa trong `TransactionFormScreen` (prefill; đổi loại → `CategorySelectField` coi danh mục cũ là chưa chọn; xử lý `ConcurrencyConflict` → tải lại dữ liệu mới, `RecordGone` → thông báo + về sổ) trong `src/lib/features/transactions/presentation/screens/transaction_form_screen.dart`
- [ ] T027 [US3] Điểm vào sửa từ sổ (tap dòng giao dịch → `/txn/edit`) trong `src/lib/features/transactions/presentation/screens/ledger_screen.dart`

**Checkpoint**: US1 + US2 + US3 hoạt động độc lập.

---

## Phase 7: User Story 4 - Xóa giao dịch có xác nhận (Priority: P4)

**Goal**: Xóa luôn qua cảnh báo/xác nhận; số dư hoàn tác đúng; xử lý đã-bị-xóa-trước. (UC-TRK-05)

**Independent Test**: quickstart kịch bản #19–#21 (dialog xác nhận; hủy không đổi gì; xóa xong số dư hoàn tác, sổ mọi người cập nhật).

- [ ] T028 [P] [US4] Use case `DeleteTransaction` trong `src/lib/features/transactions/domain/usecases/delete_transaction.dart`
- [ ] T029 [US4] Nút xóa từ sổ/form + dialog cảnh báo xác nhận (nêu rõ vĩnh viễn — SC-005); `RecordGone` → thông báo nhẹ + làm tươi sổ trong `src/lib/features/transactions/presentation/widgets/delete_transaction_dialog.dart`

**Checkpoint**: Cả 5 user story hoạt động độc lập.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Test tự động, kiểm chứng end-to-end, đồng bộ tài liệu.

- [ ] T030 [P] Unit test (validate AddTransaction/UpdateTransaction: amount/mô tả/ngày; phân biệt conflict) trong `src/test/unit/transactions_domain_test.dart` + widget test (form chặn lưu khi thiếu danh mục/tài khoản; ledger hiển thị tên người nhập) trong `src/test/widget_test.dart`
- [ ] T031 Chạy kiểm chứng `quickstart.md` (23 kịch bản — UI Chrome + API e2e đa thành viên #11/#13/#16–#18/#21 với Alice/Bob/Carol như cách feature 001; đối chiếu số dư #22 bằng truy vấn `account_balances`; #23 kiểm chống trùng khi thử lại)
- [ ] T032 [P] Cập nhật tài liệu & khối References/History các artifact theo quy tắc lan truyền `CLAUDE.md` (BR-002 → implemented nếu đủ; entity-model/data-model nếu lệch; chạy `/speckit-analyze`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: bắt đầu ngay.
- **Foundational (Phase 2)**: phụ thuộc Setup. **BLOCKS mọi user story.** Trong phase: **T003 (users — ĐẦU TIÊN)** → T005 → T006 → T007 (nhánh migration, tuần tự); T004 sau T003; T008–T010 song song; T011 phụ thuộc T004; T012 phụ thuộc T009+T011; T013 song song.
- **US5 (Phase 3)**: ngay sau Foundational — nền tảng định danh cho mọi story còn lại.
- **US1–US4 (Phase 4–7)**: sau US5; theo thứ tự ưu tiên hoặc song song (US3/US4 cần US2 làm điểm vào UI nhưng use case/repo độc lập).
- **Polish (Phase 8)**: sau các story mong muốn.

### User Story Dependencies

- **US5 (P1 nền tảng)**: chỉ phụ thuộc Foundational — làm TRƯỚC TIÊN.
- **US1 (P1)**: phụ thuộc US5; là MVP cùng US5.
- **US2 (P2)**: phụ thuộc US5; độc lập US1 (dữ liệu test có thể seed).
- **US3 (P3)** / **US4 (P4)**: phụ thuộc US5; điểm vào UI từ US2 (T027, T029 chạm `ledger_screen.dart` — sau T023).

### Parallel Opportunities

- Setup: T001, T002 song song.
- Foundational: T008, T010, T013 song song với nhánh migration; T004 song song T005/T006.
- Use case các story (T016, T021, T025, T028) đều [P] — khác file, có thể viết song song trước khi ghép UI.
- Sau US5: Dev A → US1 (form), Dev B → US2 (ledger); US3/US4 nối sau US2.

---

## Parallel Example: Foundational

```bash
# Nhánh migration (tuần tự, users ĐẦU TIÊN):
Task: "T003 áp 0011_users.sql qua setup_dev.sql + verify API"
Task: "T005 0012_accounts.sql" → "T006 0013_transactions_v2.sql" → "T007 gộp setup_dev.sql"
# Song song nhánh Dart:
Task: "T008 entities in src/lib/features/transactions/domain/entities/"
Task: "T010 DTO models in src/lib/features/transactions/data/models/transaction_model.dart"
Task: "T013 routes in src/lib/core/router/app_router.dart"
```

---

## Implementation Strategy

### MVP First (US5 + US1)

1. Phase 1 Setup → 2. Phase 2 Foundational (**T003 users trước tiên**) → 3. Phase 3 US5 (định danh) → 4. Phase 4 US1 (nhập giao dịch) → 5. **DỪNG & kiểm chứng** quickstart #1–#10 → demo MVP.

### Incremental Delivery

US5 (định danh) → US1 (nhập — MVP) → US2 (sổ chung + realtime) → US3 (sửa + concurrency) → US4 (xóa) → Polish (test + e2e 23 kịch bản + docs). Mỗi story kiểm chứng độc lập theo nhóm kịch bản quickstart của nó trước khi sang story sau.

---

## Notes

- [P] = khác file, không phụ thuộc nhau. Nhãn [US#] gắn task với user story (US5 = story nền tảng).
- **T003 là tiền đề bắt buộc** — users độc lập phải xong & verify trước mọi thứ (yêu cầu 2026-07-06).
- Số dư KHÔNG lưu cột riêng — view `account_balances` (R16); mọi kiểm chứng số dư đối chiếu qua view.
- Quản lý tài khoản đầy đủ (CRUD, chuyển tiền) thuộc BR-005; quản lý hộ thuộc BR riêng — ở đây chỉ seed mặc định.
- Commit sau mỗi task hoặc nhóm logic; dừng ở mỗi Checkpoint để kiểm chứng story độc lập.
