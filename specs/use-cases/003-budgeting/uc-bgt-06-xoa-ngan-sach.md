# UC-BGT-06: Xóa ngân sách (có xác nhận)

## Metadata
- **ID:** UC-BGT-06
- **Bounded Context:** Budgeting (thiết lập & theo dõi ngân sách của hộ)
- **Liên quan tới BR:** BR-003 (vòng đời suy luận — FR-012/013)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Ngân sách)
- **Last updated:** 2026-07-13

## Actor
- Thành viên hộ gia đình (bất kỳ — ngang quyền, kể cả với ngân sách do người khác tạo).

## Trigger (Khi nào use case bắt đầu)
- Thành viên chọn một ngân sách trong danh sách (UC-BGT-03) và bấm **Xóa**.

## Preconditions
- Thành viên đã đăng nhập và thuộc hộ (UC-TRK-01).
- Ngân sách cần xóa đang tồn tại trong danh sách của hộ (UC-BGT-03).

## Main Flow
1. Thành viên chọn một ngân sách và bấm **Xóa**.
2. Hệ thống yêu cầu **xác nhận** xóa.
3. Thành viên xác nhận.
4. Hệ thống xóa ngân sách cùng các cảnh báo liên quan; ngân sách biến mất khỏi danh sách của **mọi thành viên** trong hộ — mục tiêu đạt được.

## Alternative Flows
- **2a. Hủy xác nhận (bước 2):** Thành viên chọn Hủy — không gì thay đổi, ngân sách vẫn còn. Use case kết thúc.

## Exceptions
- **E1. Ngân sách đã bị thành viên khác xóa (bước 4):** Hệ thống thông báo ngân sách không còn tồn tại và làm mới danh sách. Use case kết thúc.

## Postconditions
- **Thành công:** Ngân sách và các cảnh báo liên quan không còn trong danh sách của bất kỳ thành viên nào; dữ liệu giao dịch và danh mục nguồn **không** bị ảnh hưởng.
- **Thất bại (hủy hoặc lỗi):** Ngân sách vẫn tồn tại nguyên vẹn; danh sách không đổi.

## Acceptance Criteria
### AC-1: Xóa có xác nhận
Given: một ngân sách bất kỳ
When: thành viên bấm Xóa
Then: hệ thống yêu cầu xác nhận
And: xác nhận xong ngân sách và cảnh báo liên quan biến mất khỏi danh sách của mọi thành viên
### AC-2: Hủy không thay đổi gì
Given: hộp xác nhận xóa đang hiển thị
When: thành viên chọn Hủy
Then: không gì thay đổi, ngân sách vẫn còn trong danh sách
### AC-3: Ngang quyền xóa ngân sách người khác tạo
Given: Bob mở xóa ngân sách do Alice tạo
When: Bob xác nhận xóa
Then: ngân sách bị xóa thành công (mọi thành viên ngang quyền)

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-BGT-03 «extend» (điểm vào từ danh sách).
- **Downstream UC:** Không.
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-012, FR-013 ([spec 003](../../003-budgeting/spec.md)) · **Acceptance** US5 #2 · **SC** SC-006.
- Sửa/xóa ngân sách là **vòng đời tối thiểu suy luận** (BR-003 In Scope không liệt kê — cần nghiệp vụ xác nhận; Assumptions spec 003).
- Xóa ngân sách chỉ gỡ ngân sách + cảnh báo của nó; giao dịch và danh mục nguồn giữ nguyên (ngân sách chỉ đọc dữ liệu nguồn).

## Link file specs change
- [spec 003 §US5, FR-012/013](../../003-budgeting/spec.md)

## History
- v1 (2026-07-13, claude): initial từ US5 của spec 003.
