# UC-BGT-02: Tạo ngân sách tổng chi tiêu

## Metadata
- **ID:** UC-BGT-02
- **Bounded Context:** Budgeting (thiết lập & theo dõi ngân sách của hộ)
- **Liên quan tới BR:** BR-003 (BR-BGT-002, BR-BGT-003, BR-BGT-004)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Ngân sách)
- **Last updated:** 2026-07-13

## Actor
- Thành viên hộ gia đình (quyền ngang nhau).

## Trigger (Khi nào use case bắt đầu)
- Thành viên mở màn hình tạo ngân sách và chọn loại **tổng chi tiêu**.

## Preconditions
- Thành viên đã đăng nhập và thuộc một hộ (UC-TRK-01).

## Main Flow
1. Thành viên mở màn hình tạo ngân sách.
2. Thành viên chọn loại **tổng chi tiêu** (không gắn danh mục cụ thể).
3. Thành viên nhập **số tiền giới hạn** (lớn hơn 0).
4. Thành viên chọn **kỳ áp dụng**: hàng tháng, hàng tuần, hoặc một lần.
5. Thành viên bấm **Lưu**.
6. Hệ thống xác thực dữ liệu (giới hạn là số dương; kỳ hợp lệ; hộ chưa có ngân sách tổng đang hoạt động trong cùng kỳ).
7. Hệ thống lưu ngân sách kèm **người tạo**; ngân sách tổng xuất hiện trong danh sách của hộ với **tiến độ bằng tổng mọi giao dịch Chi của hộ từ đầu kỳ hiện tại**, hoạt động song song và độc lập với các ngân sách theo danh mục — mục tiêu đạt được.

## Alternative Flows
- **4a. Kỳ một lần (bước 4):** Thành viên nhập khoảng ngày xác định (ngày kết thúc không trước ngày bắt đầu). Use case tiếp tục ở bước 5.

## Exceptions
- **E1. Giới hạn không hợp lệ (bước 6):** 0, số âm hoặc không phải số — hệ thống chặn lưu và thông báo "giới hạn phải lớn hơn 0". Use case tiếp tục ở bước 3.
- **E2. Trùng ngân sách tổng trong kỳ (bước 6):** Hộ đã có một ngân sách tổng đang hoạt động trong cùng kỳ — hệ thống chặn và chỉ tới ngân sách hiện có (tối đa một ngân sách tổng đang hoạt động mỗi kỳ). Use case tiếp tục ở bước 2.
- **E3. Khoảng ngày không hợp lệ (bước 6):** Với kỳ một lần, ngày kết thúc trước ngày bắt đầu — hệ thống chặn và thông báo. Use case tiếp tục ở bước 4 (luồng 4a).

## Postconditions
- **Thành công:** Ngân sách tổng tồn tại trong danh sách chung của hộ với giới hạn > 0 và kỳ hợp lệ, đúng người tạo; tiến độ bằng tổng chi tiêu của hộ trong kỳ hiện tại (SC-002).
- **Thất bại:** Không ngân sách nào được tạo; danh sách ngân sách không đổi.

## Acceptance Criteria
### AC-1: Tiến độ bằng tổng chi tiêu của hộ
Given: hộ có nhiều giao dịch Chi thuộc nhiều danh mục
When: tạo ngân sách tổng theo tháng
Then: tiến độ bằng tổng toàn bộ chi tiêu trong tháng
### AC-2: Chặn ngân sách tổng trùng trong kỳ
Given: đã có ngân sách tổng hàng tháng đang hoạt động
When: tạo thêm ngân sách tổng hàng tháng
Then: hệ thống chặn (tối đa một ngân sách tổng đang hoạt động mỗi kỳ)
### AC-3: Song song và độc lập với ngân sách danh mục
Given: ngân sách tổng và ngân sách danh mục cùng tồn tại
When: một giao dịch Chi được nhập
Then: tiến độ của cả hai đều cập nhật
And: cảnh báo của từng ngân sách hoạt động độc lập
### AC-4: Chặn giới hạn không hợp lệ
Given: thành viên nhập giới hạn 0 hoặc số âm
When: bấm Lưu
Then: hệ thống chặn và thông báo giới hạn phải lớn hơn 0

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); dùng chung bộ xác thực với UC-BGT-01.
- **Downstream UC:** UC-BGT-03 (theo dõi tiến độ); UC-BGT-04 (cảnh báo); UC-BGT-05/06 (sửa/xóa ngân sách).
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-002, FR-003, FR-004, FR-010, FR-011 ([spec 003](../../003-budgeting/spec.md)) · **Acceptance** US4 #1–#3 · **SC** SC-001, SC-002.
- Ngân sách tổng tính trên **mọi giao dịch loại Chi** của hộ trong kỳ, không phân biệt danh mục; giao dịch Thu không bao giờ tính vào (Assumptions spec 003).
- Ngân sách tổng luôn tồn tại song song, độc lập với các ngân sách theo danh mục (BR-BGT-002; giải nghĩa Open Question "một danh mục ≤ 1 ngân sách").

## Link file specs change
- [spec 003 §US4, FR-002/003/004/010/011](../../003-budgeting/spec.md)

## History
- v1 (2026-07-13, claude): initial từ US4 của spec 003.
