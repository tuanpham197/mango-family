# Contract: REST API — Báo Cáo (feature 005)

**Date**: 2026-07-15 · **Plan**: [../plan.md](../plan.md) · **Data model**: [../data-model.md](../data-model.md) · **Nền tảng**: [transaction-api.md (002)](../../002-transaction-tracking/contracts/transaction-api.md) · [overview-api.md (004)](../../004-monthly-income-expense-overview/contracts/overview-api.md)

> Mở rộng hợp đồng 001/002/004: Auth, bọc `app_response`, cookie JWT, phạm vi hộ → 404 ngoài hộ **giữ nguyên**.
> Feature 005 **chỉ đọc** — thêm 2 endpoint tổng hợp; **không** endpoint ghi, **không** DB schema delta/migration.

## Reports (tổng hợp giá trị suy ra theo khoảng)

| Method & Path | Query | Response | Truy vết |
|---|---|---|---|
| `GET /api/reports/overview` | `from`, `to` (YYYY-MM-DD, inclusive) | `200 {data: ReportOverview}` | FR-001…004, FR-007…009 |
| `GET /api/reports/category/:id` | `from`, `to` | `200 {data: CategoryReport}` | FR-005, FR-006 |

**ReportOverview** =
```json
{
  "from": "2026-07-01", "to": "2026-07-31", "group_unit": "day",
  "income": 20000000, "expense": 12000000, "net": 8000000,
  "category_breakdown": [
    { "category_id": "…", "category_name": "Ăn uống", "category_hidden": false, "amount": 3500000, "percent": 29 }
  ],
  "trend": [
    { "bucket": "2026-07-01", "income": 0, "expense": 45000 },
    { "bucket": "2026-07-02", "income": 18000000, "expense": 120000 }
  ]
}
```

**CategoryReport** =
```json
{
  "category_id": "…", "category_name": "Ăn uống",
  "from": "2026-07-01", "to": "2026-07-31", "group_unit": "day",
  "total": 3500000,
  "transactions": [
    { "id": "…", "amount": 45000, "type": "EXPENSE", "category_name": "Ăn uống",
      "account_name": "Tiền mặt", "created_by_name": "Alice", "transaction_date": "…", "description": "Bún bò Huế" }
  ],
  "trend": [ { "bucket": "2026-07-01", "amount": 45000 } ]
}
```

- `group_unit` ∈ {`day`,`week`,`month`} do server chọn theo độ dài [from,to] (D36); `trend` đã **điền đủ mốc** trong khoảng (mốc trống = 0).
- `category_breakdown`: chỉ EXPENSE, gộp **danh mục con một cấp** vào cha (D37); `percent = round(amount/expense×100)`; sắp giảm dần.
- `CategoryReport.total`/`transactions`/`trend`: theo danh mục `:id` **+ danh mục con một cấp** (D38); `transactions` dạng `ListItem` của 002, mới nhất trước.

**Mã lỗi / quy ước**:

| Tình huống | Kết quả | Truy vết |
|---|---|---|
| `to < from` | `400 INVALID_REQUEST` (kèm field) | FR-008 |
| thiếu/không hợp lệ `from`/`to` | `400 INVALID_REQUEST` | FR-001 |
| `:id` danh mục ngoài hộ / không tồn tại | `404 RECORD_GONE` | FR-007, D5/001 |
| khoảng trống (không giao dịch) | `200` với `income/expense/net = 0`, `category_breakdown: []`, `trend` các mốc = 0 | FR-009, SC-004 |

## Bất biến phía server (biz)

1. Mọi giá trị **suy ra khi đọc** — không lưu/cache; khớp 100% giao dịch theo [from,to] (SC-001).
2. Mọi truy vấn **filter `household_id`** của phiên (phạm vi hộ — D5/001); không lộ dữ liệu hộ khác.
3. Tổng hợp bằng **SQL** (`SUM`/`GROUP BY date_trunc`), gộp danh mục con vào cha; chỉ EXPENSE cho phân bổ/chi tiết danh mục (D37/D38/D41).
4. `to ≥ from`; `group_unit` chọn theo độ dài khoảng (D36); trend điền mốc trống = 0.
5. Server **chỉ SELECT/SUM** — không mutate.

## WebSocket (dùng lại 001/002/004)

Không thêm topic mới. Màn Báo cáo có thể nghe `transactions_changed`/`categories_changed` (đã có) để refetch khi đang mở (D40); báo cáo chủ yếu tính **theo yêu cầu** (đổi khoảng / mở màn).

## History

- v1 (2026-07-15): tạo mới — `GET /api/reports/overview` + `GET /api/reports/category/:id` (giá trị suy ra theo [from,to]; group_unit tự chọn; phân bổ danh mục gộp con; trend điền mốc; chi tiết danh mục + subtree). Chỉ đọc, không schema delta.
