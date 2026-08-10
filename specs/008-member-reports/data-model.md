# Data Model — Feature 008: Báo Cáo Thu Chi Theo Thành Viên (Phase 1)

**Nguồn**: [research.md](./research.md) · [spec.md](./spec.md) · [entity-model.md](../entities/entity-model.md).

> ⚠️ **Không thực thể/bảng/migration mới.** Tất cả là **DTO suy ra** (giá trị đọc-tính), song song `ReportOverview`/`CategoryReport` của 005. Nguồn dữ liệu: `transactions` (002), `users` + `household_members` (nền tảng) — chỉ ĐỌC.

## Thực thể nguồn (chỉ đọc — đã tồn tại)

| Thực thể | Trường dùng | Vai trò trong feature |
|----------|-------------|------------------------|
| `transactions` | `household_id`, `created_by`, `type` (INCOME/EXPENSE), `amount`, `transaction_date`, (+ trường cho ListItem: category, account, description) | Nguồn tổng hợp; nhóm theo `created_by` |
| `household_members` | `household_id`, `user_id` | Xác định thành viên **hiện tại** của hộ (current ↔ former) |
| `users` | `id`, `display_name`, `email` | Tên hiển thị (fallback email — FR-009) |

Không thêm cột/bảng/view. `created_by` đã là `NOT NULL FK → users(id)` (migration 00004).

## DTO suy ra (mới — `src/api/module/report/model`)

### `MemberRow` — một dòng thành viên trong báo cáo tổng hợp

| Field (JSON) | Kiểu | Ý nghĩa | Ràng buộc / nguồn |
|--------------|------|---------|-------------------|
| `member_id` | string | UUID thành viên; hoặc sentinel `"former"` cho dòng gộp người đã rời | khóa drill-in; **không** hiển thị thô (FR-009/D46) |
| `display_name` | string | Tên hiển thị; `former` → nhãn "Thành viên cũ / Đã rời hộ" | fallback email khi rỗng (D46) |
| `is_former` | bool | true nếu là dòng gộp người đã rời hộ | chỉ một dòng như vậy (D44) |
| `income` | number | Tổng Thu trong khoảng | `SUM(amount) WHERE type=INCOME` (D42); 0 nếu không có |
| `expense` | number | Tổng Chi trong khoảng | `SUM(amount) WHERE type=EXPENSE` (D42); 0 nếu không có |
| `net` | number | Ròng = income − expense | suy ở biz |

### `MembersReport` — response `GET /api/reports/members`

| Field (JSON) | Kiểu | Ý nghĩa |
|--------------|------|---------|
| `from`, `to` | string (YYYY-MM-DD) | khoảng đã giải (preset→from/to hoặc từ query) |
| `members` | `MemberRow[]` | mọi thành viên hiện tại (kể cả 0/0/0) + dòng `former` nếu có; sắp theo `expense` desc (D43/D47) |
| `totals` | `{ income, expense, net }` | tổng toàn hộ trong khoảng — **phải bằng** `Σ members` (đối soát SC-001) |

**Bất biến (kiểm thử)**: `Σ members[].income == totals.income` và `Σ members[].expense == totals.expense`; `totals` cũng bằng income/expense của `GET /api/reports/overview` cùng khoảng.

### `MemberReport` — response `GET /api/reports/member/:id` (drill-in, phân trang)

| Field (JSON) | Kiểu | Ý nghĩa |
|--------------|------|---------|
| `member_id` | string | UUID hoặc `former` |
| `display_name` | string | tên hiển thị (hoặc "Thành viên cũ") |
| `is_former` | bool | dòng gộp người đã rời |
| `from`, `to` | string | khoảng |
| `income`, `expense`, `net` | number | tổng của thành viên trong khoảng (khớp `MemberRow`) |
| `transactions` | `transactionmodel.ListItem[]` | trang giao dịch hiện tại — loại/số tiền/danh mục/mô tả/ngày giờ/tài khoản/người nhập (FR-006) |
| `page`, `page_size`, `total` | number | phân trang (`common.Paging`, FR-007) |

Không có trend cho drill-in thành viên (BR-MBR-006 chỉ yêu cầu danh sách giao dịch).

## Quy tắc dẫn xuất (biz)

- **Nhóm & tách**: quét toàn bộ giao dịch hộ trong `[from, to)`; hàng có `created_by ∈ household_members` → gộp theo `member_id`; còn lại → gộp vào `former` (D42/D44).
- **Đủ thành viên**: seed từ `household_members` để thành viên 0 giao dịch có dòng 0/0/0 (D43/FR-004).
- **Net**: `income − expense`.
- **Sort**: `expense` desc, tie-break `display_name` asc (Assumptions).
- **Đối soát**: `totals` tính độc lập (SUM toàn hộ, không nhóm) rồi assert `== Σ members` ở test.
- **Drill-in `former`**: lọc `created_by NOT IN (SELECT user_id FROM household_members WHERE household_id=?)` (D47).

## Thay đổi module lân cận (đọc, tương thích ngược)

- `transaction/storage.ListFilter` **thêm** `CreatedBy *uuid.UUID` (nil = không lọc). Không đổi hành vi hiện có; phục vụ drill-in một thành viên (D47).

## Không đổi

- `entity-model.md`, migrations, module transaction/category/users (hành vi ghi) — **không đổi** (D49).
