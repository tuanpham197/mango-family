# Use Case: Tạo danh mục mới

## Overview

**Use Case ID:** UC-CAT-02
**Use Case Name:** Tạo danh mục mới
**Primary Actor:** Người dùng
**Goal:** Tạo một danh mục Thu/Chi mới phù hợp với thói quen chi tiêu/thu nhập của người dùng để tăng tỷ lệ phân loại đúng khi nhập giao dịch.
**Status:** Draft

> **Tính năng:** Phân loại giao dịch ([spec.md](../001-transaction-categorization/spec.md)) · **User Story:** US2 (P2) · **Sơ đồ:** [use-cases.puml](../diagrams/use-cases.puml)

## Preconditions

- Người dùng đã đăng nhập.
- Người dùng đang ở màn hình quản lý danh mục, hoặc đang trong luồng nhập giao dịch và cần tạo nhanh một danh mục.

## Main Success Scenario

1. Người dùng chọn "Thêm danh mục" ở màn hình quản lý danh mục.
2. Hệ thống hiển thị biểu mẫu tạo danh mục.
3. Người dùng chọn loại danh mục — đúng một trong hai: Thu hoặc Chi.
4. Người dùng nhập tên danh mục.
5. Người dùng chọn biểu tượng cho danh mục (tùy chọn).
6. Người dùng xác nhận lưu.
7. Hệ thống kiểm tra dữ liệu: đã chọn loại và tên hợp lệ, không trùng trong cùng loại và cùng cấp cha.
8. Hệ thống tạo danh mục mới, gắn với người dùng và mang đúng loại đã chọn.
9. Hệ thống hiển thị danh mục mới trong nhóm loại tương ứng, sẵn sàng để chọn khi nhập giao dịch cùng loại.

## Alternative Flows

### A1: Tên danh mục bị trùng

**Trigger:** Tên nhập vào trùng với một danh mục khác trong cùng loại và cùng cấp cha (step 7)
**Flow:**

1. Hệ thống cảnh báo tên bị trùng để tránh nhầm lẫn.
2. Người dùng đổi sang tên khác, hoặc chấp nhận tiếp tục với tên hiện tại.
3. Use case continues at step 8.

### A2: Chưa chọn loại Thu/Chi

**Trigger:** Người dùng xác nhận lưu khi chưa chọn loại danh mục (step 7)
**Flow:**

1. Hệ thống chặn lưu và yêu cầu chọn loại Thu/Chi (loại là bắt buộc).
2. Người dùng chọn loại.
3. Use case continues at step 6.

### A3: Tạo nhanh từ luồng nhập giao dịch

**Trigger:** Người dùng tạo danh mục ngay trong luồng nhập giao dịch khi chưa có danh mục phù hợp (step 3)
**Flow:**

1. Hệ thống tự đặt loại danh mục mới bằng loại Thu/Chi của giao dịch đang nhập và bỏ qua bước chọn loại thủ công.
2. Use case continues at step 4.

### A4: Hủy tạo danh mục

**Trigger:** Người dùng đóng hoặc hủy biểu mẫu trước khi lưu (step 6)
**Flow:**

1. Hệ thống loại bỏ dữ liệu đang nhập và đóng biểu mẫu.
2. Use case ends.

## Postconditions

### Success Postconditions

- Một danh mục mới tồn tại, thuộc đúng một người dùng và mang đúng một loại Thu/Chi cố định.
- Danh mục mới xuất hiện trong nhóm loại tương ứng và có thể được chọn khi nhập giao dịch cùng loại.
- Nếu được tạo nhanh từ luồng nhập giao dịch, danh mục mới được chọn ngay cho giao dịch đang nhập (liên kết UC-CAT-07).

### Failure Postconditions

- Không có danh mục mới nào được tạo.
- Hệ thống giữ người dùng ở biểu mẫu và hiển thị lý do (chưa chọn loại, hoặc tên không hợp lệ) để người dùng sửa; trạng thái danh mục hiện có không thay đổi.

## Business Rules

### BR-CAT-002: Cấu trúc danh mục mới

Người dùng tạo danh mục mới bằng cách chọn đúng một loại (Thu hoặc Chi), đặt tên và chọn biểu tượng tùy chọn. *(FR-003)*

### BR-CAT-008: Loại Thu/Chi cố định, đúng một loại

Mỗi danh mục mang đúng một thuộc tính loại cố định là Thu hoặc Chi; loại không thể thay đổi sau khi tạo — muốn chuyển loại, người dùng tạo danh mục mới ở loại kia và gán lại giao dịch. *(FR-004, FR-005)*

### BR-CAT-009: Cảnh báo trùng tên

Hệ thống cảnh báo khi tên danh mục trùng trong cùng một loại và cùng cấp cha; người dùng vẫn có thể xác nhận tiếp tục. *(FR-017)*

### BR-CAT-010: Danh mục riêng theo từng người dùng

Danh mục thuộc về riêng từng người dùng; không chia sẻ hay đồng bộ bộ danh mục giữa nhiều người dùng. *(FR-018)*

## Acceptance Criteria

Truy vết tới Acceptance Scenarios của US2 trong [spec.md](../001-transaction-categorization/spec.md).

- **AC1 (US2 #1) — Tạo và dùng được danh mục mới**: *Given* người dùng ở màn hình quản lý danh mục, *When* thêm danh mục mới, chọn loại Chi và đặt tên "Thú cưng", *Then* danh mục "Thú cưng" xuất hiện trong nhóm Chi và có thể được chọn khi nhập giao dịch Chi.
- **AC2 — Loại là bắt buộc**: *Given* người dùng chưa chọn loại Thu/Chi, *When* bấm Lưu, *Then* hệ thống chặn lưu và yêu cầu chọn loại. *(A2, FR-004)*
- **AC3 — Cảnh báo trùng tên**: *Given* tên trùng với danh mục khác trong cùng loại và cùng cấp cha, *When* người dùng lưu, *Then* hệ thống cảnh báo nhưng vẫn cho phép xác nhận tiếp tục. *(A1, FR-017)*
- **AC4 — Tạo nhanh kế thừa loại giao dịch**: *Given* người dùng tạo nhanh danh mục trong luồng nhập giao dịch loại Chi, *When* tạo xong, *Then* danh mục mới mang loại Chi và được chọn ngay cho giao dịch. *(A3, UC-CAT-07)*
- **AC5 — Hiệu năng (SC-005)**: Người dùng tạo được một danh mục mới trong tối đa 3 bước thao tác và dưới 30 giây.

---

## Truy vết

- **FR**: FR-003, FR-004, FR-005, FR-017, FR-018
- **Acceptance**: US2 #1
- **Success Criteria**: SC-005 (tạo danh mục ≤ 3 bước, < 30 giây)
- **BR (nghiệp vụ)**: BR-CAT-002, BR-CAT-008
- **Liên kết UC**: UC-CAT-03 «extend» (tạo danh mục con), UC-CAT-07 (gán danh mục cho giao dịch)
