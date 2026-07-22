# UC-BGT-03: Theo dõi tiến độ ngân sách

## Metadata
- **ID:** UC-BGT-03
- **Bounded Context:** Budgeting (thiết lập & theo dõi ngân sách của hộ)
- **Liên quan tới BR:** BR-003 (BR-BGT-005)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Ngân sách)
- **Last updated:** 2026-07-13

## Actor
- Thành viên hộ gia đình (quyền ngang nhau).

## Trigger (Khi nào use case bắt đầu)
- Thành viên mở danh sách/màn hình ngân sách của hộ; hoặc một giao dịch Chi liên quan được thêm/sửa/xóa làm tiến độ thay đổi.

## Preconditions
- Thành viên đã đăng nhập và thuộc một hộ (UC-TRK-01).
- Hộ có ít nhất một ngân sách đang hoạt động (UC-BGT-01 hoặc UC-BGT-02).

## Main Flow
1. Thành viên mở danh sách ngân sách của hộ.
2. Với mỗi ngân sách, hệ thống hiển thị: tên/danh mục, giới hạn, **số đã chi trong kỳ hiện tại** và **phần trăm tiến độ** — ví dụ "3.500.000/5.000.000 VNĐ (70%)".
3. Hệ thống tính tiến độ bằng **tổng các giao dịch Chi liên quan trong kỳ**: theo danh mục (bao gồm danh mục con) với ngân sách danh mục, hoặc toàn bộ chi tiêu của hộ với ngân sách tổng.
4. Khi một giao dịch Chi liên quan được thêm/sửa/xóa hoặc đổi danh mục/loại (UC-TRK-02/04/05), hệ thống tính lại tiến độ đúng và cập nhật hiển thị trên màn hình của mọi thành viên đang mở — mục tiêu đạt được.

## Alternative Flows
- **3a. Giao dịch thuộc danh mục con (bước 3):** Giao dịch Chi thuộc danh mục con của danh mục có ngân sách được tính vào tiến độ của ngân sách danh mục **cha**. Use case tiếp tục ở bước 4.
- **4a. Giao dịch đổi giữa hai danh mục có ngân sách (bước 4):** Giao dịch chuyển từ danh mục A (có ngân sách) sang danh mục B (có ngân sách khác) — tiến độ của **cả hai** ngân sách được tính lại đúng. Use case tiếp tục ở bước 4.
- **4b. Sang kỳ mới (bước 4):** Khi bước sang kỳ tháng/tuần mới, ngân sách lặp lại tự động với cùng giới hạn; tiến độ và trạng thái cảnh báo bắt đầu lại từ 0. Use case tiếp tục ở bước 2.

## Exceptions
- **E1. Giao dịch đổi loại Chi → Thu (bước 4):** Giao dịch không còn được tính vào bất kỳ ngân sách nào; tiến độ liên quan được tính lại (giảm tương ứng). Use case tiếp tục ở bước 2.
- **E2. Sửa/xóa giao dịch quá khứ trong kỳ (bước 4):** Tiến độ được tính lại đúng theo dữ liệu mới, kể cả khi giao dịch nằm trong quá khứ của kỳ hiện tại. Use case tiếp tục ở bước 2.
- **E3. Kỳ chưa có giao dịch nào (bước 3):** Tiến độ hiển thị 0%; không có cảnh báo. Use case tiếp tục ở bước 2.
- **E4. Danh mục có ngân sách bị ẩn (bước 2):** Ngân sách vẫn theo dõi bình thường; danh sách hiển thị kèm trạng thái danh mục (đã ẩn) để thành viên biết. Use case tiếp tục ở bước 2.
- **E5. Ngân sách một lần đã kết thúc kỳ (bước 2):** Hiển thị trạng thái đã kết thúc, giữ lại để xem lại; không tính tiến độ mới, không phát cảnh báo. Use case tiếp tục ở bước 2.

## Postconditions
- **Thành công:** Tiến độ mỗi ngân sách khớp đúng tổng giao dịch Chi liên quan trong kỳ hiện tại tại mọi thời điểm đối chiếu (SC-003); thay đổi hiển thị cho các thành viên khác cùng hộ trong vòng 5 giây (SC-006).
- **Thất bại:** Không áp dụng — đây là use case đọc; nếu không tải được dữ liệu, hệ thống thông báo lỗi và không hiển thị số liệu sai.

## Acceptance Criteria
### AC-1: Hiển thị tiến độ đúng định dạng
Given: ngân sách "Ăn uống" 5.000.000/tháng với tổng chi "Ăn uống" trong tháng là 3.500.000
When: mở danh sách ngân sách
Then: hiển thị "3.500.000/5.000.000 (70%)"
### AC-2: Cập nhật khi thêm giao dịch
Given: thành viên nhập một giao dịch Chi thuộc danh mục có ngân sách
When: giao dịch được lưu
Then: tiến độ ngân sách cập nhật đúng số tiền mới
### AC-3: Tính lại khi sửa/xóa giao dịch quá khứ
Given: một giao dịch Chi trong quá khứ thuộc kỳ hiện tại bị sửa số tiền hoặc bị xóa
When: thao tác hoàn tất
Then: tiến độ ngân sách được tính lại đúng
### AC-4: Đổi danh mục cập nhật cả hai ngân sách
Given: một giao dịch đổi từ danh mục A (có ngân sách) sang danh mục B (có ngân sách khác)
When: lưu
Then: tiến độ của cả hai ngân sách đều cập nhật đúng
### AC-5: Đồng bộ realtime giữa thành viên
Given: Alice đang mở màn hình ngân sách
When: Bob nhập một giao dịch ảnh hưởng ngân sách
Then: tiến độ trên màn hình của Alice cập nhật trong vòng 5 giây
### AC-6: Danh mục con tính vào ngân sách cha
Given: giao dịch Chi thuộc danh mục con của danh mục có ngân sách
When: lưu
Then: giao dịch được tính vào tiến độ ngân sách của danh mục cha

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-BGT-01/02 (ngân sách tồn tại); đọc giao dịch của UC-TRK-02/04/05.
- **Downstream UC:** UC-BGT-04 (cảnh báo — kích hoạt bởi thay đổi tiến độ).
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-005, FR-006 ([spec 003](../../003-budgeting/spec.md)) · **Acceptance** US2 #1–#6 · **SC** SC-003, SC-006.
- Tiến độ là **giá trị suy ra** từ giao dịch (không lưu cứng), nhất quán cơ chế số dư suy ra của feature 002.
- Đồng bộ realtime ≤ 5 giây dùng chung cơ chế WebSocket của feature 001/002 (SC-006).
- Ngân sách chỉ **đọc** giao dịch & danh mục; không thay đổi dữ liệu nguồn (Assumptions spec 003).

## Link file specs change
- [spec 003 §US2, FR-005/006](../../003-budgeting/spec.md)

## History
- v1 (2026-07-13, claude): initial từ US2 của spec 003.
