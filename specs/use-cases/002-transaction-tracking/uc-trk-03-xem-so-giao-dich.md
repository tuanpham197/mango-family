# UC-TRK-03: Xem sổ giao dịch chung của hộ

## Metadata
- **ID:** UC-TRK-03
- **Bounded Context:** Transaction Tracking (sổ chung minh bạch của hộ)
- **Liên quan tới BR:** BR-002 (Goal); liên kết BR-CAT-009 (hiển thị người nhập)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Giao dịch)
- **Last updated:** 2026-07-09

## Actor
- Thành viên hộ gia đình (quyền ngang nhau).

## Trigger (Khi nào use case bắt đầu)
- Thành viên mở sổ giao dịch của hộ.

## Preconditions
- Thành viên đã đăng nhập và thuộc một hộ (UC-TRK-01).

## Main Flow
1. Thành viên mở sổ giao dịch.
2. Hệ thống liệt kê các giao dịch **của hộ mình** (và chỉ của hộ mình), sắp xếp theo ngày giờ **mới nhất trước**.
3. Hệ thống hiển thị trên mỗi dòng: số tiền, loại Thu/Chi, danh mục, ngày giờ và **tên thành viên đã nhập** (tên hiển thị dễ đọc).
4. Khi một thành viên khác thêm/sửa/xóa giao dịch, hệ thống làm tươi sổ của thành viên đang xem trong vòng 5 giây (SC-006).
5. Thành viên chọn một giao dịch để xem chi tiết, chỉnh sửa (UC-TRK-04) hoặc xóa (UC-TRK-05) — mục tiêu (sổ chung minh bạch làm điểm vào thao tác) đạt được.

## Alternative Flows
- **3a. Người nhập chưa có tên hiển thị (bước 3):** Hệ thống hiển thị email thay cho tên; không bao giờ hiển thị mã định danh thô. Use case tiếp tục ở bước 4.

## Exceptions
- **E1. Sổ chưa có giao dịch nào (bước 2):** Hệ thống hiển thị trạng thái rỗng và hướng thành viên đến việc nhập giao dịch đầu tiên (UC-TRK-02). Use case kết thúc.

## Postconditions
- **Thành công:** Thành viên nhìn thấy đúng và đủ giao dịch của hộ mình, kèm tên người nhập; không thấy dữ liệu của hộ khác.
- **Thất bại:** Không dữ liệu nào của hộ khác bị lộ.

## Acceptance Criteria
### AC-1: Sổ chung trong hộ
Given: Alice (hộ A) vừa nhập một giao dịch
When: Bob (cùng hộ A) mở sổ giao dịch
Then: Bob thấy giao dịch đó với đầy đủ số tiền, danh mục, ngày
And: người nhập hiển thị là Alice
### AC-2: Thứ tự mới nhất trước
Given: sổ có nhiều giao dịch
When: mở danh sách
Then: giao dịch sắp xếp theo ngày giờ mới nhất trước
### AC-3: Cô lập giữa các hộ
Given: Carol thuộc hộ B
When: Carol mở sổ
Then: Carol không thấy bất kỳ giao dịch nào của hộ A

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-TRK-02 (có giao dịch để hiển thị).
- **Downstream UC:** UC-TRK-04 (sửa) và UC-TRK-05 (xóa) «extend» từ sổ này.
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-008, FR-012, FR-013, FR-015 ([spec 002](../../002-transaction-tracking/spec.md)) · **Acceptance** US2 #1–#3 · **SC** SC-006.
- Sổ dùng chung trong hộ, cô lập giữa các hộ (FR-012); hiển thị **tên** người nhập — minh bạch sổ chung (BR-CAT-009).

## History
- v1 (2026-07-06, claude): initial từ US2 của spec 002.
- v2 (2026-07-06, claude): chuyển sang cấu trúc `specs/use-cases/template.md`.
- v3 (2026-07-09, claude): chốt ngưỡng làm tươi sổ ≤ 5 giây theo SC-006 (spec 002 v3).
