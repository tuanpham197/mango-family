# UC-TRK-01: Đăng nhập và định danh người dùng

## Metadata
- **ID:** UC-TRK-01
- **Bounded Context:** Identity (định danh người dùng & phạm vi hộ)
- **Liên quan tới BR:** BR-002 (tiền đề mọi thao tác); liên kết BR-CAT-009 (ghi người nhập)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Giao dịch)
- **Last updated:** 2026-07-10

## Actor
- Người dùng đã có tài khoản (trước khi được định danh là thành viên hộ).

## Trigger (Khi nào use case bắt đầu)
- Người dùng mở ứng dụng khi chưa có phiên đăng nhập.

## Preconditions
- Người dùng có một tài khoản hợp lệ (email + mật khẩu).
- Danh sách người dùng của hệ thống đã được dựng (FR-015 — bước nền tảng, làm ĐẦU TIÊN).

## Main Flow
1. Người dùng mở ứng dụng.
2. Hệ thống hiển thị màn hình đăng nhập (không hiển thị bất kỳ dữ liệu hộ nào).
3. Người dùng nhập email và mật khẩu rồi xác nhận đăng nhập.
4. Hệ thống xác thực thông tin đăng nhập.
5. Hệ thống đối chiếu phiên đăng nhập với **hồ sơ người dùng** (tên hiển thị) và xác định **hộ** mà người dùng là thành viên.
6. Hệ thống đưa người dùng vào sổ của hộ mình; từ đây mọi bản ghi mới (giao dịch, danh mục) đều ghi đúng người tạo theo định danh người dùng — mục tiêu đạt được.

## Alternative Flows
- **2a. Đã có phiên đăng nhập từ trước (bước 2):** Hệ thống bỏ qua màn đăng nhập. Use case tiếp tục ở bước 6.
- **5a. Tài khoản chưa có hồ sơ tên hiển thị (bước 5):** Hệ thống tự bổ sung hồ sơ từ email; trong lúc chưa có tên, giao diện hiển thị email thay cho tên (không bao giờ hiển thị mã định danh thô). Use case tiếp tục ở bước 6.

## Exceptions
- **E1. Sai email hoặc mật khẩu (bước 4):** Hệ thống từ chối với thông báo rõ ràng, không tiết lộ thông tin nhạy cảm (không nói rõ sai trường nào). Use case tiếp tục ở bước 3.
- **E2. Người dùng chưa thuộc hộ nào (bước 5):** Hệ thống thông báo cần được thêm vào một hộ (tiền đề Quản lý hộ — feature riêng) và không hiển thị dữ liệu. Use case kết thúc.

## Postconditions
- **Thành công:** Phiên đăng nhập hoạt động; định danh người dùng và hộ hiện tại được xác định; mọi thao tác tiếp theo chạy trong phạm vi hộ đó.
- **Thất bại:** Không có phiên đăng nhập; không dữ liệu nào của bất kỳ hộ nào bị lộ.

## Acceptance Criteria
### AC-1: Đăng nhập thành công vào đúng hộ
Given: một tài khoản người dùng hợp lệ và là thành viên hộ A
When: đăng nhập đúng email và mật khẩu
Then: người dùng vào được sổ của hộ A
And: hồ sơ hiển thị đúng tên của người dùng
### AC-2: Chưa đăng nhập thì không thấy dữ liệu
Given: người dùng chưa đăng nhập
When: mở ứng dụng
Then: được đưa về màn đăng nhập
And: không dữ liệu nào của hộ được hiển thị
### AC-3: Sai mật khẩu bị từ chối an toàn
Given: người dùng nhập sai mật khẩu
When: thử đăng nhập
Then: bị từ chối kèm thông báo rõ ràng
And: thông báo không tiết lộ thông tin nhạy cảm
### AC-4: Bản ghi mới ghi đúng người tạo
Given: Bob đã đăng nhập
When: Bob nhập một giao dịch
Then: sổ của mọi thành viên hiển thị giao dịch với người nhập là "Bob" (tên hiển thị, không phải mã định danh)

## Dependencies
- **Upstream UC:** UC-HH-* (Quản lý hộ — người dùng phải đã là thành viên một hộ; feature riêng, chưa có BR).
- **Downstream UC:** UC-TRK-02…05 và toàn bộ UC-CAT-01…08 (mọi use case đều yêu cầu phiên đăng nhập + định danh).
- **External Systems:** Không — xác thực (mật khẩu/phiên) là chức năng của hệ thống; chi tiết quản lý không thuộc ý nghĩa nghiệp vụ của thực thể Người dùng.

## Notes
- Danh sách người dùng là **nguồn định danh duy nhất** (FR-015): thành viên hộ và mọi trường "người nhập/người tạo" tham chiếu về đây. Cách nối phiên đăng nhập với hồ sơ người dùng là **chi tiết hiện thực** (đã đổi theo re-platform 2026-07-10; trước đây đối chiếu qua email trên stack cũ).
- Truy vết: **FR** FR-015, FR-016 ([spec 002](../../002-transaction-tracking/spec.md)) · **Acceptance** US5 #1–#4 · **SC** SC-006 (gián tiếp).
- Vòng đời tài khoản đầy đủ (tự đăng ký, quên mật khẩu, đổi email) ngoài phạm vi — BR riêng.

## History
- v1 (2026-07-06, claude): initial từ US5 của spec 002.
- v2 (2026-07-06, claude): chuyển sang cấu trúc `specs/use-cases/template.md` (Metadata/Exceptions/AC/Dependencies/History).
- v3 (2026-07-10, claude): re-platform Go+Vue — gỡ mô tả cơ chế đối chiếu email (chi tiết hiện thực của stack cũ) khỏi Notes/External Systems; nghiệp vụ không đổi.
