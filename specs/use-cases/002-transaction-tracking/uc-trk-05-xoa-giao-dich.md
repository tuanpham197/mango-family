# UC-TRK-05: Xóa giao dịch (có xác nhận)

## Metadata
- **ID:** UC-TRK-05
- **Bounded Context:** Transaction Tracking (ghi chép thu chi của hộ)
- **Liên quan tới BR:** BR-002 (BR-TRK-009, BR-TRK-010)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Giao dịch)
- **Last updated:** 2026-07-06

## Actor
- Thành viên hộ gia đình (bất kỳ — ngang quyền, kể cả với giao dịch do người khác nhập).

## Trigger (Khi nào use case bắt đầu)
- Thành viên chọn một giao dịch trong sổ (UC-TRK-03) và bấm Xóa.

## Preconditions
- Thành viên đã đăng nhập và thuộc hộ (UC-TRK-01).
- Giao dịch cần xóa tồn tại trong sổ của hộ (UC-TRK-03).

## Main Flow
1. Thành viên chọn một giao dịch trong sổ và bấm **Xóa**.
2. Hệ thống hiển thị **cảnh báo xác nhận**, nêu rõ việc xóa là vĩnh viễn.
3. Thành viên xác nhận xóa.
4. Hệ thống xóa giao dịch, **hoàn tác số dư** tài khoản liên quan đúng phần của giao dịch; sổ của mọi thành viên trong hộ cập nhật (giao dịch biến mất) — mục tiêu đạt được.

## Alternative Flows
- **3a. Thành viên hủy tại bước xác nhận (bước 3):** Không có gì thay đổi — giao dịch và số dư giữ nguyên. Use case kết thúc.

## Exceptions
- **E1. Giao dịch đã bị thành viên khác xóa trước đó (bước 4):** Hệ thống thông báo giao dịch không còn tồn tại và làm tươi sổ. Use case kết thúc.

## Postconditions
- **Thành công:** Giao dịch không còn trong sổ của bất kỳ thành viên nào; số dư tài khoản khớp đúng tổng các bút toán còn lại (SC-004).
- **Thất bại/Hủy:** Giao dịch và số dư giữ nguyên.

## Acceptance Criteria
### AC-1: Luôn có xác nhận trước khi xóa
Given: một giao dịch bất kỳ
When: thành viên bấm Xóa
Then: hệ thống hiển thị cảnh báo xác nhận trước khi xóa vĩnh viễn
### AC-2: Hủy thì không thay đổi
Given: hộp xác nhận đang hiển thị
When: thành viên hủy
Then: giao dịch và số dư giữ nguyên
### AC-3: Xóa xong số dư hoàn tác đúng
Given: thành viên xác nhận xóa một giao dịch Chi 50.000
When: xóa hoàn tất
Then: giao dịch biến mất khỏi sổ của mọi thành viên
And: số dư tài khoản liên quan tăng lại 50.000
### AC-4: Ngang quyền
Given: Bob xóa giao dịch do Alice nhập
When: xác nhận
Then: xóa thành công (ngang quyền)
And: sổ của Alice cập nhật theo

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-TRK-03 «extend» (điểm vào từ sổ).
- **Downstream UC:** Không.
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-010, FR-011, FR-013, FR-014 ([spec 002](../../002-transaction-tracking/spec.md)) · **Acceptance** US4 #1–#4 · **SC** SC-004, SC-005, SC-007.
- Không có xóa bằng một thao tác duy nhất (SC-005); thao tác đồng thời không gây lỗi khó hiểu (FR-014).

## History
- v1 (2026-07-06, claude): initial từ US4 của spec 002.
- v2 (2026-07-06, claude): chuyển sang cấu trúc `specs/use-cases/template.md`.
