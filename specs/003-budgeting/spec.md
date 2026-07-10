# Feature Specification: Thiết Lập Ngân Sách (Budgeting)

**Feature Branch**: `003-budgeting`

**Created**: 2026-07-10

**Status**: Draft

**Input**: User description: "@specs/business-requirements/BR-003.md — Thiết lập ngân sách: đặt giới hạn chi tiêu theo danh mục hoặc theo tổng trong một khoảng thời gian, theo dõi tiến độ thực tế và nhận cảnh báo khi đến ngưỡng hoặc vượt ngân sách."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Tạo ngân sách theo danh mục (Priority: P1)

Một thành viên trong hộ tạo ngân sách cho một danh mục chi tiêu: chọn danh mục loại Chi của hộ, nhập số tiền giới hạn (lớn hơn 0), chọn kỳ áp dụng (hàng tháng, hàng tuần, hoặc một lần với khoảng ngày xác định). Ngân sách là dữ liệu chung của hộ — mọi thành viên đều thấy và theo dõi được. Nếu thiếu hoặc sai trường bắt buộc, hệ thống chặn lưu và chỉ rõ lỗi.

**Why this priority**: Không tạo được ngân sách thì không có gì để theo dõi hay cảnh báo — đây là cửa ngõ của toàn bộ BR-003. Cùng US2 tạo thành MVP.

**Independent Test**: Tạo ngân sách "Ăn uống" 5.000.000/tháng → ngân sách xuất hiện trong danh sách của hộ với tiến độ tính từ chi tiêu hiện có trong tháng; thử lưu với số tiền 0/âm/bỏ trống danh mục → bị chặn kèm thông báo rõ ràng.

**Acceptance Scenarios**:

1. **Given** thành viên chọn danh mục Chi "Ăn uống", nhập giới hạn 5.000.000, kỳ hàng tháng, **When** bấm Lưu, **Then** ngân sách được tạo và xuất hiện trong danh sách ngân sách của hộ, tiến độ phản ánh chi tiêu "Ăn uống" từ đầu kỳ hiện tại.
2. **Given** thành viên nhập giới hạn 0 hoặc số âm, **When** bấm Lưu, **Then** hệ thống chặn và thông báo giới hạn phải lớn hơn 0.
3. **Given** thành viên chưa chọn danh mục (với ngân sách theo danh mục), **When** bấm Lưu, **Then** hệ thống chặn và yêu cầu chọn danh mục; danh sách chọn chỉ gồm danh mục loại Chi của hộ.
4. **Given** danh mục "Ăn uống" đã có ngân sách hàng tháng đang hoạt động, **When** thành viên tạo thêm ngân sách hàng tháng cho chính danh mục đó, **Then** hệ thống chặn và chỉ tới ngân sách hiện có (mỗi danh mục tối đa một ngân sách đang hoạt động trong cùng kỳ).
5. **Given** Alice (hộ A) vừa tạo một ngân sách, **When** Bob (cùng hộ A) mở danh sách ngân sách, **Then** Bob thấy ngân sách đó; Carol (hộ B) không thấy.

---

### User Story 2 - Theo dõi tiến độ ngân sách (Priority: P1)

Thành viên xem danh sách ngân sách của hộ, mỗi ngân sách hiển thị tiến độ kỳ hiện tại dạng "đã chi/giới hạn (phần trăm)" — ví dụ "3.500.000/5.000.000 VNĐ (70%)". Tiến độ luôn phản ánh đúng tổng giao dịch Chi liên quan trong kỳ (gồm cả danh mục con), kể cả khi giao dịch quá khứ bị sửa/xóa hoặc đổi danh mục.

**Why this priority**: Hiển thị tiến độ là giá trị cốt lõi của ngân sách (BR-BGT-005); không có nó thì giới hạn chỉ là con số chết. Cùng US1 tạo thành MVP.

**Independent Test**: Với ngân sách "Ăn uống" 5.000.000/tháng và chi tiêu hiện có 3.500.000 → hiển thị "3.500.000/5.000.000 (70%)"; nhập thêm giao dịch Chi "Ăn uống" 500.000 → tiến độ thành 4.000.000 (80%); xóa giao dịch đó → quay về 70%.

**Acceptance Scenarios**:

1. **Given** ngân sách "Ăn uống" 5.000.000/tháng với tổng chi "Ăn uống" trong tháng là 3.500.000, **When** mở danh sách ngân sách, **Then** hiển thị "3.500.000/5.000.000 (70%)".
2. **Given** thành viên nhập một giao dịch Chi thuộc danh mục có ngân sách, **When** giao dịch được lưu, **Then** tiến độ ngân sách cập nhật đúng số tiền mới.
3. **Given** một giao dịch Chi trong quá khứ thuộc kỳ hiện tại bị sửa số tiền hoặc bị xóa, **When** thao tác hoàn tất, **Then** tiến độ ngân sách được tính lại đúng.
4. **Given** một giao dịch được đổi từ danh mục A (có ngân sách) sang danh mục B (có ngân sách khác), **When** lưu, **Then** tiến độ của cả hai ngân sách đều cập nhật đúng.
5. **Given** Alice đang mở màn hình ngân sách, **When** Bob nhập một giao dịch ảnh hưởng ngân sách, **Then** tiến độ trên màn hình của Alice cập nhật trong vòng 5 giây.
6. **Given** giao dịch Chi thuộc danh mục con của danh mục có ngân sách, **When** lưu, **Then** giao dịch được tính vào tiến độ ngân sách của danh mục cha.

---

### User Story 3 - Nhận cảnh báo ngưỡng và vượt ngân sách (Priority: P2)

Khi chi tiêu của một ngân sách đạt ngưỡng cảnh báo (80% giới hạn), mọi thành viên trong hộ nhận cảnh báo trong ứng dụng. Khi chi tiêu vượt giới hạn, cảnh báo hiển thị rõ số tiền vượt. Mỗi mức cảnh báo chỉ phát một lần trong kỳ (không dội bom thông báo); nếu tiến độ tụt xuống dưới mức (do sửa/xóa giao dịch) rồi vượt lên lại, cảnh báo phát lại.

**Why this priority**: Cảnh báo là phần "chủ động" của ngân sách (BR-BGT-006/007) — giúp điều chỉnh hành vi trước khi vượt. Phụ thuộc US1 + US2.

**Independent Test**: Ngân sách 1.000.000/tháng, chi tiêu 750.000 → nhập thêm giao dịch 100.000 (85%) → cảnh báo "đạt 80%"; nhập thêm 200.000 (105%) → cảnh báo "vượt 50.000"; nhập thêm giao dịch nữa → không lặp lại cảnh báo cùng mức.

**Acceptance Scenarios**:

1. **Given** ngân sách với tiến độ dưới 80%, **When** một giao dịch mới đẩy tiến độ đạt/vượt 80% (nhưng chưa quá 100%), **Then** cảnh báo "đạt ngưỡng 80%" hiển thị trong ứng dụng cho mọi thành viên của hộ.
2. **Given** ngân sách với tiến độ dưới 100%, **When** một giao dịch đẩy tổng chi vượt giới hạn, **Then** cảnh báo "vượt ngân sách" hiển thị kèm đúng số tiền vượt.
3. **Given** cảnh báo 80% đã phát trong kỳ và tiến độ vẫn trên 80%, **When** có thêm giao dịch, **Then** không phát lại cảnh báo 80%.
4. **Given** tiến độ đã tụt xuống dưới 80% do một giao dịch bị xóa, **When** giao dịch mới đẩy tiến độ vượt 80% lần nữa, **Then** cảnh báo 80% phát lại.
5. **Given** Bob gây ra giao dịch làm vượt ngưỡng, **When** cảnh báo phát, **Then** Alice (cùng hộ) cũng thấy cảnh báo đó.

---

### User Story 4 - Ngân sách tổng chi tiêu (Priority: P2)

Thành viên tạo ngân sách cho **tổng** chi tiêu của hộ trong một kỳ (không gắn danh mục cụ thể). Tiến độ tính trên mọi giao dịch Chi của hộ trong kỳ; hoạt động song song với các ngân sách theo danh mục và dùng chung cơ chế cảnh báo.

**Why this priority**: Bao quát "bức tranh lớn" (BR-BGT-002) nhưng giá trị chỉ trọn vẹn khi đã có nhập liệu + tiến độ + cảnh báo chạy ổn.

**Independent Test**: Tạo ngân sách tổng 20.000.000/tháng → tiến độ bằng tổng mọi giao dịch Chi trong tháng của hộ, độc lập với các ngân sách danh mục đang có.

**Acceptance Scenarios**:

1. **Given** hộ có nhiều giao dịch Chi thuộc nhiều danh mục, **When** tạo ngân sách tổng theo tháng, **Then** tiến độ bằng tổng toàn bộ chi tiêu trong tháng.
2. **Given** đã có ngân sách tổng hàng tháng đang hoạt động, **When** tạo thêm ngân sách tổng hàng tháng, **Then** hệ thống chặn (tối đa một ngân sách tổng đang hoạt động mỗi kỳ).
3. **Given** ngân sách tổng và ngân sách danh mục cùng tồn tại, **When** một giao dịch Chi được nhập, **Then** tiến độ của cả hai đều cập nhật; cảnh báo của từng ngân sách hoạt động độc lập.

---

### User Story 5 - Quản lý ngân sách: sửa và xóa (Priority: P3)

Bất kỳ thành viên nào cũng có thể sửa (giới hạn, kỳ, danh mục) hoặc xóa một ngân sách của hộ, kể cả do thành viên khác tạo (ngang quyền). Sửa phải qua cùng bộ xác thực như khi tạo; xóa luôn có bước xác nhận. Trạng thái cảnh báo được tính lại theo giới hạn mới.

**Why this priority**: Vòng đời tối thiểu để dữ liệu ngân sách không bị "đóng băng"; tần suất thấp hơn tạo/theo dõi. *(BR-003 không liệt kê rõ — xem Assumptions)*

**Independent Test**: Sửa giới hạn từ 5.000.000 thành 4.000.000 khi đã chi 3.500.000 → tiến độ nhảy từ 70% lên 87,5% và cảnh báo 80% phát; xóa ngân sách → hộp xác nhận hiện ra, xác nhận xong ngân sách biến mất khỏi danh sách của mọi thành viên.

**Acceptance Scenarios**:

1. **Given** ngân sách 5.000.000 đã chi 3.500.000 (70%), **When** sửa giới hạn thành 4.000.000, **Then** tiến độ hiển thị 87,5% và cảnh báo ngưỡng 80% phát theo trạng thái mới.
2. **Given** một ngân sách bất kỳ, **When** thành viên bấm Xóa, **Then** hệ thống yêu cầu xác nhận; xác nhận xong ngân sách và cảnh báo liên quan biến mất khỏi danh sách của mọi thành viên; hủy thì không gì thay đổi.
3. **Given** Bob mở sửa ngân sách do Alice tạo, **When** Bob lưu thay đổi hợp lệ, **Then** lưu thành công (mọi thành viên ngang quyền).
4. **Given** hai thành viên cùng sửa một ngân sách, **When** người thứ hai lưu sau, **Then** hệ thống không ghi đè thầm lặng — người lưu sau được thông báo dữ liệu đã thay đổi.

---

### Edge Cases

- **Sửa/xóa giao dịch quá khứ làm tiến độ tụt xuống dưới ngưỡng** → tiến độ và trạng thái cảnh báo tính lại đúng; vượt ngưỡng lần nữa sẽ cảnh báo lại (US3 #4).
- **Giao dịch đổi loại Chi → Thu** → không còn được tính vào bất kỳ ngân sách nào; tiến độ liên quan tính lại.
- **Danh mục có ngân sách bị ẩn** (feature 001 cho phép ẩn danh mục) → ngân sách vẫn theo dõi bình thường; danh sách ngân sách hiển thị kèm trạng thái danh mục để thành viên biết.
- **Ngân sách một lần kết thúc kỳ** → chuyển trạng thái đã kết thúc, giữ để xem lại, không phát cảnh báo nữa.
- **Ngân sách tạo giữa kỳ** → tiến độ tính toàn bộ chi tiêu từ đầu kỳ hiện tại (không chỉ từ lúc tạo).
- **Kỳ chưa có giao dịch nào** → tiến độ 0%, không cảnh báo.
- **Giới hạn nhập không phải số / bỏ trống** → chặn lưu với thông báo cụ thể.
- **Hai thành viên cùng sửa một ngân sách gần như đồng thời** → không ghi đè thầm lặng (nhất quán cơ chế của feature 002).
- **Ngân sách bị thành viên khác xóa trong lúc mình đang mở form sửa** → khi lưu, hệ thống báo ngân sách không còn tồn tại.
- **Sang kỳ mới (tháng/tuần)** → ngân sách lặp lại tự động với cùng giới hạn; tiến độ và trạng thái cảnh báo bắt đầu lại từ 0.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Thành viên MUST tạo được ngân sách cho một danh mục cụ thể; danh mục MUST thuộc hộ và thuộc loại Chi. *(BR-BGT-001)*
- **FR-002**: Thành viên MUST tạo được ngân sách cho tổng chi tiêu của hộ trong một kỳ, không gắn danh mục. *(BR-BGT-002)*
- **FR-003**: Mỗi ngân sách MUST có số tiền giới hạn là giá trị số dương (> 0); giá trị 0, số âm hoặc không phải số MUST bị từ chối kèm thông báo rõ ràng. *(BR-BGT-003)*
- **FR-004**: Mỗi ngân sách MUST thuộc đúng một kỳ áp dụng: hàng tháng (theo tháng dương lịch), hàng tuần (bắt đầu Thứ Hai), hoặc một lần (khoảng ngày xác định, ngày kết thúc không trước ngày bắt đầu); ngân sách hàng tháng/hàng tuần MUST tự lặp lại mỗi kỳ với cùng giới hạn, tiến độ bắt đầu lại từ 0. *(BR-BGT-004; xem Assumptions)*
- **FR-005**: Danh sách ngân sách MUST hiển thị cho mỗi ngân sách: tên/danh mục, giới hạn, số đã chi trong kỳ hiện tại và phần trăm tiến độ (ví dụ "3.500.000/5.000.000 VNĐ (70%)"). *(BR-BGT-005)*
- **FR-006**: Tiến độ ngân sách MUST luôn bằng tổng các giao dịch Chi liên quan trong kỳ — theo danh mục (bao gồm danh mục con) với ngân sách danh mục, hoặc toàn bộ chi tiêu của hộ với ngân sách tổng — và MUST được tính lại đúng khi giao dịch được thêm/sửa/xóa hoặc đổi danh mục/loại, kể cả giao dịch trong quá khứ. *(BR-BGT-005; Constraint BR-003)*
- **FR-007**: Hệ thống MUST phát cảnh báo trong ứng dụng cho mọi thành viên của hộ khi tiến độ một ngân sách đạt ngưỡng 80% giới hạn. *(BR-BGT-006; ngưỡng cố định — xem Assumptions)*
- **FR-008**: Hệ thống MUST phát cảnh báo khi chi tiêu vượt giới hạn ngân sách, hiển thị rõ số tiền vượt. *(BR-BGT-007)*
- **FR-009**: Mỗi mức cảnh báo (ngưỡng 80%, vượt 100%) MUST chỉ phát tối đa một lần trong một kỳ khi tiến độ duy trì trên mức đó; nếu tiến độ tụt xuống dưới mức rồi vượt lên lại, cảnh báo MUST phát lại.
- **FR-010**: Mỗi danh mục MUST có tối đa một ngân sách đang hoạt động trong cùng một kỳ; hộ MUST có tối đa một ngân sách tổng đang hoạt động mỗi kỳ; vi phạm MUST bị chặn kèm chỉ dẫn tới ngân sách hiện có.
- **FR-011**: Ngân sách MUST thuộc về một hộ gia đình, dùng chung giữa các thành viên trong hộ và cô lập giữa các hộ khác nhau; mỗi ngân sách MUST ghi nhận thành viên đã tạo. *(kế thừa mô hình BR-001/BR-002)*
- **FR-012**: Thành viên MUST sửa được (giới hạn, kỳ, danh mục) và xóa được ngân sách của hộ với quyền ngang nhau, kể cả ngân sách do thành viên khác tạo; sửa MUST qua cùng quy tắc xác thực như khi tạo; xóa MUST luôn có bước xác nhận; trạng thái cảnh báo MUST được tính lại theo thay đổi. *(suy luận vòng đời tối thiểu — xem Assumptions)*
- **FR-013**: Khi nhiều thành viên sửa/xóa cùng một ngân sách gần như đồng thời, hệ thống MUST không ghi đè thầm lặng — người lưu sau MUST được thông báo bản ghi đã thay đổi hoặc đã bị xóa. *(nhất quán feature 002)*

### Key Entities *(include if feature involves data)*

- **Ngân sách (Budget)**: Giới hạn chi tiêu của hộ trong một kỳ. Thuộc tính: loại (theo danh mục / tổng), danh mục liên kết (bắt buộc với loại danh mục — loại Chi, cùng hộ), số tiền giới hạn (> 0), kỳ áp dụng (hàng tháng / hàng tuần / một lần kèm khoảng ngày), trạng thái (đang hoạt động / đã kết thúc), hộ sở hữu, người tạo. Quan hệ: thuộc một hộ; tham chiếu một danh mục (nếu loại danh mục); tiến độ là **giá trị suy ra** từ giao dịch (không lưu cứng).
- **Cảnh báo ngân sách (Budget Alert)**: Ghi nhận một lần phát cảnh báo của ngân sách trong một kỳ. Thuộc tính: ngân sách liên quan, mức (ngưỡng 80% / vượt giới hạn), số tiền vượt (nếu có), thời điểm phát, kỳ áp dụng. Dùng để hiển thị trong ứng dụng và chống phát trùng trong kỳ (FR-009).
- **Giao dịch / Danh mục / Hộ gia đình & Thành viên** *(định nghĩa ở feature 001/002)*: Ngân sách chỉ **đọc** dữ liệu giao dịch và danh mục để tính tiến độ; không thay đổi chúng.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Thành viên tạo xong một ngân sách trong ≤ 30 giây với tối đa 3 bước thao tác chính.
- **SC-002**: 100% ngân sách đã lưu có giới hạn > 0, kỳ hợp lệ và (với loại danh mục) đúng danh mục Chi thuộc hộ; không tồn tại ngân sách thiếu trường bắt buộc.
- **SC-003**: 100% thời điểm đối chiếu, tiến độ ngân sách khớp đúng tổng giao dịch Chi liên quan trong kỳ — kể cả sau khi sửa/xóa/đổi danh mục giao dịch quá khứ.
- **SC-004**: 100% lần tiến độ chạm ngưỡng 80% hoặc vượt giới hạn đều có cảnh báo; 0 cảnh báo trùng cùng mức trong một kỳ khi tiến độ duy trì trên mức.
- **SC-005**: 100% cảnh báo vượt ngân sách hiển thị đúng số tiền vượt.
- **SC-006**: Thay đổi tiến độ/cảnh báo do một thành viên gây ra hiển thị cho các thành viên khác cùng hộ trong vòng 5 giây. *(nhất quán SC-006 feature 002)*
- **SC-007**: ≥ 40% hộ đang hoạt động có ít nhất một ngân sách sau 1 tháng ra mắt tính năng. *(KPI theo Success Metrics BR-003 — baseline cần nghiệp vụ xác nhận)*

## Assumptions

- **Mô hình sổ chung hộ gia đình** *(kế thừa BR-001/BR-002)*: Ngân sách là dữ liệu chung của hộ; mọi thành viên ngang quyền tạo/xem/sửa/xóa; cô lập giữa các hộ. BR-003 viết "người dùng" nhưng toàn hệ thống theo mô hình hộ — cần nghiệp vụ xác nhận nếu muốn ngân sách cá nhân.
- **Ngưỡng cảnh báo cố định 80%** *(trả lời Open Question bằng mặc định an toàn)*: MVP dùng ngưỡng 80% cố định cho mọi ngân sách; cho phép tự cấu hình ngưỡng là mở rộng sau, cần cập nhật BR trước.
- **Kênh cảnh báo: trong ứng dụng (in-app)** *(trả lời Open Question)*: MVP chỉ cảnh báo trong ứng dụng; push notification/email thuộc mở rộng sau (phụ thuộc hạ tầng thông báo riêng).
- **Tự khởi tạo lại mỗi kỳ: có** *(trả lời Open Question)*: Ngân sách hàng tháng/hàng tuần tự lặp lại với cùng giới hạn; tiến độ và trạng thái cảnh báo bắt đầu lại từ 0 mỗi kỳ; không rollover phần chưa dùng (Out of Scope BR-003).
- **Một danh mục ≤ 1 ngân sách hoạt động trong cùng kỳ** *(trả lời Open Question)*: Tránh nhập nhằng "một khoản chi tính vào ngân sách nào"; ngân sách tổng luôn tồn tại song song và độc lập với ngân sách danh mục.
- **Ngân sách chỉ tính chi tiêu (loại Chi)**: Giao dịch Thu không bao giờ tính vào ngân sách; ngân sách theo danh mục bao gồm cả danh mục con của danh mục đó (nhất quán cây danh mục feature 001).
- **Kỳ theo lịch**: Tháng theo tháng dương lịch; tuần bắt đầu Thứ Hai; ngân sách tạo giữa kỳ tính toàn bộ chi tiêu từ đầu kỳ hiện tại.
- **Sửa/xóa ngân sách là suy luận** *(BR-003 In Scope không liệt kê)*: Spec bổ sung vòng đời tối thiểu (FR-012) để ngân sách không bị đóng băng sau khi tạo; cần nghiệp vụ xác nhận khi review BR.
- **Một loại tiền tệ**: Theo Out of Scope BR-003; không xử lý đa tiền tệ.
- **Phụ thuộc dữ liệu**: Tiến độ ngân sách phụ thuộc giao dịch (BR-002/feature 002) và danh mục (BR-001/feature 001); ngân sách chỉ đọc, không sửa dữ liệu nguồn. Feature 002 cần được triển khai trước hoặc song song để ngân sách có dữ liệu thật.
- **Owner & Target Quarter**: TBD trong BR-003; không ảnh hưởng phạm vi chức năng.

## References / Truy vết

> Giữ các liên kết này cập nhật mỗi khi một artifact thay đổi (xem quy tắc lan truyền trong [`CLAUDE.md`](../../CLAUDE.md)).

- **Nguồn (BR)**: [`BR-003`](../business-requirements/BR-003.md) — Thiết lập ngân sách.
- **Tiền đề**: [`BR-001`](../business-requirements/BR-001.md) / feature [`001-transaction-categorization`](../001-transaction-categorization/spec.md) (danh mục, mô hình hộ) · [`BR-002`](../business-requirements/BR-002.md) / feature [`002-transaction-tracking`](../002-transaction-tracking/spec.md) (giao dịch, số dư).
- **Use cases**: chưa tạo — dự kiến `UC-BGT-01` (tạo ngân sách), `UC-BGT-02` (theo dõi & cảnh báo) tại `specs/use-cases/003-budgeting/` (bước sau).
- **Entity model**: [`specs/entities/entity-model.md`](../entities/entity-model.md) — cần bổ sung BUDGET/BUDGET_ALERT khi cập nhật (bước sau).
- **Dẫn xuất (design)**: `plan.md` · `research.md` · `data-model.md` · `contracts/` · `quickstart.md` · `tasks.md` (chưa tạo — `/speckit-plan`).
- **Checklist chất lượng**: [`checklists/requirements.md`](checklists/requirements.md).

## History

- v1 (2026-07-10): tạo spec từ BR-003 (mô hình sổ chung hộ gia đình kế thừa BR-001/BR-002); 4 Open Question của BR-003 (ngưỡng cấu hình, kênh cảnh báo, tự lặp kỳ, danh mục nhiều ngân sách) chốt bằng mặc định an toàn trong Assumptions — chờ nghiệp vụ xác nhận; bổ sung FR-012 sửa/xóa ngân sách (suy luận vòng đời tối thiểu).
