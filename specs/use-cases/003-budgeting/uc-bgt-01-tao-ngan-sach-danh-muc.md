# UC-BGT-01: Tạo ngân sách theo danh mục

## Metadata
- **ID:** UC-BGT-01
- **Bounded Context:** Budgeting (thiết lập & theo dõi ngân sách của hộ)
- **Liên quan tới BR:** BR-003 (BR-BGT-001, BR-BGT-003, BR-BGT-004)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Ngân sách)
- **Last updated:** 2026-07-13

## Actor
- Thành viên hộ gia đình (quyền ngang nhau).

## Trigger (Khi nào use case bắt đầu)
- Thành viên mở màn hình tạo ngân sách và chọn loại **theo danh mục**.

## Preconditions
- Thành viên đã đăng nhập và thuộc một hộ (UC-TRK-01).
- Hộ có ít nhất một danh mục loại **Chi** (BR-001 / feature 001).

## Main Flow
1. Thành viên mở màn hình tạo ngân sách.
2. Thành viên chọn loại **theo danh mục**.
3. Thành viên chọn một danh mục loại **Chi** của hộ (danh sách chỉ gồm danh mục Chi).
4. Thành viên nhập **số tiền giới hạn** (lớn hơn 0).
5. Thành viên chọn **kỳ áp dụng**: hàng tháng, hàng tuần, hoặc một lần.
6. Thành viên bấm **Lưu**.
7. Hệ thống xác thực dữ liệu (danh mục thuộc hộ và loại Chi; giới hạn là số dương; kỳ hợp lệ; chưa có ngân sách đang hoạt động cho cùng danh mục trong cùng kỳ).
8. Hệ thống lưu ngân sách kèm **người tạo** (định danh người dùng); ngân sách xuất hiện trong danh sách ngân sách của hộ với **tiến độ tính từ toàn bộ chi tiêu của danh mục từ đầu kỳ hiện tại** — mục tiêu đạt được.

## Alternative Flows
- **5a. Kỳ một lần (bước 5):** Thành viên nhập khoảng ngày xác định (ngày kết thúc không trước ngày bắt đầu). Use case tiếp tục ở bước 6.
- **3a. Danh mục có ngân sách con tính vào (bước 3):** Ngân sách theo danh mục cha bao gồm cả danh mục con của danh mục đó khi tính tiến độ (nhất quán cây danh mục feature 001). Use case tiếp tục ở bước 4.

## Exceptions
- **E1. Giới hạn không hợp lệ (bước 7):** 0, số âm hoặc không phải số — hệ thống chặn lưu và thông báo "giới hạn phải lớn hơn 0" ngay tại trường giới hạn. Use case tiếp tục ở bước 4.
- **E2. Chưa chọn danh mục (bước 7):** Hệ thống chặn lưu và yêu cầu chọn danh mục Chi. Use case tiếp tục ở bước 3.
- **E3. Trùng ngân sách trong kỳ (bước 7):** Danh mục đã có một ngân sách đang hoạt động trong cùng kỳ — hệ thống chặn và chỉ tới ngân sách hiện có (mỗi danh mục tối đa một ngân sách đang hoạt động mỗi kỳ). Use case tiếp tục ở bước 3.
- **E4. Khoảng ngày không hợp lệ (bước 7):** Với kỳ một lần, ngày kết thúc trước ngày bắt đầu — hệ thống chặn và thông báo. Use case tiếp tục ở bước 5 (luồng 5a).

## Postconditions
- **Thành công:** Ngân sách theo danh mục tồn tại trong danh sách chung của hộ với đủ trường bắt buộc (danh mục Chi, giới hạn > 0, kỳ hợp lệ), đúng người tạo; tiến độ phản ánh đúng chi tiêu của danh mục trong kỳ hiện tại (SC-002).
- **Thất bại:** Không ngân sách nào được tạo; danh sách ngân sách không đổi.

## Acceptance Criteria
### AC-1: Tạo ngân sách danh mục hợp lệ
Given: thành viên chọn danh mục Chi "Ăn uống", nhập giới hạn 5.000.000, kỳ hàng tháng
When: bấm Lưu
Then: ngân sách được tạo và xuất hiện trong danh sách ngân sách của hộ
And: tiến độ phản ánh chi tiêu "Ăn uống" từ đầu kỳ hiện tại
### AC-2: Chặn giới hạn không hợp lệ
Given: thành viên nhập giới hạn 0 hoặc số âm
When: bấm Lưu
Then: hệ thống chặn và thông báo giới hạn phải lớn hơn 0
### AC-3: Bắt buộc chọn danh mục Chi
Given: thành viên chưa chọn danh mục (với ngân sách theo danh mục)
When: bấm Lưu
Then: hệ thống chặn và yêu cầu chọn danh mục
And: danh sách chọn chỉ gồm danh mục loại Chi của hộ
### AC-4: Chặn ngân sách trùng danh mục trong kỳ
Given: danh mục "Ăn uống" đã có ngân sách hàng tháng đang hoạt động
When: thành viên tạo thêm ngân sách hàng tháng cho chính danh mục đó
Then: hệ thống chặn và chỉ tới ngân sách hiện có
### AC-5: Dùng chung trong hộ, cô lập giữa các hộ
Given: Alice (hộ A) vừa tạo một ngân sách
When: Bob (cùng hộ A) mở danh sách ngân sách
Then: Bob thấy ngân sách đó
And: Carol (hộ B) không thấy

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); đọc danh mục Chi của feature 001.
- **Downstream UC:** UC-BGT-03 (theo dõi tiến độ); UC-BGT-04 (cảnh báo); UC-BGT-05/06 (sửa/xóa ngân sách).
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-001, FR-003, FR-004, FR-010, FR-011 ([spec 003](../../003-budgeting/spec.md)) · **Acceptance** US1 #1–#5 · **SC** SC-001, SC-002.
- Ngân sách chỉ **đọc** danh mục & giao dịch để tính tiến độ; không thay đổi dữ liệu nguồn (Assumptions spec 003).
- Ngân sách hàng tháng/hàng tuần tự lặp lại mỗi kỳ với cùng giới hạn, tiến độ bắt đầu lại từ 0 (FR-004; xử lý sang kỳ mới ở UC-BGT-03).
- Mục tiêu hiệu năng: tạo xong ≤ 30 giây, tối đa 3 bước thao tác chính (SC-001).

## Link file specs change
- [spec 003 §US1, FR-001/003/004/010/011](../../003-budgeting/spec.md)

## History
- v1 (2026-07-13, claude): initial từ US1 của spec 003.
