# Phase 0 — Research: Thiết Lập Ngân Sách (Go + Vue)

**Date**: 2026-07-13 · **Feature**: 003-budgeting · **Plan**: [plan.md](./plan.md)

> Kế thừa toàn bộ **D1–D12** của [research 001](../001-transaction-categorization/research.md) và
> **D13–D19** của [research 002](../002-transaction-tracking/research.md) (stack Go+Vue, không RLS/trigger,
> bất biến ở biz, pubsub→wshub, mốc lạc quan `updated_at`). Dưới đây chỉ là điểm MỚI của 003, đánh số
> tiếp **D20–D26**. Ngân sách chỉ **đọc** giao dịch/danh mục — không thêm ràng buộc lên dữ liệu nguồn.

## D20. Mô hình ngân sách & kỳ áp dụng

- **Decision**: Bảng `budgets` (00008): `id, household_id, type ∈ {CATEGORY,TOTAL}, category_id (null khi TOTAL), limit_amount numeric>0, period_type ∈ {MONTHLY,WEEKLY,ONE_TIME}, start_date, end_date (chỉ ONE_TIME), status ∈ {ACTIVE,ENDED}, created_by, created_at, updated_at`. **Không lưu tiến độ**. Kỳ hiện tại suy ra từ `now()` + `period_type`:
  - MONTHLY → `period_key = 'YYYY-MM'` (tháng dương lịch);
  - WEEKLY → `period_key = 'IYYY-IW'` (tuần ISO, bắt đầu Thứ Hai);
  - ONE_TIME → `period_key = 'once'`, cửa sổ = `[start_date, end_date]`, `status=ENDED` khi `now() > end_date`.
  **Tự lặp lại** (MONTHLY/WEEKLY) = **cùng một hàng** `budgets`, `period_key` đổi theo lịch → tiến độ & cảnh báo reset tự nhiên vì đều khóa theo `period_key`.
- **Rationale**: Auto-repeat không cần nhân bản hàng mỗi kỳ (FR-004; Assumptions "tự khởi tạo lại"); `period_key` là khóa duy nhất để cả tiến độ lẫn chống-trùng-cảnh-báo cùng "reset" khi sang kỳ (edge case "sang kỳ mới"). ONE_TIME có `status=ENDED` để "giữ xem lại, không cảnh báo nữa" (edge case).
- **Alternatives**: nhân bản một hàng/kỳ (bác — bùng nổ dữ liệu, khó sửa "giới hạn cho mọi kỳ"); lưu cứng `spent` (bác — vi phạm FR-006 "giá trị suy ra", drift khi sửa giao dịch quá khứ).

## D21. Tiến độ — giá trị suy ra, tính ở biz khi đọc

- **Decision**: `spent` tính ở **biz** khi đọc (`GET /api/budgets`), KHÔNG lưu, KHÔNG dùng DB view: tổng `transactions.amount` với `type='EXPENSE'`, `household_id` của hộ, `transaction_date` trong cửa sổ kỳ hiện tại (D20), và:
  - **CATEGORY**: `category_id ∈ {danh mục ngân sách} ∪ {danh mục con một cấp của nó}` (cây danh mục 001 chỉ một cấp — D10/001) → dùng `WHERE category_id IN (:id, :child_ids…)`;
  - **TOTAL**: mọi giao dịch EXPENSE của hộ trong kỳ.
  `percent = round(spent / limit_amount × 100)`. Trả kèm `spent, percent, period_key, alerts[]`.
- **Rationale**: Cửa sổ kỳ phụ thuộc `now()`+`period_type` và subtree danh mục — logic này rõ ràng, unit-test dễ ở Go hơn là nhồi vào một view khổng lồ; đọc lại tổng đảm bảo 100% khớp sau sửa/xóa/đổi danh mục quá khứ (SC-003); loại EXPENSE-only theo Assumptions ("giao dịch Thu không bao giờ tính"). Quy mô vài nghìn giao dịch/hộ → truy vấn `SUM` có index `idx_transactions_account`/category đủ nhanh.
- **Alternatives**: DB view kiểu `account_balances` (bác — điều kiện kỳ theo `now()` + subtree khó biểu diễn gọn, kém test); cột `spent` cập nhật khi giao dịch đổi (bác — 4 đường mutate giao dịch × mọi ngân sách liên quan → khó chứng minh SC-003, giống lý do bác cột balance ở D14).

## D22. Ràng buộc duy nhất ngân sách trong kỳ (FR-010)

- **Decision**: Mỗi hộ tối đa **một ngân sách ACTIVE mỗi (danh mục, period_type)** và **một ngân sách TOTAL ACTIVE mỗi period_type**. Thực thi 2 lớp: **partial unique index** (phòng tuyến cuối) + **biz check** (thông báo `BUDGET_DUPLICATE` kèm id ngân sách hiện có để UI chỉ tới — UC-BGT-01 E3 / UC-BGT-02 E2):
  - `unique (household_id, category_id, period_type) where status='ACTIVE' and type='CATEGORY'`;
  - `unique (household_id, period_type) where status='ACTIVE' and type='TOTAL'`.
- **Rationale**: Tránh nhập nhằng "một khoản chi tính vào ngân sách nào" (Assumptions); "cùng kỳ" hiểu là **cùng `period_type`** (ngân sách tổng vẫn song song, độc lập với ngân sách danh mục — FR-002/FR-010). Partial index chỉ chặn trên hàng ACTIVE nên ngân sách ONE_TIME đã ENDED không cản tạo mới.
- **Alternatives**: chặn chỉ ở biz (bác — race giữa 2 request đồng thời có thể lọt; index là phòng tuyến cuối, nhất quán D3/001); duy nhất theo `period_key` cụ thể (bác — với MONTHLY tự lặp sẽ cho tạo trùng ở kỳ kế, phá ý nghĩa "một ngân sách/danh mục").

## D23. Máy trạng thái cảnh báo + bảng `budget_alerts` (FR-007/008/009)

- **Decision**: Bảng `budget_alerts` (00009): `id, budget_id (FK, on delete cascade), period_key, level ∈ {THRESHOLD_80, OVER_100}, over_amount numeric (null trừ OVER_100), fired_at`. Chống trùng: `unique (budget_id, period_key, level)`. Đánh giá lại (D24) cho mỗi ngân sách ACTIVE: tính `percent`;
  - `percent ≥ 80` và chưa có hàng `(budget, period_key, THRESHOLD_80)` → **insert** (phát 80%);
  - `percent > 100` và chưa có hàng OVER_100 → **insert** với `over_amount = spent − limit` (phát vượt);
  - `percent < 80` mà tồn tại hàng THRESHOLD_80 của kỳ → **delete** (nạp lại — re-arm); tương tự `percent ≤ 100` xóa OVER_100.
  Cảnh báo đang hoạt động của kỳ hiện tại được **embed vào `GET /api/budgets`** để UI hiển thị (badge amber ≥80%, đỏ khi vượt) cho mọi thành viên.
- **Rationale**: `unique(budget, period_key, level)` cho "tối đa một lần/kỳ"; delete-khi-tụt-dưới cho "phát lại nếu vượt lại" (US3 #4); `period_key` đảm bảo sang kỳ mới trạng thái tự sạch; `over_amount` phục vụ SC-005 (hiển thị đúng số tiền vượt). Cascade xóa theo budget → xóa ngân sách gỡ luôn cảnh báo (UC-BGT-06).
- **Alternatives**: cột `notified_80 bool`/`notified_over bool` trên `budgets` (bác — không giữ được `over_amount` theo kỳ, khó reset per period_key, khó truy vết thời điểm phát); tính cảnh báo thuần khi đọc không lưu (bác — mọi lần mở màn sẽ hiện lại, không "chống trùng"/không phân biệt đã tụt-dưới-rồi-vượt-lại).

## D24. Kích hoạt đánh giá cảnh báo — subscriber pubsub (giữ chiều phụ thuộc)

- **Decision**: Một **subscriber budget** (goroutine) `Subscribe()` pubsub sẵn có (D8). Khi nhận `transactions_changed` (do mọi mutation giao dịch của 002 phát — đã có) hoặc sự kiện CRUD ngân sách (module budget tự phát nội bộ), nó **recompute toàn phần** tiến độ + máy trạng thái cảnh báo (D23) cho các ngân sách ACTIVE của hộ liên quan trong một DB transaction, rồi `Publish(budgets_changed)`. Subscriber **không** phản ứng với `budgets_changed` (tránh vòng lặp) — topic đó chỉ để subscriber-WS (`component/subscriber`) đẩy về client.
- **Rationale**: Giữ đúng chiều phụ thuộc — module transaction KHÔNG biết budget (budget là hạ nguồn, chỉ đọc); tái dùng pubsub→wshub đã kiểm chứng (D8) thay vì cơ chế mới. Recompute **idempotent**: nếu pubsub lỡ bỏ một message lúc quá tải (semantics "drop if slow" — pubsub.go), lần mutation kế tiếp vẫn phát lại cảnh báo còn treo (percent vẫn trên mức, level chưa có hàng) → tự lành; kênh buffer 64 + tần suất theo hộ khiến drop gần như không xảy ra thực tế.
- **Alternatives**: gọi đồng bộ budget từ transaction biz (bác — đảo chiều phụ thuộc, transaction phải biết budget); dependency-inversion `BudgetNotifier` interface tiêm vào transaction (ghi nhận là phương án nâng cấp nếu sau này cần đảm bảo phát cảnh báo tức thời tuyệt đối, đổi lại thêm khớp nối); cron quét định kỳ (bác — trễ, phá SC-006 ≤ 5s).

## D25. Sửa/xóa ngân sách đồng thời *(dùng lại D6/D17)*

- **Decision**: `budgets.updated_at` (GORM hook) làm mốc lạc quan; `PATCH`/`DELETE` kèm `expected_updated_at`; storage conditional UPDATE/DELETE — 0 hàng → tra tồn tại để phân biệt `CONCURRENCY_CONFLICT` (409, kèm bản mới nhất) vs `RECORD_GONE` (404). Sửa chạy **cùng bộ xác thực như tạo** (D21/D22 + limit>0 + kỳ hợp lệ); sau khi sửa giới hạn/kỳ/danh mục → phát sự kiện để evaluator tính lại cảnh báo theo giá trị mới (UC-BGT-05 AC-1).
- **Rationale**: Nhất quán cơ chế đã kiểm chứng ở 001 (#16) và 002 (D17); UI tái dùng đúng pattern xử lý 409/404 (UC-BGT-05 E3/E4). Xóa luôn qua dialog xác nhận ở UI (UC-BGT-06).
- **Alternatives**: version int (cột thừa — như D17); khóa bi quan (kém UX).

## D26. Realtime & hiển thị cảnh báo in-app *(mở rộng D8/D16)*

- **Decision**: Thêm hằng `TopicBudgetsChanged = "budgets_changed"` vào `common/const.go`. `component/subscriber` (pubsub→wshub) tự động đẩy topic mới (không cần sửa — nó forward mọi message). Client `budgets.ts` nghe `budgets_changed` (và `transactions_changed`) → refetch `GET /api/budgets` (đã embed spent/percent/alerts) → cập nhật thanh tiến độ + badge cảnh báo cho **mọi thành viên** đang mở, mục tiêu ≤ 5s (SC-006). Cảnh báo là **in-app** (badge/banner + widget Tổng quan) — không push/email (Assumptions).
- **Rationale**: Kênh realtime đã có; embed alerts vào cùng payload budgets tránh thêm endpoint; nhất quán cách 002 refetch theo `transactions_changed` (D16).
- **Alternatives**: endpoint `/api/budget-alerts` riêng (bác — thừa; alerts luôn gắn ngân sách hiện tại); SSE/polling (bác — đã chuẩn hóa WS).

## Tổng hợp

Toàn bộ điểm mở đã chốt (D20–D26 + kế thừa D1–D19), không còn NEEDS CLARIFICATION. Migrations thêm:
`00008_budgets.sql`, `00009_budget_alerts.sql`. Điểm chờ nghiệp vụ (không chặn triển khai MVP, đã chốt
mặc định an toàn trong spec Assumptions): ngưỡng 80% cố định (có thể cấu hình sau — cần đổi BR trước);
kênh cảnh báo chỉ in-app (push/email sau); sửa/xóa ngân sách là vòng đời suy luận (FR-012 — cần nghiệp
vụ xác nhận); "cùng kỳ" của FR-010 hiểu theo `period_type` (D22 — ghi rõ để review). Baseline SC-007
(≥40% hộ có ≥1 ngân sách) cần nghiệp vụ xác nhận.

## History

- v1 (2026-07-13, claude): Phase 0 cho stack Go+Vue — D20 (mô hình ngân sách + period_key tự lặp),
  D21 (tiến độ suy ra ở biz, không view), D22 (partial unique index + biz cho FR-010), D23 (máy trạng
  thái cảnh báo + bảng budget_alerts), D24 (evaluator qua subscriber pubsub, idempotent), D25 (mốc lạc
  quan dùng lại D6/D17), D26 (topic budgets_changed + in-app). Kế thừa D1–D19.
