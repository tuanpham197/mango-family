# UC-BGT-04: Nhận cảnh báo ngưỡng & vượt ngân sách

## Metadata
- **ID:** UC-BGT-04
- **Bounded Context:** Budgeting (thiết lập & theo dõi ngân sách của hộ)
- **Liên quan tới BR:** BR-003 (BR-BGT-006, BR-BGT-007)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Ngân sách)
- **Last updated:** 2026-07-13

## Actor
- Hệ thống (kích hoạt bởi thay đổi giao dịch của một thành viên); người nhận cảnh báo là **mọi thành viên trong hộ**.

## Trigger (Khi nào use case bắt đầu)
- Một giao dịch Chi được thêm/sửa/xóa (UC-TRK-02/04/05) làm tiến độ một ngân sách đạt/vượt ngưỡng 80% hoặc vượt giới hạn (100%).

## Preconditions
- Hộ có ít nhất một ngân sách đang hoạt động (UC-BGT-01 hoặc UC-BGT-02).
- Tiến độ ngân sách được tính lại chính xác (UC-BGT-03).

## Main Flow
1. Một thành viên thực hiện giao dịch làm thay đổi tiến độ một ngân sách.
2. Hệ thống tính lại tiến độ của ngân sách liên quan (UC-BGT-03).
3. Nếu tiến độ đạt/vượt **80%** giới hạn (nhưng chưa quá 100%) và mức 80% **chưa được phát trong kỳ hiện tại**, hệ thống phát cảnh báo "đạt ngưỡng 80%" trong ứng dụng cho **mọi thành viên** của hộ.
4. Nếu tiến độ **vượt giới hạn (trên 100%)** và mức vượt **chưa được phát trong kỳ hiện tại**, hệ thống phát cảnh báo "vượt ngân sách" kèm **đúng số tiền vượt** cho mọi thành viên của hộ.
5. Hệ thống ghi nhận mỗi mức cảnh báo đã phát trong kỳ để chống phát trùng — mục tiêu đạt được.

## Alternative Flows
- **3a. Mức cảnh báo đã phát và tiến độ vẫn trên mức (bước 3/4):** Có thêm giao dịch nhưng tiến độ vẫn ở trên mức đã cảnh báo — hệ thống **không** phát lại cảnh báo cùng mức. Use case tiếp tục ở bước 5.
- **4a. Tiến độ tụt xuống dưới mức rồi vượt lên lại (bước 4):** Do sửa/xóa giao dịch, tiến độ tụt xuống dưới mức đã cảnh báo; hệ thống đặt lại trạng thái mức đó; khi một giao dịch mới đẩy tiến độ vượt mức lần nữa, cảnh báo cùng mức được **phát lại**. Use case tiếp tục ở bước 3.

## Exceptions
- **E1. Ngân sách một lần đã kết thúc kỳ (bước 2):** Không tính tiến độ mới và không phát cảnh báo cho ngân sách đã kết thúc. Use case kết thúc.
- **E2. Sang kỳ mới (bước 2):** Trạng thái cảnh báo của mọi mức được đặt lại về chưa phát khi bước sang kỳ tháng/tuần mới. Use case tiếp tục ở bước 3.

## Postconditions
- **Thành công:** Mọi lần tiến độ chạm ngưỡng 80% hoặc vượt giới hạn đều có đúng một cảnh báo tương ứng hiển thị cho mọi thành viên của hộ; không có cảnh báo trùng cùng mức trong một kỳ khi tiến độ duy trì trên mức (SC-004); cảnh báo vượt hiển thị đúng số tiền vượt (SC-005).
- **Thất bại:** Không có cảnh báo sai (không phát khi chưa chạm mức, không phát trùng); trạng thái ngân sách không đổi.

## Acceptance Criteria
### AC-1: Cảnh báo đạt ngưỡng 80%
Given: ngân sách với tiến độ dưới 80%
When: một giao dịch mới đẩy tiến độ đạt/vượt 80% (nhưng chưa quá 100%)
Then: cảnh báo "đạt ngưỡng 80%" hiển thị trong ứng dụng cho mọi thành viên của hộ
### AC-2: Cảnh báo vượt ngân sách kèm số tiền vượt
Given: ngân sách với tiến độ dưới 100%
When: một giao dịch đẩy tổng chi vượt giới hạn
Then: cảnh báo "vượt ngân sách" hiển thị kèm đúng số tiền vượt
### AC-3: Không phát lại cùng mức trong kỳ
Given: cảnh báo 80% đã phát trong kỳ và tiến độ vẫn trên 80%
When: có thêm giao dịch
Then: không phát lại cảnh báo 80%
### AC-4: Phát lại sau khi tụt xuống dưới mức
Given: tiến độ đã tụt xuống dưới 80% do một giao dịch bị xóa
When: giao dịch mới đẩy tiến độ vượt 80% lần nữa
Then: cảnh báo 80% phát lại
### AC-5: Cảnh báo tới mọi thành viên
Given: Bob gây ra giao dịch làm vượt ngưỡng
When: cảnh báo phát
Then: Alice (cùng hộ) cũng thấy cảnh báo đó

## Dependencies
- **Upstream UC:** UC-BGT-03 (tiến độ được tính lại); kích hoạt bởi UC-TRK-02/04/05.
- **Downstream UC:** UC-BGT-05 (sửa giới hạn làm tính lại trạng thái cảnh báo).
- **External Systems:** Không (MVP chỉ cảnh báo in-app; push/email để sau — Assumptions spec 003).

## Notes
- Truy vết: **FR** FR-007, FR-008, FR-009 ([spec 003](../../003-budgeting/spec.md)) · **Acceptance** US3 #1–#5 · **SC** SC-004, SC-005, SC-006.
- Ngưỡng cố định **80%** ở MVP; cho tự cấu hình là mở rộng sau (Assumptions spec 003).
- Kênh cảnh báo: **trong ứng dụng (in-app)** cho mọi thành viên; đồng bộ realtime ≤ 5 giây (SC-006), dùng chung cơ chế WebSocket feature 001/002.
- Cảnh báo ngân sách (Budget Alert) ghi nhận: ngân sách, mức (80% / vượt), số tiền vượt (nếu có), thời điểm phát, kỳ — dùng để hiển thị và chống phát trùng (FR-009; Key Entities spec 003).

## Link file specs change
- [spec 003 §US3, FR-007/008/009](../../003-budgeting/spec.md)

## History
- v1 (2026-07-13, claude): initial từ US3 của spec 003.
