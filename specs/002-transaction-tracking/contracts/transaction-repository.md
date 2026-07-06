# Contract — TransactionRepository & CurrentUser (domain interfaces)

**Feature**: 002-transaction-tracking · **Plan**: [../plan.md](../plan.md) ·
**Data model**: [../data-model.md](../data-model.md)

Hợp đồng tầng domain mà `presentation` phụ thuộc và `data` (Supabase) hiện thực. Đây là
**hợp đồng hành vi** — không phải code cài đặt. Mọi thao tác chạy trong ngữ cảnh **hộ hiện tại**
và **người dùng hiện tại** (users độc lập — phiên đối chiếu qua email, R15). Tiền điều kiện chung:
**đã đăng nhập và là thành viên hộ** (UC-TRK-01).

> Kiểu dữ liệu:
> `UserProfile { id, email, displayName }`
> `Account { id, householdId, name, type: CASH|BANK|EWALLET|CREDIT }`
> `TransactionEntry { id, householdId, accountId, categoryId, type: INCOME|EXPENSE, amount, description?, transactionDate, createdBy, createdByName, updatedAt }`
> `TransactionDraft { accountId, categoryId, type, amount, description?, transactionDate }`

## CurrentUser (core/auth)

| Thao tác | Mô tả | Hậu điều kiện / Lỗi | Truy vết |
|----------|-------|----------------------|----------|
| `currentUser()` | Hồ sơ người dùng hiện tại (đối chiếu phiên qua email) | Trả `UserProfile`; chưa đăng nhập → `NotAuthenticated`; chưa có hồ sơ → tự tạo từ email rồi trả (UC-TRK-01 5a) | FR-015, FR-016, R15 |

## AccountRepository (chỉ đọc trong phạm vi 002)

| Thao tác | Mô tả | Hậu điều kiện / Lỗi | Truy vết |
|----------|-------|----------------------|----------|
| `listAccounts()` | Tài khoản của hộ hiện tại | Luôn ≥ 1 (tài khoản mặc định — R17); kèm số dư từ `account_balances` | FR-006, FR-011 |

## TransactionRepository

| Thao tác | Mô tả | Tiền điều kiện | Hậu điều kiện / Lỗi | Truy vết |
|----------|-------|----------------|----------------------|----------|
| `listTransactions({page})` | Sổ của hộ, mới nhất trước, kèm `createdByName`/danh mục/tài khoản | Đã đăng nhập | Chỉ giao dịch hộ mình (RLS); phân trang 50/trang | UC-TRK-03, FR-008, FR-012 |
| `addTransaction(draft)` | Nhập giao dịch mới | Draft qua xác thực client | Lưu với `created_by` = người dùng hiện tại (trigger DB tự gán); số dư suy ra tự khớp. **Lỗi**: `AmountInvalid` (≤ 0); `CategoryRequired`; `TypeMismatch` (khác loại — hàng rào DB); `AccountRequired`; `HouseholdMismatch`; `FutureDateNotAllowed`; `DescriptionTooLong` (> 255) | UC-TRK-02, FR-001…007, FR-011, FR-013 |
| `updateTransaction(id, draft, {expectedUpdatedAt})` | Sửa mọi trường, xác thực như nhập mới | Giao dịch tồn tại; mốc `expectedUpdatedAt` là bản đang xem | Ghi có điều kiện mốc (R20): 0 hàng khớp → còn bản ghi ⇒ `ConcurrencyConflict` (đã đổi — tải lại); không còn ⇒ `RecordGone` (đã bị xóa). Đổi loại ⇒ danh mục phải cùng loại mới (`TypeMismatch` nếu không) | UC-TRK-04, FR-009, FR-011, FR-014 |
| `deleteTransaction(id)` | Xóa vĩnh viễn (UI đã xác nhận) | Giao dịch tồn tại | Xóa xong số dư suy ra tự khớp; 0 hàng (người khác xóa trước) → `RecordGone` (thông báo nhẹ + làm tươi sổ) | UC-TRK-05, FR-010, FR-011, FR-014 |

**Ghi chú hợp đồng**: bước **xác nhận xóa** là trách nhiệm của presentation (FR-010, SC-005) —
repository không tự xác nhận. Gán danh mục trong form dùng lại hợp đồng
[CategoryRepository của 001](../../001-transaction-categorization/contracts/category-repository.md)
(`listCategories` lọc theo loại, `suggestCategory`).

## Ánh xạ lỗi nghiệp vụ ↔ hàng rào DB

| Lỗi domain | Hàng rào DB (contracts/db-schema.sql) |
|------------|----------------------------------------|
| `AmountInvalid` | `check (amount > 0)` (001) |
| `CategoryRequired` / `TypeMismatch` | `category_id NOT NULL` + `enforce_txn_rules` (001) |
| `AccountRequired` / `HouseholdMismatch` | `account_id NOT NULL` + `enforce_txn_rules` mở rộng (0013) |
| `FutureDateNotAllowed` | `enforce_txn_rules` — chặn ngày tương lai (0013, R18) |
| `DescriptionTooLong` | `check (char_length(description) <= 255)` (001) |
| `ConcurrencyConflict` / `RecordGone` | ghi/xóa có điều kiện `updated_at`/`id` — 0 hàng khớp (R20) |
| `NotAuthenticated` | RLS + `current_user_id()` trả null (0011) |
| Số dư luôn khớp | view `account_balances` — giá trị suy ra (0012, R16) |
