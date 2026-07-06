# UC-CAT-03: Tạo danh mục con (một cấp)

| Thuộc tính | Giá trị |
|------------|---------|
| **Mã** | UC-CAT-03 |
| **Tính năng** | Phân loại giao dịch ([spec.md](../../001-transaction-categorization/spec.md)) |
| **User Story** | US3 (P3) |
| **Tác nhân chính** | Người dùng |
| **Mức độ ưu tiên** | Thấp |
| **Quan hệ** | «extend» UC-CAT-02 |

## Mục tiêu

Người dùng tổ chức chi tiêu chi tiết hơn bằng cách tạo danh mục con nằm dưới một
danh mục cha (ví dụ "Ăn uống" → "Ăn ngoài", "Đi chợ"), giúp báo cáo phân bổ chính xác hơn.

## Tiền điều kiện

- Người dùng đã đăng nhập.
- Đã tồn tại một danh mục cha (mặc định hoặc tự tạo).

## Kích hoạt (Trigger)

Người dùng chọn "Thêm danh mục con" dưới một danh mục cha.

## Luồng sự kiện chính

1. Người dùng chọn một danh mục cha và yêu cầu tạo danh mục con.
2. Người dùng nhập **tên** danh mục con và chọn **biểu tượng** (tùy chọn).
3. Hệ thống tự động **kế thừa loại Thu/Chi của danh mục cha** cho danh mục con (không cho phép chọn loại khác).
4. Người dùng xác nhận lưu.
5. Hệ thống tạo danh mục con lồng dưới danh mục cha và hiển thị trong cùng nhóm loại.

## Luồng thay thế / ngoại lệ

- **1a. Tạo con dưới một danh mục con**: Nếu danh mục được chọn đã là danh mục con,
  hệ thống KHÔNG cho phép tạo thêm cấp con (chỉ hỗ trợ một cấp).
- **2a. Tên trùng**: Nếu tên trùng với danh mục khác trong cùng loại và cùng cấp cha,
  hệ thống cảnh báo (FR-017).
- **3a. Cố đổi loại con**: Nếu người dùng cố gán loại khác với cha, hệ thống từ chối.

## Hậu điều kiện

- Một danh mục con tồn tại, tham chiếu đúng một danh mục cha, mang đúng loại Thu/Chi của cha.

## Quy tắc nghiệp vụ

- Hệ thống chỉ hỗ trợ danh mục con lồng **đúng một cấp** dưới danh mục cha (FR-010).
- Danh mục con **bắt buộc kế thừa** loại của cha và không thể mang loại khác (FR-011).
- Cả danh mục cha (kể cả khi đã có con) lẫn danh mục con đều có thể được giao dịch tham chiếu trực tiếp (xem UC-CAT-07, báo cáo roll-up).

## Truy vết

- **FR**: FR-010, FR-011, FR-017
- **Acceptance**: US3 #1, #2
- **Success Criteria**: SC-008 (100% danh mục con mang đúng loại của cha)
- **BR**: BR-CAT-005
