# UC-CAT-04: Chỉnh sửa tên & biểu tượng danh mục

| Thuộc tính | Giá trị |
|------------|---------|
| **Mã** | UC-CAT-04 |
| **Tính năng** | Phân loại giao dịch ([spec.md](../001-transaction-categorization/spec.md)) |
| **User Story** | US2 (P2) |
| **Tác nhân chính** | Người dùng |
| **Mức độ ưu tiên** | Trung bình |

## Mục tiêu

Người dùng đổi tên và/hoặc biểu tượng của bất kỳ danh mục nào (mặc định hoặc tự tạo)
để phản ánh đúng cách họ muốn gọi danh mục đó.

## Tiền điều kiện

- Người dùng đã đăng nhập.
- Danh mục cần sửa đang tồn tại.

## Kích hoạt (Trigger)

Người dùng chọn "Chỉnh sửa" trên một danh mục.

## Luồng sự kiện chính

1. Người dùng chọn một danh mục (mặc định hoặc tự tạo) để chỉnh sửa.
2. Người dùng đổi **tên** và/hoặc **biểu tượng**.
3. Người dùng xác nhận lưu.
4. Hệ thống cập nhật tên/biểu tượng và **phản ánh thay đổi ở mọi nơi danh mục được
   tham chiếu**, bao gồm giao dịch lịch sử và báo cáo.

## Luồng thay thế / ngoại lệ

- **2a. Tên trùng**: Nếu tên mới trùng danh mục khác trong cùng loại và cùng cấp cha,
  hệ thống cảnh báo (FR-017).
- **Không cho đổi loại**: Màn hình chỉnh sửa KHÔNG cho phép đổi loại Thu/Chi (loại cố
  định sau khi tạo — FR-005).

## Hậu điều kiện

- Tên/biểu tượng mới được áp dụng nhất quán cho danh mục, kể cả giao dịch lịch sử và báo cáo.
- Loại Thu/Chi và các tham chiếu giao dịch không thay đổi.

## Quy tắc nghiệp vụ

- Người dùng được sửa cả danh mục mặc định lẫn tự tạo (FR-006).
- Giao dịch **tham chiếu** danh mục theo định danh (không sao chép tên), nên đổi tên
  tự động cập nhật hiển thị xuyên suốt mà không cần gán lại (FR-016).
- Cảnh báo trùng tên trong cùng loại và cùng cấp cha (FR-017).

## Truy vết

- **FR**: FR-006, FR-016, FR-017
- **Acceptance**: US2 #2
- **BR**: BR-CAT-003
