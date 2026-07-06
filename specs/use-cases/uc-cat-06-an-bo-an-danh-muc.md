# UC-CAT-06: Ẩn / bỏ ẩn danh mục

| Thuộc tính | Giá trị |
|------------|---------|
| **Mã** | UC-CAT-06 |
| **Tính năng** | Phân loại giao dịch ([spec.md](../001-transaction-categorization/spec.md)) |
| **User Story** | — (Edge case, chốt tại Clarifications 2026-06-24) |
| **Tác nhân chính** | Người dùng |
| **Mức độ ưu tiên** | Thấp |

## Mục tiêu

Người dùng tạm gỡ một danh mục khỏi danh sách chọn khi nhập giao dịch mới mà **không
xóa**, giữ nguyên dữ liệu lịch sử; sau này có thể bỏ ẩn để dùng lại.

## Tiền điều kiện

- Người dùng đã đăng nhập.
- Danh mục cần ẩn/bỏ ẩn đang tồn tại.

## Kích hoạt (Trigger)

Người dùng chọn "Ẩn" hoặc "Bỏ ẩn" trên một danh mục.

## Luồng sự kiện chính — Ẩn danh mục

1. Người dùng chọn ẩn một danh mục (mặc định hoặc tự tạo).
2. Hệ thống đánh dấu danh mục là "đã ẩn".
3. Từ đó, danh mục **không xuất hiện** trong danh sách chọn khi nhập giao dịch mới.
4. Giao dịch lịch sử và báo cáo **vẫn giữ nguyên** danh mục đã ẩn.

## Luồng sự kiện phụ — Bỏ ẩn danh mục

1. Người dùng chọn bỏ ẩn một danh mục đang ẩn.
2. Hệ thống bỏ đánh dấu ẩn; danh mục lại xuất hiện trong danh sách chọn khi nhập giao dịch.

## Luồng thay thế / ngoại lệ

- **Ẩn danh mục cha có danh mục con**: Hành vi áp dụng nhất quán cho danh mục cha; danh
  mục con tiếp tục tuân theo trạng thái ẩn/hiện của chính nó (không bị xóa).
- Ẩn KHÔNG yêu cầu gán lại giao dịch (khác với Xóa — UC-CAT-05).

## Hậu điều kiện

- Danh mục đã ẩn không còn được chọn cho giao dịch mới nhưng vẫn còn nguyên cho lịch
  sử và báo cáo; trạng thái ẩn có thể được đảo ngược bất cứ lúc nào.

## Quy tắc nghiệp vụ

- Áp dụng cho cả danh mục mặc định lẫn tự tạo (FR-020).
- Ẩn là thao tác **không phá hủy** dữ liệu, có thể đảo ngược (FR-020).

## Truy vết

- **FR**: FR-020
- **Edge cases**: "Ẩn danh mục đang dùng" (spec §Edge Cases)
- **Assumptions**: "Danh mục mặc định ứng xử như danh mục tự tạo"
