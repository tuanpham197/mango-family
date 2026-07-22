# Phase 1 — Data Model: Báo Cáo và Phân Tích Trực Quan (Go + Vue)

**Date**: 2026-07-15 · **Feature**: 005-reports · **Plan**: [plan.md](./plan.md) ·
**Nguồn**: [spec.md](./spec.md) · [research.md](./research.md) · [entity-model.md](../entities/entity-model.md)

> **KHÔNG có bảng/migration mới.** Feature 005 chỉ **đọc** `transactions` (002) + `categories` (001) và trả
> về **giá trị suy ra** (tổng hợp theo khoảng thời gian). Không RLS/trigger; phạm vi hộ ở API middleware (D5/001).

## Thực thể đọc (không thay đổi)

| Thực thể | Dùng cho | Cách đọc |
|----------|----------|----------|
| `transactions` (002) | tổng thu/chi/ròng, phân bổ danh mục, trend, danh sách chi tiết | `SUM`/`SELECT` theo hộ + [from,to] + loại/danh mục |
| `categories` (001) | tên/ẩn danh mục, cây cha–con (gộp con vào cha) | `SELECT` + JOIN quy về `COALESCE(parent.id, cat.id)` |

## DTO suy ra (không map bảng — tính ở biz khi đọc)

### ReportOverview — response `GET /api/reports/overview`

| Field | Kiểu | Định nghĩa |
|-------|------|-----------|
| from / to | date | Khoảng đã áp dụng (inclusive) |
| group_unit | string | Đơn vị gom nhóm trend đã chọn: `day` \| `week` \| `month` (D36) |
| income | number | `Σ amount` INCOME trong [from,to], cùng hộ |
| expense | number | `Σ amount` EXPENSE trong [from,to], cùng hộ |
| net | number | `income − expense` |
| category_breakdown | CategoryBreakdown[] | Phân bổ chi theo danh mục cha (gộp con), sắp giảm dần (D37) |
| trend | TrendPoint[] | Chuỗi thu/chi theo mốc thời gian, đã điền mốc trống = 0 (D36) |

### CategoryBreakdown

| Field | Kiểu | Định nghĩa |
|-------|------|-----------|
| category_id | uuid | Danh mục cha (Chi) |
| category_name | string | Tên danh mục |
| category_hidden | bool | Đang ẩn (feature 001) — vẫn tính, hiển thị trạng thái |
| amount | number | `Σ chi (cha + con một cấp)` trong khoảng |
| percent | int | `round(amount / expense × 100)` (0 nếu expense = 0) |

### TrendPoint (overview)

| Field | Kiểu | Định nghĩa |
|-------|------|-----------|
| bucket | string | Nhãn mốc thời gian (vd `2026-07-15` / `2026-W29` / `2026-07`) theo `group_unit` |
| income | number | `Σ` INCOME trong mốc |
| expense | number | `Σ` EXPENSE trong mốc |

### CategoryReport — response `GET /api/reports/category/:id`

| Field | Kiểu | Định nghĩa |
|-------|------|-----------|
| category_id | uuid | Danh mục được xem (thuộc hộ; ngoài hộ → 404) |
| category_name | string | Tên danh mục |
| from / to / group_unit | — | như overview |
| total | number | `Σ chi (danh mục + con một cấp)` trong khoảng |
| transactions | TransactionListItem[] | Mọi giao dịch của danh mục (+con) trong khoảng, mới nhất trước (dùng lại `ListItem` 002) |
| trend | CategoryTrendPoint[] | `{bucket, amount}` — chi của danh mục theo mốc thời gian (điền mốc 0) |

### TransactionListItem *(dùng lại 002)*

Như `module/transaction/model.ListItem`: giao dịch + `category_name` + `account_name` + `created_by_name`. Report chỉ **đọc**.

## Quy tắc toàn vẹn / bất biến (tóm tắt → truy vết)

| Bất biến | Thực thi | Truy vết |
|----------|----------|----------|
| Số liệu suy ra khớp 100% giao dịch theo khoảng | biz đọc `SUM`/`GROUP BY` khi gọi (không lưu, không cache) | SC-001 |
| Chỉ tính giao dịch của hộ đang đăng nhập | middleware phạm vi hộ (D5/001) + filter `household_id` | FR-007 |
| Khoảng hợp lệ: `to ≥ from` | biz validate (BE) + FE chặn ngày | FR-008 |
| Phân bổ/chi tiết gộp danh mục con vào cha; chỉ EXPENSE | `COALESCE(parent.id,cat.id)` + `type='EXPENSE'` (D37/D38) | FR-003/005, SC-001 |
| Trend gom nhóm theo đơn vị hợp lý + điền mốc trống | `date_trunc(unit,...)` + fill 0 ở biz (D36) | FR-004/006 |
| Gom nhóm theo lịch nhất quán toàn hệ thống | quy ước ranh giới như 003/004 (không theo tz thiết bị) | Assumptions spec |
| Khoảng trống → 0/empty, không lỗi | biz trả DTO rỗng-an toàn; FE empty state | FR-009, SC-004 |
| Chỉ đọc — không mutate | storage report chỉ `SELECT`/`SUM` | Assumptions spec |
| Báo cáo 1 tháng ≤ 2s | tổng hợp bằng SQL + index sẵn có (D41) | SC-002 |

## History

- v1 (2026-07-15): tạo mới — DTO suy ra cho báo cáo (ReportOverview + CategoryBreakdown + TrendPoint + CategoryReport); không bảng/migration; chỉ đọc 001/002. Kế thừa entity-model hiện có (không đổi).
