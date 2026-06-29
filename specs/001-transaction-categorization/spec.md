# Feature Specification: Phân Loại Giao Dịch (Transaction Categorization)

**Feature Branch**: `001-transaction-categorization`

**Created**: 2026-06-24

**Status**: Draft

**Input**: User description: "@specs/business-requirements/BR-001.md — Phân loại giao dịch: quản lý hệ thống danh mục Thu/Chi linh hoạt với danh mục con, đảm bảo mọi giao dịch đều được gán một danh mục phù hợp khi nhập."

## Clarifications

### Session 2026-06-24

- Q: Khi danh mục cha có danh mục con, người dùng gán giao dịch như thế nào? → A: Gán được cả cha và con — có thể chọn danh mục cha trực tiếp hoặc một danh mục con; báo cáo gộp (roll-up) giao dịch của danh mục con vào danh mục cha.
- Q: Có cho phép đổi loại Thu/Chi của một danh mục sau khi tạo không? → A: Không — loại cố định sau khi tạo; muốn chuyển loại thì tạo danh mục mới và gán lại giao dịch.
- Q: Người dùng được làm gì với danh mục mặc định? → A: Sửa, ẩn và xóa như danh mục tự tạo (kèm cơ chế gán lại khi còn giao dịch).
- Q: Cơ chế gợi ý danh mục (BR-CAT-007) dựa trên gì? → A: Kết hợp khớp từ khóa theo quy tắc trên mô tả VÀ lịch sử phân loại của người dùng (không dùng AI/ML; luôn ghi đè được).

## User Scenarios & Testing *(mandatory)*

<!--
  Các user story dưới đây được sắp xếp theo mức độ quan trọng (P1 cao nhất).
  Mỗi story là một lát cắt giá trị độc lập, có thể kiểm thử riêng.
-->

### User Story 1 - Bắt buộc phân loại giao dịch theo đúng loại Thu/Chi (Priority: P1)

Khi người dùng nhập một giao dịch, họ phải chọn một danh mục trước khi lưu. Hệ thống cung cấp sẵn một bộ danh mục mặc định đã phân theo loại Thu/Chi, và chỉ hiển thị các danh mục cùng loại với loại Thu/Chi mà người dùng đang nhập. Nhờ vậy, mọi giao dịch đều được gán đúng danh mục ngay từ đầu mà người dùng không cần tự tạo danh mục.

**Why this priority**: Đây là giá trị cốt lõi và là mục tiêu chính của requirement — bảo đảm dữ liệu giao dịch luôn được phân loại nhất quán, làm nền tảng cho Ngân sách (BR-003) và Báo cáo (BR-004). Chỉ riêng story này, với bộ danh mục mặc định, đã tạo ra một sản phẩm dùng được (MVP).

**Independent Test**: Có thể kiểm thử độc lập bằng cách: chọn loại Chi khi nhập giao dịch → xác nhận chỉ thấy danh mục Chi; thử lưu mà chưa chọn danh mục → hệ thống chặn; chọn một danh mục mặc định và lưu → giao dịch được gán danh mục thành công.

**Acceptance Scenarios**:

1. **Given** người dùng đang nhập một giao dịch loại Chi, **When** mở danh sách danh mục, **Then** chỉ các danh mục thuộc loại Chi được hiển thị (không có danh mục Thu nào xuất hiện).
2. **Given** người dùng đã nhập đầy đủ thông tin giao dịch nhưng chưa chọn danh mục, **When** bấm Lưu, **Then** hệ thống chặn lưu và yêu cầu chọn một danh mục.
3. **Given** người dùng chọn loại Thu, **When** mở danh sách danh mục, **Then** chỉ các danh mục thuộc loại Thu được hiển thị.
4. **Given** ứng dụng vừa được dùng lần đầu (chưa có danh mục tự tạo), **When** người dùng nhập giao dịch, **Then** vẫn có sẵn bộ danh mục mặc định đã phân theo Thu/Chi để chọn.

---

### User Story 2 - Tự tạo và quản lý danh mục cá nhân (Priority: P2)

Người dùng có thể thêm danh mục mới (chọn loại Thu/Chi, đặt tên, chọn biểu tượng tùy chọn), chỉnh sửa tên và biểu tượng của bất kỳ danh mục nào (mặc định hoặc tự tạo), và xóa danh mục. Khi xóa một danh mục còn giao dịch liên quan, hệ thống cảnh báo và yêu cầu người dùng gán lại các giao dịch đó sang một danh mục khác (cùng loại) hoặc xóa các giao dịch đó.

**Why this priority**: Cho phép người dùng cá nhân hóa hệ thống danh mục cho phù hợp với thói quen chi tiêu, làm tăng tỷ lệ phân loại đúng. Phụ thuộc vào US1 (đã có khái niệm danh mục và loại Thu/Chi) nhưng bổ sung khả năng quản lý.

**Independent Test**: Tạo một danh mục Chi mới với tên và biểu tượng → xác nhận nó xuất hiện khi nhập giao dịch Chi; sửa tên danh mục → xác nhận tên mới hiển thị; xóa một danh mục đang có giao dịch → xác nhận hệ thống buộc gán lại hoặc xóa giao dịch trước khi hoàn tất.

**Acceptance Scenarios**:

1. **Given** người dùng ở màn hình quản lý danh mục, **When** thêm danh mục mới và chọn loại Chi, đặt tên "Thú cưng", **Then** danh mục "Thú cưng" xuất hiện trong nhóm Chi và có thể chọn khi nhập giao dịch Chi.
2. **Given** một danh mục (mặc định hoặc tự tạo), **When** người dùng đổi tên và biểu tượng, **Then** tên và biểu tượng mới được áp dụng ở mọi nơi danh mục đó xuất hiện, kể cả giao dịch lịch sử.
3. **Given** một danh mục đang có giao dịch liên quan, **When** người dùng yêu cầu xóa, **Then** hệ thống cảnh báo và yêu cầu chọn: gán lại các giao dịch sang một danh mục khác cùng loại, hoặc xóa các giao dịch đó.
4. **Given** người dùng chọn gán lại khi xóa danh mục, **When** chọn danh mục đích, **Then** chỉ các danh mục cùng loại Thu/Chi được phép làm đích gán lại.
5. **Given** một danh mục không còn giao dịch nào, **When** người dùng xóa, **Then** danh mục bị xóa ngay mà không cần bước gán lại.

---

### User Story 3 - Tổ chức bằng danh mục con (Priority: P3)

Người dùng có thể tạo danh mục con nằm dưới một danh mục cha (một cấp), ví dụ "Ăn uống" → "Ăn ngoài", "Đi chợ". Danh mục con tự động kế thừa loại Thu/Chi của danh mục cha và không thể mang loại khác.

**Why this priority**: Giúp người dùng tổ chức chi tiêu chi tiết hơn và báo cáo phân bổ chính xác hơn, nhưng không bắt buộc cho việc phân loại cơ bản. Phụ thuộc vào việc đã có danh mục cha (US1/US2).

**Independent Test**: Tạo danh mục con "Ăn ngoài" dưới "Ăn uống" → xác nhận nó kế thừa loại Chi; nhập giao dịch Chi → xác nhận có thể chọn "Ăn ngoài"; thử gán loại Thu cho danh mục con → hệ thống không cho phép.

**Acceptance Scenarios**:

1. **Given** danh mục cha "Ăn uống" thuộc loại Chi, **When** người dùng tạo danh mục con "Ăn ngoài", **Then** danh mục con tự động thuộc loại Chi và không cho phép đổi sang Thu.
2. **Given** một danh mục con, **When** người dùng cố tạo thêm một cấp con bên dưới nó, **Then** hệ thống không cho phép (chỉ hỗ trợ một cấp danh mục con).
3. **Given** một danh mục cha có danh mục con và đang được xóa, **When** người dùng xác nhận xóa, **Then** hệ thống xử lý các danh mục con (gán lại hoặc xóa) với cùng cơ chế bảo vệ giao dịch như khi xóa danh mục thường.
4. **Given** danh mục cha "Ăn uống" có các danh mục con, **When** người dùng nhập một giao dịch Chi, **Then** người dùng có thể gán trực tiếp vào "Ăn uống" hoặc vào một danh mục con của nó, và báo cáo gộp giao dịch của danh mục con vào danh mục cha.

---

### User Story 4 - Gợi ý danh mục khi nhập giao dịch (Priority: P4)

Khi người dùng nhập giao dịch, hệ thống gợi ý một danh mục phù hợp dựa trên mô tả giao dịch và/hoặc lịch sử phân loại trước đó của người dùng (ví dụ mô tả "Grab" → gợi ý "Di chuyển"). Gợi ý chỉ mang tính tham khảo; người dùng luôn có thể chấp nhận hoặc chọn danh mục khác.

**Why this priority**: Tăng tốc độ nhập liệu và nâng tỷ lệ phân loại đúng, nhưng là cải tiến tiện ích chứ không phải điều kiện bắt buộc để phân loại. Phụ thuộc vào việc đã có danh mục và lịch sử giao dịch.

**Independent Test**: Nhập một giao dịch có mô tả chứa "Grab" → xác nhận hệ thống gợi ý danh mục "Di chuyển"; chấp nhận gợi ý → danh mục được điền sẵn; chọn danh mục khác → gợi ý bị ghi đè và giao dịch lưu theo lựa chọn của người dùng.

**Acceptance Scenarios**:

1. **Given** người dùng nhập giao dịch với mô tả gợi liên tưởng tới một danh mục đã biết, **When** hệ thống nhận diện được, **Then** một danh mục được đề xuất sẵn cùng loại Thu/Chi đang chọn.
2. **Given** một gợi ý danh mục được hiển thị, **When** người dùng chọn một danh mục khác, **Then** lựa chọn của người dùng được ưu tiên và được lưu cùng giao dịch.
3. **Given** mô tả giao dịch không khớp quy tắc hoặc lịch sử nào, **When** người dùng nhập, **Then** không có gợi ý nào được hiển thị và người dùng tự chọn danh mục.

---

### Edge Cases

- **Lưu khi chưa chọn danh mục**: Giao dịch không thể được lưu; hệ thống hiển thị thông báo yêu cầu chọn danh mục (không có cơ chế gán "Chưa phân loại" tự động).
- **Không có danh mục phù hợp khi nhập**: Nếu người dùng chọn một loại (Thu/Chi) mà chưa có danh mục nào thuộc loại đó, hệ thống hướng người dùng tạo nhanh một danh mục cùng loại trước khi lưu.
- **Xóa danh mục đang có giao dịch**: Bắt buộc gán lại sang danh mục khác cùng loại hoặc xóa các giao dịch — không bao giờ để lại giao dịch không có danh mục.
- **Xóa danh mục cha có danh mục con**: Hệ thống phải xử lý cả các danh mục con (gán lại hoặc xóa) theo cùng cơ chế bảo vệ.
- **Gán lại sai loại**: Hệ thống không cho phép gán lại giao dịch sang danh mục khác loại Thu/Chi.
- **Trùng tên danh mục**: Khi tạo/sửa trùng tên trong cùng một loại và cùng cấp cha, hệ thống cảnh báo để tránh nhầm lẫn.
- **Đổi tên/biểu tượng danh mục đang dùng**: Giao dịch lịch sử và báo cáo tự động phản ánh tên/biểu tượng mới (giao dịch tham chiếu danh mục, không sao chép tên).
- **Nhu cầu đổi loại Thu/Chi của danh mục đã dùng**: Loại của danh mục cố định sau khi tạo; muốn "đổi loại" người dùng tạo danh mục mới ở loại kia và gán lại giao dịch.
- **Gán vào danh mục cha có danh mục con**: Người dùng được phép gán giao dịch trực tiếp vào danh mục cha ngay cả khi nó đã có danh mục con; báo cáo gộp giao dịch của con vào cha.
- **Ẩn danh mục đang dùng**: Khi một danh mục bị ẩn, nó không xuất hiện để chọn cho giao dịch mới nhưng các giao dịch lịch sử và báo cáo vẫn giữ nguyên danh mục đó; có thể bỏ ẩn để dùng lại.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Hệ thống MUST cung cấp sẵn một bộ danh mục mặc định, đã được phân loại sẵn theo loại Thu hoặc Chi (ví dụ Chi: Ăn uống, Di chuyển, Hóa đơn; Thu: Lương). *(BR-CAT-001)*
- **FR-002**: Người dùng MUST be able to xem danh sách danh mục, được nhóm/lọc theo loại Thu/Chi. *(BR-CAT-001, BR-CAT-008)*
- **FR-003**: Người dùng MUST be able to tạo danh mục mới bằng cách chọn loại (đúng một trong hai: Thu hoặc Chi), đặt tên và chọn biểu tượng tùy chọn. *(BR-CAT-002)*
- **FR-004**: Mỗi danh mục MUST mang đúng một thuộc tính **loại** cố định là Thu hoặc Chi. *(BR-CAT-008)*
- **FR-005**: Loại Thu/Chi của một danh mục MUST không thể thay đổi sau khi tạo; để chuyển sang loại khác, người dùng tạo danh mục mới ở loại đó và gán lại giao dịch. *(BR-CAT-008; chốt tại Clarifications 2026-06-24)*
- **FR-006**: Người dùng MUST be able to chỉnh sửa tên và biểu tượng của bất kỳ danh mục nào, kể cả danh mục mặc định lẫn tự tạo. *(BR-CAT-003)*
- **FR-007**: Người dùng MUST be able to xóa danh mục mặc định hoặc tự tạo (cả hai được đối xử như nhau). *(BR-CAT-004; chốt tại Clarifications 2026-06-24)*
- **FR-008**: Khi xóa một danh mục còn giao dịch liên quan, hệ thống MUST cảnh báo và yêu cầu người dùng hoặc gán lại các giao dịch đó sang một danh mục khác, hoặc xóa các giao dịch đó. *(BR-CAT-004)*
- **FR-009**: Khi gán lại giao dịch trong lúc xóa danh mục, hệ thống MUST chỉ cho phép chọn danh mục đích cùng loại Thu/Chi với giao dịch. *(suy luận — toàn vẹn dữ liệu)*
- **FR-010**: Hệ thống MUST hỗ trợ danh mục con lồng đúng một cấp dưới một danh mục cha. *(BR-CAT-005; ngoài phạm vi: nhiều hơn một cấp)*
- **FR-011**: Danh mục con MUST kế thừa loại Thu/Chi của danh mục cha và MUST không thể mang loại khác. *(BR-CAT-005)*
- **FR-012**: Khi xóa một danh mục cha có danh mục con, hệ thống MUST xử lý các danh mục con (gán lại hoặc xóa) theo cùng cơ chế bảo vệ giao dịch ở FR-008. *(suy luận)*
- **FR-013**: Hệ thống MUST yêu cầu chọn danh mục khi nhập giao dịch và MUST chặn việc lưu giao dịch nếu chưa chọn danh mục. *(BR-CAT-006)*
- **FR-014**: Khi nhập giao dịch, hệ thống MUST chỉ hiển thị các danh mục có loại trùng với loại Thu/Chi mà giao dịch đang chọn. *(BR-CAT-008; liên kết BR-002: BR-TRK-002 → BR-TRK-003)*
- **FR-015**: Hệ thống MUST gợi ý danh mục bằng cách KẾT HỢP khớp từ khóa theo quy tắc trên mô tả giao dịch VÀ lịch sử phân loại của chính người dùng (không dùng AI/ML); người dùng MUST be able to chấp nhận hoặc ghi đè gợi ý. *(BR-CAT-007; chốt tại Clarifications 2026-06-24)*
- **FR-016**: Khi đổi tên hoặc biểu tượng một danh mục, hệ thống MUST phản ánh thay đổi ở mọi nơi danh mục được tham chiếu, bao gồm giao dịch lịch sử và báo cáo. *(suy luận)*
- **FR-017**: Hệ thống MUST cảnh báo khi tạo hoặc sửa danh mục có tên trùng trong cùng một loại và cùng cấp cha. *(suy luận)*
- **FR-018**: Danh mục MUST thuộc về từng người dùng riêng; không có chia sẻ hay đồng bộ bộ danh mục giữa nhiều người dùng. *(BR Out of Scope)*
- **FR-019**: Khi nhập giao dịch, người dùng MUST be able to gán giao dịch vào một danh mục cha trực tiếp HOẶC vào một danh mục con của nó; báo cáo MUST gộp (roll-up) giao dịch của danh mục con vào danh mục cha. *(chốt tại Clarifications 2026-06-24)*
- **FR-020**: Người dùng MUST be able to ẩn bất kỳ danh mục nào (mặc định hoặc tự tạo) khỏi danh sách chọn khi nhập giao dịch mới mà không xóa; danh mục đã ẩn MUST giữ nguyên cho giao dịch lịch sử và báo cáo, và MUST có thể được bỏ ẩn. *(chốt tại Clarifications 2026-06-24)*

### Key Entities *(include if feature involves data)*

- **Danh mục (Category)**: Đại diện cho một nhóm phân loại giao dịch. Thuộc tính chính: tên, loại (Thu hoặc Chi — cố định, đúng một loại), biểu tượng (tùy chọn), cờ phân biệt mặc định/tự tạo, cờ ẩn/hiện, chủ sở hữu (người dùng), tham chiếu danh mục cha (rỗng nếu là danh mục gốc). Quan hệ: thuộc về một người dùng; có thể có một danh mục cha; được nhiều giao dịch tham chiếu — cả danh mục cha (kể cả khi đã có danh mục con) lẫn danh mục con đều có thể được giao dịch tham chiếu trực tiếp.
- **Danh mục con (Subcategory)**: Một Danh mục có tham chiếu tới một danh mục cha (chỉ một cấp). Kế thừa loại Thu/Chi của cha.
- **Giao dịch (Transaction)** *(định nghĩa ở BR-002, tham chiếu tại đây)*: Mỗi giao dịch tham chiếu đúng một danh mục; loại Thu/Chi của giao dịch phải trùng loại của danh mục được gán.
- **Gợi ý danh mục (Category Suggestion)**: Liên kết suy ra từ mô tả giao dịch (theo từ khóa/quy tắc) và/hoặc lịch sử phân loại của người dùng tới một danh mục được đề xuất. Mang tính tham khảo, luôn có thể ghi đè.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: ≥ 98% giao dịch mới được gán một danh mục cụ thể (không rơi vào danh mục "Khác/Chưa phân loại"). *(baseline cần xác nhận với nghiệp vụ)*
- **SC-002**: 100% giao dịch không thể được lưu nếu chưa chọn danh mục (không có ngoại lệ).
- **SC-003**: 100% danh mục hiển thị trong lúc nhập giao dịch trùng loại Thu/Chi đang chọn (không có danh mục khác loại lọt vào danh sách).
- **SC-004**: ≥ 60% gợi ý danh mục được người dùng chấp nhận khi tính năng gợi ý bật. *(baseline cần xác nhận)*
- **SC-005**: Người dùng có thể tạo một danh mục mới trong vòng tối đa 3 bước thao tác và dưới 30 giây.
- **SC-006**: Người dùng tìm và chọn được danh mục đúng trong lúc nhập giao dịch trong vòng dưới 10 giây.
- **SC-007**: 100% thao tác xóa danh mục đang có giao dịch kết thúc mà không để lại giao dịch nào không có danh mục (gán lại hoặc xóa luôn hoàn tất).
- **SC-008**: 100% danh mục con mang đúng loại Thu/Chi của danh mục cha.

## Assumptions

- **Danh mục theo từng người dùng**: Danh mục là dữ liệu cá nhân của mỗi người dùng; không chia sẻ hay đồng bộ giữa nhiều người dùng (theo Out of Scope của BR-001).
- **Danh mục con một cấp**: Chỉ hỗ trợ một cấp danh mục con (cha → con). Nhiều hơn một cấp nằm ngoài phạm vi (theo BR-001).
- **Loại Thu/Chi cố định sau khi tạo** *(đã chốt — xem Clarifications 2026-06-24)*: Loại của một danh mục không đổi sau khi tạo để giữ báo cáo và ngân sách nhất quán; muốn chuyển loại thì tạo danh mục mới và gán lại giao dịch.
- **Danh mục mặc định ứng xử như danh mục tự tạo** *(đã chốt — xem Clarifications 2026-06-24)*: Người dùng có thể sửa tên/biểu tượng, ẩn và xóa danh mục mặc định, với cùng cơ chế gán lại khi danh mục còn giao dịch.
- **Cơ chế gợi ý** *(đã chốt — xem Clarifications 2026-06-24)*: Kết hợp khớp từ khóa theo quy tắc trên mô tả giao dịch VÀ học theo lịch sử phân loại của chính người dùng; luôn mang tính tham khảo và có thể ghi đè. Không dùng AI/ML nâng cao (theo Out of Scope).
- **Gán cha/con** *(đã chốt — xem Clarifications 2026-06-24)*: Giao dịch có thể gán trực tiếp vào danh mục cha hoặc một danh mục con; báo cáo gộp giao dịch của con vào cha.
- **Bộ danh mục mặc định** *(Open Question còn mở)*: Dựa trên các ví dụ trong BR-001 cộng các danh mục thường gặp (Chi: Ăn uống, Di chuyển, Hóa đơn, Mua sắm, Giải trí, Sức khỏe; Thu: Lương, Thưởng…). Danh sách cuối cùng cần nghiệp vụ xác nhận.
- **Tham chiếu thay vì sao chép**: Giao dịch tham chiếu danh mục theo định danh, nên đổi tên danh mục sẽ cập nhật cách hiển thị xuyên suốt lịch sử và báo cáo mà không cần gán lại.
- **Danh mục "Khác/Other"**: Có thể tồn tại như một danh mục để người dùng chủ động chọn, nhưng không được dùng làm giá trị mặc định tự động; mục tiêu là giảm thiểu việc rơi vào danh mục này (xem SC-001).
- **Phụ thuộc luồng nhập giao dịch (BR-002)**: Việc gán danh mục diễn ra trong luồng nhập giao dịch định nghĩa ở BR-002; spec này chỉ bao phủ phần liên quan đến danh mục của luồng đó.
- **Là nền tảng cho các requirement khác**: Hệ thống danh mục này là cơ sở dữ liệu nền cho Ngân sách (BR-003) và Báo cáo (BR-004).
- **Owner & Target Quarter**: Chưa xác định (TBD trong BR-001); không ảnh hưởng phạm vi chức năng.
