# Feature Specification: Màn Tổng Quan Đầy Đủ (Dashboard) & Trang Mặc Định

**Feature Branch**: `004-monthly-income-expense-overview`

**Created**: 2026-07-14

**Status**: Draft

**Input**: User description: "tôi muốn hiện tại Thu / chi tháng này lên đầu tiên của trang tổng quan, đồng thời mặc định vào page đầu tiên là trang tổng quan" → mở rộng: "thêm hiển thị chi tiêu của từng danh mục trước budget, follow chính xác 100% bố cục của `specs/design/dashboard.png`".

> **Bố cục nguồn**: màn Tổng quan phải khớp **100%** wireframe [`specs/design/dashboard.png`](../design/dashboard.png). Thứ tự từ trên xuống: **(1) Lời chào + avatar → (2) Tổng tài sản ròng → (3) Thu nhập / Chi phí → (4) Chi tiêu theo danh mục (MỚI, trước ngân sách) → (5) Ngân sách tháng (per-category) → (6) Giao dịch gần đây**; thanh điều hướng dưới cùng: **Tổng quan · Giao dịch · (＋) · Ngân sách · Báo cáo**.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Vào ứng dụng mặc định ở màn Tổng quan (Priority: P1)

Khi một thành viên mở ứng dụng hoặc đăng nhập thành công (không kèm đích cụ thể), hệ thống đưa họ tới **màn Tổng quan** làm trang mặc định. Các màn khác vẫn truy cập được qua điều hướng.

**Why this priority**: Màn Tổng quan là "trang chủ" sức khỏe tài chính; phải là điểm đến đầu tiên thì các phần tóm tắt bên dưới mới phát huy giá trị.

**Independent Test**: Đăng nhập → màn đầu tiên là Tổng quan; mở lại app khi đã đăng nhập → cũng vào Tổng quan; từ đó điều hướng tới các màn khác vẫn được.

**Acceptance Scenarios**:

1. **Given** thành viên có phiên hợp lệ, **When** mở app ở địa chỉ gốc (không kèm liên kết cụ thể), **Then** màn Tổng quan hiển thị đầu tiên.
2. **Given** thành viên chưa đăng nhập, **When** đăng nhập thành công không kèm đích chuyển hướng, **Then** được đưa tới màn Tổng quan.
3. **Given** thành viên mở liên kết trỏ tới màn cụ thể (vd sổ giao dịch), **When** app tải, **Then** đích liên kết được tôn trọng (mặc định Tổng quan chỉ áp dụng khi vào địa chỉ gốc/sau đăng nhập không kèm đích).

---

### User Story 2 - Tóm tắt Thu/Chi tháng này (Priority: P1)

Màn Tổng quan hiển thị hai thẻ **Thu nhập** và **Chi phí** của hộ trong tháng dương lịch hiện tại (theo bố cục dashboard.png: nằm ngay dưới thẻ Tổng tài sản ròng), cho biết tổng thu và tổng chi tháng này.

**Why this priority**: Đây là cặp số cốt lõi "tháng này thu bao nhiêu / chi bao nhiêu" người dùng cần thấy ngay.

**Independent Test**: Hộ có giao dịch Thu/Chi trong tháng → hai thẻ hiển thị đúng tổng Thu và tổng Chi; thêm một giao dịch Chi → thẻ Chi phí cập nhật đúng.

**Acceptance Scenarios**:

1. **Given** hộ có tổng Thu 18.200.000 và tổng Chi 9.400.000 trong tháng, **When** mở màn Tổng quan, **Then** thẻ Thu nhập hiển thị tổng Thu và thẻ Chi phí hiển thị tổng Chi của tháng.
2. **Given** đang mở màn Tổng quan, **When** một giao dịch Thu/Chi của tháng hiện tại được thêm/sửa/xóa (kể cả giao dịch quá khứ trong tháng), **Then** thẻ Thu nhập/Chi phí tính lại đúng.
3. **Given** hai thẻ Thu nhập/Chi phí, **When** hiển thị, **Then** chúng nằm **ngay dưới** thẻ Tổng tài sản ròng và **trên** phần Chi tiêu theo danh mục (khớp dashboard.png).

---

### User Story 3 - Chi tiêu theo danh mục (trước Ngân sách) (Priority: P1)

Trước phần Ngân sách, màn Tổng quan hiển thị một mục **"Chi tiêu theo danh mục"**: liệt kê mức **chi (Chi)** của từng danh mục trong tháng hiện tại (mọi danh mục có phát sinh chi, không chỉ danh mục có ngân sách), kèm số tiền và **tỷ trọng %** trên tổng chi tháng, sắp xếp từ cao xuống thấp.

**Why this priority**: Người dùng muốn biết tiền chi đi đâu (danh mục nào tốn nhất) trước khi xét tới hạn mức ngân sách — đây là yêu cầu mới người dùng nêu rõ ("thêm hiển thị chi tiêu của từng danh mục trước budget").

**Independent Test**: Hộ có chi ở nhiều danh mục trong tháng → mục Chi tiêu theo danh mục liệt kê từng danh mục kèm số tiền và % đúng, sắp giảm dần, đặt ngay trước phần Ngân sách; thêm một giao dịch Chi → mục cập nhật.

**Acceptance Scenarios**:

1. **Given** hộ chi Ăn uống 3.500.000, Di chuyển 900.000, Mua sắm 1.200.000 trong tháng, **When** mở màn Tổng quan, **Then** mục Chi tiêu theo danh mục liệt kê ba danh mục này kèm số tiền và tỷ trọng % trên tổng chi, sắp giảm dần (Ăn uống trước).
2. **Given** một danh mục không phát sinh chi nào trong tháng, **When** hiển thị, **Then** danh mục đó không xuất hiện trong danh sách.
3. **Given** chi thuộc danh mục con, **When** tính, **Then** được gộp vào danh mục cha (nhất quán cách gộp cây danh mục của feature 003).
4. **Given** mục Chi tiêu theo danh mục và phần Ngân sách, **When** hiển thị, **Then** Chi tiêu theo danh mục nằm **ngay trước** phần Ngân sách (khớp yêu cầu người dùng).
5. **Given** đang mở màn, **When** giao dịch Chi thay đổi, **Then** số tiền và % của danh mục liên quan tính lại đúng.

---

### User Story 4 - Tổng tài sản ròng ở đầu màn (Priority: P2)

Trên cùng màn Tổng quan (dưới lời chào) hiển thị thẻ **Tổng tài sản ròng**: tổng số dư mọi tài khoản của hộ, kèm **tỷ lệ thay đổi so với tháng trước**.

**Why this priority**: Là "headline" của dashboard theo mockup; giá trị cao nhưng phụ thuộc dữ liệu tài khoản (đã có) — đứng sau các tóm tắt Thu/chi cốt lõi về mặt triển khai.

**Independent Test**: Hộ có nhiều tài khoản với số dư → thẻ hiển thị tổng số dư đúng và một chỉ số thay đổi so với tháng trước; số dư đổi khi có giao dịch → thẻ cập nhật.

**Acceptance Scenarios**:

1. **Given** hộ có các tài khoản với tổng số dư 24.560.000, **When** mở màn Tổng quan, **Then** thẻ Tổng tài sản ròng hiển thị 24.560.000 và một chỉ số thay đổi so với tháng trước.
2. **Given** một giao dịch làm số dư đổi, **When** lưu, **Then** Tổng tài sản ròng cập nhật (realtime cho thành viên khác ≤ 5s).
3. **Given** thẻ Tổng tài sản ròng, **When** hiển thị, **Then** nó là thẻ **đầu tiên** dưới lời chào (khớp dashboard.png).

---

### User Story 5 - Giao dịch gần đây trên màn Tổng quan (Priority: P2)

Cuối màn Tổng quan hiển thị **Giao dịch gần đây**: một vài giao dịch mới nhất của hộ (tên/mô tả, danh mục · tài khoản, số tiền có dấu và màu theo Thu/Chi), giúp xem nhanh không cần mở sổ đầy đủ.

**Why this priority**: Hoàn thiện bố cục mockup; tiện lợi nhưng đã có màn sổ đầy đủ nên ưu tiên sau.

**Independent Test**: Hộ có giao dịch → mục Giao dịch gần đây hiển thị vài giao dịch mới nhất đúng thứ tự và định dạng; nhập giao dịch mới → xuất hiện ở đầu danh sách.

**Acceptance Scenarios**:

1. **Given** hộ có nhiều giao dịch, **When** mở màn Tổng quan, **Then** mục Giao dịch gần đây hiển thị một số giao dịch mới nhất (Chi màu đỏ có dấu −, Thu màu xanh có dấu +), kèm danh mục · tài khoản.
2. **Given** một giao dịch mới được nhập, **When** lưu, **Then** nó xuất hiện đầu danh sách Giao dịch gần đây (realtime ≤ 5s cho thành viên khác).
3. **Given** mục Giao dịch gần đây, **When** hiển thị, **Then** nằm cuối cùng (dưới phần Ngân sách) khớp dashboard.png.

---

### User Story 6 - Bố cục & điều hướng khớp dashboard.png (Priority: P2)

Toàn bộ màn Tổng quan tuân theo **100%** bố cục dashboard.png: lời chào theo buổi + avatar ở trên cùng; thứ tự các phần đúng như mockup; thanh điều hướng dưới cùng gồm **Tổng quan · Giao dịch · (＋) · Ngân sách · Báo cáo**.

**Why this priority**: Bảo đảm trải nghiệm khớp thiết kế đã duyệt; là "khung" chứa các phần trên.

**Independent Test**: Mở màn Tổng quan → thấy lời chào + avatar; các phần xuất hiện đúng thứ tự mockup; thanh điều hướng có 5 mục đúng nhãn với nút ＋ ở giữa.

**Acceptance Scenarios**:

1. **Given** thành viên mở màn Tổng quan, **When** hiển thị, **Then** trên cùng là lời chào theo buổi (sáng/chiều/tối) kèm tên thành viên và avatar.
2. **Given** màn Tổng quan, **When** hiển thị, **Then** thứ tự các phần đúng: Tổng tài sản ròng → Thu nhập/Chi phí → Chi tiêu theo danh mục → Ngân sách → Giao dịch gần đây.
3. **Given** thanh điều hướng, **When** hiển thị, **Then** gồm 5 mục **Tổng quan · Giao dịch · ＋ · Ngân sách · Báo cáo** với ＋ (thêm giao dịch) nổi bật ở giữa.
4. **Given** phần Ngân sách trên Tổng quan, **When** hiển thị, **Then** có tiêu đề "Ngân sách tháng [N]" và lối **"Xem tất cả ›"** dẫn tới danh sách ngân sách (giữ nguyên hành vi feature 003).

---

### Edge Cases

- **Tháng chưa có giao dịch** → Thu 0, Chi 0; Chi tiêu theo danh mục trống (thông báo nhẹ); Giao dịch gần đây trống; không lỗi.
- **Chưa có ngân sách** → phần Ngân sách hiển thị trạng thái trống (như feature 003), các phần khác vẫn hiển thị.
- **Chỉ có Thu hoặc chỉ có Chi** → thẻ tương ứng bằng 0; Chi tiêu theo danh mục trống nếu không có Chi.
- **Giao dịch quá khứ trong tháng bị sửa/xóa/đổi loại** → mọi tóm tắt (Thu/Chi, chi theo danh mục, tài sản ròng) tính lại đúng.
- **Giao dịch ngoài tháng hiện tại** → không tính vào Thu/Chi tháng và Chi tiêu theo danh mục (nhưng vẫn ảnh hưởng Tổng tài sản ròng — số dư là luỹ kế).
- **Danh mục Chi bị ẩn nhưng có chi trong tháng** → vẫn tính vào Chi tiêu theo danh mục (kèm dấu hiệu đã ẩn), không biến mất khỏi tổng.
- **Sang tháng mới** → Thu/Chi, Chi tiêu theo danh mục reset theo tháng dương lịch mới; Tổng tài sản ròng là luỹ kế nên không reset.
- **"Báo cáo" chưa có** → mục điều hướng Báo cáo hiển thị theo mockup nhưng là placeholder (BR-004) — xem Assumptions.
- **Cô lập hộ** → mọi tóm tắt chỉ tính dữ liệu của hộ đang đăng nhập.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Khi thành viên mở app ở địa chỉ gốc hoặc đăng nhập thành công **không kèm đích chuyển hướng**, hệ thống MUST đưa họ tới **màn Tổng quan** làm trang mặc định; các liên kết trực tiếp tới màn cụ thể MUST vẫn được tôn trọng.
- **FR-002**: Màn Tổng quan MUST hiển thị **Tổng tài sản ròng** = tổng số dư mọi tài khoản của hộ, kèm chỉ số **thay đổi so với tháng trước**, là thẻ đầu tiên dưới lời chào.
- **FR-003**: Màn Tổng quan MUST hiển thị tổng **Thu nhập** và tổng **Chi phí** của hộ trong **tháng dương lịch hiện tại** (hai thẻ, dưới Tổng tài sản ròng).
- **FR-004**: Màn Tổng quan MUST hiển thị mục **Chi tiêu theo danh mục**: với mỗi danh mục Chi có phát sinh trong tháng, hiển thị **số tiền đã chi** và **tỷ trọng %** trên tổng chi tháng; sắp xếp giảm dần; gộp danh mục con vào danh mục cha; đặt **ngay trước** phần Ngân sách.
- **FR-005**: Màn Tổng quan MUST hiển thị **tóm tắt Ngân sách** tháng hiện tại (tiến độ per-category + lối "Xem tất cả") — giữ nguyên hành vi feature 003 — đặt **sau** mục Chi tiêu theo danh mục.
- **FR-006**: Màn Tổng quan MUST hiển thị **Giao dịch gần đây**: một số giao dịch mới nhất của hộ với mô tả/tên, danh mục · tài khoản, và số tiền có dấu/màu theo loại; đặt cuối màn. Khối này MUST có lối **"Xem tất cả ›"** dẫn tới màn Giao dịch (sổ đầy đủ, `/ledger`).
- **FR-007**: Màn Tổng quan MUST hiển thị **lời chào** theo buổi trong ngày kèm tên thành viên và avatar ở trên cùng.
- **FR-008**: Bố cục màn Tổng quan MUST khớp **100%** thứ tự dashboard.png: lời chào → Tổng tài sản ròng → Thu nhập/Chi phí → Chi tiêu theo danh mục → Ngân sách → Giao dịch gần đây.
- **FR-009**: Thanh điều hướng MUST gồm 5 mục theo mockup: **Tổng quan · Giao dịch · ＋(thêm giao dịch) · Ngân sách · Báo cáo**, với ＋ nổi bật ở giữa; các màn hiện có MUST vẫn truy cập được.
- **FR-010**: Mọi tóm tắt suy ra (Thu/Chi tháng, Chi tiêu theo danh mục, Tổng tài sản ròng, Giao dịch gần đây) MUST chỉ tính dữ liệu của **hộ đang đăng nhập** và MUST được tính lại đúng khi giao dịch được thêm/sửa/xóa/đổi loại/đổi danh mục — kể cả giao dịch quá khứ trong tháng.
- **FR-011**: Thay đổi do một thành viên gây ra MUST hiển thị trên màn Tổng quan của thành viên khác cùng hộ trong vòng 5 giây.

### Key Entities *(include if feature involves data)*

- **Giao dịch (Transaction)** *(feature 002)*: nguồn để suy ra Thu/Chi tháng, Chi tiêu theo danh mục, Giao dịch gần đây; chỉ **đọc**.
- **Tài khoản & Số dư (Account / Balance)** *(feature 002)*: nguồn để suy ra Tổng tài sản ròng; chỉ **đọc**.
- **Danh mục (Category)** *(feature 001)*: nhóm chi theo danh mục (gộp cây cha–con); chỉ **đọc**.
- **Ngân sách (Budget)** *(feature 003)*: tóm tắt ngân sách tháng; chỉ **đọc** (dùng lại widget 003).
- **Tóm tắt Tổng quan (Overview summaries)** *(giá trị suy ra)*: Thu tháng, Chi tháng, chi theo danh mục (+%), tài sản ròng + thay đổi tháng, danh sách giao dịch gần đây — **không lưu**, tính khi đọc.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% thời điểm đối chiếu, Thu tháng / Chi tháng trên Tổng quan khớp tổng giao dịch Thu/Chi của hộ trong tháng hiện tại (kể cả sau sửa/xóa/đổi loại quá khứ trong tháng).
- **SC-002**: 100% thời điểm đối chiếu, số tiền và tỷ trọng % từng danh mục trong mục Chi tiêu theo danh mục khớp tổng chi thực tế của danh mục đó (gồm danh mục con) trong tháng.
- **SC-003**: 100% thời điểm đối chiếu, Tổng tài sản ròng khớp tổng số dư mọi tài khoản của hộ.
- **SC-004**: 100% lần mở app ở địa chỉ gốc hoặc đăng nhập không kèm đích đều đưa người dùng tới màn Tổng quan.
- **SC-005**: Bố cục màn Tổng quan khớp thứ tự các phần và thanh điều hướng của dashboard.png (đối chiếu trực quan: 6 phần đúng thứ tự + 5 mục nav).
- **SC-006**: Thay đổi do một thành viên gây ra hiển thị trên màn Tổng quan của thành viên khác cùng hộ trong vòng 5 giây.
- **SC-007**: Ngay khi mở app, người dùng thấy được Tổng tài sản ròng và Thu/Chi tháng mà không cần thao tác điều hướng thêm.

## Assumptions

- **Mô hình sổ chung hộ gia đình** *(BR-001/BR-002)*: mọi tóm tắt là dữ liệu chung của hộ; cô lập giữa các hộ; một loại tiền tệ.
- **Kỳ theo tháng dương lịch**: "tháng này" = tháng dương lịch hiện tại; Thu/Chi và Chi tiêu theo danh mục reset đầu mỗi tháng (nhất quán kỳ MONTHLY feature 003). Tổng tài sản ròng là luỹ kế (không reset theo tháng).
- **Trang mặc định = điểm vào ứng dụng**: áp dụng cho mở ở địa chỉ gốc và sau đăng nhập khi **không có đích chuyển hướng**; màn sổ giao dịch **giữ nguyên địa chỉ hiện tại** và vẫn mở được qua điều hướng/liên kết trực tiếp (bảo toàn luồng + e2e 001/002/003).
- **Tổng tài sản ròng** = tổng số dư mọi tài khoản (dùng lại số dư suy ra của feature 002). **"% so với tháng trước"** = so sánh tài sản ròng hiện tại với tài sản ròng tại **cuối tháng trước** (suy ra từ số dư đầu kỳ + giao dịch tới thời điểm đó); nếu không đủ dữ liệu lịch sử để tính, hiển thị trung tính (không có mũi tên). *(Cách tính chính xác chốt ở /plan.)*
- **Chi tiêu theo danh mục** = chỉ giao dịch **Chi (EXPENSE)** của tháng; gộp danh mục con vào cha; % = số chi danh mục / tổng chi tháng; chỉ liệt kê danh mục có chi > 0, sắp giảm dần. Số lượng hiển thị (tất cả hay top N + "khác") chốt ở /plan; mặc định hiển thị tất cả danh mục có chi.
- **Giao dịch gần đây** = một vài (mặc định ~3–5) giao dịch mới nhất của hộ, chỉ xem nhanh (không sửa tại chỗ); mở sổ đầy đủ để thao tác.
- **Lời chào**: theo buổi trong ngày (sáng/chiều/tối) + tên hiển thị; avatar là chỗ dành sẵn (ảnh đại diện chưa thuộc phạm vi — hiển thị placeholder).
- **Mục điều hướng "Báo cáo"**: theo mockup, nhưng phân hệ Báo cáo thuộc **BR-004** chưa xây — hiển thị như **placeholder** ("sắp có") để giữ đúng bố cục; **Quản lý danh mục** (feature 001) vẫn phải truy cập được qua một lối phụ (vd trong luồng tạo danh mục nhanh hoặc mục phụ) — cách bố trí cụ thể chốt ở /plan. *(Điểm cần xác nhận nghiệp vụ nếu muốn thay hẳn Danh mục bằng Báo cáo.)*
- **Hiện thực hóa placeholder feature 003**: các thẻ trước đây là placeholder trên màn Tổng quan (tài sản ròng, Thu/chi tháng, giao dịch gần đây) nay được hiện thực hóa bằng dữ liệu thật; widget Ngân sách của feature 003 được giữ và đặt sau mục Chi tiêu theo danh mục.
- **Realtime dùng lại cơ chế sẵn có**: cập nhật ≤ 5s dùng chung kênh realtime các feature trước đã dùng cho giao dịch/tài khoản/ngân sách.
- **Ngoài phạm vi**: báo cáo nâng cao (biểu đồ xu hướng nhiều tháng, so sánh kỳ, phân tích sâu) thuộc BR-004; chỉnh sửa avatar/hồ sơ; đa tiền tệ.

## References / Truy vết

> Giữ các liên kết này cập nhật mỗi khi một artifact thay đổi (xem quy tắc lan truyền trong [`CLAUDE.md`](../../CLAUDE.md)).

- **Nguồn bố cục (wireframe)**: [`specs/design/dashboard.png`](../design/dashboard.png) — màn **1 · Tổng quan** (bố cục phải khớp 100%).
- **Tiền đề**: feature [`002-transaction-tracking`](../002-transaction-tracking/spec.md) (giao dịch Thu/Chi, tài khoản & số dư) · feature [`003-budgeting`](../003-budgeting/spec.md) (màn Tổng quan `/overview` + widget ngân sách, FR-014) · feature [`001-transaction-categorization`](../001-transaction-categorization/spec.md) (danh mục, cây cha–con).
- **Liên quan**: `BR-004` (Báo cáo — mục điều hướng "Báo cáo" là placeholder cho tới khi BR-004 được xây).

## History

- v1 (2026-07-14): tạo spec — Thu/chi tháng này ở đầu Tổng quan + đặt Tổng quan làm trang mặc định.
- v2 (2026-07-14): mở rộng theo yêu cầu người dùng + wireframe [`dashboard.png`](../design/dashboard.png) — bổ sung **Tổng tài sản ròng**, **Chi tiêu theo danh mục** (mục MỚI, trước Ngân sách), **Giao dịch gần đây**, lời chào + thanh điều hướng 5 mục; toàn màn khớp 100% bố cục mockup. Thu/chi tháng giữ nguyên nhưng theo bố cục mockup (dưới Tổng tài sản ròng). Trang mặc định không đổi.
- v3 (2026-07-15): sau triển khai — bổ sung lối **"Xem tất cả ›"** cho khối Giao dịch gần đây (mở màn Giao dịch `/ledger`) theo yêu cầu người dùng (FR-006). Component `RecentTransactionsList` + Vitest + e2e #17 cập nhật.
- v4 (2026-07-15): mục điều hướng **"Báo cáo" (`/reports`)** — vốn là placeholder ở feature 004 — đã được **feature 005** hiện thực hóa thành màn Báo cáo thật (`ReportsView` thay `ReportsPlaceholderView`). Phạm vi 004 không đổi.
