# UC-MBR-01: Xem báo cáo thu/chi theo từng thành viên

## Metadata
- **ID:** UC-MBR-01
- **Bounded Context:** Reports (báo cáo chỉ đọc, suy ra từ giao dịch của hộ)
- **Liên quan tới BR:** BR-008 (Goal; BR-MBR-001…005, BR-MBR-008, BR-MBR-009); mở rộng báo cáo BR-006 / feature 005-reports
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Báo cáo)
- **Last updated:** 2026-08-10

## Actor
- Thành viên hộ gia đình (quyền ngang nhau).

## Trigger (Khi nào use case bắt đầu)
- Thành viên mở màn Báo cáo và xem phần tổng hợp **theo thành viên** cho một khoảng thời gian.

## Preconditions
- Thành viên đã đăng nhập và thuộc một hộ (UC-TRK-01).
- Màn Báo cáo tổng quan của hộ đã sẵn sàng (feature 005-reports).

## Main Flow
1. Thành viên mở báo cáo và chọn khoảng thời gian (kế thừa bộ lọc của báo cáo hiện tại: tuần, tháng, quý, năm hoặc khoảng tùy chỉnh) *(BR-MBR-002)*.
2. Hệ thống tổng hợp giao dịch **của hộ mình** trong khoảng đã chọn và nhóm theo thành viên đã nhập giao dịch, gán trọn giá trị mỗi giao dịch cho đúng một thành viên *(BR-MBR-003)*.
3. Hệ thống hiển thị danh sách **tất cả thành viên hiện tại** của hộ; mỗi dòng gồm tên hiển thị, **tổng Thu**, **tổng Chi** và **ròng = Thu − Chi** trong khoảng *(BR-MBR-001)*.
4. Với thành viên không có giao dịch trong khoảng, hệ thống hiển thị tổng Thu = 0, tổng Chi = 0 và ròng = 0 *(BR-MBR-004)*.
5. Hệ thống bảo đảm tổng Thu và tổng Chi khi cộng mọi dòng thành viên khớp với tổng Thu và tổng Chi toàn hộ trên báo cáo hiện tại trong cùng khoảng *(BR-MBR-005)*.
6. Thành viên đổi khoảng thời gian; hệ thống tính lại toàn bộ số liệu theo thành viên theo cùng ranh giới thời gian với tổng hợp toàn hộ *(BR-MBR-002)*.
7. Thành viên chọn một thành viên để xem các giao dịch cấu thành số liệu (UC-MBR-02) — mục tiêu (so sánh đóng góp thu/chi của từng thành viên) đạt được.

## Alternative Flows
- **1a. Chưa chọn khoảng (bước 1):** Hệ thống dùng khoảng mặc định của báo cáo hiện tại. Use case tiếp tục ở bước 2.
- **3a. Người nhập chưa có tên hiển thị (bước 3):** Hệ thống hiển thị email thay cho tên; không bao giờ hiển thị mã định danh thô *(BR-MBR-009)*. Use case tiếp tục ở bước 4.
- **3b. Giao dịch do người đã rời hộ nhập (bước 3):** Hệ thống gộp toàn bộ giá trị vào một dòng bổ sung **"Thành viên cũ / Đã rời hộ"** (chỉ hiện khi có dữ liệu như vậy) để bất biến đối soát ở bước 5 luôn đúng; dòng này drill-down được như một thành viên (UC-MBR-02). *(chốt Option A — FR-010 / BR-008 #1)*. Use case tiếp tục ở bước 5.

## Exceptions
- **E1. Hộ không có giao dịch nào trong khoảng (bước 2):** Hệ thống hiển thị mọi thành viên hiện tại với 0/0/0 (hoặc trạng thái rỗng của báo cáo). Use case kết thúc.

## Postconditions
- **Thành công:** Thành viên thấy tổng Thu/Chi/ròng của từng thành viên hiện tại trong hộ cho khoảng đã chọn; tổng các dòng khớp tổng toàn hộ; không thấy dữ liệu của hộ khác.
- **Thất bại:** Không dữ liệu nào của hộ khác bị lộ; không hiển thị số liệu suy ra sai lệch.

## Acceptance Criteria
### AC-1: Tổng hợp theo từng thành viên
Given: trong hộ A, Alice đã nhập vài giao dịch và Bob đã nhập vài giao dịch trong tháng này
When: một thành viên hộ A mở báo cáo theo thành viên cho tháng này
Then: mỗi thành viên hiển thị một dòng với tổng Thu, tổng Chi và ròng = Thu − Chi
And: giao dịch Alice nhập được tính cho Alice, giao dịch Bob nhập được tính cho Bob
### AC-2: Đối soát với tổng toàn hộ
Given: báo cáo theo thành viên cho một khoảng
When: cộng tổng Thu và tổng Chi của mọi dòng thành viên
Then: kết quả bằng đúng tổng Thu và tổng Chi toàn hộ trên báo cáo tổng quan cùng khoảng
### AC-3: Thành viên không có giao dịch
Given: Carol thuộc hộ A nhưng chưa nhập giao dịch nào trong khoảng
When: mở báo cáo theo thành viên
Then: Carol vẫn xuất hiện với tổng Thu = 0, tổng Chi = 0 và ròng = 0
### AC-4: Đổi khoảng thời gian tính lại
Given: đang xem báo cáo theo thành viên cho tháng này
When: đổi sang quý này
Then: mọi số liệu theo thành viên được tính lại theo cùng ranh giới thời gian với tổng hợp toàn hộ
### AC-5: Cô lập giữa các hộ
Given: Dave thuộc hộ B
When: Dave mở báo cáo theo thành viên
Then: Dave chỉ thấy thành viên và số liệu của hộ B, không thấy bất kỳ dữ liệu nào của hộ A

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-TRK-02 (có giao dịch với `created_by` để tổng hợp); báo cáo tổng quan feature 005-reports (bộ lọc thời gian & tổng toàn hộ).
- **Downstream UC:** UC-MBR-02 (drill-down danh sách giao dịch của một thành viên) «extend» từ báo cáo này.
- **External Systems:** Không.

## Notes
- Truy vết: **BR-008** BR-MBR-001…005, BR-MBR-008 (cô lập hộ, xác thực thành viên thuộc hộ), BR-MBR-009 (tên hiển thị/fallback email) · **Nguồn dữ liệu** `transactions.created_by` (BR-002 / entity-model: `TRANSACTION.created_by → USER.id`, tư cách thành viên qua `HOUSEHOLD_MEMBER`).
- "Thu/chi của thành viên" = thành viên **đã nhập** giao dịch (`created_by`), không khẳng định người thực tế nhận/chi tiền (BR-008 Background & Out of Scope).
- Báo cáo **chỉ đọc**, suy ra theo yêu cầu từ giao dịch trong `household_id` + khoảng; không bảng tổng hợp mới (BR-008 Constraints).
- **Giao dịch của người đã rời hộ (chốt Option A, spec 008 FR-010)**: gộp vào dòng "Thành viên cũ / Đã rời hộ" để đối soát tổng hộ (BR-MBR-005) luôn đúng; giải quyết BR-008 Open Question #1.
- **Quyết định cấu trúc:** UC đặt tại `specs/use-cases/008-member-reports/` (prefix `MBR` theo BR-008); feature spec dự kiến mở rộng `specs/005-reports/` hoặc feature kế tiếp khi lập kế hoạch — chờ chốt.

## Link file specs change
- [BR-008](../../business-requirements/BR-008.md) · README chỉ mục: [`../README.md`](../README.md) · sơ đồ: [`../../diagrams/use-cases.puml`](../../diagrams/use-cases.puml)

## History
- v1 (2026-08-10, claude): initial — dẫn xuất từ BR-008 (UC-MBR-01); tổng Thu/Chi/ròng theo từng thành viên hiện tại, đối soát tổng hộ, thành viên không giao dịch = 0, đổi khoảng tính lại; ghi nhận open question thành viên đã rời hộ.
