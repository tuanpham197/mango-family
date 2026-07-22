# Contract: REST API — Màn Tổng Quan (feature 004)

**Date**: 2026-07-14 · **Plan**: [../plan.md](../plan.md) · **Data model**: [../data-model.md](../data-model.md) · **Nền tảng**: [transaction-api.md (002)](../../002-transaction-tracking/contracts/transaction-api.md) · [budget-api.md (003)](../../003-budgeting/contracts/budget-api.md)

> Mở rộng hợp đồng 001/002/003: Auth (`/api/auth/*`, `/api/me`), quy ước chung (bọc `app_response`, cookie JWT,
> phạm vi hộ → 404 ngoài hộ, mã lỗi chung) **giữ nguyên**. Feature 004 **chỉ đọc** — thêm đúng một endpoint
> tổng hợp; **không** endpoint sửa dữ liệu, **không** DB schema delta, **không** migration.

## Overview (tổng hợp giá trị suy ra)

| Method & Path | Query | Response | Truy vết |
|---|---|---|---|
| `GET /api/overview` | — | `200 {data: OverviewSummary}` | FR-002…004, FR-006, FR-010 |

**OverviewSummary** =
```json
{
  "net_worth": 24560000,
  "net_worth_change_percent": 5.2,        // null nếu không đủ dữ liệu tháng trước / mẫu số 0
  "month": { "income": 18200000, "expense": 9400000, "net": 8800000 },
  "category_spending": [
    { "category_id": "…", "category_name": "Ăn uống", "category_hidden": false, "amount": 3500000, "percent": 37 },
    { "category_id": "…", "category_name": "Mua sắm", "category_hidden": false, "amount": 1200000, "percent": 13 }
  ],
  "recent_transactions": [
    { "id": "…", "amount": 45000, "type": "EXPENSE", "category_name": "Ăn uống",
      "account_name": "Tiền mặt", "created_by_name": "Minh Anh", "transaction_date": "…", "description": "Bún bò Huế" }
  ]
}
```

- `net_worth` = tổng số dư mọi tài khoản của hộ (view `account_balances` — 002/D14).
- `net_worth_change_percent` = % thay đổi so tài sản ròng cuối tháng trước (D28); `null` khi không tính được.
- `month.{income,expense,net}` = tổng Thu / Chi / chênh lệch của hộ trong **tháng dương lịch hiện tại**.
- `category_spending[]` = chi theo **danh mục cha** (gộp con một cấp), tháng hiện tại, `amount>0`, **sắp giảm dần**; `percent = round(amount/month.expense×100)` (D29).
- `recent_transactions[]` = 5 giao dịch mới nhất của hộ, dạng `ListItem` của 002 (D30).

**Tóm tắt ngân sách trên Tổng quan**: KHÔNG nằm trong endpoint này — client dùng lại `GET /api/budgets` (contract 003) cho widget ngân sách.

## Bất biến phía server (biz)

1. Mọi giá trị **suy ra khi đọc** — không lưu, không cache; khớp 100% dữ liệu nguồn tại thời điểm gọi (SC-001/002/003, D21).
2. Mọi truy vấn **filter theo `household_id`** của phiên (phạm vi hộ — D5/001); không lộ dữ liệu hộ khác.
3. Cửa sổ tháng suy từ `now()` (dùng lại `model.ResolvePeriod(MONTHLY)` của module budget — D20/003).
4. Chi theo danh mục gộp **danh mục con một cấp** vào danh mục cha (D21/003, D29); chỉ loại `EXPENSE`.
5. Server **chỉ SELECT/SUM** trên transactions/accounts/categories — không mutate (bất biến "overview chỉ đọc").

## WebSocket (dùng lại 001/002/003)

Không thêm topic mới. Client màn Tổng quan nghe các topic sẵn có `transactions_changed`, `accounts_changed`,
`budgets_changed`, `categories_changed` → refetch `GET /api/overview` (và `GET /api/budgets` cho widget) → cập
nhật realtime ≤ 5s (SC-006, D33).

## History

- v1 (2026-07-14): tạo mới — `GET /api/overview` trả về OverviewSummary (net worth + % tháng trước, Thu/Chi
  tháng, chi theo danh mục + %, giao dịch gần đây); chỉ đọc, không schema delta; realtime dùng lại topic sẵn có.
