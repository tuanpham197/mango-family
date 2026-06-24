# Phân Rã Chức Năng Cốt Lõi và Yêu Cầu Nghiệp Vụ (Business Requirements) cho App Quản Lý Tài Chính Cơ Bản

Để phát triển một ứng dụng quản lý tài chính hiệu quả, việc phân rã các chức năng cốt lõi thành các tác vụ nhỏ hơn và định nghĩa rõ ràng các yêu cầu nghiệp vụ (Business Requirements - BR) là vô cùng quan trọng. Điều này giúp đảm bảo rằng mọi thành phần của ứng dụng đều đáp ứng được mục tiêu kinh doanh và nhu cầu của người dùng. Tài liệu này sẽ trình bày chi tiết các chức năng cốt lõi đã được xác định và các BR tương ứng cho từng tác vụ.

## 1. Chức Năng: Ghi Chép Thu Nhập và Chi Phí (Income & Expense Tracking)

Đây là trái tim của mọi ứng dụng quản lý tài chính, cho phép người dùng ghi lại mọi giao dịch tiền tệ.

### 1.1. Tác Vụ: Nhập Liệu Giao Dịch Thủ Công (Manual Transaction Entry)

Người dùng cần một cách nhanh chóng và dễ dàng để ghi lại các khoản thu và chi.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-TRK-001:** Người dùng phải có khả năng nhập số tiền của giao dịch. | Số tiền phải là một giá trị số dương. |
| **BR-TRK-002:** Người dùng phải có khả năng chọn loại giao dịch. | Các lựa chọn bao gồm: "Thu nhập" hoặc "Chi phí". |
| **BR-TRK-003:** Người dùng phải có khả năng chọn một danh mục cho giao dịch. | Danh mục phải được chọn từ danh sách các danh mục đã có hoặc danh mục mới được tạo. |
| **BR-TRK-004:** Người dùng phải có khả năng thêm mô tả ngắn gọn cho giao dịch. | Mô tả là tùy chọn, tối đa 255 ký tự. |
| **BR-TRK-005:** Người dùng phải có khả năng chọn ngày và giờ thực hiện giao dịch. | Mặc định là thời gian hiện tại. |
| **BR-TRK-006:** Người dùng phải có khả năng chọn tài khoản liên quan đến giao dịch. | Ví dụ: "Tiền mặt", "Tài khoản ngân hàng A", "Ví điện tử B". |
| **BR-TRK-007:** Hệ thống phải xác thực dữ liệu nhập vào. | Đảm bảo số tiền là hợp lệ, danh mục và tài khoản được chọn. |

### 1.2. Tác Vụ: Chỉnh Sửa/Xóa Giao Dịch (Edit/Delete Transaction)

Người dùng cần linh hoạt để sửa lỗi hoặc loại bỏ các giao dịch không chính xác.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-TRK-008:** Người dùng phải có khả năng chỉnh sửa tất cả thông tin của một giao dịch đã nhập. | Bao gồm số tiền, loại, danh mục, mô tả, ngày/giờ, và tài khoản. |
| **BR-TRK-009:** Người dùng phải có khả năng xóa một giao dịch đã nhập. | Hệ thống phải hiển thị cảnh báo xác nhận trước khi xóa vĩnh viễn. |
| **BR-TRK-010:** Sau khi chỉnh sửa/xóa, số dư tài khoản liên quan phải được cập nhật tự động. | Đảm bảo tính nhất quán của dữ liệu tài chính. |

## 2. Chức Năng: Phân Loại Giao Dịch (Categorization)

Giúp người dùng tổ chức và hiểu rõ hơn về dòng tiền của mình.

### 2.1. Tác Vụ: Quản Lý Danh Mục (Category Management)

Cung cấp sự linh hoạt cho người dùng trong việc định nghĩa các danh mục chi tiêu/thu nhập.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-CAT-001:** Người dùng phải có khả năng xem danh sách các danh mục mặc định. | Ví dụ: "Ăn uống", "Di chuyển", "Hóa đơn", "Lương". |
| **BR-CAT-002:** Người dùng phải có khả năng thêm danh mục mới. | Cho phép đặt tên và chọn biểu tượng (tùy chọn) cho danh mục. |
| **BR-CAT-003:** Người dùng phải có khả năng chỉnh sửa tên và biểu tượng của danh mục hiện có. | Áp dụng cho cả danh mục mặc định và danh mục tự tạo. |
| **BR-CAT-004:** Người dùng phải có khả năng xóa danh mục. | Nếu danh mục có giao dịch liên quan, hệ thống phải cảnh báo và yêu cầu người dùng gán lại giao dịch đó vào danh mục khác hoặc xóa các giao dịch đó. |
| **BR-CAT-005:** Hệ thống phải hỗ trợ danh mục con (sub-categories). | Ví dụ: "Ăn uống" có thể có "Ăn ngoài", "Đi chợ". |

### 2.2. Tác Vụ: Gán Danh Mục Cho Giao Dịch (Assign Category to Transaction)

Đảm bảo mọi giao dịch đều được phân loại để phục vụ báo cáo.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-CAT-006:** Khi nhập giao dịch, người dùng phải chọn một danh mục. | Không cho phép lưu giao dịch nếu chưa chọn danh mục. |
| **BR-CAT-007:** Hệ thống có thể gợi ý danh mục dựa trên mô tả giao dịch hoặc lịch sử. | Ví dụ: nếu mô tả là "Grab", gợi ý danh mục "Di chuyển". |

## 3. Chức Năng: Thiết Lập Ngân Sách (Budgeting)

Giúp người dùng kiểm soát chi tiêu và đạt được mục tiêu tài chính.

### 3.1. Tác Vụ: Tạo Ngân Sách (Create Budget)

Cho phép người dùng định nghĩa các giới hạn chi tiêu.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-BGT-001:** Người dùng phải có khả năng tạo ngân sách cho một danh mục cụ thể. | Ví dụ: "Ngân sách ăn uống tháng 7: 5.000.000 VNĐ". |
| **BR-BGT-002:** Người dùng phải có khả năng tạo ngân sách cho tổng chi tiêu trong một khoảng thời gian. | Ví dụ: "Tổng chi tiêu tháng 7: 15.000.000 VNĐ". |
| **BR-BGT-003:** Người dùng phải có khả năng đặt số tiền giới hạn cho ngân sách. | Số tiền phải là giá trị số dương. |
| **BR-BGT-004:** Người dùng phải có khả năng chọn khoảng thời gian áp dụng ngân sách. | Ví dụ: Hàng tháng, hàng tuần, một lần. |

### 3.2. Tác Vụ: Theo Dõi Ngân Sách (Budget Tracking)

Cung cấp cái nhìn tổng quan về tình hình chi tiêu so với ngân sách.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-BGT-005:** Hệ thống phải hiển thị tiến độ chi tiêu so với ngân sách đã đặt. | Ví dụ: "Đã chi 3.500.000/5.000.000 VNĐ (70%)". |
| **BR-BGT-006:** Hệ thống phải cảnh báo người dùng khi chi tiêu đạt đến một ngưỡng nhất định. | Ví dụ: Cảnh báo khi đạt 80% ngân sách. |
| **BR-BGT-007:** Hệ thống phải cảnh báo người dùng khi chi tiêu vượt quá ngân sách. | Hiển thị rõ ràng số tiền đã vượt. |

## 4. Chức Năng: Báo Cáo và Phân Tích Trực Quan (Visualized Reports)

Biến dữ liệu thô thành thông tin hữu ích, dễ hiểu để người dùng đưa ra quyết định tài chính tốt hơn.

### 4.1. Tác Vụ: Xem Báo Cáo Tổng Quan (Overview Report)

Cung cấp cái nhìn nhanh về tình hình tài chính chung.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-RPT-001:** Hệ thống phải hiển thị tổng thu nhập, tổng chi phí và số dư ròng trong một khoảng thời gian. | Người dùng có thể chọn khoảng thời gian (tuần, tháng, quý, năm, tùy chỉnh). |
| **BR-RPT-002:** Hệ thống phải hiển thị biểu đồ phân bổ chi tiêu theo danh mục. | Sử dụng biểu đồ tròn (pie chart) hoặc biểu đồ cột (bar chart) để trực quan hóa. |
| **BR-RPT-003:** Hệ thống phải hiển thị biểu đồ xu hướng thu nhập và chi phí theo thời gian. | Sử dụng biểu đồ đường (line chart) để thể hiện sự thay đổi. |

### 4.2. Tác Vụ: Xem Báo Cáo Chi Tiết Theo Danh Mục (Detailed Category Report)

Cho phép người dùng đào sâu vào từng khoản mục chi tiêu cụ thể.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-RPT-004:** Người dùng phải có khả năng xem danh sách tất cả giao dịch trong một danh mục cụ thể. | Hiển thị chi tiết từng giao dịch (số tiền, mô tả, ngày). |
| **BR-RPT-005:** Hệ thống phải hiển thị tổng chi tiêu của một danh mục trong một khoảng thời gian. | Ví dụ: "Tổng chi cho Ăn uống tháng 7: 4.800.000 VNĐ". |
| **BR-RPT-006:** Hệ thống phải hiển thị biểu đồ xu hướng chi tiêu của một danh mục theo thời gian. | Giúp người dùng nhận biết sự thay đổi trong thói quen chi tiêu. |

## 5. Chức Năng: Quản Lý Tài Khoản/Ví (Account Management)

Cho phép người dùng quản lý các nguồn tiền khác nhau của mình.

### 5.1. Tác Vụ: Thêm/Chỉnh Sửa/Xóa Tài Khoản (Add/Edit/Delete Account)

Cung cấp khả năng quản lý các nguồn tiền đa dạng.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-ACC-001:** Người dùng phải có khả năng thêm tài khoản mới. | Các loại tài khoản: Tiền mặt, Tài khoản ngân hàng, Ví điện tử, Thẻ tín dụng. |
| **BR-ACC-002:** Người dùng phải có khả năng đặt tên và số dư ban đầu cho tài khoản. | Số dư ban đầu có thể là 0 hoặc một giá trị dương. |
| **BR-ACC-003:** Người dùng phải có khả năng chỉnh sửa tên và số dư hiện tại của tài khoản. | Việc chỉnh sửa số dư phải được ghi nhận là một giao dịch điều chỉnh. |
| **BR-ACC-004:** Người dùng phải có khả năng xóa tài khoản. | Nếu tài khoản có giao dịch liên quan, hệ thống phải cảnh báo và yêu cầu người dùng chuyển các giao dịch đó sang tài khoản khác hoặc xóa chúng. |

### 5.2. Tác Vụ: Chuyển Tiền Giữa Các Tài Khoản (Transfer between Accounts)

Ghi nhận sự luân chuyển tiền giữa các nguồn khác nhau mà không ảnh hưởng đến tổng tài sản ròng.

| Yêu Cầu Nghiệp Vụ (BR) | Mô Tả Chi Tiết |
| :--- | :--- |
| **BR-ACC-005:** Người dùng phải có khả năng ghi nhận việc chuyển tiền từ tài khoản này sang tài khoản khác. | Ví dụ: Chuyển 1.000.000 VNĐ từ "Tài khoản ngân hàng A" sang "Tiền mặt". |
| **BR-ACC-006:** Hệ thống phải tự động cập nhật số dư của cả hai tài khoản liên quan. | Số dư tài khoản nguồn giảm, số dư tài khoản đích tăng. |
| **BR-ACC-007:** Giao dịch chuyển tiền không được tính vào tổng thu nhập hoặc chi phí. | Đây là giao dịch nội bộ, không làm thay đổi tổng tài sản ròng. |

---

**Tài Liệu Tham Khảo**
[1] [Personal Finance App Development: features, benefits & costs](https://www.purrweb.com/blog/personal-finance-app-development-features-benefits-costs/)
[2] [Creating a Personal Finance App That People Will Want To Use](https://medium.com/@Shakuro/creating-a-personal-finance-app-that-people-will-want-to-use-67af415e8f55)
[3] [How to Build a Personal Finance App: Steps, Requirements & Features](https://artkai.io/blog/finance-app-development-ultimate-guide)
