# Contract — Member Report API (feature 008)

**Base**: `/api` · Auth: JWT cookie (như 001–005) · Household scope: suy từ phiên (`middleware.HouseholdID`) — **không** nhận household/user id từ client. Read-only. Nhất quán envelope `common.WriteOK` / `common.WriteError` của 005.

Mở rộng module `report` (song song `GET /api/reports/overview` · `/api/reports/category/:id` của 005).

---

## GET /api/reports/members

Tổng hợp Thu/Chi/ròng theo từng thành viên hiện tại của hộ trong khoảng, kèm dòng gộp "Thành viên cũ" nếu có, và totals đối soát.

**Query**:
| Tham số | Bắt buộc | Định dạng | Ghi chú |
|---------|:--------:|-----------|---------|
| `from` | ✓ | `YYYY-MM-DD` | inclusive (đầu ngày) |
| `to` | ✓ | `YYYY-MM-DD` | inclusive (nội bộ dùng nửa mở `< to+1day`), phải ≥ `from` |

**200 OK**:
```json
{
  "data": {
    "from": "2026-08-01",
    "to": "2026-08-31",
    "members": [
      { "member_id": "6f1c…", "display_name": "Alice", "is_former": false, "income": 12000000, "expense": 4500000, "net": 7500000 },
      { "member_id": "9a2d…", "display_name": "bob@example.com", "is_former": false, "income": 0, "expense": 3200000, "net": -3200000 },
      { "member_id": "c3e0…", "display_name": "Carol", "is_former": false, "income": 0, "expense": 0, "net": 0 },
      { "member_id": "former", "display_name": "Thành viên cũ / Đã rời hộ", "is_former": true, "income": 0, "expense": 800000, "net": -800000 }
    ],
    "totals": { "income": 12000000, "expense": 8500000, "net": 3500000 }
  }
}
```

**Bất biến**: `Σ members[].income == totals.income`, `Σ members[].expense == totals.expense`; `totals` bằng `income`/`expense` của `GET /api/reports/overview` cùng `from/to` (SC-001).

**Quy tắc**:
- `members` gồm **mọi thành viên hiện tại** (kể cả income=expense=0 — FR-004); dòng `member_id:"former"` **chỉ** xuất hiện khi có giao dịch của người không còn là thành viên (FR-010).
- Sắp theo `expense` giảm dần, tie-break `display_name` (Assumptions). Dòng `former` xếp cuối.
- `display_name` fallback email khi trống; không trả UUID nào khác `member_id` (FR-009).

**Lỗi**:
| HTTP | Khi nào |
|------|---------|
| 400 | thiếu/sai `from`/`to`, hoặc `to < from` |
| 401 | chưa đăng nhập |

---

## GET /api/reports/member/:id

Drill-in: danh sách giao dịch (phân trang) cấu thành số liệu của một thành viên (hoặc bucket former) trong khoảng.

**Path**:
| Tham số | Định dạng | Ghi chú |
|---------|-----------|---------|
| `id` | UUID **hoặc** `former` | UUID phải là thành viên **hiện tại** của hộ; `former` = bucket người đã rời |

**Query**: `from`, `to` (như trên) · `page` (mặc định 1) · `page_size` (mặc định theo `common.Paging`, ví dụ 20)

**200 OK**:
```json
{
  "data": {
    "member_id": "6f1c…",
    "display_name": "Alice",
    "is_former": false,
    "from": "2026-08-01",
    "to": "2026-08-31",
    "income": 12000000,
    "expense": 4500000,
    "net": 7500000,
    "transactions": [
      {
        "id": "…", "type": "EXPENSE", "amount": 250000,
        "category_name": "Ăn uống", "account_name": "Tiền mặt",
        "description": "Chợ", "transaction_date": "2026-08-12T09:30:00+07:00",
        "created_by_name": "Alice"
      }
    ],
    "page": 1, "page_size": 20, "total": 37
  }
}
```

**Quy tắc**:
- Giao dịch = các bản ghi của hộ trong `[from,to]` có `created_by = :id` (hoặc `created_by ∉ thành viên hiện tại` khi `:id="former"`), sắp mới nhất trước (như sổ 002).
- Mỗi giao dịch phơi tối thiểu loại/số tiền/danh mục/mô tả/ngày giờ/tài khoản (FR-006), tái dùng `transactionmodel.ListItem`.
- `income/expense/net` là tổng của thành viên trong khoảng (khớp dòng ở `/members`).
- Trạng thái trống: `transactions: []`, `total: 0` (FR-007) — không lỗi.

**Lỗi**:
| HTTP | Khi nào |
|------|---------|
| 400 | thiếu/sai `from`/`to`, `to < from`, hoặc `page`/`page_size` không hợp lệ |
| 401 | chưa đăng nhập |
| 404 | `:id` không phải UUID hợp lệ, **hoặc** UUID không là thành viên hiện tại của hộ (không phân biệt "không tồn tại" vs "khác hộ" — chống dò, FR-008/D48) |

---

## Ghi chú triển khai (không thuộc hợp đồng)

- Household scope luôn từ phiên; endpoint không đọc household/user id thân client ngoài `:id` (và `:id` được kiểm tra thuộc hộ).
- Không cache/side-effect; đọc thuần. Không realtime bắt buộc (tính theo yêu cầu — Assumptions).
