# UC-XXX: <Tên use case ngắn gọn theo ngôn ngữ nghiệp vụ>
## Metadata
- **ID:** UC-XXX
- **Bounded Context:** <vd: Checkout, Shipping, Billing>
- **Liên quan tới BR:** BR-YYY
- **Status:** draft | reviewed | implemented | deprecated
- **Owner:** <PO/Dev chịu trách nhiệm>
- **Last updated:** YYYY-MM-DD
## Actor
- Ai khởi tạo use case này: người dùng, admin, system, scheduled job,
webhook...
## Trigger (Khi nào use case bắt đầu)
 - ....
## Preconditions
- <Điều kiện phải đúng trước khi use case chạy>
-
...
## Main Flow
1. <Bước 1>
2. <Bước 2>
3. ...
## Alternative Flows
- **<N>a. <Tên biến thể>:** <mô tả>
-
...
## Exceptions
- **E1. <Tên exception>:** <mô tả + cách hệ thống xử lý>
-
...
## Postconditions
- <Trạng thái hệ thống sau khi use case thành công>
## Acceptance Criteria
### AC-1: <Tên ngắn gọn>
Given: <context>
When: <action>
Then: <expected outcome>
And: <expected outcome bổ sung>
### AC-2: ...
## Dependencies
- **Upstream UC:** <UC khác phải xong trước>
- **Downstream UC:** <UC khác phụ thuộc vào UC này>
- **External Systems:** <vd: Stripe, Sendgrid, internal LDAP>
## Notes
 - Context bổ sung: quyết định lịch sử, lý do nghiệp vụ, link tới ADR, open
question
## History
- v1 (YYYY-MM-DD, <author>): initial
- v2 (YYYY-MM-DD, <author>): added AC-4 for idempotency
- v3 (YYYY-MM-DD, <author>): clarified exception E2