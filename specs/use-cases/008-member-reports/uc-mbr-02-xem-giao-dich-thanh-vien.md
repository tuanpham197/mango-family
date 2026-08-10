# UC-MBR-02: Xem danh sách giao dịch của một thành viên

## Metadata
- **ID:** UC-MBR-02
- **Bounded Context:** Reports (drill-down chỉ đọc từ báo cáo theo thành viên)
- **Liên quan tới BR:** BR-008 (BR-MBR-006, BR-MBR-007, BR-MBR-008, BR-MBR-009)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Báo cáo)
- **Last updated:** 2026-08-10

## Actor
- Thành viên hộ gia đình (quyền ngang nhau).

## Trigger (Khi nào use case bắt đầu)
- Từ báo cáo theo thành viên (UC-MBR-01), thành viên chọn một thành viên để xem các giao dịch cấu thành số liệu của người đó.

## Preconditions
- Thành viên đã đăng nhập và thuộc một hộ (UC-TRK-01).
- Đang xem báo cáo theo thành viên với một khoảng thời gian đã chọn (UC-MBR-01).

## Main Flow
1. Thành viên chọn một thành viên trên báo cáo theo thành viên.
2. Hệ thống xác thực thành viên được chọn thuộc đúng hộ hiện tại trước khi truy vấn *(BR-MBR-008)*.
3. Hệ thống liệt kê các giao dịch được tính vào số liệu của thành viên đó (theo `created_by`) trong **khoảng thời gian hiện tại** *(BR-MBR-006)*.
4. Trên mỗi giao dịch, hệ thống hiển thị tối thiểu: loại Thu/Chi, số tiền, danh mục, mô tả, ngày giờ và tài khoản *(BR-MBR-006)*.
5. Khi danh sách dài, hệ thống phân trang danh sách giao dịch *(BR-MBR-007)*.
6. Thành viên quay lại báo cáo theo thành viên; hệ thống giữ nguyên khoảng thời gian đã chọn *(BR-MBR-007)* — mục tiêu (soi các giao dịch cấu thành số liệu của một thành viên) đạt được.

## Alternative Flows
- **3a. Thành viên được chọn không có giao dịch trong khoảng (bước 3):** Hệ thống hiển thị trạng thái không có dữ liệu và cho phép quay lại báo cáo *(BR-MBR-007)*. Use case tiếp tục ở bước 6.
- **4a. Người nhập chưa có tên hiển thị (bước 4):** Nơi hiển thị tên thành viên dùng email thay cho tên; không hiển thị mã định danh thô *(BR-MBR-009)*. Use case tiếp tục ở bước 5.
- **5a. Danh sách còn trang tiếp theo (bước 5):** Thành viên chuyển trang; hệ thống nạp trang kế trong cùng khoảng và cùng thành viên. Use case tiếp tục ở bước 6.

## Exceptions
- **E1. Thành viên được chọn không thuộc hộ hiện tại (bước 2):** Hệ thống từ chối truy vấn và không trả về giao dịch nào — backend không nhận trực tiếp một user ID ngoài hộ *(BR-MBR-008)*. Use case kết thúc.

## Postconditions
- **Thành công:** Thành viên xem được đúng và đủ các giao dịch tính cho thành viên được chọn trong khoảng hiện tại; quay lại báo cáo giữ nguyên khoảng.
- **Thất bại:** Không truy xuất được giao dịch của thành viên ngoài hộ; không dữ liệu hộ khác bị lộ.

## Acceptance Criteria
### AC-1: Drill-down đúng giao dịch của thành viên
Given: đang xem báo cáo theo thành viên hộ A cho tháng này, Alice có nhiều giao dịch
When: chọn Alice
Then: hệ thống liệt kê đúng các giao dịch Alice đã nhập trong tháng này
And: mỗi giao dịch hiển thị loại Thu/Chi, số tiền, danh mục, mô tả, ngày giờ và tài khoản
### AC-2: Giữ khoảng khi quay lại
Given: đang xem giao dịch của Alice trong khoảng "quý này"
When: quay lại báo cáo theo thành viên
Then: khoảng thời gian vẫn là "quý này", không bị đặt lại
### AC-3: Trạng thái không có dữ liệu
Given: Carol không có giao dịch trong khoảng đã chọn
When: chọn Carol
Then: hệ thống hiển thị trạng thái không có dữ liệu và cho phép quay lại
### AC-4: Phân trang danh sách dài
Given: một thành viên có số lượng giao dịch vượt một trang
When: mở danh sách giao dịch của thành viên đó
Then: danh sách được phân trang và có thể chuyển sang trang tiếp theo trong cùng khoảng và cùng thành viên
### AC-5: Chặn thành viên ngoài hộ
Given: một yêu cầu drill-down tới một thành viên không thuộc hộ hiện tại
When: hệ thống nhận yêu cầu
Then: hệ thống từ chối và không trả về giao dịch nào

## Dependencies
- **Upstream UC:** UC-MBR-01 (báo cáo theo thành viên — điểm vào & khoảng thời gian); UC-TRK-01 (định danh & tư cách thành viên).
- **Downstream UC:** Không (drill-down chỉ đọc; không sửa/xóa từ đây — sửa/xóa giao dịch vẫn qua UC-TRK-04/05 trên sổ chung).
- **External Systems:** Không.

## Notes
- Truy vết: **BR-008** BR-MBR-006 (trường tối thiểu hiển thị), BR-MBR-007 (empty state, phân trang, giữ khoảng khi quay lại), BR-MBR-008 (xác thực thành viên thuộc hộ; không nhận user ID ngoài hộ), BR-MBR-009 (tên/fallback email).
- Nhóm theo `transactions.created_by` với cùng điều kiện thời gian/loại/cô lập hộ như endpoint báo cáo hiện tại (BR-008 Constraints); cân nhắc index `(household_id, created_by, transaction_date)` sau khi đo bằng execution plan.
- Báo cáo **chỉ đọc**: use case này chỉ hiển thị giao dịch, không thay đổi quy thuộc (`created_by`) hay dữ liệu giao dịch (BR-008 Out of Scope).

## Link file specs change
- [BR-008](../../business-requirements/BR-008.md) · README chỉ mục: [`../README.md`](../README.md) · sơ đồ: [`../../diagrams/use-cases.puml`](../../diagrams/use-cases.puml)

## History
- v1 (2026-08-10, claude): initial — dẫn xuất từ BR-008 (UC-MBR-02); drill-down danh sách giao dịch của một thành viên trong khoảng hiện tại, trường tối thiểu, empty state + phân trang, giữ khoảng khi quay lại, xác thực thành viên thuộc hộ.
