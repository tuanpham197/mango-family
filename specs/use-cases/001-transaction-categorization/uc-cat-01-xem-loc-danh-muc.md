# UC-CAT-01: Xem & lọc danh mục theo loại Thu/Chi

| Thuộc tính | Giá trị |
|------------|---------|
| **Mã** | UC-CAT-01 |
| **Tính năng** | Phân loại giao dịch ([spec.md](../../001-transaction-categorization/spec.md)) |
| **User Story** | US1 (P1) |
| **Tác nhân chính** | Người dùng |
| **Mức độ ưu tiên** | Cao |

## Mục tiêu

Người dùng xem được toàn bộ danh mục của mình, được nhóm và lọc theo loại Thu/Chi,
để nắm hệ thống phân loại hiện có và chọn đúng danh mục khi cần.

## Tiền điều kiện

- Người dùng đã đăng nhập.
- Hệ thống đã có sẵn bộ danh mục mặc định đã phân loại theo Thu/Chi (FR-001).

## Kích hoạt (Trigger)

Người dùng mở màn hình danh sách / quản lý danh mục.

## Luồng sự kiện chính

1. Người dùng mở danh sách danh mục.
2. Hệ thống hiển thị danh mục, **nhóm theo loại Thu và loại Chi**.
3. Với mỗi danh mục cha có danh mục con, hệ thống hiển thị danh mục con lồng dưới cha (một cấp).
4. Người dùng chọn bộ lọc theo loại (Thu hoặc Chi).
5. Hệ thống chỉ hiển thị các danh mục thuộc loại đã chọn.

## Luồng thay thế / ngoại lệ

- **2a. Danh mục đã ẩn**: Danh mục bị ẩn được đánh dấu rõ "đã ẩn" trong màn hình
  quản lý nhưng KHÔNG xuất hiện trong danh sách chọn khi nhập giao dịch (xem UC-CAT-06).
- **4a. Loại chưa có danh mục nào**: Nếu loại được lọc chưa có danh mục, hệ thống
  hiển thị trạng thái rỗng và gợi ý người dùng tạo danh mục mới (xem UC-CAT-02).

## Hậu điều kiện

- Người dùng thấy đúng tập danh mục theo bộ lọc; không có danh mục khác loại lọt vào.

## Quy tắc nghiệp vụ

- Mỗi danh mục mang đúng một loại cố định Thu hoặc Chi (FR-004).
- Khi dùng trong luồng nhập giao dịch, chỉ danh mục cùng loại với giao dịch được hiển thị (FR-014 — xem UC-CAT-07).

## Truy vết

- **FR**: FR-002, FR-004, FR-014
- **Acceptance**: US1 #1, #3, #4
- **Success Criteria**: SC-003
- **BR**: BR-CAT-001, BR-CAT-008
