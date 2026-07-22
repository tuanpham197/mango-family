# Feature Specification: Báo Cáo và Phân Tích Trực Quan (Visualized Reports)

**Feature Branch**: `005-reports`

**Created**: 2026-07-15

**Status**: Draft

**Input**: BR: `specs/business-requirements/BR-006.md` (nội dung mang tiêu đề nội bộ "BR-004: Báo Cáo và Phân Tích Trực Quan") — báo cáo tổng quan + báo cáo chi tiết theo danh mục, trực quan hóa bằng biểu đồ, theo khoảng thời gian người dùng chọn.

> ⚠️ **Ghi chú định danh BR**: file `BR-006.md` chứa requirement có ID nội bộ ghi là **BR-004** và mã con **BR-RPT-***. Repo không có file `BR-004.md` riêng. Spec này tham chiếu tới file thực tế `BR-006.md`; cần nghiệp vụ thống nhất lại số hiệu BR (đổi tên file hoặc sửa ID nội bộ) — xem Assumptions.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Báo cáo tổng quan theo khoảng thời gian (Priority: P1)

Một thành viên mở màn Báo cáo, chọn một khoảng thời gian (tuần này / tháng này / quý này / năm nay / tùy chỉnh), và xem **tổng thu nhập, tổng chi phí, số dư ròng** (thu − chi) của hộ trong khoảng đó. Đổi khoảng thời gian thì các con số cập nhật ngay trên cùng màn.

**Why this priority**: Con số tổng thu/chi/ròng theo kỳ là giá trị cốt lõi của báo cáo (BR-RPT-001) — nền tảng để mọi biểu đồ và chi tiết bên dưới có ngữ cảnh. Không có nó thì không có "báo cáo".

**Independent Test**: Với hộ có giao dịch trong tháng, mở Báo cáo → chọn "Tháng này" → thấy tổng thu, tổng chi, số dư ròng đúng bằng tổng giao dịch trong tháng; đổi sang "Tuần này" → các con số cập nhật theo tuần.

**Acceptance Scenarios**:

1. **Given** hộ có giao dịch trong khoảng đã chọn, **When** mở Báo cáo với khoảng "Tháng này", **Then** hiển thị tổng thu nhập, tổng chi phí và số dư ròng (thu − chi) đúng cho tháng hiện tại.
2. **Given** đang xem báo cáo, **When** đổi khoảng thời gian (tuần/tháng/quý/năm), **Then** tổng thu/chi/ròng cập nhật theo khoảng mới mà không rời màn.
3. **Given** thành viên chọn **khoảng tùy chỉnh** với ngày bắt đầu và ngày kết thúc, **When** ngày kết thúc ≥ ngày bắt đầu, **Then** báo cáo tính trên đúng khoảng đó; nếu ngày kết thúc < bắt đầu thì bị chặn với thông báo rõ ràng.
4. **Given** khoảng đã chọn không có giao dịch nào, **When** xem báo cáo, **Then** hiển thị 0 cho thu/chi/ròng (trạng thái trống rõ ràng, không lỗi).

---

### User Story 2 - Biểu đồ phân bổ chi tiêu theo danh mục (Priority: P1)

Trong báo cáo tổng quan, thành viên xem một **biểu đồ phân bổ chi tiêu theo danh mục** (dạng tròn hoặc cột) cho khoảng thời gian đã chọn: mỗi danh mục Chi chiếm bao nhiêu phần trong tổng chi, kèm số tiền và tỷ trọng %.

**Why this priority**: Trực quan hóa "tiền đi đâu" là giá trị chính người dùng tìm ở báo cáo (BR-RPT-002); cùng US1 tạo nên báo cáo tổng quan có ý nghĩa.

**Independent Test**: Với hộ chi ở nhiều danh mục trong kỳ, mở Báo cáo → thấy biểu đồ phân bổ với mỗi danh mục kèm số tiền + %, tổng các phần khớp tổng chi của kỳ.

**Acceptance Scenarios**:

1. **Given** hộ chi ở nhiều danh mục trong kỳ, **When** xem báo cáo tổng quan, **Then** biểu đồ phân bổ hiển thị từng danh mục Chi với số tiền và tỷ trọng % trên tổng chi, sắp theo mức chi giảm dần.
2. **Given** chi thuộc danh mục con, **When** tính phân bổ, **Then** được gộp vào danh mục cha (nhất quán cách gộp cây danh mục của các tính năng trước).
3. **Given** kỳ không có chi tiêu, **When** xem biểu đồ, **Then** hiển thị trạng thái trống rõ ràng.

---

### User Story 3 - Biểu đồ xu hướng thu nhập & chi phí theo thời gian (Priority: P2)

Thành viên xem một **biểu đồ đường** thể hiện xu hướng tổng thu nhập và tổng chi phí của hộ theo thời gian trong khoảng đã chọn (gom nhóm theo đơn vị thời gian phù hợp với độ dài khoảng).

**Why this priority**: Cho thấy diễn biến tăng/giảm theo thời gian (BR-RPT-003) — bổ trợ cho ảnh chụp tĩnh của US1/US2; giá trị cao nhưng đến sau bức tranh tổng và phân bổ.

**Independent Test**: Với hộ có giao dịch rải theo nhiều ngày/tháng trong kỳ, mở Báo cáo → thấy hai đường (thu, chi) theo mốc thời gian; mỗi mốc bằng tổng giao dịch trong mốc đó.

**Acceptance Scenarios**:

1. **Given** hộ có giao dịch trải theo thời gian trong kỳ, **When** xem báo cáo, **Then** biểu đồ xu hướng hiển thị hai đường thu & chi theo các mốc thời gian, mỗi mốc khớp tổng giao dịch trong mốc.
2. **Given** khoảng thời gian ngắn (tuần/tháng) so với dài (quý/năm), **When** hiển thị, **Then** đơn vị gom nhóm được chọn phù hợp (ví dụ theo ngày cho khoảng ngắn, theo tháng cho khoảng dài).

---

### User Story 4 - Báo cáo chi tiết theo danh mục (Priority: P2)

Thành viên chọn một danh mục cụ thể (từ biểu đồ phân bổ hoặc danh sách danh mục) để xem **báo cáo chi tiết**: tổng chi tiêu của danh mục đó trong kỳ và **danh sách tất cả giao dịch** thuộc danh mục đó (số tiền, mô tả, ngày).

**Why this priority**: Cho phép đào sâu từ tổng quan xuống chi tiết (BR-RPT-004/005) — trả lời "khoản chi này gồm những gì". Phụ thuộc dữ liệu tổng quan.

**Independent Test**: Chọn danh mục "Ăn uống" trong kỳ → thấy tổng chi "Ăn uống" và danh sách mọi giao dịch "Ăn uống" trong kỳ với số tiền/mô tả/ngày.

**Acceptance Scenarios**:

1. **Given** đang xem báo cáo tổng quan của một kỳ, **When** chọn một danh mục Chi, **Then** hiển thị tổng chi của danh mục đó trong kỳ và danh sách mọi giao dịch thuộc danh mục (kèm danh mục con) với số tiền, mô tả, ngày.
2. **Given** đang xem chi tiết một danh mục, **When** đổi khoảng thời gian, **Then** tổng và danh sách giao dịch cập nhật theo khoảng mới.
3. **Given** danh mục không có giao dịch trong kỳ, **When** xem chi tiết, **Then** tổng = 0 và danh sách trống (trạng thái rõ ràng).

---

### User Story 5 - Biểu đồ xu hướng chi tiêu của một danh mục (Priority: P3)

Trong báo cáo chi tiết theo danh mục, thành viên xem một **biểu đồ đường** thể hiện xu hướng chi tiêu của riêng danh mục đó theo thời gian trong khoảng đã chọn.

**Why this priority**: Hoàn thiện phần đào sâu (BR-RPT-006); hữu ích nhưng là lớp trên cùng, ưu tiên sau danh sách chi tiết.

**Independent Test**: Xem chi tiết "Ăn uống" trong một quý → thấy đường xu hướng chi "Ăn uống" theo các mốc thời gian trong quý.

**Acceptance Scenarios**:

1. **Given** đang xem chi tiết một danh mục trong kỳ, **When** hiển thị, **Then** biểu đồ đường thể hiện chi tiêu của danh mục đó theo các mốc thời gian, mỗi mốc khớp tổng chi của danh mục trong mốc.

---

### Edge Cases

- **Khoảng thời gian không có giao dịch** → mọi con số = 0, biểu đồ/danh sách hiển thị trạng thái trống, không lỗi.
- **Khoảng tùy chỉnh ngày kết thúc < ngày bắt đầu** → bị chặn với thông báo rõ ràng.
- **Chỉ có thu hoặc chỉ có chi trong kỳ** → biểu đồ/tổng hiển thị đúng phần có dữ liệu.
- **Danh mục bị ẩn nhưng có giao dịch trong kỳ** → vẫn được tính vào tổng và phân bổ (kèm dấu hiệu đã ẩn khi liệt kê).
- **Đổi danh mục / sửa / xóa giao dịch trong lúc đang xem** → khi làm mới báo cáo, số liệu phản ánh đúng dữ liệu hiện tại.
- **Khoảng thời gian rất dài (năm/tùy chỉnh dài)** → báo cáo vẫn hiển thị được; gom nhóm theo đơn vị lớn hơn để dễ đọc và giữ hiệu năng.
- **Ranh giới múi giờ khi gom nhóm theo ngày/tháng** → gom nhóm theo lịch nhất quán với toàn hệ thống (xem Assumptions).
- **Cô lập hộ** → mọi báo cáo chỉ tính dữ liệu của hộ đang đăng nhập.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Màn Báo cáo MUST cho phép chọn khoảng thời gian: **tuần này, tháng này, quý này, năm nay, hoặc tùy chỉnh** (ngày bắt đầu–kết thúc). *(BR-RPT-001)*
- **FR-002**: Báo cáo tổng quan MUST hiển thị **tổng thu nhập, tổng chi phí và số dư ròng** (thu − chi) của hộ trong khoảng đã chọn. *(BR-RPT-001)*
- **FR-003**: Báo cáo tổng quan MUST hiển thị **biểu đồ phân bổ chi tiêu theo danh mục** (tròn hoặc cột) cho khoảng đã chọn: mỗi danh mục Chi kèm số tiền và tỷ trọng % trên tổng chi; gộp danh mục con vào danh mục cha. *(BR-RPT-002)*
- **FR-004**: Báo cáo tổng quan MUST hiển thị **biểu đồ xu hướng thu nhập & chi phí theo thời gian** (đường) trong khoảng đã chọn, gom nhóm theo đơn vị thời gian phù hợp với độ dài khoảng. *(BR-RPT-003)*
- **FR-005**: Thành viên MUST xem được **báo cáo chi tiết theo một danh mục**: tổng chi của danh mục trong kỳ và **danh sách mọi giao dịch** thuộc danh mục đó (số tiền, mô tả, ngày), gồm cả danh mục con. *(BR-RPT-004, BR-RPT-005)*
- **FR-006**: Báo cáo chi tiết theo danh mục MUST hiển thị **biểu đồ xu hướng chi tiêu của danh mục đó theo thời gian** trong khoảng đã chọn. *(BR-RPT-006)*
- **FR-007**: Mọi số liệu báo cáo MUST được tính đúng theo khoảng thời gian & danh mục đã chọn, **chỉ trong phạm vi hộ đang đăng nhập**, và phản ánh đúng dữ liệu hiện tại khi làm mới (sau khi giao dịch/danh mục thay đổi hoặc khi đổi khoảng thời gian).
- **FR-008**: Khoảng tùy chỉnh MUST yêu cầu **ngày kết thúc ≥ ngày bắt đầu**; vi phạm MUST bị chặn kèm thông báo rõ ràng.
- **FR-009**: Mọi báo cáo/biểu đồ MUST xử lý **khoảng trống** (không có giao dịch) một cách rõ ràng (hiển thị 0 / trạng thái trống, không lỗi).

### Key Entities *(include if feature involves data)*

- **Giao dịch (Transaction)** *(feature 002)*: nguồn để tổng hợp thu/chi, phân bổ danh mục, xu hướng, chi tiết; chỉ **đọc**.
- **Danh mục (Category)** *(feature 001)*: nhóm chi theo danh mục (gộp cây cha–con); chỉ **đọc**.
- **Báo cáo (Report aggregates)** *(giá trị suy ra)*: tổng thu/chi/ròng, phân bổ theo danh mục, chuỗi xu hướng theo thời gian, tổng & danh sách theo danh mục — **suy ra** từ giao dịch theo khoảng thời gian, không lưu.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% thời điểm đối chiếu, tổng thu/chi/ròng và tổng theo danh mục trên báo cáo khớp đúng tổng giao dịch tương ứng của hộ trong khoảng đã chọn.
- **SC-002**: Báo cáo cho khoảng **một tháng** hiển thị kết quả trong **≤ 2 giây**. *(Success Metric BR-006)*
- **SC-003**: Thành viên đổi được khoảng thời gian (tuần/tháng/quý/năm/tùy chỉnh) và thấy báo cáo cập nhật ngay trên cùng màn, không cần rời màn.
- **SC-004**: 100% khoảng có dữ liệu hiển thị đúng biểu đồ phân bổ + xu hướng; 100% khoảng trống hiển thị trạng thái trống rõ ràng (0 lỗi).
- **SC-005**: Từ báo cáo tổng quan, thành viên đào sâu xuống chi tiết một danh mục trong ≤ 2 thao tác.
- **SC-006** *(KPI, baseline cần nghiệp vụ xác nhận)*: ≥ 50% hộ đang hoạt động xem báo cáo ≥ 1 lần/tháng. *(Success Metric BR-006)*

## Assumptions

- **Mô hình sổ chung hộ gia đình** *(kế thừa BR-001/BR-002)*: báo cáo là dữ liệu chung của hộ; cô lập giữa các hộ; một loại tiền tệ.
- **Định danh BR**: file `BR-006.md` chứa requirement báo cáo (ID nội bộ ghi "BR-004", mã con BR-RPT-*). Spec dùng file này làm nguồn; **cần nghiệp vụ thống nhất số hiệu** (không sửa trong phạm vi feature này).
- **Khoảng thời gian**: "tuần" bắt đầu Thứ Hai, "tháng/quý/năm" theo dương lịch (nhất quán cách tính kỳ của feature 003/004). Khoảng tùy chỉnh: bất kỳ ngày bắt đầu–kết thúc với end ≥ start; **không đặt trần cứng** cho độ dài, nhưng mục tiêu hiệu năng SC-002 áp cho khoảng ~1 tháng.
- **Múi giờ gom nhóm** *(trả lời Open Question)*: gom nhóm theo **lịch nhất quán toàn hệ thống** (cùng quy ước ranh giới ngày/tháng như feature 003/004), không theo múi giờ thiết bị riêng lẻ — để số liệu khớp giữa các thành viên.
- **Đơn vị gom nhóm xu hướng**: chọn theo độ dài khoảng (ví dụ theo ngày cho tuần/tháng, theo tháng cho quý/năm); ngưỡng cụ thể chốt ở /plan.
- **Cập nhật dữ liệu** *(trả lời Open Question)*: báo cáo tính **theo yêu cầu** (khi mở màn / đổi khoảng); khi đang mở, có thể làm mới nếu dữ liệu nền thay đổi (tái dùng cơ chế đồng bộ sẵn có). Không yêu cầu realtime nghiêm ngặt cho báo cáo.
- **Phân bổ danh mục** = chỉ giao dịch **Chi (EXPENSE)**; gộp danh mục con vào cha (nhất quán feature 003/004).
- **Màn Báo cáo** thay thế **placeholder `/reports`** đã dựng ở feature 004 (mục điều hướng "Báo cáo").
- **Ngoài phạm vi** *(theo Out of Scope BR-006)*: xuất báo cáo PDF/Excel/CSV; báo cáo dự báo (forecasting); so sánh/benchmark với người dùng khác; phân tích thu nhập chi tiết (feature tập trung phân tích **chi tiêu** theo danh mục).

## References / Truy vết

> Giữ các liên kết này cập nhật mỗi khi một artifact thay đổi (xem quy tắc lan truyền trong [`CLAUDE.md`](../../CLAUDE.md)).

- **Nguồn (BR)**: [`BR-006.md`](../business-requirements/BR-006.md) — Báo Cáo và Phân Tích Trực Quan (ID nội bộ ghi "BR-004"; mã con BR-RPT-001…006).
- **Tiền đề**: feature [`002-transaction-tracking`](../002-transaction-tracking/spec.md) (giao dịch Thu/Chi, ngày) · feature [`001-transaction-categorization`](../001-transaction-categorization/spec.md) (danh mục, cây cha–con) · feature [`004-monthly-income-expense-overview`](../004-monthly-income-expense-overview/spec.md) (mục điều hướng "Báo cáo" `/reports` — nay hiện thực hóa).
- **Use cases**: BR-006 nêu UC-008 (báo cáo tổng quan) · UC-009 (báo cáo chi tiết theo danh mục) — sẽ tạo ở bước use case.

## History

- v1 (2026-07-15): tạo spec từ `BR-006.md` (Báo cáo & phân tích trực quan) — báo cáo tổng quan (thu/chi/ròng + phân bổ danh mục + xu hướng) theo khoảng thời gian chọn được, và báo cáo chi tiết theo danh mục (tổng + danh sách giao dịch + xu hướng danh mục); hiện thực hóa mục "Báo cáo" (`/reports`) của feature 004. 4 Open Question của BR chốt bằng mặc định an toàn trong Assumptions (xuất file ngoài phạm vi; không trần cứng khoảng tùy chỉnh; gom nhóm theo lịch hệ thống; tính theo yêu cầu). Ghi nhận sai lệch số hiệu BR (file BR-006 ↔ ID nội bộ BR-004).
