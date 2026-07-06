# UC-CAT-05: Xóa danh mục (gán lại / xóa giao dịch)

| Thuộc tính | Giá trị |
|------------|---------|
| **Mã** | UC-CAT-05 |
| **Tính năng** | Phân loại giao dịch ([spec.md](../../001-transaction-categorization/spec.md)) |
| **User Story** | US2 (P2), US3 (P3) |
| **Tác nhân chính** | Người dùng |
| **Mức độ ưu tiên** | Trung bình |

## Mục tiêu

Người dùng xóa một danh mục không còn cần dùng, đồng thời **không bao giờ để lại
giao dịch nào không có danh mục**.

## Tiền điều kiện

- Người dùng đã đăng nhập.
- Danh mục cần xóa đang tồn tại.

## Kích hoạt (Trigger)

Người dùng chọn "Xóa" trên một danh mục.

## Luồng sự kiện chính (danh mục KHÔNG còn giao dịch)

1. Người dùng yêu cầu xóa một danh mục.
2. Hệ thống xác định danh mục không có giao dịch liên quan và không có danh mục con.
3. Hệ thống xóa danh mục ngay, không cần bước gán lại.

## Luồng thay thế / ngoại lệ

- **2a. Danh mục còn giao dịch liên quan**:
  1. Hệ thống cảnh báo và yêu cầu người dùng chọn một trong hai:
     - **Gán lại** các giao dịch sang một danh mục khác **cùng loại** Thu/Chi, hoặc
     - **Xóa** các giao dịch đó.
  2. Nếu chọn gán lại, hệ thống chỉ cho phép chọn danh mục đích **cùng loại** với giao dịch.
  3. Sau khi gán lại/xóa giao dịch xong, hệ thống mới hoàn tất việc xóa danh mục.
- **2b. Danh mục cha có danh mục con**:
  1. Hệ thống xử lý cả các **danh mục con** theo cùng cơ chế bảo vệ giao dịch (gán lại hoặc xóa).
  2. Việc xóa chỉ hoàn tất khi mọi giao dịch của cha và của các con đã được gán lại hoặc xóa.
- **Gán lại sai loại**: Hệ thống không cho phép chọn danh mục đích khác loại Thu/Chi.

## Hậu điều kiện

- Danh mục (và các danh mục con của nó, nếu có) bị xóa.
- Không tồn tại giao dịch nào bị mất danh mục: tất cả đã được gán lại sang danh mục
  cùng loại hoặc đã bị xóa.

## Quy tắc nghiệp vụ

- Danh mục mặc định và tự tạo được đối xử như nhau khi xóa (FR-007).
- Bắt buộc gán lại hoặc xóa giao dịch trước khi hoàn tất xóa (FR-008).
- Chỉ cho phép gán lại sang danh mục **cùng loại** Thu/Chi (FR-009).
- Xóa danh mục cha kéo theo xử lý danh mục con với cùng cơ chế (FR-012).

## Truy vết

- **FR**: FR-007, FR-008, FR-009, FR-012
- **Acceptance**: US2 #3, #4, #5; US3 #3
- **Success Criteria**: SC-007 (100% thao tác xóa kết thúc không để lại giao dịch mất danh mục)
- **BR**: BR-CAT-004
