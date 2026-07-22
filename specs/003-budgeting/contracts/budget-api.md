# Contract: REST API + WebSocket — Thiết Lập Ngân Sách (feature 003)

**Date**: 2026-07-13 · **Plan**: [../plan.md](../plan.md) · **Schema**: [db-schema.sql](./db-schema.sql) · **Nền tảng**: [category-api.md (001)](../../001-transaction-categorization/contracts/category-api.md) · [transaction-api.md (002)](../../002-transaction-tracking/contracts/transaction-api.md)

> Mở rộng hợp đồng 001/002: Auth (`/api/auth/*`, `/api/me`), quy ước chung (bọc `app_response`, cookie
> JWT, phạm vi hộ → 404 ngoài hộ, mã lỗi chung `CONCURRENCY_CONFLICT`/`RECORD_GONE`) **giữ nguyên** — không
> lặp lại ở đây. Ngân sách chỉ đọc transactions/categories; không có endpoint sửa dữ liệu nguồn.

## Budgets (vòng đời đầy đủ)

**Budget** = `{id, type, category_id, category_name, category_hidden, limit_amount, period_type,
start_date, end_date, status, period_key, spent, percent, alerts, created_by, created_by_name,
updated_at}` — trong đó `spent`/`percent`/`period_key`/`alerts` là **giá trị suy ra kỳ hiện tại** (D21/D23),
`alerts = [{level, over_amount, fired_at}]`. `category_hidden` cho biết danh mục gắn ngân sách đang bị ẩn
(edge case — hiển thị trạng thái để thành viên biết).

| Method & Path | Body / Query | Response | Truy vết |
|---|---|---|---|
| `POST /api/budgets` | `{type, category_id?, limit_amount, period_type, start_date?, end_date?}` — `created_by` server gán từ phiên | `201 {data:Budget}` | FR-001…004, FR-010, FR-011 |
| `GET /api/budgets` | `?status=` (mặc định ACTIVE; `all` để xem cả ENDED) | `{data:[Budget]}` — mỗi ngân sách kèm spent/percent/alerts kỳ hiện tại | FR-005, FR-006 |
| `GET /api/budgets/:id` | — | `{data:Budget}` (prefill form sửa) | FR-012 |
| `PATCH /api/budgets/:id` | `{limit_amount?, period_type?, category_id?, start_date?, end_date?, expected_updated_at}` — validate như tạo | `{data:Budget}` (tiến độ & cảnh báo tính lại theo giá trị mới) | FR-012, FR-013 |
| `DELETE /api/budgets/:id` | `{expected_updated_at?}` — UI luôn hiện dialog xác nhận trước khi gọi | `204` (cảnh báo liên quan xóa cascade) | FR-012, FR-013 |

**Mã lỗi nghiệp vụ** (422 trừ khi ghi khác; bổ sung trên bộ lỗi 001/002):

| Code | Khi nào | Truy vết |
|---|---|---|
| `LIMIT_INVALID` | `limit_amount` ≤ 0 / không phải số | FR-003 |
| `CATEGORY_REQUIRED` | type=CATEGORY nhưng thiếu `category_id` | FR-001, UC-BGT-01 E2 |
| `CATEGORY_TYPE_MISMATCH` | danh mục không phải loại EXPENSE | FR-001 |
| `CATEGORY_HOUSEHOLD_MISMATCH` | danh mục khác hộ | FR-001, FR-011 |
| `CATEGORY_NOT_ALLOWED_FOR_TOTAL` | type=TOTAL nhưng gửi kèm `category_id` | FR-002 |
| `PERIOD_INVALID` | ONE_TIME với `end_date < start_date` / thiếu ngày | FR-004, UC-BGT-01 E4 |
| `BUDGET_DUPLICATE` | đã có ngân sách ACTIVE cùng (danh mục, kỳ) hoặc (tổng, kỳ); response kèm `{existing_budget_id}` để UI chỉ tới | FR-010, UC-BGT-01 E3 / UC-BGT-02 E2 |
| `CONCURRENCY_CONFLICT` (409) | `expected_updated_at` lệch — đã bị thành viên khác sửa; kèm `{data:Budget}` mới nhất | FR-013, D25, UC-BGT-05 E3 |
| `RECORD_GONE` (404) | ngân sách đã bị xóa (sửa/xóa đè) → client thông báo + làm tươi danh sách | FR-013, D25, UC-BGT-05 E4 / UC-BGT-06 E1 |

## WebSocket (mở rộng 001/002)

| Path | Sự kiện server → client | Ghi chú |
|---|---|---|
| `WS /ws` | `{"type": "budgets_changed"}` (+ `transactions_changed`, `accounts_changed`, `categories_changed` từ 001/002) | Hub theo household. Phát khi: (a) CRUD ngân sách; (b) evaluator cảnh báo đổi trạng thái do giao dịch thay đổi (D24). Client nghe `budgets_changed` **và** `transactions_changed` → refetch `GET /api/budgets` → cập nhật tiến độ + cảnh báo ≤ 5s (SC-006, D26) |

## Bất biến phía server (biz — trong DB transaction)

1. Mọi validate của POST chạy lại nguyên vẹn cho PATCH (FR-012).
2. `created_by` không bao giờ nhận từ client (server gán từ phiên — FR-011).
3. `spent`/`percent` KHÔNG bao giờ ghi tay — luôn tính từ giao dịch khi đọc (SC-003, D21); chỉ loại EXPENSE; danh mục con một cấp gộp vào ngân sách cha.
4. Đánh giá cảnh báo (insert/delete `budget_alerts`) chạy trong **evaluator** kích hoạt bởi `transactions_changed` + CRUD ngân sách; recompute idempotent (D24); `over_amount = spent − limit` cho OVER_100 (SC-005).
5. Conditional write theo `updated_at`; GORM hook làm mới `updated_at` mỗi lần update (D25).
6. Ngân sách ONE_TIME hết kỳ → `status=ENDED`, không đánh giá cảnh báo nữa (edge case).

## History

- v1 (2026-07-13, claude): tạo mới (re-platform Go+Vue) — REST `/api/budgets/*` + WS `budgets_changed`;
  spent/percent/alerts suy ra embed trong Budget; bộ mã lỗi nghiệp vụ ngân sách (LIMIT_INVALID,
  BUDGET_DUPLICATE, PERIOD_INVALID, …) trên nền lỗi chung 001/002.
