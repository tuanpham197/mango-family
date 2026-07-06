# UC-TRK-02: Nhập giao dịch thu/chi

## Metadata
- **ID:** UC-TRK-02
- **Bounded Context:** Transaction Tracking (ghi chép thu chi của hộ)
- **Liên quan tới BR:** BR-002 (BR-TRK-001…007, BR-TRK-010)
- **Status:** draft
- **Owner:** TBD (PO phụ trách phân hệ Giao dịch)
- **Last updated:** 2026-07-06

## Actor
- Thành viên hộ gia đình (quyền ngang nhau).

## Trigger (Khi nào use case bắt đầu)
- Thành viên mở màn hình nhập giao dịch.

## Preconditions
- Thành viên đã đăng nhập và thuộc một hộ (UC-TRK-01).
- Hộ có ít nhất một tài khoản để ghi giao dịch (tài khoản mặc định — phụ thuộc BR-005).
- Đã có danh mục phân theo Thu/Chi (BR-001 / feature 001).

## Main Flow
1. Thành viên mở màn hình nhập giao dịch.
2. Hệ thống hiển thị form với **ngày giờ mặc định là hiện tại** và **tài khoản của hộ được chọn sẵn** (khi hộ chỉ có một tài khoản).
3. Thành viên nhập số tiền (lớn hơn 0).
4. Thành viên chọn loại **Thu** hoặc **Chi**.
5. Thành viên gán danh mục cùng loại cho giao dịch (**UC-CAT-07** — danh sách chỉ gồm danh mục cùng loại).
6. Thành viên nhập mô tả (tùy chọn, tối đa 255 ký tự).
7. Thành viên bấm **Lưu**.
8. Hệ thống xác thực toàn bộ dữ liệu (số tiền hợp lệ; loại, danh mục, tài khoản bắt buộc; mô tả trong giới hạn; ngày không ở tương lai).
9. Hệ thống lưu giao dịch kèm **người nhập** (định danh người dùng), cập nhật **số dư tài khoản** liên quan; giao dịch xuất hiện trong sổ của mọi thành viên trong hộ — mục tiêu đạt được.

## Alternative Flows
- **5a. Có gợi ý danh mục (bước 5):** Hệ thống đề xuất một danh mục theo mô tả/lịch sử chung của hộ (UC-CAT-08); thành viên chấp nhận hoặc chọn danh mục khác. Use case tiếp tục ở bước 6.
- **6a. Chỉnh ngày giờ (bước 6):** Thành viên đổi ngày giờ về một thời điểm trong quá khứ trước khi lưu. Use case tiếp tục ở bước 7.

## Exceptions
- **E1. Số tiền không hợp lệ (bước 8):** 0, số âm hoặc không phải số — hệ thống chặn lưu, thông báo "số tiền phải lớn hơn 0" ngay tại trường số tiền. Use case tiếp tục ở bước 3.
- **E2. Chưa chọn danh mục (bước 8):** Hệ thống chặn lưu và yêu cầu chọn danh mục (luồng 5a của UC-CAT-07). Use case tiếp tục ở bước 5.
- **E3. Mô tả vượt 255 ký tự (bước 8):** Hệ thống chặn/giới hạn ngay khi nhập kèm thông báo giới hạn. Use case tiếp tục ở bước 6.
- **E4. Ngày ở tương lai (bước 8):** Hệ thống từ chối (giao dịch dự kiến ngoài phạm vi) kèm thông báo. Use case tiếp tục ở bước 6.
- **E5. Mất kết nối khi lưu (bước 9):** Hệ thống thông báo rõ ràng; khi thử lại không tạo bản ghi trùng. Use case tiếp tục ở bước 7.

## Postconditions
- **Thành công:** Giao dịch tồn tại trong sổ chung với đủ trường bắt buộc, đúng người nhập; số dư tài khoản phản ánh đúng tổng bút toán (SC-003, SC-004).
- **Thất bại:** Không giao dịch nào được lưu; số dư không đổi.

## Acceptance Criteria
### AC-1: Nhập giao dịch Chi hợp lệ
Given: thành viên đang nhập giao dịch Chi 50.000 và đã chọn danh mục Chi hợp lệ
When: bấm Lưu
Then: giao dịch được lưu và xuất hiện trong sổ chung
And: số dư tài khoản liên quan giảm 50.000
### AC-2: Chặn số tiền không hợp lệ
Given: thành viên nhập số tiền bằng 0 hoặc số âm
When: bấm Lưu
Then: hệ thống chặn và thông báo số tiền phải lớn hơn 0
### AC-3: Chặn khi thiếu danh mục
Given: thành viên chưa chọn danh mục
When: bấm Lưu
Then: hệ thống chặn và yêu cầu chọn danh mục (nhất quán BR-001)
### AC-4: Giới hạn mô tả
Given: thành viên nhập mô tả dài hơn 255 ký tự
When: bấm Lưu
Then: hệ thống chặn hoặc giới hạn ngay khi nhập, kèm thông báo giới hạn độ dài
### AC-5: Ngày mặc định & ngày quá khứ
Given: thành viên mở form nhập
When: không chỉnh ngày giờ
Then: giao dịch được ghi với ngày giờ hiện tại
And: khi chỉnh sang một ngày trong quá khứ, giao dịch ghi theo ngày đã chọn
### AC-6: Lọc danh mục theo loại
Given: giao dịch loại Thu
When: mở danh sách danh mục
Then: chỉ danh mục loại Thu hiển thị (liên kết BR-001/BR-CAT-008)

## Dependencies
- **Upstream UC:** UC-TRK-01 (đăng nhập/định danh); UC-CAT-07 «include» (gán danh mục); UC-CAT-08 «extend» (gợi ý).
- **Downstream UC:** UC-TRK-03 (giao dịch hiển thị trong sổ); UC-TRK-04/05 (sửa/xóa giao dịch đã nhập).
- **External Systems:** Không.

## Notes
- Truy vết: **FR** FR-001…FR-007, FR-011…FR-013 ([spec 002](../../002-transaction-tracking/spec.md)) · **Acceptance** US1 #1–#6 · **SC** SC-001…SC-004, SC-006.
- Tài khoản là phụ thuộc BR-005 (chưa triển khai) — hộ có tài khoản mặc định; số dư là giá trị suy ra nhất quán (Assumptions spec 002).
- Mục tiêu hiệu năng: nhập xong ≤ 15 giây (SC-001).

## History
- v1 (2026-07-06, claude): initial từ US1 của spec 002.
- v2 (2026-07-06, claude): chuyển sang cấu trúc `specs/use-cases/template.md`; tách Exceptions khỏi Alternative Flows.
