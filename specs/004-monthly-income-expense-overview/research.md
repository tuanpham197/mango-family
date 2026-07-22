# Phase 0 — Research: Màn Tổng Quan Đầy Đủ (Go + Vue)

**Date**: 2026-07-14 · **Feature**: 004-monthly-income-expense-overview · **Plan**: [plan.md](./plan.md)

> Kế thừa **D1–D12** (research 001), **D13–D19** (research 002), **D20–D26** (research 003): stack Go+Vue, không RLS/trigger, bất biến/giá trị suy ra ở biz, pubsub→wshub realtime, phạm vi hộ ở middleware. Dưới đây là điểm MỚI của 004 (**D27–D33**). Feature chỉ **đọc** dữ liệu 001/002/003 — không thêm ràng buộc/bảng.

## D27. Một endpoint tổng hợp `GET /api/overview`

- **Decision**: Thêm module đọc `module/overview` phơi **một** endpoint `GET /api/overview` trả về mọi giá trị suy ra của màn Tổng quan trong một round-trip: `net_worth`, `net_worth_change_percent` (so tháng trước), `month.income`, `month.expense`, `month.net`, `category_spending[]`, `recent_transactions[]`. Tóm tắt **ngân sách** KHÔNG nhồi vào đây — dùng lại `GET /api/budgets` (widget 003 đã tự fetch).
- **Rationale**: Một round-trip cho toàn màn (trừ ngân sách vốn đã có endpoint riêng) → hiển thị nhanh khi mở app (SC-007); nhất quán triết lý "giá trị suy ra ở biz khi đọc" (D21). Tách ngân sách vì widget 003 đã hoàn chỉnh, tránh trùng lặp + giữ module độc lập.
- **Alternatives**: gọi nhiều endpoint rời (net worth, month, category, recent) — bác (nhiều round-trip, khó đồng bộ một thời điểm); nhồi cả budget vào overview — bác (lặp logic 003, phá tính tự chứa của module budget).

## D28. Tổng tài sản ròng + % so với tháng trước

- **Decision**: `net_worth` = `Σ account_balances.balance` của hộ (view suy ra của 002 — D14). **% so tháng trước** = `(net_worth − net_worth_prev_end) / |net_worth_prev_end| × 100`, trong đó `net_worth_prev_end` = `Σ initial_balance của mọi tài khoản + Σ bút toán có dấu (INCOME +, EXPENSE −) có transaction_date < đầu tháng hiện tại`. Nếu `net_worth_prev_end = 0` (hoặc hộ chưa có dữ liệu tháng trước) → trả `net_worth_change_percent = null` (UI ẩn mũi tên/không hiển thị %).
- **Rationale**: Số dư hiện tại đã có sẵn (view). Tài sản ròng "cuối tháng trước" suy ra thuần từ giao dịch trước mốc đầu tháng + số dư đầu kỳ tài khoản — không cần lưu lịch sử số dư. Nhất quán "giá trị suy ra". `null` khi mẫu số 0 để tránh chia 0 và hiển thị sai.
- **Alternatives**: lưu snapshot số dư theo tháng (bác — thêm bảng/độ phức tạp, YAGNI); dùng chênh lệch Thu−Chi tháng làm "% thay đổi" (bác — sai ngữ nghĩa "tài sản ròng so tháng trước", vì tài sản ròng còn phụ thuộc số dư đầu kỳ tài khoản).

## D29. Chi tiêu theo danh mục (mục MỚI, trước Ngân sách)

- **Decision**: `category_spending[]` = với mỗi **danh mục cha** có phát sinh **Chi (EXPENSE)** trong tháng hiện tại của hộ: `{category_id, category_name, category_hidden, amount, percent}` với `amount = Σ chi của danh mục cha + các danh mục con một cấp` (gộp cây — nhất quán D21/003), `percent = round(amount / tổng_chi_tháng × 100)`. Chỉ liệt kê danh mục có `amount > 0`, **sắp giảm dần** theo amount. Mặc định trả **tất cả** danh mục có chi (không cắt top-N ở API; UI có thể rút gọn hiển thị nếu cần). Danh mục ẩn vẫn tính, kèm cờ `category_hidden`.
- **Rationale**: Trả lời trực tiếp yêu cầu "chi tiêu của từng danh mục trước budget"; gộp con vào cha để con số khớp cách ngân sách tính (tránh nhầm lẫn giữa hai màn). Sắp giảm dần đưa danh mục tốn nhất lên đầu. Trả tất cả để nguồn dữ liệu đầy đủ; quyết định rút gọn hiển thị thuộc UI.
- **Alternatives**: liệt kê từng danh mục con riêng (bác — không khớp cách gộp của ngân sách, rối); cắt top-N ở API (bác — mất dữ liệu cho các cách hiển thị khác, để UI quyết định).

## D30. Giao dịch gần đây embed trong overview

- **Decision**: `recent_transactions[]` = **5** giao dịch mới nhất của hộ (mới nhất trước), dùng lại truy vấn/`ListItem` của storage transaction 002 (embed tên danh mục · tài khoản · người nhập). Chỉ để xem nhanh (read-only); thao tác đầy đủ ở sổ `/ledger`.
- **Rationale**: Tái dùng đúng dạng dữ liệu sổ 002 (đã có join tên) → nhất quán hiển thị; embed trong overview để render một màn không cần gọi thêm. Con số 5 khớp mật độ wireframe.
- **Alternatives**: gọi `GET /api/transactions?page_size=5` riêng từ FE (chấp nhận được, nhưng thêm round-trip); embed toàn bộ sổ (bác — thừa, đã có màn sổ).

## D31. Trang mặc định & định tuyến (`/` = Tổng quan)

- **Decision**: `/` = **DashboardView** (Tổng quan) làm trang chủ; sổ giao dịch chuyển sang **`/ledger`**. Guard định tuyến: vào ứng dụng ở gốc / sau đăng nhập không kèm đích → Tổng quan (`/`). Mọi màn vẫn truy cập qua nav/liên kết trực tiếp; các liên kết có đích cụ thể được tôn trọng (FR-001).
- **Rationale**: Trung thành FR-001 "địa chỉ gốc → Tổng quan" và yêu cầu "mặc định vào page đầu tiên là trang tổng quan". Đây là **thay đổi so với quyết định tạm ở feature 003** (khi đó giữ `/` = ledger, Tổng quan ở `/overview` để không đụng e2e 001/002) — nay người dùng yêu cầu rõ Tổng quan là trang chủ, nên **đảo lại**: chấp nhận **cập nhật điều hướng trong e2e 001/002/003** (đổi `goto('/')` xem sổ → `goto('/ledger')`, khẳng định URL sau lưu giao dịch → `/ledger`, helper login đáp ở Tổng quan — DashboardView đã có `household-name`).
- **Alternatives**: giữ `/` = ledger, chỉ redirect sau đăng nhập → `/overview` (bác — vào gốc `/` vẫn ra sổ, không thỏa "địa chỉ gốc → Tổng quan"; hơn nữa vẫn phải sửa nhiều e2e vì đăng nhập không còn đáp ở sổ); route `/overview` song song `/` (bác — hai trang chủ, khó hiểu). Route `/overview` cũ có thể **redirect → `/`** để không vỡ liên kết cũ.

## D32. Thanh điều hướng 5 mục & lối vào Quản lý danh mục

- **Decision**: Nav dưới cùng theo wireframe: **Tổng quan (`/`) · Giao dịch (`/ledger`) · ＋ (thêm giao dịch) · Ngân sách (`/budgets`) · Báo cáo**. **Báo cáo** là **placeholder** (BR-004 chưa xây) → mở ra màn "sắp có". **Quản lý danh mục** (feature 001) giữ nguyên route `/categories` và vẫn truy cập được qua **lối phụ** (link ở phần đầu màn Tổng quan hoặc trong luồng tạo danh mục nhanh của form giao dịch — đã có), không nằm trên thanh nav chính.
- **Rationale**: Trung thành 100% nhãn nav wireframe mà không mất chức năng Danh mục (vốn còn cần). Không xóa route/chức năng danh mục — chỉ đổi vị trí lối vào.
- **Alternatives**: bỏ hẳn Danh mục khỏi app (bác — mất chức năng 001, cần nghiệp vụ xác nhận); giữ 6 mục nav (Danh mục + Báo cáo) (bác — lệch wireframe "follow 100%"). *(Nếu nghiệp vụ muốn bỏ hẳn Danh mục khỏi điều hướng chính vĩnh viễn — cần xác nhận; hiện chỉ chuyển thành lối phụ.)*

## D33. Realtime cho màn Tổng quan

- **Decision**: Store `overview.ts` refetch `GET /api/overview` khi nhận **bất kỳ** topic ảnh hưởng: `transactions_changed`, `accounts_changed`, `budgets_changed`, `categories_changed` (dùng lại `useInvalidation` — D8/D26). Widget ngân sách vẫn tự refetch theo cơ chế 003. Mục tiêu ≤ 5s (SC-006).
- **Rationale**: Mọi phần của overview đều suy ra từ giao dịch/tài khoản/danh mục/ngân sách; nghe đủ các topic đảm bảo mọi con số cập nhật realtime cho mọi thành viên. Kênh WS đã kiểm chứng ở 001/002/003.
- **Alternatives**: chỉ nghe `transactions_changed` (bác — bỏ lỡ đổi tên danh mục/ngân sách); polling (bác — trễ, đã chuẩn hóa WS).

## Tổng hợp

Toàn bộ điểm mở đã chốt (D27–D33 + kế thừa D1–D26), không còn NEEDS CLARIFICATION. **Không migration** (chỉ đọc). Điểm để review nghiệp vụ (không chặn triển khai): cách hiển thị rút gọn "Chi tiêu theo danh mục" (top-N vs tất cả — mặc định tất cả); có bỏ hẳn "Quản lý danh mục" khỏi điều hướng chính khi thêm "Báo cáo" không (mặc định giữ lối phụ); công thức "% tài sản ròng so tháng trước" (đã chốt D28, ghi rõ để review).

## History

- v1 (2026-07-14): Phase 0 cho stack Go+Vue — D27 (endpoint tổng hợp `/api/overview`), D28 (tài sản ròng + % so tháng trước suy ra), D29 (chi tiêu theo danh mục gộp cây, trước ngân sách), D30 (giao dịch gần đây embed), D31 (đổi trang mặc định `/`=Tổng quan, ledger→`/ledger`; đảo quyết định tạm của 003), D32 (nav 5 mục + Báo cáo placeholder + lối phụ Danh mục), D33 (realtime đa topic). Kế thừa D1–D26.
