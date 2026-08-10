# Quickstart — Feature 008: Báo Cáo Thu Chi Theo Thành Viên

Kịch bản kiểm chứng end-to-end cho báo cáo theo thành viên. Tham chiếu hợp đồng ở [`contracts/member-report-api.md`](./contracts/member-report-api.md) và DTO ở [`data-model.md`](./data-model.md). Không lặp lại mã triển khai (thuộc `tasks.md`).

## Prerequisites

- Nền tảng 001/002/005 đã chạy; DB Postgres có seed hộ nhiều thành viên với giao dịch của **≥2 thành viên** trong tháng hiện tại.
- Dev workflow (CLAUDE.md): `cd src && make dev && make seed`, rồi `make api` + `make web`.
- Đăng nhập bằng một thành viên của hộ seed.

## Kịch bản

### QS-1 — Báo cáo theo thành viên (US1 · FR-001/002/003)
1. Mở `/reports`, chọn "Tháng này".
2. **Kỳ vọng**: khối "Theo thành viên" liệt kê mỗi thành viên với Thu / Chi / ròng; giao dịch mỗi người tính đúng theo người **đã nhập** (`created_by`).
3. Đổi sang "Tuần này".
4. **Kỳ vọng**: số liệu mỗi thành viên tính lại theo tuần, không rời màn.

Kiểm chứng API: `GET /api/reports/members?from=&to=` trả `members[]` + `totals`.

### QS-2 — Đối soát tổng hộ (US3 · FR-005 · SC-001)
1. Với cùng khoảng, gọi `GET /api/reports/members` và `GET /api/reports/overview`.
2. **Kỳ vọng**: `Σ members[].income == totals.income == overview.income`; tương tự với expense. (Bất biến chính — có test integration.)

### QS-3 — Thành viên không có giao dịch (US3 · FR-004 · SC-003)
1. Bảo đảm hộ có một thành viên chưa nhập giao dịch nào trong khoảng.
2. **Kỳ vọng**: thành viên đó vẫn xuất hiện với Thu=0, Chi=0, ròng=0.

### QS-4 — Drill-in danh sách giao dịch của một thành viên (US2 · FR-006/007)
1. Bấm vào một thành viên trên bảng.
2. **Kỳ vọng**: panel liệt kê giao dịch của người đó trong khoảng — mỗi dòng có loại Thu/Chi, số tiền, danh mục, mô tả, ngày giờ, tài khoản; danh sách **phân trang** khi dài.
3. Bấm quay lại.
4. **Kỳ vọng**: về bảng theo thành viên, **giữ nguyên khoảng** đã chọn.

Kiểm chứng API: `GET /api/reports/member/:id?from=&to=&page=1`.

### QS-5 — "Thành viên cũ / Đã rời hộ" (FR-010 · Option A)
1. (Nếu môi trường mô phỏng được) có giao dịch với `created_by` không còn là thành viên hộ.
2. **Kỳ vọng**: xuất hiện một dòng "Thành viên cũ / Đã rời hộ" gộp các giao dịch đó; đối soát QS-2 **vẫn đúng**. Drill-in dòng này (`:id=former`) liệt kê các giao dịch tương ứng.
3. Nếu không có dữ liệu như vậy: **không** có dòng former (chỉ hiện khi cần).

### QS-6 — Cô lập hộ & xác thực thành viên (FR-008 · SC-006)
1. Đăng nhập bằng thành viên hộ khác (hoặc gọi `GET /api/reports/member/:id` với `:id` là user **không** thuộc hộ hiện tại).
2. **Kỳ vọng**: chỉ thấy thành viên & số liệu của hộ mình; drill-in `:id` ngoài hộ → **404** (không rò rỉ).

### QS-7 — Khoảng tùy chỉnh & trạng thái trống
1. Chọn khoảng tùy chỉnh với `to < from`.
2. **Kỳ vọng**: bị chặn (400 / thông báo rõ ràng).
3. Chọn khoảng không có giao dịch nào.
4. **Kỳ vọng**: mọi thành viên hiện tại hiển thị 0/0/0; drill-in cho trạng thái trống rõ ràng, không lỗi.

## Test tương ứng (định hướng — chi tiết ở tasks.md)
- **Go unit/integration**: tách current↔former, seed thành viên 0 GD, đối soát `Σ = totals = overview`, phân trang drill-in, cô lập hộ, 404 ngoài hộ, sentinel `former`.
- **Vitest**: store members + drill-in, format tên fallback email, giữ khoảng khi quay lại.
- **Playwright**: QS-1…QS-7.
