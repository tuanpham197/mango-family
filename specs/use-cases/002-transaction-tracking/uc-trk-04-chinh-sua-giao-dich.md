# UC-TRK-04: Chỉnh sửa giao dịch

## Metadata
- **ID:** UC-TRK-04
- **Bounded Context:** Transaction Tracking (ghi chép thu chi của hộ)
- **Liên quan tới BR:** BR-002 (BR-TRK-008, BR-TRK-010)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Giao dịch)
- **Last updated:** 2026-07-06

## Actor
- Thành viên hộ gia đình (bất kỳ — ngang quyền, kể cả với giao dịch do người khác nhập).

## Trigger (Khi nào use case bắt đầu)
- Thành viên chọn một giao dịch trong sổ (UC-TRK-03) và mở chỉnh sửa.

## Preconditions
- Thành viên đã đăng nhập và thuộc hộ (UC-TRK-01).
- Giao dịch cần sửa tồn tại trong sổ của hộ (UC-TRK-03).

## Main Flow
1. Thành viên chọn một giao dịch trong sổ và bấm Sửa.
2. Hệ thống mở form với dữ liệu hiện tại của giao dịch.
3. Thành viên chỉnh sửa các trường cần thiết: số tiền, loại, danh mục, mô tả, ngày giờ, tài khoản.
4. Thành viên bấm **Lưu**.
5. Hệ thống xác thực dữ liệu theo đúng quy tắc như khi nhập mới (UC-TRK-02 bước 8).
6. Hệ thống lưu thay đổi, **tính lại số dư** các tài khoản liên quan (kể cả khi đổi số tiền, đổi loại, đổi tài khoản hoặc ngày trong quá khứ); sổ của mọi thành viên hiển thị bản cập nhật — mục tiêu đạt được.

## Alternative Flows
- **3a. Đổi loại Thu ⇄ Chi (bước 3):** Hệ thống bỏ danh mục cũ (khác loại) và yêu cầu chọn lại danh mục **cùng loại mới** (UC-CAT-07). Use case tiếp tục ở bước 4.

## Exceptions
- **E1. Dữ liệu không hợp lệ (bước 5):** Hệ thống chặn lưu và chỉ rõ lỗi theo từng trường. Use case tiếp tục ở bước 3.
- **E2. Giao dịch đã bị thành viên khác thay đổi trong lúc mình sửa (bước 6):** Hệ thống không ghi đè thầm lặng — thông báo bản ghi đã thay đổi và hiển thị dữ liệu mới nhất. Use case tiếp tục ở bước 3.
- **E3. Giao dịch đã bị thành viên khác xóa (bước 6):** Hệ thống thông báo giao dịch không còn tồn tại và quay về sổ. Use case kết thúc.

## Postconditions
- **Thành công:** Giao dịch mang giá trị mới, vẫn đủ trường bắt buộc và danh mục cùng loại; số dư mọi tài khoản liên quan khớp đúng tổng bút toán (SC-004).
- **Thất bại:** Giao dịch giữ nguyên như trước; số dư không đổi.

## Acceptance Criteria
### AC-1: Sửa số tiền — số dư điều chỉnh đúng chênh lệch
Given: một giao dịch Chi 50.000 đã lưu
When: thành viên sửa số tiền thành 80.000 và lưu
Then: số dư tài khoản liên quan được điều chỉnh thêm 30.000 chênh lệch
And: đối chiếu tổng thể luôn khớp
### AC-2: Đổi loại buộc chọn lại danh mục
Given: một giao dịch loại Chi
When: thành viên đổi loại sang Thu
Then: danh mục cũ (loại Chi) không còn hợp lệ
And: hệ thống yêu cầu chọn danh mục loại Thu trước khi lưu
### AC-3: Ngang quyền
Given: Bob mở sửa giao dịch do Alice nhập
When: Bob lưu thay đổi hợp lệ
Then: lưu thành công (mọi thành viên ngang quyền)
### AC-4: Không ghi đè thầm lặng
Given: hai thành viên cùng mở sửa một giao dịch
When: người thứ hai lưu sau khi người thứ nhất đã lưu
Then: hệ thống thông báo dữ liệu đã thay đổi
And: hiển thị dữ liệu mới nhất, không ghi đè thầm lặng
### AC-5: Sửa ngày quá khứ vẫn nhất quán
Given: giao dịch được sửa ngày về một ngày trong quá khứ
When: lưu
Then: sổ và số dư vẫn nhất quán (số dư phản ánh đúng tổng mọi giao dịch)

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-TRK-03 «extend» (điểm vào từ sổ); UC-CAT-07 «include» khi đổi loại.
- **Downstream UC:** Không.
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-009, FR-011, FR-013, FR-014 ([spec 002](../../002-transaction-tracking/spec.md)) · **Acceptance** US3 #1–#5 · **SC** SC-004, SC-007.
- Cơ chế chống ghi đè thầm lặng nhất quán với feature 001 (cập nhật lạc quan — R13).

## History
- v1 (2026-07-06, claude): initial từ US3 của spec 002.
- v2 (2026-07-06, claude): chuyển sang cấu trúc `specs/use-cases/template.md`; tách Exceptions khỏi Alternative Flows.
