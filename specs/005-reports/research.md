# Phase 0 — Research: Báo Cáo và Phân Tích Trực Quan (Go + Vue)

**Date**: 2026-07-15 · **Feature**: 005-reports · **Plan**: [plan.md](./plan.md)

> Kế thừa **D1–D33** (research 001/002/003/004): stack Go+Vue, không RLS/trigger, giá trị suy ra ở biz, phạm vi hộ ở middleware, cửa sổ kỳ `ResolvePeriod`, gộp danh mục con vào cha (D29). Dưới đây là điểm MỚI của 005 (**D34–D41**). Chỉ **đọc** giao dịch/danh mục — không bảng/migration.

## D34. Module report + 2 endpoint (read-only)

- **Decision**: Module `module/report` phơi hai endpoint: `GET /api/reports/overview?from=&to=` (tổng thu/chi/ròng + phân bổ danh mục + trend thu/chi) và `GET /api/reports/category/:id?from=&to=` (tổng + danh sách giao dịch + trend của một danh mục). Cả hai chỉ SELECT/SUM trên transactions/categories, scope theo hộ (middleware D5/001). Thay màn placeholder `/reports` của feature 004 bằng màn thật.
- **Rationale**: Hai màn/nhu cầu rõ rệt (tổng quan vs chi tiết một danh mục) → hai endpoint tách bạch, mỗi cái một round-trip. Nhất quán triết lý "giá trị suy ra ở biz khi đọc" (D21/004).
- **Alternatives**: một endpoint đa năng (bác — payload phình, khó cache/tối ưu riêng); GraphQL (bác — thừa, dự án dùng REST).

## D35. Mô hình khoảng thời gian (preset + tùy chỉnh)

- **Decision**: API nhận **`from`/`to`** (ngày, inclusive) đã giải sẵn; preset (tuần/tháng/quý/năm) do **FE** quy về [from,to] rồi gọi API. Backend validate `to ≥ from` (lỗi rõ ràng nếu sai — FR-008). Ranh giới: tuần bắt đầu Thứ Hai; tháng/quý/năm theo dương lịch (dùng lại logic kiểu `ResolvePeriod` — mở rộng cho quý/năm ở biz report).
- **Rationale**: FE nắm ngữ cảnh "hôm nay" của người dùng để suy preset; API nhận [from,to] thuần → đơn giản, dễ test, dùng lại cho cả tùy chỉnh. Validate end≥start ở BE là phòng tuyến (SC + FR-008).
- **Alternatives**: gửi preset tên xuống API (bác — nhân đôi logic lịch ở BE+FE; nhưng BE vẫn có helper preset cho tiện test); chỉ tùy chỉnh, bỏ preset (bác — trái FR-001).

## D36. Chọn đơn vị gom nhóm chuỗi thời gian (trend)

- **Decision**: Đơn vị gom nhóm cho biểu đồ xu hướng chọn theo **độ dài khoảng**: span ≤ 31 ngày → **theo ngày**; ≤ 92 ngày (≈ quý) → **theo tuần** (ISO, Thứ Hai); còn lại → **theo tháng**. Gom nhóm bằng SQL `date_trunc(unit, transaction_date)` (hoặc tương đương cho tuần ISO), trả về danh sách mốc `{bucket, income, expense}` (overview) hoặc `{bucket, amount}` (category), **điền đủ mốc trống = 0** ở biz để đường liền mạch.
- **Rationale**: Giữ số điểm trên biểu đồ hợp lý (dễ đọc + nhẹ) bất kể khoảng dài/ngắn; điền mốc 0 để biểu đồ không "nhảy" khi thiếu dữ liệu. Gom nhóm ở SQL để đạt SC-002 (≤2s) với khoảng dài.
- **Alternatives**: luôn theo ngày (bác — khoảng năm → 365 điểm, nặng & rối); để FE tự gom (bác — kéo toàn bộ giao dịch về FE, phá SC-002).

## D37. Phân bổ chi tiêu theo danh mục (overview)

- **Decision**: Chỉ giao dịch **EXPENSE** trong [from,to], gộp **danh mục con một cấp vào cha** (quy về `COALESCE(parent.id, cat.id)` như D29/004), trả `{category_id, category_name, category_hidden, amount, percent}`, `percent = round(amount/tổng_chi × 100)`, sắp giảm dần. Dùng cho biểu đồ tròn/donut + chú giải.
- **Rationale**: Nhất quán tuyệt đối với "Chi tiêu theo danh mục" của feature 004 (cùng cách gộp/percent) — người dùng thấy số khớp giữa hai màn. 
- **Alternatives**: liệt kê danh mục con riêng (bác — lệch 004, rối biểu đồ tròn); top-N + "khác" (ghi nhận là tùy chọn hiển thị FE, mặc định trả tất cả).

## D38. Báo cáo chi tiết theo một danh mục

- **Decision**: `GET /api/reports/category/:id?from=&to=` trả: `total` (Σ chi của danh mục + **subtree một cấp** trong khoảng), `transactions[]` (dùng lại `ListItem` của 002 — embed tên danh mục · tài khoản · người nhập, lọc theo `category_id IN (id + con)` + khoảng, mới nhất trước), và `trend[]` (chi của danh mục theo mốc thời gian — D36). `id` phải thuộc hộ (ngoài hộ → 404).
- **Rationale**: Đào sâu từ phân bổ (D37) xuống giao dịch cụ thể (BR-RPT-004/005) + xu hướng danh mục (BR-RPT-006); tái dùng subtree + ListItem để khớp các màn khác.
- **Alternatives**: phân trang danh sách giao dịch (ghi nhận — nếu một danh mục có quá nhiều giao dịch trong khoảng dài, thêm phân trang sau; MVP trả trọn khoảng, giới hạn hợp lý theo hiệu năng).

## D39. Biểu đồ — thêm dependency `chart.js` (quyết định của người dùng)

- **Decision**: Thêm **`chart.js`** (FE) làm thư viện biểu đồ, bọc bằng **wrapper Vue mỏng tự viết** (`DonutChart.vue`, `LineChart.vue`) — KHÔNG thêm `vue-chartjs` (giảm số dependency). Đăng ký tree-shake chỉ các controller cần (Doughnut/Pie, Line, cùng scale/element) để giữ bundle nhỏ. Đây là **dependency FE mới đầu tiên** kể từ re-platform.
- **Rationale**: BR-006 nêu rõ "cần thư viện biểu đồ phù hợp"; người dùng chọn dùng thư viện thay vì tự vẽ SVG để có trục/nhãn/tooltip/legend hoàn chỉnh với ít mã và ít lỗi. Wrapper mỏng cô lập chart.js sau một API nội bộ (dễ thay/test).
- **Alternatives**: tự vẽ SVG (bác theo lựa chọn người dùng — nhiều mã, dễ sai, kém trau chuốt); vue-chartjs (bác — thêm dependency thứ hai không cần thiết khi wrapper mỏng là đủ); ECharts/ApexCharts (bác — nặng hơn nhu cầu).

## D40. Độ tươi dữ liệu (freshness)

- **Decision**: Báo cáo tính **theo yêu cầu** (khi mở màn / đổi khoảng / chọn danh mục). Khi đang mở, store `reports.ts` **có thể** refetch khi nhận `transactions_changed`/`categories_changed` (dùng lại `useInvalidation` — D8/D33) để số liệu không cũ; không yêu cầu realtime nghiêm ngặt (báo cáo là read-heavy, tần suất đổi thấp trong lúc xem).
- **Rationale**: Trả lời Open Question BR (realtime hay chu kỳ) bằng "on-demand + làm mới cơ hội"; tránh tải server bằng recompute liên tục cho khoảng lớn.
- **Alternatives**: realtime đầy đủ như 004 (bác — báo cáo nặng hơn, ít giá trị realtime); chỉ tải một lần không refetch (bác — dễ hiển thị số cũ sau khi người khác nhập giao dịch).

## D41. Hiệu năng & tổng hợp bằng SQL

- **Decision**: Mọi tổng hợp (SUM thu/chi, phân bổ danh mục, trend theo bucket) chạy bằng **SQL** (`SUM ... GROUP BY ...`, `date_trunc`) với index sẵn có (`idx_transactions_*`, `category_id`); KHÔNG kéo giao dịch thô về app để tính (trừ danh sách giao dịch của một danh mục — D38). Mục tiêu ≤ 2s cho khoảng 1 tháng (SC-002).
- **Rationale**: Gom nhóm/tổng ở DB tận dụng index + tránh truyền dữ liệu lớn; khoảng năm vẫn nhẹ vì chỉ trả các bucket đã tổng hợp.
- **Alternatives**: materialized view/bảng tổng hợp trước (bác — thêm hạ tầng/độ phức tạp, dữ liệu mỗi hộ nhỏ, YAGNI); tính ở app (bác — phá SC-002 với khoảng dài).

## Tổng hợp

Chốt D34–D41 + kế thừa D1–D33; không còn NEEDS CLARIFICATION. **Không migration.** Dependency mới duy nhất: `chart.js` (FE). Điểm để nghiệp vụ xác nhận (không chặn): số hiệu BR (file BR-006 ↔ ID "BR-004"); baseline SC-006; ngưỡng đơn vị gom nhóm (đã chốt mặc định D36 — có thể tinh chỉnh); có cần phân trang danh sách giao dịch theo danh mục cho khoảng rất dài (D38).

## History

- v1 (2026-07-15): Phase 0 cho stack Go+Vue — D34 (module report + 2 endpoint), D35 (khoảng preset+tùy chỉnh, from/to), D36 (chọn đơn vị gom nhóm trend + điền mốc 0), D37 (phân bổ danh mục gộp con, nhất quán 004), D38 (chi tiết theo danh mục: tổng+danh sách+trend, subtree), D39 (**thêm chart.js** + wrapper mỏng — quyết định người dùng), D40 (on-demand + làm mới cơ hội), D41 (tổng hợp bằng SQL cho SC-002). Kế thừa D1–D33.
