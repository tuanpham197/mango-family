# Phase 1 — Data Model: Màn Tổng Quan Đầy Đủ (Go + Vue)

**Date**: 2026-07-14 · **Feature**: 004-monthly-income-expense-overview · **Plan**: [plan.md](./plan.md) ·
**Nguồn**: [spec.md](./spec.md) · [research.md](./research.md) · [entity-model.md](../entities/entity-model.md)

> **KHÔNG có bảng/migration mới.** Feature 004 chỉ **đọc** dữ liệu 001/002/003 và trả về **giá trị suy ra**.
> Bảng/thực thể liên quan (chỉ đọc): `transactions`, `accounts` + view `account_balances` (002), `categories`
> (001), `budgets` (003 — qua endpoint riêng). Không RLS/trigger; phạm vi hộ ở API middleware (D5/001).

## Thực thể đọc (không thay đổi)

| Thực thể | Dùng cho | Cách đọc |
|----------|----------|----------|
| `transactions` (002) | Thu/Chi tháng, chi theo danh mục, giao dịch gần đây | `SUM`/`SELECT` theo hộ + cửa sổ tháng + loại |
| `account_balances` (view, 002) | Tổng tài sản ròng hiện tại | `Σ balance` theo hộ |
| `accounts.initial_balance` + `transactions` (002) | Tài sản ròng cuối tháng trước | `Σ initial_balance + Σ bút toán có dấu (date < đầu tháng)` |
| `categories` (001) | Tên/ẩn danh mục, cây cha–con (gộp con vào cha) | `SELECT` theo hộ |
| `budgets` (003) | Tóm tắt ngân sách trên Tổng quan | **không đọc ở đây** — dùng `GET /api/budgets` |

## DTO suy ra (không map bảng — tính ở biz khi đọc)

### OverviewSummary — response của `GET /api/overview`

| Field | Kiểu | Định nghĩa |
|-------|------|-----------|
| net_worth | number | `Σ account_balances.balance` của hộ (D28) |
| net_worth_change_percent | number \| null | `(net_worth − net_worth_prev_end)/|net_worth_prev_end|×100`; `null` khi mẫu số 0 / thiếu dữ liệu (D28) |
| month.income | number | `Σ amount` giao dịch `type=INCOME`, cùng hộ, `transaction_date ∈ tháng hiện tại` |
| month.expense | number | `Σ amount` giao dịch `type=EXPENSE`, cùng hộ, tháng hiện tại |
| month.net | number | `month.income − month.expense` |
| category_spending | CategorySpending[] | Chi theo danh mục cha (gộp con), tháng hiện tại, sắp giảm dần, chỉ `amount>0` (D29) |
| recent_transactions | TransactionListItem[] | 5 giao dịch mới nhất của hộ (dạng `ListItem` 002: embed tên danh mục · tài khoản · người nhập) (D30) |

### CategorySpending

| Field | Kiểu | Định nghĩa |
|-------|------|-----------|
| category_id | uuid | Danh mục cha (Chi) |
| category_name | string | Tên danh mục |
| category_hidden | bool | Danh mục đang ẩn (feature 001) — vẫn tính, hiển thị trạng thái |
| amount | number | `Σ chi của danh mục cha + danh mục con một cấp` trong tháng (D21/003, D29) |
| percent | int | `round(amount / month.expense × 100)` (0 nếu tổng chi tháng = 0) |

### TransactionListItem *(dùng lại 002)*

Cấu trúc như `module/transaction/model.ListItem`: giao dịch + `category_name` + `account_name` + `created_by_name`. Overview chỉ **đọc**, không sửa.

## Quy tắc toàn vẹn / bất biến (tóm tắt → truy vết)

| Bất biến | Thực thi | Truy vết |
|----------|----------|----------|
| Mọi số liệu là giá trị suy ra, khớp 100% dữ liệu nguồn tại mọi thời điểm | biz đọc `SUM`/`SELECT` khi gọi (không lưu, không cache) | SC-001/002/003 |
| Chỉ tính giao dịch của hộ đang đăng nhập | middleware phạm vi hộ (D5/001) + mọi truy vấn filter `household_id` | FR-010 |
| Thu/Chi & chi theo danh mục theo **tháng dương lịch hiện tại** | cửa sổ tháng suy từ `now()` (dùng lại `model.ResolvePeriod(MONTHLY)` của budget) | FR-003/004 |
| Chi theo danh mục gộp danh mục con vào cha | `category_id IN (cha + con một cấp)` (D21/003, D29) | FR-004, SC-002 |
| Tài sản ròng là luỹ kế (không reset theo tháng) | `Σ balance` view; % so tháng trước suy từ mốc đầu tháng (D28) | FR-002, SC-003 |
| Chỉ đọc — không mutate dữ liệu nguồn | storage overview chỉ `SELECT`/`SUM` | Assumptions spec |
| Thay đổi hiển thị cho thành viên khác ≤ 5s | store nghe `transactions/accounts/budgets/categories_changed` → refetch (D33) | SC-006 |

## History

- v1 (2026-07-14): tạo mới — DTO suy ra cho màn Tổng quan (OverviewSummary + CategorySpending), không bảng/migration mới; chỉ đọc dữ liệu 001/002/003. Kế thừa entity-model hiện có (không đổi).
