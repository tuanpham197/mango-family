# UC-BGT-05: Chỉnh sửa ngân sách

## Metadata
- **ID:** UC-BGT-05
- **Bounded Context:** Budgeting (thiết lập & theo dõi ngân sách của hộ)
- **Liên quan tới BR:** BR-003 (BR-BGT-003, BR-BGT-004; vòng đời suy luận — FR-012/013)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Ngân sách)
- **Last updated:** 2026-07-13

## Actor
- Thành viên hộ gia đình (bất kỳ — ngang quyền, kể cả với ngân sách do người khác tạo).

## Trigger (Khi nào use case bắt đầu)
- Thành viên chọn một ngân sách trong danh sách (UC-BGT-03) và mở chỉnh sửa.

## Preconditions
- Thành viên đã đăng nhập và thuộc hộ (UC-TRK-01).
- Ngân sách cần sửa đang tồn tại trong danh sách của hộ (UC-BGT-03).

## Main Flow
1. Thành viên chọn một ngân sách và bấm **Sửa**.
2. Hệ thống mở form với dữ liệu hiện tại của ngân sách.
3. Thành viên chỉnh sửa các trường cần thiết: **giới hạn**, **kỳ**, hoặc **danh mục** (với ngân sách theo danh mục).
4. Thành viên bấm **Lưu**.
5. Hệ thống xác thực dữ liệu theo đúng quy tắc như khi tạo (UC-BGT-01/02 bước xác thực: giới hạn > 0; kỳ hợp lệ; không trùng ngân sách hiện có cùng danh mục+kỳ / tổng+kỳ).
6. Hệ thống lưu thay đổi; **tiến độ và trạng thái cảnh báo được tính lại theo giới hạn/kỳ/danh mục mới** (UC-BGT-03/04); danh sách của mọi thành viên hiển thị bản cập nhật — mục tiêu đạt được.

## Alternative Flows
- **3a. Đổi danh mục sang danh mục Chi khác (bước 3):** Hệ thống kiểm tra danh mục mới thuộc hộ và loại Chi, và chưa có ngân sách trùng trong cùng kỳ. Use case tiếp tục ở bước 4.
- **6a. Giới hạn mới thấp hơn khiến tiến độ vượt ngưỡng (bước 6):** Trạng thái cảnh báo được tính lại theo giới hạn mới; nếu tiến độ chạm/vượt mức chưa phát trong kỳ, cảnh báo tương ứng phát (UC-BGT-04). Use case tiếp tục ở bước 6.

## Exceptions
- **E1. Dữ liệu không hợp lệ (bước 5):** Hệ thống chặn lưu và chỉ rõ lỗi theo từng trường (ví dụ giới hạn ≤ 0, khoảng ngày sai). Use case tiếp tục ở bước 3.
- **E2. Trùng ngân sách hiện có sau khi đổi danh mục/kỳ (bước 5):** Hệ thống chặn và chỉ tới ngân sách hiện có (FR-010). Use case tiếp tục ở bước 3.
- **E3. Ngân sách đã bị thành viên khác thay đổi trong lúc mình sửa (bước 6):** Hệ thống không ghi đè thầm lặng — thông báo bản ghi đã thay đổi và hiển thị dữ liệu mới nhất. Use case tiếp tục ở bước 3.
- **E4. Ngân sách đã bị thành viên khác xóa (bước 6):** Hệ thống thông báo ngân sách không còn tồn tại và quay về danh sách. Use case kết thúc.

## Postconditions
- **Thành công:** Ngân sách mang giá trị mới, vẫn đủ trường bắt buộc; tiến độ và trạng thái cảnh báo phản ánh đúng giới hạn/kỳ/danh mục mới (SC-002, SC-003).
- **Thất bại:** Ngân sách giữ nguyên như trước; tiến độ và trạng thái cảnh báo không đổi.

## Acceptance Criteria
### AC-1: Sửa giới hạn — tiến độ & cảnh báo tính lại
Given: ngân sách 5.000.000 đã chi 3.500.000 (70%)
When: sửa giới hạn thành 4.000.000
Then: tiến độ hiển thị 87,5%
And: cảnh báo ngưỡng 80% phát theo trạng thái mới
### AC-2: Ngang quyền sửa ngân sách người khác tạo
Given: Bob mở sửa ngân sách do Alice tạo
When: Bob lưu thay đổi hợp lệ
Then: lưu thành công (mọi thành viên ngang quyền)
### AC-3: Không ghi đè thầm lặng
Given: hai thành viên cùng sửa một ngân sách
When: người thứ hai lưu sau khi người thứ nhất đã lưu
Then: hệ thống thông báo dữ liệu đã thay đổi
And: hiển thị dữ liệu mới nhất, không ghi đè thầm lặng
### AC-4: Sửa qua cùng bộ xác thực như tạo
Given: thành viên sửa giới hạn thành 0 hoặc số âm
When: bấm Lưu
Then: hệ thống chặn và thông báo giới hạn phải lớn hơn 0

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-BGT-03 «extend» (điểm vào từ danh sách); dùng chung xác thực UC-BGT-01/02.
- **Downstream UC:** UC-BGT-03 (tiến độ tính lại); UC-BGT-04 (cảnh báo tính lại theo giới hạn mới).
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-012, FR-013 ([spec 003](../../003-budgeting/spec.md)) · **Acceptance** US5 #1, #3, #4 · **SC** SC-002, SC-003, SC-006.
- Sửa/xóa ngân sách là **vòng đời tối thiểu suy luận** (BR-003 In Scope không liệt kê — cần nghiệp vụ xác nhận; Assumptions spec 003).
- Cơ chế chống ghi đè thầm lặng nhất quán với feature 002 (cập nhật lạc quan theo mốc thời gian sửa đổi).

## Link file specs change
- [spec 003 §US5, FR-012/013](../../003-budgeting/spec.md)

## History
- v1 (2026-07-13, claude): initial từ US5 của spec 003.
