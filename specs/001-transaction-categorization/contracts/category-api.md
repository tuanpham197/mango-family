# Contract: REST API + WebSocket — Phân Loại Giao Dịch (feature 001)

**Date**: 2026-07-10 · **Plan**: [plan.md](../plan.md) · **Schema**: [db-schema.sql](./db-schema.sql)

> Thay thế `category-repository.md` (hợp đồng domain Flutter cũ — đã xóa). Đây là hợp đồng
> giữa `src/web` (Vue) và `src/api` (Go/Gin). Format lỗi/response theo `common/app_response`
> (mẫu learn_go): thành công `{ "data": ... }`; lỗi `{ "error": { "code", "message", "field?" } }`.
> Mọi endpoint (trừ login) yêu cầu cookie JWT (middleware `authenticate`) và chạy trong
> phạm vi hộ của người dùng (middleware household scope — research D5); bản ghi ngoài hộ → **404**.

## Auth (nền tảng — D4)

| Method & Path | Body | Response | Ghi chú |
|---|---|---|---|
| `POST /api/auth/login` | `{email, password}` | `200 {data:{user}}` + Set-Cookie JWT (HttpOnly, SameSite=Lax, 7d) | Sai thông tin → `401 INVALID_CREDENTIALS` (không nói rõ trường nào — UC-TRK-01 E1) |
| `POST /api/auth/logout` | — | `204` + xóa cookie | |
| `GET /api/me` | — | `{data:{user:{id,email,display_name}, household:{id,name}}}` | Chưa thuộc hộ nào → `409 NO_HOUSEHOLD` (UC-TRK-01 E2) |

## Categories

| Method & Path | Body / Query | Response | Truy vết |
|---|---|---|---|
| `GET /api/categories` | `?type=INCOME\|EXPENSE&include_hidden=true\|false` (mặc định false) | `{data:[CategoryTree]}` — mảng danh mục gốc, mỗi mục kèm `children[]` | FR-002, FR-014, FR-020 |
| `POST /api/categories` | `{name, type, icon?, parent_id?}` | `201 {data:Category}` | FR-003, FR-010/011 |
| `PATCH /api/categories/:id` | `{name?, icon?, is_hidden?, expected_updated_at}` | `{data:Category}` | FR-006, FR-020, D6 |
| `DELETE /api/categories/:id` | `{mode?: "reassign"\|"delete_transactions", target_category_id?, expected_updated_at}` | `204` | FR-007/008/009/012, D12 |
| `GET /api/categories/suggest` | `?description=...&type=EXPENSE` | `{data:{category_id, name} \| null}` | FR-015, D7 |

**Category** = `{id, household_id, name, type, icon, parent_id, is_default, is_hidden, created_by, updated_at}`.

**Quy tắc lỗi nghiệp vụ** (HTTP 422 trừ khi ghi khác):

| Code | Khi nào | Truy vết |
|---|---|---|
| `TYPE_IMMUTABLE` | PATCH gửi `type` | FR-005 |
| `NESTING_TOO_DEEP` | `parent_id` trỏ tới một danh mục con | FR-010 |
| `PARENT_TYPE_MISMATCH` | tạo con với type ≠ type cha (hoặc gửi type khác cha) | FR-011 |
| `NAME_DUPLICATE_WARNING` | trùng tên trong (hộ, loại, cha) — **cảnh báo**: client hiện confirm, gửi lại với `confirm_duplicate: true` để vẫn tạo | FR-017 |
| `CATEGORY_HAS_TRANSACTIONS` | DELETE không kèm `mode` khi danh mục (hoặc con của nó) còn giao dịch — response kèm `{transaction_count, children_count}` để client hiện dialog | FR-008 |
| `REASSIGN_TYPE_MISMATCH` | `target_category_id` khác loại hoặc khác hộ hoặc chính nó/con của nó | FR-009 |
| `CONCURRENCY_CONFLICT` (409) | `expected_updated_at` lệch — bản ghi đã bị thành viên khác sửa | D6 |
| `RECORD_GONE` (404) | bản ghi đã bị xóa trước đó | D6 |

## Transactions (001 — luồng tối thiểu kiểm chứng phân loại)

| Method & Path | Body / Query | Response | Truy vết |
|---|---|---|---|
| `POST /api/transactions` | `{amount, type, category_id, description?, transaction_date?}` | `201 {data:Transaction}` — `created_by` do server gán từ phiên | FR-013/014/019/022 |
| `GET /api/transactions` | `?page=&page_size=` | `{data:[Transaction+{category_name, created_by_name}], paging}` | FR-022 (sổ chung + authorship) |

Lỗi: `CATEGORY_REQUIRED` (thiếu category_id — FR-013), `CATEGORY_TYPE_MISMATCH` (type giao dịch ≠ type danh mục — FR-014), `AMOUNT_INVALID` (≤ 0), `DESCRIPTION_TOO_LONG` (> 255).

> Khi lưu giao dịch có `description`, biz upsert `categorization_rules` (tăng `match_count`) — cơ chế "học" của gợi ý (D7). Feature 002 mở rộng endpoint này (account, sửa/xóa, chặn ngày tương lai).

## WebSocket

| Path | Sự kiện server → client | Ghi chú |
|---|---|---|
| `WS /ws` | `{"type": "categories_changed"}` · `{"type": "transactions_changed"}` | Auth qua cookie khi handshake; hub theo household (D8); client refetch store tương ứng; fallback refetch khi tab focus |

## History

- 2026-07-10: Tạo mới thay `category-repository.md` (re-platform Go + Vue); ánh xạ toàn bộ hành vi domain cũ sang REST + WS.
