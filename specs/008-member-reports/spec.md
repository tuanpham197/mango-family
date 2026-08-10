# Feature Specification: Báo Cáo Thu Chi Theo Thành Viên (Per-Member Income/Expense Report)

**Feature Branch**: `008-member-reports`

**Created**: 2026-08-10

**Status**: Draft

**Input**: BR: `specs/business-requirements/BR-008.md` — mở rộng màn Báo cáo hiện tại (feature `005-reports`) với tổng Thu/Chi/ròng **theo từng thành viên** trong hộ theo khoảng thời gian chọn được, và drill-down danh sách giao dịch của một thành viên. Quy thuộc giao dịch theo **thành viên đã nhập** (`transactions.created_by`).

> ℹ️ **Ghi chú định danh**: "thu/chi của thành viên" = thành viên **đã nhập** giao dịch (`created_by`), **không** khẳng định người đó thực tế nhận/chi tiền. Feature này **chỉ đọc**, suy ra từ giao dịch — không đổi mô hình giao dịch, không thêm bảng tổng hợp.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Xem & so sánh Thu/Chi/ròng theo từng thành viên (Priority: P1)

Một thành viên mở màn Báo cáo, chọn một khoảng thời gian (tuần này / tháng này / quý này / năm nay / tùy chỉnh), và xem danh sách **tất cả thành viên hiện tại** của hộ, mỗi thành viên kèm **tổng Thu**, **tổng Chi** và **số dư ròng = Thu − Chi** trong khoảng đó. Đổi khoảng thời gian thì mọi con số theo thành viên được tính lại ngay trên cùng màn.

**Why this priority**: Đây là giá trị cốt lõi của requirement (BR-MBR-001/002) — biết mỗi thành viên đóng góp bao nhiêu vào thu/chi của hộ. Không có phần này thì không có "báo cáo theo thành viên".

**Independent Test**: Với hộ có giao dịch do nhiều thành viên nhập trong tháng, mở Báo cáo → phần theo thành viên → thấy mỗi thành viên có tổng Thu/Chi/ròng đúng; đổi sang "Tuần này" → các con số cập nhật theo tuần.

**Acceptance Scenarios**:

1. **Given** trong hộ A, Alice và Bob mỗi người đã nhập vài giao dịch trong tháng, **When** một thành viên hộ A xem báo cáo theo thành viên cho "Tháng này", **Then** mỗi thành viên hiển thị một dòng với tổng Thu, tổng Chi và ròng (Thu − Chi); giao dịch Alice nhập tính cho Alice, giao dịch Bob nhập tính cho Bob.
2. **Given** đang xem báo cáo theo thành viên, **When** đổi khoảng thời gian (tuần/tháng/quý/năm/tùy chỉnh), **Then** toàn bộ số liệu theo thành viên được tính lại theo cùng ranh giới thời gian với tổng hợp toàn hộ, không rời màn.
3. **Given** một giao dịch thuộc danh mục con, **When** tính tổng Thu/Chi của thành viên, **Then** giá trị vẫn được gán trọn cho thành viên đã nhập (một giao dịch không bị chia cho nhiều thành viên).

---

### User Story 2 - Xem danh sách giao dịch của một thành viên (drill-down) (Priority: P2)

Từ báo cáo theo thành viên, thành viên chọn một thành viên để xem danh sách các giao dịch cấu thành số liệu của người đó trong khoảng đang chọn; mỗi giao dịch hiển thị tối thiểu loại Thu/Chi, số tiền, danh mục, mô tả, ngày giờ và tài khoản. Quay lại báo cáo giữ nguyên khoảng đã chọn.

**Why this priority**: Cho phép soi "vì sao con số của một thành viên như vậy" (BR-MBR-006/007) — bổ trợ trực tiếp cho US1, nhưng US1 vẫn dùng được độc lập nếu chưa có drill-down.

**Independent Test**: Trên báo cáo theo thành viên, chọn Alice → thấy đúng các giao dịch Alice đã nhập trong khoảng, có phân trang khi dài; bấm quay lại → vẫn ở cùng khoảng thời gian.

**Acceptance Scenarios**:

1. **Given** đang xem báo cáo theo thành viên hộ A cho "Tháng này" và Alice có nhiều giao dịch, **When** chọn Alice, **Then** hệ thống liệt kê đúng các giao dịch Alice đã nhập trong tháng, mỗi giao dịch kèm loại Thu/Chi, số tiền, danh mục, mô tả, ngày giờ và tài khoản.
2. **Given** danh sách giao dịch của một thành viên dài hơn một trang, **When** mở danh sách, **Then** danh sách được phân trang trong cùng khoảng và cùng thành viên.
3. **Given** thành viên được chọn không có giao dịch trong khoảng, **When** mở danh sách, **Then** hiển thị trạng thái không có dữ liệu rõ ràng và cho phép quay lại.
4. **Given** đang xem giao dịch của một thành viên với khoảng "Quý này", **When** quay lại báo cáo theo thành viên, **Then** khoảng vẫn là "Quý này" (không bị đặt lại).

---

### User Story 3 - Tính đầy đủ & đối soát với tổng toàn hộ (Priority: P2)

Báo cáo theo thành viên phải **đầy đủ** (mọi thành viên hiện tại đều xuất hiện, kể cả người không có giao dịch) và **đối soát được** (cộng mọi dòng thành viên khớp đúng tổng Thu/Chi toàn hộ cùng khoảng), để người dùng tin được con số.

**Why this priority**: Nếu số liệu theo thành viên không khớp tổng hộ hoặc bỏ sót thành viên, báo cáo mất giá trị và niềm tin — đây là ràng buộc đúng đắn (BR-MBR-004/005), tách riêng để kiểm thử độc lập.

**Independent Test**: Với hộ có thành viên chưa nhập giao dịch nào trong kỳ, mở báo cáo → người đó vẫn hiện với 0/0/0; cộng tổng Thu và tổng Chi mọi dòng → bằng đúng tổng Thu/Chi toàn hộ trên báo cáo tổng quan cùng kỳ.

**Acceptance Scenarios**:

1. **Given** Carol thuộc hộ A nhưng chưa nhập giao dịch nào trong khoảng, **When** xem báo cáo theo thành viên, **Then** Carol vẫn xuất hiện với tổng Thu = 0, tổng Chi = 0 và ròng = 0.
2. **Given** một khoảng bất kỳ có dữ liệu (kể cả có giao dịch do người đã rời hộ nhập), **When** cộng tổng Thu và tổng Chi của mọi dòng thành viên hiện tại **cộng nhóm "Thành viên cũ"** (nếu có), **Then** kết quả bằng đúng tổng Thu và tổng Chi toàn hộ trên báo cáo tổng quan (feature 005) cùng khoảng.
3. **Given** hộ không có giao dịch nào trong khoảng, **When** xem báo cáo, **Then** mọi thành viên hiện tại hiển thị 0/0/0 (trạng thái trống rõ ràng, không lỗi).

---

### Edge Cases

- **Giao dịch do người đã rời hộ nhập**: một giao dịch có người nhập từng thuộc hộ nhưng nay không còn là thành viên. Để không phá vỡ đối soát tổng hộ (US3/AC2), toàn bộ giá trị được **gộp vào nhóm "Thành viên cũ / Đã rời hộ"** hiển thị như một dòng bổ sung chỉ khi tồn tại dữ liệu như vậy (FR-010).
- **Khoảng tùy chỉnh có ngày kết thúc < ngày bắt đầu**: bị chặn với thông báo rõ ràng (nhất quán feature 005).
- **Thành viên duy nhất trong hộ**: báo cáo hiển thị một dòng khớp đúng tổng hộ.
- **Yêu cầu drill-down tới một thành viên không thuộc hộ hiện tại**: hệ thống từ chối, không trả về giao dịch nào (không nhận trực tiếp user ID ngoài hộ).
- **Người nhập chưa đặt tên hiển thị**: hiển thị email thay cho tên; không bao giờ hiển thị mã định danh thô.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Hệ thống MUST hiển thị danh sách **tất cả thành viên hiện tại** của hộ trong báo cáo, mỗi thành viên gồm tên hiển thị, **tổng Thu**, **tổng Chi** và **số dư ròng = Thu − Chi** trong khoảng thời gian đang chọn. *(BR-MBR-001)*
- **FR-002**: Hệ thống MUST kế thừa bộ lọc thời gian của báo cáo hiện tại (tuần, tháng, quý, năm, khoảng tùy chỉnh); khi đổi khoảng, MUST tính lại toàn bộ số liệu theo thành viên theo **cùng ranh giới thời gian** với tổng hợp toàn hộ. *(BR-MBR-002)*
- **FR-003**: Hệ thống MUST gán trọn giá trị mỗi giao dịch cho **đúng một** thành viên theo thành viên đã nhập giao dịch; MUST NOT chia một giao dịch cho nhiều thành viên và MUST NOT suy luận người thực tế nhận/chi tiền. *(BR-MBR-003)*
- **FR-004**: Hệ thống MUST hiển thị thành viên hiện tại không có giao dịch trong khoảng với tổng Thu = 0, tổng Chi = 0 và ròng = 0. *(BR-MBR-004)*
- **FR-005**: Tổng Thu và tổng Chi khi cộng mọi dòng thành viên (thành viên hiện tại **và** nhóm "Thành viên cũ" nếu có — xem FR-010) MUST khớp với tổng Thu và tổng Chi toàn hộ trên báo cáo hiện tại trong cùng khoảng; hệ thống MUST NOT lưu cứng số liệu báo cáo mà suy ra từ giao dịch theo yêu cầu. *(BR-MBR-005)*
- **FR-006**: Người dùng MUST be able to chọn một thành viên để xem danh sách các giao dịch được tính vào số liệu của thành viên đó trong khoảng hiện tại; mỗi giao dịch MUST hiển thị tối thiểu loại Thu/Chi, số tiền, danh mục, mô tả, ngày giờ và tài khoản. *(BR-MBR-006)*
- **FR-007**: Drill-down MUST hỗ trợ trạng thái không có dữ liệu và phân trang khi danh sách dài; quay lại báo cáo MUST giữ nguyên khoảng thời gian đã chọn. *(BR-MBR-007)*
- **FR-008**: Mọi thành viên trong hộ MUST có quyền ngang nhau khi xem báo cáo theo thành viên; dữ liệu MUST được cô lập giữa các hộ; hệ thống MUST xác thực thành viên được chọn (drill-down) thuộc đúng hộ hiện tại và MUST NOT nhận trực tiếp một user ID ngoài hộ. *(BR-MBR-008)*
- **FR-009**: Tên hiển thị thành viên MUST theo quy tắc định danh: dùng tên hiển thị, **fallback sang email** khi chưa có tên, và MUST NOT hiển thị mã định danh thô. *(BR-MBR-009)*
- **FR-010**: Với giao dịch có người nhập **không còn là thành viên hiện tại** của hộ, hệ thống MUST gộp toàn bộ giá trị vào một nhóm **"Thành viên cũ / Đã rời hộ"** và hiển thị nhóm này như **một dòng bổ sung chỉ khi** tồn tại dữ liệu như vậy — để tổng các dòng (thành viên hiện tại + nhóm Thành viên cũ) luôn khớp tổng Thu/Chi toàn hộ (FR-005). Nhóm "Thành viên cũ" MUST hỗ trợ drill-down xem danh sách giao dịch như một dòng thành viên bình thường (FR-006/FR-007); trong drill-down, mỗi giao dịch vẫn hiển thị tên/định danh người đã nhập theo quy tắc FR-009 nếu còn truy được, ngược lại gắn nhãn "Thành viên cũ". *(giải quyết Open Question BR-008 #1 — Option A)*

### Key Entities *(include if feature involves data)*

- **Giao dịch (Transaction)** *(feature 002, chỉ đọc)*: nguồn để tổng hợp; mang loại Thu/Chi, số tiền, danh mục, tài khoản, ngày giờ và **thành viên đã nhập** (`created_by`) dùng làm khóa nhóm.
- **Người dùng / Thành viên hộ (User / Household member)** *(nền tảng, chỉ đọc)*: nguồn tên hiển thị (fallback email) và xác định tư cách thành viên hiện tại của hộ; là đối tượng được nhóm theo trong báo cáo.
- **Số liệu theo thành viên (Per-member aggregate)** *(giá trị suy ra)*: tổng Thu, tổng Chi, ròng của từng thành viên trong khoảng — suy ra từ giao dịch theo `household_id` + khoảng, **không lưu**.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% thời điểm đối chiếu, tổng Thu và tổng Chi cộng mọi dòng thành viên hiện tại **và nhóm "Thành viên cũ"** (nếu có) khớp đúng tổng Thu/Chi toàn hộ trong cùng khoảng và cùng tập dữ liệu. *(Success Metric BR-008)*
- **SC-002**: Báo cáo theo thành viên cho khoảng **một tháng** hiển thị kết quả trong **≤ 2 giây**, nhất quán với báo cáo hiện tại. *(Success Metric BR-008; baseline TBD)*
- **SC-003**: 100% thành viên hiện tại của hộ xuất hiện trong báo cáo cho mọi khoảng (kể cả người không có giao dịch → hiển thị 0/0/0).
- **SC-004**: Từ báo cáo theo thành viên, người dùng đào sâu xuống danh sách giao dịch của một thành viên trong ≤ 2 thao tác, và quay lại giữ nguyên khoảng đã chọn.
- **SC-005**: 100% khoảng có dữ liệu tính đúng theo thành viên; 100% khoảng trống hiển thị trạng thái trống rõ ràng (0 lỗi).
- **SC-006**: Cô lập hộ tuyệt đối — 0 trường hợp một thành viên thấy dữ liệu hoặc thành viên của hộ khác.
- **SC-007** *(KPI adoption, baseline cần nghiệp vụ xác nhận)*: tỷ lệ hộ đang hoạt động xem báo cáo theo thành viên ≥ 1 lần/tháng: từ TBD → TBD. *(Success Metric BR-008)*

## Assumptions

- **Mô hình sổ chung hộ gia đình** *(kế thừa BR-001/BR-002)*: báo cáo là dữ liệu chung của hộ; mọi thành viên quyền ngang nhau; cô lập giữa các hộ; một loại tiền tệ.
- **Quy thuộc theo `created_by`** *(BR-008 Background)*: "thu/chi của thành viên" = thành viên đã nhập giao dịch; không suy người thực tế nhận/chi; không thêm trường "chủ sở hữu giao dịch".
- **Chỉ đọc, suy ra**: không tạo bảng tổng hợp; nhóm theo thành viên dùng **cùng** điều kiện thời gian, loại giao dịch và quy tắc cô lập hộ với endpoint báo cáo hiện tại (feature 005).
- **Ranh giới thời gian**: "tuần" bắt đầu Thứ Hai; "tháng/quý/năm" theo dương lịch; gom nhóm theo **lịch nhất quán toàn hệ thống** (như feature 003/004/005) để số liệu khớp giữa các thành viên. Khoảng tùy chỉnh: end ≥ start.
- **Cập nhật dữ liệu**: tính theo yêu cầu (khi mở màn / đổi khoảng); khi đang mở có thể làm mới nếu dữ liệu nền thay đổi (tái dùng cơ chế đồng bộ sẵn có) — không yêu cầu realtime nghiêm ngặt.
- **Trình bày mặc định** *(trả lời Open Question BR-008)*: bảng danh sách thành viên là dạng chính, có thể kèm biểu đồ cột so sánh; **thứ tự mặc định theo tổng Chi giảm dần** (chốt cụ thể ở /plan/UI). Bộ lọc chọn nhiều thành viên trên màn tổng quan **ngoài phạm vi v1** (chỉ drill-down một thành viên).
- **Vị trí màn**: mở rộng màn Báo cáo hiện tại của feature `005-reports` (`/reports`), thêm phần/khối "theo thành viên"; không thay đổi báo cáo tổng quan hiện có.
- **Ngoài phạm vi** *(theo Out of Scope BR-008)*: xác định người thực nhận/chi khác `created_by`; đổi quy thuộc giao dịch; chia một giao dịch cho nhiều thành viên; hạn mức/ngân sách/mục tiêu riêng theo thành viên; xếp hạng/chấm điểm hành vi; giới hạn quyền xem theo vai trò; xuất PDF/Excel/CSV.
- **Giao dịch của người đã rời hộ** *(chốt — Option A, giải quyết BR-008 Open Question #1)*: gộp vào nhóm "Thành viên cũ / Đã rời hộ" (một dòng bổ sung chỉ khi có dữ liệu) để đối soát tổng hộ (SC-001) luôn đúng; nhóm này drill-down được như một dòng thành viên (FR-010).
- **Owner & Target Quarter**: TBD (BR-008 Open Question) — không chặn spec.

## References / Truy vết

> Giữ các liên kết này cập nhật mỗi khi một artifact thay đổi (xem quy tắc lan truyền trong [`CLAUDE.md`](../../CLAUDE.md)).

- **Nguồn (BR)**: [`BR-008.md`](../business-requirements/BR-008.md) — Báo Cáo Thu Chi Theo Thành Viên (mã con BR-MBR-001…009).
- **Mở rộng báo cáo**: feature [`005-reports`](../005-reports/spec.md) / [`BR-006.md`](../business-requirements/BR-006.md) — kế thừa bộ lọc thời gian, tổng Thu/Chi/ròng toàn hộ và quy tắc tổng hợp.
- **Tiền đề**: feature [`002-transaction-tracking`](../002-transaction-tracking/spec.md) (giao dịch, `created_by`, ngày) · feature [`001-transaction-categorization`](../001-transaction-categorization/spec.md) (danh mục, cây cha–con).
- **Mô hình thực thể**: [`entity-model.md`](../entities/entity-model.md) — `TRANSACTION.created_by → USER.id`; tư cách thành viên qua `HOUSEHOLD_MEMBER`.
- **Use cases**: [`UC-MBR-01`](../use-cases/008-member-reports/uc-mbr-01-xem-bao-cao-theo-thanh-vien.md) (báo cáo theo thành viên) · [`UC-MBR-02`](../use-cases/008-member-reports/uc-mbr-02-xem-giao-dich-thanh-vien.md) (drill-down giao dịch một thành viên).

## History

- v1 (2026-08-10): tạo spec từ `BR-008.md` — báo cáo Thu/Chi/ròng theo từng thành viên hiện tại theo khoảng chọn được (US1), drill-down danh sách giao dịch của một thành viên (US2), tính đầy đủ & đối soát tổng hộ (US3). Quy thuộc theo `created_by`; chỉ đọc/suy ra. Open Question BR còn 1 điểm cần nghiệp vụ chốt: xử lý giao dịch của người đã rời hộ (FR-010, ảnh hưởng SC-001) — đánh dấu NEEDS CLARIFICATION; các Open Question UI (trình bày/thứ tự/bộ lọc) chốt bằng mặc định an toàn trong Assumptions.
- v2 (2026-08-10): chốt FR-010 theo **Option A** (giải quyết BR-008 Open Question #1) — gộp giao dịch của người đã rời hộ vào nhóm "Thành viên cũ / Đã rời hộ" (drill-down được) để đối soát tổng hộ luôn đúng; cập nhật FR-005, US3/AC2, SC-001 và Assumptions. Hết NEEDS CLARIFICATION.
