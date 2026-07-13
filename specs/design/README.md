# UI Design — Wireframes (lo-fi)

**File**: [`design.html`](design.html) — canvas wireframe (mở bằng trình duyệt), mobile-first · lo-fi · tiếng Việt · VNĐ. 7 màn hình cốt lõi bám theo các Business Requirement.

> Giữ các liên kết này cập nhật mỗi khi một artifact thay đổi (xem quy tắc lan truyền trong [`CLAUDE.md`](../../CLAUDE.md)).

## Bản đồ màn hình → spec / BR

| # | Màn hình | Feature spec / BR liên quan |
|---|----------|------------------------------|
| 1 | Tổng quan (Dashboard) | Tổng hợp — chưa có spec riêng. Giao dịch gần đây → [`002-transaction-tracking`](../002-transaction-tracking/spec.md) · widget ngân sách → [`003-budgeting`](../003-budgeting/spec.md) · tổng tài sản → [`BR-005`](../business-requirements/BR-005.md) |
| 2 | Nhập giao dịch | [`002-transaction-tracking`](../002-transaction-tracking/spec.md) · trường danh mục & gợi ý → [`001-transaction-categorization`](../001-transaction-categorization/spec.md) |
| 3 | Danh sách giao dịch | [`002-transaction-tracking`](../002-transaction-tracking/spec.md) |
| 4 | Ngân sách | [`003-budgeting`](../003-budgeting/spec.md) — cảnh báo amber ≥80%, đỏ khi vượt mức (BR-BGT-006/007) |
| 5 | Báo cáo & phân tích | [`BR-004`](../business-requirements/BR-004.md) — chưa có feature spec |
| 6 | Tài khoản / Ví | [`002-transaction-tracking`](../002-transaction-tracking/spec.md) (accounts tối thiểu + số dư) · phạm vi đầy đủ → [`BR-005`](../business-requirements/BR-005.md) |
| 7 | Quản lý danh mục | [`001-transaction-categorization`](../001-transaction-categorization/spec.md) |

## Ghi chú thiết kế (post-it trong file)

- Nút **＋** giữa bottom nav mở thẳng màn Nhập giao dịch với bàn phím số — mục tiêu ghi chép < 5 giây (khớp SC-006 của spec 002).
- Ngân sách: cảnh báo amber khi ≥80%, đỏ khi vượt mức (BR-BGT-006/007).
- Chuyển tiền giữa ví KHÔNG tính vào Thu/Chi — chỉ đổi số dư (BR-ACC-007, thuộc BR-005).

## References / Truy vết

- **Nguồn (BR)**: [`BR-001`](../business-requirements/BR-001.md) · [`BR-002`](../business-requirements/BR-002.md) · [`BR-003`](../business-requirements/BR-003.md) · [`BR-004`](../business-requirements/BR-004.md) · [`BR-005`](../business-requirements/BR-005.md).
- **Được tham chiếu bởi**: References/Truy vết của spec [`001`](../001-transaction-categorization/spec.md) · [`002`](../002-transaction-tracking/spec.md) · [`003`](../003-budgeting/spec.md).

## History

- v1 (2026-07-13): tạo chỉ mục cho `design.html`; liên kết wireframes vào khối References của spec 001/002/003.
