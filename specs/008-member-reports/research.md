# Research — Feature 008: Báo Cáo Thu Chi Theo Thành Viên (Phase 0)

**Nguồn**: [spec.md](./spec.md) · [BR-008](../business-requirements/BR-008.md) · [plan 005 (report module)](../005-reports/plan.md) · nền tảng D1–D41 (001–005).

> Không còn `[NEEDS CLARIFICATION]` trong Technical Context. FR-010 đã chốt (Option A). Các quyết định dưới đây tiếp tục đánh số **D42+** (kế thừa D1–D41).

## D42 — Tổng hợp theo thành viên bằng `GROUP BY created_by`

- **Decision**: Tính tổng Thu/Chi mỗi thành viên bằng một truy vấn `SUM(CASE WHEN type='INCOME' THEN amount END)` / `EXPENSE`, `GROUP BY created_by` trên giao dịch của hộ trong `[from, to)` (dùng đúng ranh giới nửa mở như 005: `>= from` và `< to+1day`). Net = income − expense suy ở biz.
- **Rationale**: Một lượt quét, nhất quán cách tổng hợp của 005 (SUM theo khoảng). `created_by` đã có index FK; đề xuất index `(household_id, created_by, transaction_date)` nếu execution plan cho thấy cần (BR-008 Constraints) — đo trước khi thêm.
- **Alternatives**: Quét từng thành viên (N truy vấn) — bị loại: N+1, chậm. View tổng hợp — bị loại: BR cấm bảng/aggregate mới, số liệu phải suy ra.

## D43 — Mọi thành viên hiện tại xuất hiện (kể cả 0 giao dịch)

- **Decision**: Lấy danh sách khởi đầu từ `household_members JOIN users` của hộ, **LEFT JOIN** sang kết quả tổng hợp D42 theo `created_by`; thành viên không khớp → income/expense = 0. Sắp mặc định theo **tổng Chi giảm dần** (Assumptions spec), tie-break theo tên.
- **Rationale**: Bảo đảm SC-003 (FR-004) — người chưa nhập giao dịch vẫn hiển thị 0/0/0 — mà không cần seed hàng rỗng ở biz.
- **Alternatives**: Chỉ liệt kê `created_by` xuất hiện trong giao dịch — bị loại: bỏ sót thành viên 0 giao dịch (vi phạm FR-004).

## D44 — "Thành viên cũ / Đã rời hộ" (Option A — bảo toàn đối soát)

- **Decision**: Giao dịch có `created_by` **không** nằm trong `household_members` hiện tại của hộ được gộp vào **một** dòng tổng hợp `former` (khóa sentinel, không phải user thật). Dòng này chỉ xuất hiện khi có ≥1 giao dịch như vậy. Do D42 quét **toàn bộ** giao dịch của hộ và chỉ tách theo "có/không là thành viên hiện tại", nên `Σ(current) + former = tổng hộ` — đối soát SC-001/FR-005 luôn đúng.
- **Rationale**: Giải quyết trực tiếp BR-008 Open Question #1 mà không phá vỡ bất biến đối soát (mâu thuẫn đã nêu trong review BR-008). Gộp một dòng thay vì liệt kê từng người đã rời → tránh hiển thị danh tính người ngoài hộ (nhất quán Regulatory BR-008).
- **Trạng thái thực tế**: Rời hộ **chưa được hiện thực** (quản lý hộ dùng dev seed — thuộc BR-005/quản lý hộ chờ nghiệp vụ). Vì vậy hôm nay bucket `former` thường rỗng; xử lý vẫn phải **đúng** để không sai đối soát khi tính năng rời hộ xuất hiện. Có test integration mô phỏng một `created_by` không có hàng `household_members`.
- **Alternatives**: (B) loại former khỏi báo cáo + nới lỏng đối soát — bị loại: phá vỡ SC-001. (C) mỗi người đã rời một dòng riêng — hoãn: lộ danh tính ngoài hộ, phức tạp hơn; có thể mở rộng sau nếu nghiệp vụ cần.

## D45 — Tái dùng bộ giải khoảng & ranh giới thời gian của 005

- **Decision**: Dùng lại `ResolvePeriod`/quy ước preset→`[from,to]` và ranh giới lịch nhất quán toàn hệ thống (tuần bắt đầu Thứ Hai; tháng/quý/năm dương lịch) đã có ở module `report` (005). Endpoint members/member nhận `?from=&to=` giống overview; validate end ≥ start tái dùng `parseRange`.
- **Rationale**: FR-002 yêu cầu "cùng ranh giới thời gian với tổng hợp toàn hộ" → phải dùng chính bộ giải khoảng của báo cáo hiện tại để số liệu khớp (SC-001).
- **Alternatives**: Bộ giải khoảng riêng — bị loại: nguy cơ lệch ranh giới → sai đối soát.

## D46 — Tên hiển thị: fallback email, không lộ mã thô

- **Decision**: Trả `display_name`; nếu rỗng dùng `email` (quy tắc entity-model USER, FR-009). Không trả `created_by` UUID cho FE hiển thị (chỉ dùng làm khóa drill-in). Tái dùng cách hiển thị người nhập của sổ giao dịch 002 (`ListItem.CreatedByName`).
- **Rationale**: Nhất quán FR-009/BR-MBR-009 và cách sổ chung 002 hiển thị người nhập.
- **Alternatives**: Hiển thị email luôn — bị loại: kém thân thiện; hiển thị UUID — bị loại: lộ mã thô (cấm).

## D47 — Drill-in tái dùng liệt kê giao dịch phân trang của module transaction

- **Decision**: Thêm trường **`CreatedBy *uuid.UUID`** vào `transaction/storage.ListFilter` (đọc, tương thích ngược — nil = không lọc). Drill-in một thành viên = `List(hid, paging, ListFilter{From, EndExcl, CreatedBy})`; drill-in `former` = biến thể lọc `created_by NOT IN (current members)` (một hàm storage riêng ở module report hoặc cờ trong filter). Trả `[]transactionmodel.ListItem` + total để phân trang (`common.Paging`), như sổ 002.
- **Rationale**: Tái dùng đúng bộ trường hiển thị (loại/số tiền/danh mục/mô tả/ngày giờ/tài khoản/người nhập) mà FR-006 yêu cầu, và cơ chế phân trang FR-007 đã có. Tránh nhân bản SQL liệt kê giao dịch.
- **Alternatives**: Báo cáo category (005) trả **toàn bộ** danh sách không phân trang — không tái dùng được cho member vì FR-007 bắt buộc phân trang. Viết truy vấn liệt kê mới trong report — bị loại: trùng lặp.

## D48 — Xác thực `:id` thuộc hộ; sentinel `former`

- **Decision**: `GET /api/reports/member/:id`: nếu `:id == "former"` → nhánh bucket former; ngược lại parse UUID và **kiểm tra là thành viên hiện tại của hộ** (tồn tại hàng `household_members(household_id, user_id)`); không phải → **404 Not Found** (không phân biệt "không tồn tại" vs "khác hộ" để tránh dò). Household scope lấy từ `middleware.HouseholdID(c)` (không nhận household/user id từ client).
- **Rationale**: FR-008/BR-MBR-008 (cô lập hộ, không nhận user ID ngoài hộ). 404 thay vì 403 để không rò rỉ sự tồn tại.
- **Alternatives**: Cho drill-in bất kỳ UUID rồi lọc theo hộ — bị loại: rò rỉ/định danh chéo hộ.

## D49 — Không bảng/migration; entity-model không đổi

- **Decision**: Toàn bộ là giá trị suy ra từ `transactions` + `users` + `household_members` (đều đã có). Không thêm bảng/cột/view/migration; `entity-model.md` không đổi.
- **Rationale**: BR-008 Constraints (không aggregate mới) + nhất quán 004/005.

## D50 — FE: mở rộng ReportsView/store, không thêm dependency

- **Decision**: Thêm khối "Theo thành viên" trong `ReportsView` với `MemberBreakdown.vue` (bảng Thu/Chi/ròng; biểu đồ **cột** so sánh **tùy chọn**, dùng lại `chart.js` đã có — thêm một wrapper `BarChart.vue` mỏng nếu chưa có, không thêm npm dep). Drill-in bằng `MemberTransactionsPanel.vue` (danh sách phân trang, empty state, nút quay lại giữ khoảng qua state store). Mở rộng `stores/reports.ts` với state `members`, `memberDetail`, actions `fetchMembers(from,to)`, `fetchMember(id,from,to,page)`.
- **Rationale**: Tái dùng hạ tầng 005 (TimeRangePicker, wrapper chart, store) → không dependency mới (Constitution/Complexity sạch). Bảng là dạng trình bày chính (Assumptions); cột chỉ bổ trợ.
- **Alternatives**: Màn riêng cho báo cáo thành viên — bị loại: BR nói "mở rộng màn Báo cáo hiện tại"; thêm `vue-chartjs` — bị loại: giữ tối thiểu như 005.

## Tổng hợp

- **Không** unknowns còn lại; **không** dependency/bảng/migration mới.
- Rủi ro chính: đối soát tổng (giảm thiểu bằng D42+D44 quét toàn bộ giao dịch một lần) và hiệu năng khoảng lớn (giảm thiểu bằng chỉ số `(household_id, created_by, transaction_date)` nếu execution plan cần — đo trước).
