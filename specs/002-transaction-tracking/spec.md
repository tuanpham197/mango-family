# Feature Specification: Ghi Chép Thu Nhập và Chi Phí (Income & Expense Tracking)

**Feature Branch**: `002-transaction-tracking`

**Created**: 2026-07-06

**Status**: Draft

**Input**: User description: "@specs/business-requirements/BR-002.md — Ghi chép thu nhập và chi phí: nhập, chỉnh sửa, xóa giao dịch thu/chi nhanh chóng, chính xác trong sổ chung hộ gia đình; số dư tài khoản liên quan luôn được cập nhật đúng."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Nhập giao dịch thu/chi nhanh và hợp lệ (Priority: P1)

Một thành viên trong hộ mở màn hình nhập giao dịch, nhập số tiền (lớn hơn 0), chọn loại Thu hoặc Chi, chọn một danh mục cùng loại (bắt buộc — theo BR-001), thêm mô tả tùy chọn (tối đa 255 ký tự); ngày giờ mặc định là thời điểm hiện tại và có thể sửa. Giao dịch được ghi vào tài khoản liên quan và số dư tài khoản đó được cập nhật ngay. Nếu thiếu hoặc sai bất kỳ trường bắt buộc nào, hệ thống chặn lưu và chỉ rõ lỗi.

**Why this priority**: Đây là "trái tim" của ứng dụng theo BR-002 — không có nhập liệu nhanh và chính xác thì ngân sách (BR-003) và báo cáo (BR-004) đều mất giá trị. Chỉ riêng story này đã là sản phẩm dùng được.

**Independent Test**: Nhập một giao dịch Chi hợp lệ → giao dịch xuất hiện trong sổ của hộ và số dư tài khoản giảm đúng số tiền; thử lưu với số tiền 0/âm/bỏ trống danh mục → bị chặn kèm thông báo rõ ràng.

**Acceptance Scenarios**:

1. **Given** thành viên đang nhập giao dịch Chi với số tiền 50.000 và đã chọn danh mục Chi hợp lệ, **When** bấm Lưu, **Then** giao dịch được lưu, xuất hiện trong sổ chung và số dư tài khoản liên quan giảm 50.000.
2. **Given** thành viên nhập số tiền bằng 0 hoặc số âm, **When** bấm Lưu, **Then** hệ thống chặn và thông báo số tiền phải lớn hơn 0.
3. **Given** thành viên chưa chọn danh mục, **When** bấm Lưu, **Then** hệ thống chặn và yêu cầu chọn danh mục (nhất quán BR-001).
4. **Given** thành viên nhập mô tả dài hơn 255 ký tự, **When** bấm Lưu, **Then** hệ thống chặn hoặc giới hạn ngay khi nhập, kèm thông báo giới hạn độ dài.
5. **Given** thành viên mở form nhập, **When** không chỉnh ngày giờ, **Then** giao dịch được ghi với ngày giờ hiện tại; **When** chỉnh sang một ngày trong quá khứ, **Then** giao dịch ghi theo ngày đã chọn.
6. **Given** giao dịch loại Thu, **When** mở danh sách danh mục, **Then** chỉ danh mục loại Thu hiển thị (liên kết BR-001/BR-CAT-008).

---

### User Story 2 - Xem sổ giao dịch chung của hộ (Priority: P2)

Mọi thành viên xem được danh sách giao dịch của hộ (mới nhất trước), mỗi dòng hiển thị số tiền, loại Thu/Chi, danh mục, ngày giờ và **thành viên đã nhập**. Đây là điểm vào để chỉnh sửa hoặc xóa một giao dịch.

**Why this priority**: Sổ chung minh bạch là giá trị cốt lõi của mô hình hộ gia đình (BR-001/BR-CAT-009); đồng thời là tiền đề thao tác cho sửa/xóa (US3, US4). Không có danh sách thì không chọn được giao dịch để sửa/xóa.

**Independent Test**: Thành viên A nhập một giao dịch → thành viên B (cùng hộ) mở sổ và thấy giao dịch đó kèm nhãn "do A nhập"; thành viên hộ khác không thấy.

**Acceptance Scenarios**:

1. **Given** Alice (hộ A) vừa nhập một giao dịch, **When** Bob (cùng hộ A) mở sổ giao dịch, **Then** Bob thấy giao dịch đó với đầy đủ số tiền, danh mục, ngày và người nhập là Alice.
2. **Given** sổ có nhiều giao dịch, **When** mở danh sách, **Then** giao dịch sắp xếp theo ngày giờ mới nhất trước.
3. **Given** Carol thuộc hộ B, **When** Carol mở sổ, **Then** Carol không thấy bất kỳ giao dịch nào của hộ A.

---

### User Story 3 - Chỉnh sửa giao dịch (Priority: P3)

Bất kỳ thành viên nào cũng có thể mở một giao dịch đã lưu (kể cả do thành viên khác nhập) và chỉnh sửa mọi thông tin: số tiền, loại, danh mục, mô tả, ngày giờ, tài khoản. Dữ liệu sửa phải qua cùng bộ xác thực như khi nhập mới; nếu đổi loại Thu/Chi thì phải chọn lại danh mục cùng loại mới. Số dư các tài khoản liên quan được tính lại đúng sau khi lưu.

**Why this priority**: Sai sót nhập liệu là thường xuyên; sửa được nhanh giúp dữ liệu chính xác. Phụ thuộc US1 (đã có giao dịch) và US2 (tìm được giao dịch).

**Independent Test**: Sửa số tiền một giao dịch từ 50.000 thành 80.000 → số dư tài khoản phản ánh chênh lệch 30.000; đổi loại từ Chi sang Thu → hệ thống buộc chọn lại danh mục loại Thu trước khi lưu.

**Acceptance Scenarios**:

1. **Given** một giao dịch Chi 50.000 đã lưu, **When** thành viên sửa số tiền thành 80.000 và lưu, **Then** số dư tài khoản liên quan được điều chỉnh thêm 30.000 chênh lệch (tổng thể luôn khớp).
2. **Given** một giao dịch loại Chi, **When** thành viên đổi loại sang Thu, **Then** danh mục cũ (loại Chi) không còn hợp lệ và hệ thống yêu cầu chọn danh mục loại Thu trước khi lưu.
3. **Given** Bob mở sửa giao dịch do Alice nhập, **When** Bob lưu thay đổi hợp lệ, **Then** lưu thành công (mọi thành viên ngang quyền).
4. **Given** hai thành viên cùng mở sửa một giao dịch, **When** người thứ hai lưu sau khi người thứ nhất đã lưu, **Then** hệ thống không ghi đè thầm lặng — người thứ hai được thông báo dữ liệu đã thay đổi và thấy dữ liệu mới nhất.
5. **Given** giao dịch được sửa ngày về một ngày trong quá khứ, **When** lưu, **Then** sổ và số dư vẫn nhất quán (số dư phản ánh đúng tổng mọi giao dịch).

---

### User Story 4 - Xóa giao dịch có xác nhận (Priority: P4)

Thành viên có thể xóa một giao dịch. Hệ thống luôn cảnh báo và yêu cầu xác nhận trước khi xóa vĩnh viễn; sau khi xóa, số dư tài khoản liên quan được hoàn tác đúng phần của giao dịch đó.

**Why this priority**: Cần thiết cho vòng đời dữ liệu nhưng tần suất thấp hơn nhập/sửa; rủi ro mất dữ liệu nên phải có xác nhận.

**Independent Test**: Xóa một giao dịch Chi 50.000 → hộp xác nhận hiện ra; xác nhận xong giao dịch biến mất khỏi sổ và số dư tài khoản tăng lại 50.000; hủy thì không có gì thay đổi.

**Acceptance Scenarios**:

1. **Given** một giao dịch bất kỳ, **When** thành viên bấm Xóa, **Then** hệ thống hiển thị cảnh báo xác nhận trước khi xóa vĩnh viễn.
2. **Given** hộp xác nhận đang hiển thị, **When** thành viên hủy, **Then** giao dịch và số dư giữ nguyên.
3. **Given** thành viên xác nhận xóa một giao dịch Chi 50.000, **When** xóa hoàn tất, **Then** giao dịch biến mất khỏi sổ của mọi thành viên và số dư tài khoản liên quan tăng lại 50.000.
4. **Given** Bob xóa giao dịch do Alice nhập, **When** xác nhận, **Then** xóa thành công (ngang quyền); sổ của Alice cập nhật theo.

---

### User Story 5 - Đăng nhập và định danh người dùng (Priority: P1 — nền tảng, thực hiện ĐẦU TIÊN)

Người dùng đăng nhập vào hệ thống bằng email và mật khẩu của tài khoản mình. Sau khi đăng nhập, hệ thống xác định được "tôi là ai" (hồ sơ người dùng với tên hiển thị) và "tôi thuộc hộ nào" (qua liên kết thành viên hộ dựa trên định danh người dùng), từ đó mọi dữ liệu hiển thị đúng phạm vi hộ và mọi bản ghi mới ghi đúng người tạo.

**Why this priority**: Đây là nguồn định danh duy nhất mà thành viên hộ, giao dịch và danh mục đều tham chiếu — mọi story khác phụ thuộc vào nó. Dù được liệt kê cuối, story này được thực hiện **trước tiên** (nền tảng).

**Independent Test**: Đăng nhập bằng tài khoản Alice → vào được sổ hộ A và thấy tên "Alice"; đăng xuất rồi đăng nhập Carol → thấy hộ B; chưa đăng nhập → không truy cập được bất kỳ dữ liệu nào.

**Acceptance Scenarios**:

1. **Given** một tài khoản người dùng hợp lệ, **When** đăng nhập đúng email và mật khẩu, **Then** người dùng vào được sổ của hộ mình và hồ sơ hiển thị đúng tên.
2. **Given** người dùng chưa đăng nhập, **When** mở ứng dụng, **Then** được đưa về màn đăng nhập và không thấy bất kỳ dữ liệu nào của hộ.
3. **Given** người dùng nhập sai mật khẩu, **When** thử đăng nhập, **Then** bị từ chối kèm thông báo rõ ràng, không lộ thông tin nhạy cảm.
4. **Given** Bob đã đăng nhập, **When** Bob nhập một giao dịch, **Then** sổ của mọi thành viên hiển thị giao dịch với người nhập là "Bob" (tên hiển thị, không phải mã định danh).

---

### Edge Cases

- **Số tiền không hợp lệ**: 0, số âm, ký tự không phải số → chặn lưu với thông báo cụ thể; không lưu giá trị mặc định thay thế.
- **Mô tả vượt 255 ký tự** → chặn/giới hạn ngay khi nhập, người dùng biết rõ giới hạn.
- **Đổi loại Thu/Chi khi sửa** → danh mục hiện tại mất hiệu lực; bắt buộc chọn lại danh mục cùng loại mới trước khi lưu (không bao giờ lưu giao dịch có danh mục khác loại).
- **Ngày trong tương lai** → không hỗ trợ ở phạm vi này (giao dịch dự kiến ngoài phạm vi); hệ thống chặn và thông báo.
- **Sửa/xóa giao dịch trong quá khứ** → số dư tài khoản luôn được tính lại nhất quán, không phụ thuộc thứ tự nhập.
- **Hai thành viên sửa cùng một giao dịch gần như đồng thời** → không ghi đè thầm lặng; người lưu sau nhận thông báo bản ghi đã thay đổi.
- **Giao dịch bị thành viên khác xóa trong lúc mình đang mở form sửa** → khi lưu, hệ thống báo giao dịch không còn tồn tại thay vì lỗi khó hiểu.
- **Hộ chưa có tài khoản nào** → hệ thống bảo đảm luôn có một tài khoản mặc định của hộ để ghi giao dịch (xem Assumptions — phụ thuộc BR-005).
- **Người dùng chưa có tên hiển thị** (tài khoản tạo trước khi có hồ sơ) → sổ hiển thị email thay cho tên, không bao giờ hiển thị mã định danh thô; hồ sơ được bổ sung tự động khi có thể.
- **Mất kết nối khi lưu** → thông báo rõ ràng, không tạo bản ghi trùng khi thử lại.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Hệ thống MUST yêu cầu số tiền cho mỗi giao dịch và MUST chỉ chấp nhận giá trị lớn hơn 0. *(BR-TRK-001)*
- **FR-002**: Mỗi giao dịch MUST mang đúng một loại: Thu hoặc Chi. *(BR-TRK-002)*
- **FR-003**: Mỗi giao dịch MUST được gán đúng một danh mục; danh sách danh mục khi nhập/sửa MUST chỉ gồm danh mục cùng loại Thu/Chi đang chọn và thuộc cùng hộ. *(BR-TRK-003; liên kết BR-001: FR-013, FR-014)*
- **FR-004**: Mô tả giao dịch là tùy chọn và MUST không vượt quá 255 ký tự. *(BR-TRK-004)*
- **FR-005**: Ngày giờ giao dịch MUST mặc định là thời điểm hiện tại và MUST cho phép sửa về một thời điểm trong quá khứ; ngày trong tương lai MUST bị từ chối. *(BR-TRK-005; xem Assumptions về giao dịch dự kiến)*
- **FR-006**: Mỗi giao dịch MUST gắn với một tài khoản của hộ; khi hộ chỉ có một tài khoản, tài khoản đó được chọn sẵn. *(BR-TRK-006; phụ thuộc BR-005 — xem Assumptions)*
- **FR-007**: Hệ thống MUST xác thực toàn bộ dữ liệu trước khi lưu (số tiền hợp lệ; loại, danh mục và tài khoản bắt buộc) và MUST hiển thị thông báo lỗi rõ ràng theo từng trường; giao dịch không hợp lệ MUST không được lưu. *(BR-TRK-007)*
- **FR-008**: Thành viên MUST be able to xem danh sách giao dịch của hộ, sắp xếp theo ngày giờ mới nhất trước, mỗi giao dịch hiển thị số tiền, loại, danh mục, ngày giờ và **tên thành viên đã nhập** (tên hiển thị dễ đọc, không phải mã định danh). *(BR-001/BR-CAT-009 — sổ chung minh bạch)*
- **FR-009**: Thành viên MUST be able to chỉnh sửa mọi thông tin của một giao dịch đã lưu (số tiền, loại, danh mục, mô tả, ngày giờ, tài khoản); dữ liệu sửa MUST qua cùng quy tắc xác thực như khi nhập mới; khi đổi loại Thu/Chi, hệ thống MUST yêu cầu chọn lại danh mục cùng loại mới. *(BR-TRK-008)*
- **FR-010**: Việc xóa giao dịch MUST luôn kèm cảnh báo và bước xác nhận rõ ràng trước khi xóa vĩnh viễn. *(BR-TRK-009)*
- **FR-011**: Sau mỗi thao tác thêm/sửa/xóa giao dịch, số dư của (các) tài khoản liên quan MUST được cập nhật chính xác — kể cả khi sửa số tiền, đổi tài khoản, đổi loại, hoặc thao tác trên giao dịch có ngày trong quá khứ. *(BR-TRK-010)*
- **FR-012**: Giao dịch MUST thuộc về một hộ gia đình, dùng chung giữa các thành viên trong hộ và cô lập giữa các hộ khác nhau. *(kế thừa mô hình BR-001 — FR-018 của feature 001)*
- **FR-013**: Mỗi giao dịch MUST ghi nhận thành viên đã nhập; mọi thành viên MUST có quyền ngang nhau khi xem/sửa/xóa giao dịch của hộ, kể cả giao dịch do thành viên khác nhập. *(kế thừa BR-CAT-009 — FR-021/FR-022 của feature 001)*
- **FR-014**: Khi nhiều thành viên sửa/xóa cùng một giao dịch gần như đồng thời, hệ thống MUST không ghi đè thầm lặng — người lưu sau MUST được thông báo bản ghi đã thay đổi hoặc đã bị xóa. *(nhất quán cơ chế đồng thời của feature 001)*
- **FR-015**: Hệ thống MUST duy trì danh sách **người dùng** với định danh duy nhất, email đăng nhập và tên hiển thị — là **nguồn định danh duy nhất** của hệ thống: tư cách thành viên hộ (liên kết người dùng ↔ hộ) và mọi trường "người nhập/người tạo" MUST tham chiếu tới định danh người dùng này. *(bước nền tảng đầu tiên của feature — xem Assumptions)*
- **FR-016**: Người dùng MUST đăng nhập bằng tài khoản của chính mình (email + mật khẩu) trước khi truy cập sổ của hộ; người chưa đăng nhập MUST không xem hoặc thao tác được bất kỳ dữ liệu nào của hộ.

### Key Entities *(include if feature involves data)*

- **Người dùng (User)**: Danh bạ người dùng **độc lập** của hệ thống — có định danh riêng (`user_id`), email (duy nhất) và tên hiển thị; **không phụ thuộc cấu trúc của cơ chế xác thực**. Phiên đăng nhập được đối chiếu với hồ sơ người dùng **qua email**; mật khẩu/thông tin xác thực do hệ thống đăng nhập quản lý riêng, không nằm trong thực thể này. Quan hệ: được **Thành viên hộ (household_members)** tham chiếu qua `user_id` để xác định ai thuộc hộ nào; được giao dịch/danh mục tham chiếu ở trường "người tạo/người nhập". Đây là thực thể nền tảng — phải tồn tại trước mọi dữ liệu khác của feature này.
- **Giao dịch (Transaction)**: Một bút toán Thu hoặc Chi của hộ. Thuộc tính: số tiền (> 0), loại (Thu/Chi), danh mục (bắt buộc, cùng loại, cùng hộ — BR-001), mô tả (tùy chọn, ≤ 255 ký tự), ngày giờ (mặc định hiện tại), tài khoản liên quan, hộ sở hữu, thành viên đã nhập (→ Người dùng). Quan hệ: thuộc một hộ; tham chiếu một danh mục và một tài khoản; do một thành viên tạo.
- **Tài khoản (Account)** *(định nghĩa đầy đủ ở BR-005, tham chiếu tại đây)*: Nguồn tiền của hộ mà giao dịch được ghi vào; có số dư phản ánh đúng tổng các giao dịch liên quan. Phạm vi này chỉ cần: mỗi giao dịch gắn một tài khoản và số dư tài khoản cập nhật đúng.
- **Danh mục (Category)** *(định nghĩa ở BR-001/feature 001)*: Nhóm phân loại bắt buộc của giao dịch, cùng loại Thu/Chi, dùng chung trong hộ.
- **Hộ gia đình & Thành viên** *(định nghĩa ở feature 001)*: Đơn vị sở hữu chung sổ giao dịch; thành viên ngang quyền.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Thành viên nhập xong một giao dịch thủ công trong ≤ 15 giây (trung bình). *(theo Success Metrics BR-002)*
- **SC-002**: ≥ 95% lượt lưu giao dịch thành công ngay lần đầu, không vướng lỗi xác thực. *(baseline cần nghiệp vụ xác nhận)*
- **SC-003**: 100% giao dịch đã lưu có đủ số tiền > 0, loại, danh mục cùng loại và tài khoản; không tồn tại giao dịch thiếu trường bắt buộc.
- **SC-004**: 100% thao tác thêm/sửa/xóa kết thúc với số dư tài khoản khớp đúng tổng bút toán liên quan (đối chiếu lại toàn bộ luôn khớp).
- **SC-005**: 100% lượt xóa giao dịch đi qua bước cảnh báo/xác nhận; không có xóa vĩnh viễn nào diễn ra chỉ bằng một thao tác.
- **SC-006**: Giao dịch do một thành viên nhập hiển thị trong sổ của các thành viên khác cùng hộ trong vòng vài giây.
- **SC-007**: 0 trường hợp ghi đè thầm lặng khi hai thành viên sửa cùng một giao dịch — người lưu sau luôn nhận được thông báo.

## Assumptions

- **Người dùng & đăng nhập là bước nền tảng ĐẦU TIÊN** *(yêu cầu 2026-07-06)*: Khi triển khai feature này, danh sách người dùng (FR-015) phải được dựng **trước mọi phần khác** — thứ tự phụ thuộc: Người dùng → Thành viên hộ → Tài khoản/Giao dịch. Người dùng hiện có (đã đăng nhập được từ trước) được bổ sung hồ sơ tự động, không cần thao tác tay. Vòng đời tài khoản đầy đủ (tự đăng ký, quên mật khẩu, đổi email, xóa tài khoản) nằm ngoài phạm vi — thuộc BR riêng về quản lý người dùng/hộ.
- **Chiến lược chuyển đổi dữ liệu hiện có** *(quyết định 2026-07-06)*: Mọi liên kết "ai" trong hệ thống (thành viên hộ, người tạo danh mục/giao dịch/hộ) được chuyển về tham chiếu danh sách người dùng mới. Vì đang ở giai đoạn phát triển (chỉ có dữ liệu thử nghiệm), môi trường dev được **làm mới (refresh)**: xóa dữ liệu app và dựng lại theo cấu trúc mới trong một bước; **tài khoản đăng nhập hiện có được giữ nguyên** và hồ sơ người dùng backfill tự động. Với môi trường có dữ liệu thật sau này, việc chuyển đổi phải giữ nguyên dữ liệu (không xóa).
- **Sổ chung hộ gia đình** *(kế thừa BR-001 cập nhật 2026-06-29)*: Giao dịch là dữ liệu chung của hộ; mọi thành viên ngang quyền; cô lập giữa các hộ. Quản lý hộ (tạo/mời/tham gia) vẫn là tiền đề thuộc feature riêng.
- **Tài khoản là phụ thuộc từ BR-005 (chưa triển khai)**: Để phạm vi này chạy độc lập, mỗi hộ được bảo đảm có sẵn một tài khoản mặc định (ví dụ "Tiền mặt") để ghi giao dịch; việc tạo/quản lý nhiều tài khoản, loại tài khoản và chuyển tiền thuộc BR-005. Khi BR-005 triển khai, form nhập cho phép chọn giữa các tài khoản của hộ.
- **Số dư là giá trị suy ra nhất quán**: "Cập nhật số dư" nghĩa là số dư tài khoản luôn phản ánh đúng tổng các giao dịch liên quan tại mọi thời điểm; không ràng buộc cách hiện thực (tính lại hay cộng dồn).
- **Không hỗ trợ ngày tương lai** *(trả lời Open Question của BR-002 bằng mặc định an toàn)*: Giao dịch dự kiến/định kỳ ngoài phạm vi (theo Out of Scope BR-002); chỉ chấp nhận ngày hiện tại hoặc quá khứ. Nếu nghiệp vụ muốn hỗ trợ, cần cập nhật BR trước.
- **Không có audit log chi tiết cho việc sửa** *(trả lời Open Question bằng mặc định MVP)*: Chỉ ghi thành viên đã nhập (người tạo); lịch sử từng lần thay đổi để giai đoạn sau nếu nghiệp vụ yêu cầu.
- **Một loại tiền tệ**: Theo Out of Scope BR-002; không xử lý đa tiền tệ/tỷ giá.
- **Danh mục dùng lại từ feature 001**: Toàn bộ quy tắc danh mục (bắt buộc, cùng loại, gợi ý, danh mục con) đã được BR-001/feature 001 định nghĩa và triển khai; spec này không định nghĩa lại.
- **Phần "nhập giao dịch tối thiểu" của feature 001**: Feature 001 đã dựng một luồng nhập giao dịch tối thiểu để kiểm chứng phân loại; feature này mở rộng thành luồng nhập/sửa/xóa đầy đủ theo BR-002 (thêm tài khoản, ngày giờ chỉnh được, xác thực đầy đủ, sổ giao dịch).
- **Owner & Target Quarter**: TBD trong BR-002; không ảnh hưởng phạm vi chức năng.
