# Phase 0 — Research: Ghi Chép Thu Nhập và Chi Phí (Go + Vue re-plan)

**Date**: 2026-07-10 · **Feature**: 002-transaction-tracking · **Plan**: [plan.md](./plan.md)

> ♻️ Thay thế bản research Supabase 2026-07-06 (R15–R21 — xem git history). Kế thừa toàn bộ
> **D1–D12** của [research 001 (Go+Vue)](../001-transaction-categorization/research.md); dưới đây
> chỉ là điểm MỚI của 002, đánh số tiếp **D13–D19**. Quyết định nghiệp vụ của bản cũ được giữ
> nguyên ngữ nghĩa và ghi chú *(carry-over Rxx)*.

## D13. Tài khoản mặc định của hộ *(carry-over R17 — thu gọn BR-005)*

- **Decision**: Bảng `accounts` (id, household_id, name, type ∈ CASH/BANK/EWALLET/CREDIT, `initial_balance` default 0, created_by, created_at). Tài khoản mặc định **"Tiền mặt" (CASH)** tạo bằng **logic app**: mở rộng `SeedDefaults` của `module/household` (cùng chỗ seed danh mục mặc định — D9); `cmd/seed` backfill cho hộ dev. KHÔNG trigger DB, KHÔNG seed trong migration.
- **Rationale**: FR-006/BR-TRK-006 cần tài khoản ngay nhưng BR-005 chưa triển khai; một chỗ duy nhất định nghĩa "hộ mới có gì" — dễ test biz, nhất quán với D9; `initial_balance` giữ chỗ cho BR-005 mà không phá công thức số dư.
- **Alternatives**: trigger after-insert on households (bản cũ — bỏ vì logic app là writer duy nhất, D3); hoãn account sang BR-005 (bác — BR-TRK-006 In Scope).

## D14. Số dư — view suy ra *(carry-over R16, giữ nguyên)*

- **Decision**: View **`account_balances`**: `balance = initial_balance + Σ(amount × CASE type WHEN 'INCOME' THEN 1 ELSE -1 END)` theo `account_id` (+ `household_id` để lọc). View thường (không materialized), tạo trong `00006_accounts.sql` (bọc `-- +goose StatementBegin/End`); GORM map model read-only; index `idx_transactions_account`.
- **Rationale**: SC-004 đòi 100% khớp tổng bút toán sau MỌI thao tác — view đúng theo định nghĩa, không rủi ro drift; thang nghìn giao dịch/hộ đủ nhanh. (Không còn `security_invoker` — RLS đã bỏ, phạm vi hộ do API lọc.)
- **Alternatives**: cột balance + cập nhật trong biz (4 đường mutate → khó chứng minh SC-004); tính ở client (sai khi nhiều thành viên).

## D15. Chặn ngày tương lai *(carry-over R18 — chuyển từ trigger về biz)*

- **Decision**: Chặn 2 lớp: client (date picker `max = hôm nay`) + **biz** `CreateTransaction`/`UpdateTransaction` từ chối `transaction_date > now() + 1 ngày` (nới lệch múi giờ) với mã `FUTURE_DATE_NOT_ALLOWED`. Không dùng trigger DB (D3).
- **Rationale**: API là writer duy nhất; lỗi từ biz có ngữ nghĩa (field-level) cho form; nới +1 ngày tránh từ chối oan múi giờ sớm hơn server.
- **Alternatives**: trigger `enforce_txn_rules` (bản cũ — thừa khi có API); CHECK với now() (semantics gây ngạc nhiên khi restore).

## D16. Sổ giao dịch — truy vấn, tên người nhập, phân trang *(carry-over R19)*

- **Decision**: `GET /api/transactions` join/preload `users(display_name)` + `categories(name, icon)` + `accounts(name)` một round-trip; `ORDER BY transaction_date DESC, id DESC`; phân trang offset 50/trang (infinite scroll); response embed `created_by_name`/`category_name`/`account_name`. Realtime: WS `transactions_changed` → store refetch trang hiện tại.
- **Rationale**: Đủ dữ liệu hiển thị FR-008 (tên, không phải mã) trong một request; offset đơn giản đủ cho hàng nghìn bản ghi.
- **Alternatives**: cursor pagination (tối ưu sau nếu cần); view join sẵn (artifact thừa).

## D17. Sửa/xóa đồng thời *(carry-over R20 — mở rộng D6 sang transactions)*

- **Decision**: `transactions.updated_at` làm mốc lạc quan (00007): PATCH/DELETE kèm `expected_updated_at`; storage chạy conditional UPDATE/DELETE — 0 hàng → tra tồn tại để phân biệt `CONCURRENCY_CONFLICT` 409 (đã đổi → client tải lại data mới) vs `RECORD_GONE` 404 (đã xóa → thông báo + về sổ — UC-TRK-04 E3/UC-TRK-05 E1). `updated_at` do GORM hook cập nhật.
- **Rationale**: Nhất quán cơ chế D6 của categories (đã kiểm chứng ở 001 #16); UI tái dùng cùng pattern xử lý lỗi.
- **Alternatives**: version int (cột thừa); khóa bi quan (kém UX).

## D18. Sửa giao dịch đổi loại Thu ⇄ Chi *(carry-over R21)*

- **Decision**: Web `TransactionFormView` chế độ sửa: đổi loại → `CategoryPicker` coi danh mục cũ (khác loại) là "chưa chọn" → buộc chọn lại; biz validate type giao dịch = type danh mục (đã có từ 001 — chạy cho cả update).
- **Rationale**: Không cần cơ chế mới; bất biến loại thực thi ở biz cho mọi đường ghi.
- **Alternatives**: cấm đổi loại khi sửa (bác — FR-009 cho phép sửa mọi trường).

## D19. Định danh người dùng *(thay R15 — đã hiện thực ở 001)*

- **Decision**: KHÔNG còn hạng mục riêng — `users` kiêm credentials (bcrypt) + JWT cookie + middleware phạm vi hộ đã dựng ở 001 (D4/D5); `created_by` do biz gán từ phiên (không trigger). US5/FR-015/FR-016 của spec 002 được 001 đáp ứng; quickstart 002 chỉ **kiểm chứng lại** (#1–#3).
- **Rationale**: Re-platform hợp nhất "danh bạ độc lập" và "đăng nhập" (không còn Supabase Auth để đối chiếu email); tránh trùng công việc giữa 2 feature.
- **Alternatives**: giữ cầu nối email như cũ (vô nghĩa khi tự quản auth).

## Tổng hợp

Toàn bộ điểm mở đã chốt (D13–D19 + kế thừa D1–D12), không còn NEEDS CLARIFICATION. Điểm chờ nghiệp vụ
(không chặn): baseline Success Metrics (BR-002); SC-006 ngưỡng 5s cần xác nhận; danh mục tài khoản
mặc định cuối cùng thuộc BR-005. Migrations sẽ thêm: `00006_accounts.sql`, `00007_transactions_v2.sql`.

## History

- 2026-07-10: Viết lại cho stack Go + Vue (re-platform) — R15→D19 (users kiêm auth từ 001), R16→D14, R17→D13 (bỏ trigger, dùng app logic), R18→D15 (trigger→biz), R19→D16, R20→D17, R21→D18.
