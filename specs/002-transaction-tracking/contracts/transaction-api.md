# Contract: REST API + WebSocket — Ghi Chép Thu Nhập và Chi Phí (feature 002)

**Date**: 2026-07-10 · **Plan**: [../plan.md](../plan.md) · **Schema**: [db-schema.sql](./db-schema.sql) · **Nền tảng**: [category-api.md (001)](../../001-transaction-categorization/contracts/category-api.md)

> Thay thế `transaction-repository.md` (hợp đồng domain Flutter cũ — đã xóa). Mở rộng hợp đồng 001:
> Auth (`/api/auth/*`, `/api/me`), Categories và quy ước chung (format `app_response`, cookie JWT,
> phạm vi hộ → 404, mã lỗi CONCURRENCY_CONFLICT/RECORD_GONE) **giữ nguyên từ 001** — không lặp lại ở đây.

## Accounts (002 chỉ đọc — quản lý đầy đủ thuộc BR-005)

| Method & Path | Query | Response | Truy vết |
|---|---|---|---|
| `GET /api/accounts` | — | `{data:[{id, name, type, balance}]}` — `balance` lấy từ view `account_balances` | FR-006, FR-011, D13/D14 |

## Transactions (vòng đời đầy đủ)

**Transaction** = `{id, household_id, amount, type, category_id, category_name, account_id, account_name, description, transaction_date, created_by, created_by_name, updated_at}`.

| Method & Path | Body / Query | Response | Truy vết |
|---|---|---|---|
| `POST /api/transactions` | `{amount, type, category_id, account_id, description?, transaction_date?}` — `transaction_date` mặc định now; `created_by` server gán từ phiên | `201 {data:Transaction}` | FR-001…007, FR-013 |
| `GET /api/transactions` | `?page=&page_size=` (mặc định 50) | `{data:[Transaction], paging:{page, page_size, total}}` — mới nhất trước, embed đủ tên hiển thị | FR-008, D16 |
| `GET /api/transactions/:id` | — | `{data:Transaction}` (prefill form sửa) | FR-009 |
| `PATCH /api/transactions/:id` | `{amount?, type?, category_id?, account_id?, description?, transaction_date?, expected_updated_at}` — validate như tạo mới | `{data:Transaction}` | FR-009, FR-014, D17/D18 |
| `DELETE /api/transactions/:id` | `{expected_updated_at?}` — UI luôn hiện dialog xác nhận trước khi gọi (SC-005) | `204` | FR-010, FR-014 |

**Mã lỗi nghiệp vụ** (422 trừ khi ghi khác; bổ sung trên bộ lỗi 001):

| Code | Khi nào | Truy vết |
|---|---|---|
| `AMOUNT_INVALID` | amount ≤ 0 / không phải số | FR-001 |
| `CATEGORY_REQUIRED` / `CATEGORY_TYPE_MISMATCH` | thiếu danh mục / khác loại (kể cả khi đổi loại lúc sửa — client coi danh mục cũ là "chưa chọn") | FR-003, FR-009, D18 |
| `ACCOUNT_REQUIRED` / `ACCOUNT_HOUSEHOLD_MISMATCH` | thiếu tài khoản / tài khoản khác hộ | FR-006 |
| `DESCRIPTION_TOO_LONG` | mô tả > 255 ký tự | FR-004 |
| `FUTURE_DATE_NOT_ALLOWED` | `transaction_date` > now + 1 ngày | FR-005, D15 |
| `CONCURRENCY_CONFLICT` (409) | `expected_updated_at` lệch — bản ghi đã bị thành viên khác sửa; response kèm `{data:Transaction}` mới nhất để client hiển thị | FR-014, D17 |
| `RECORD_GONE` (404) | bản ghi đã bị xóa (sửa/xóa đè) → client thông báo + làm tươi sổ | FR-014, D17 |

## WebSocket (mở rộng 001)

| Path | Sự kiện server → client | Ghi chú |
|---|---|---|
| `WS /ws` | `{"type": "transactions_changed"}` · `{"type": "accounts_changed"}` (+ `categories_changed` từ 001) | Hub theo household; mọi mutation transaction phát cả hai (số dư đổi theo); client refetch sổ + balances — mục tiêu hiển thị ≤ 5s (SC-006) |

## Bất biến phía server (biz — trong DB transaction)

1. Mọi validate của POST chạy lại nguyên vẹn cho PATCH (FR-009).
2. `created_by` không bao giờ nhận từ client (server gán từ phiên — FR-013).
3. Conditional write theo `updated_at`; GORM hook làm mới `updated_at` mỗi lần update (D17).
4. Số dư KHÔNG bao giờ ghi tay — chỉ đọc qua view `account_balances` (SC-004, D14).

## History

- 2026-07-10: Tạo mới thay `transaction-repository.md` (re-platform Go + Vue); ánh xạ toàn bộ hợp đồng domain cũ (TransactionRepository/AccountRepository/CurrentUser) sang REST + WS; CurrentUser → `GET /api/me` (001).
