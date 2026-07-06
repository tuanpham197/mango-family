# Phase 0 — Research: Ghi Chép Thu Nhập và Chi Phí

**Date**: 2026-07-06 · **Feature**: 002-transaction-tracking · **Plan**: [plan.md](./plan.md)

Giải quyết các điểm kỹ thuật chưa chốt trước khi thiết kế. Tái dùng các quyết định R1–R14 của
feature 001 (stack, RLS membership, concurrency R13, realtime…); dưới đây chỉ là điểm MỚI của 002,
đánh số tiếp **R15–R21**.

## R15. Người dùng độc lập & cầu nối phiên đăng nhập (đã chốt 2026-07-06)

- **Decision**: Bảng `public.users` **độc lập** (PK riêng `gen_random_uuid()`, KHÔNG FK sang
  `auth.users`); email unique là danh tính; phiên đăng nhập đối chiếu qua hàm
  `current_user_id()` (tra `users` theo email trong phiên). Mọi FK người (membership,
  `created_by`) trỏ về `users.id`; `created_by` do trigger `set_created_by` tự gán —
  client không cần biết `users.id`.
- **Rationale**: Yêu cầu người dùng (2026-07-06): bảng users là danh bạ của app, không phụ thuộc
  cấu trúc hệ xác thực; mật khẩu vẫn do Supabase Auth giữ (không tự chế lưu mật khẩu).
- **Alternatives considered**: `users.id` FK = `auth.users.id` (bị bác — muốn độc lập cấu trúc);
  tự xây tầng xác thực riêng lưu mật khẩu trong users (bác — rủi ro bảo mật, phải làm lại toàn bộ RLS).
- **Hiện trạng**: `0011_users.sql` + `setup_dev.sql` (refresh) + `current_household.dart` đã viết.

## R16. Số dư tài khoản — view suy ra, không cột materialized

- **Decision**: **View `account_balances`**: số dư = tổng có dấu của giao dịch
  (`INCOME` cộng, `EXPENSE` trừ) theo `account_id`, cộng số dư ban đầu của tài khoản.
  Không lưu cột `balance` tự cập nhật bằng trigger ở MVP.
- **Rationale**: SC-004 đòi 100% khớp tổng bút toán sau mọi thêm/sửa/xóa — view suy ra **đúng theo
  định nghĩa**, không có rủi ro trigger drift khi sửa số tiền/đổi tài khoản/đổi loại/xóa; thang dữ
  liệu (nghìn giao dịch/hộ, index theo `account_id`) đủ nhanh. Khớp Assumption spec: "số dư là giá
  trị suy ra nhất quán; không ràng buộc cách hiện thực".
- **Alternatives considered**: cột `accounts.balance` + trigger cộng dồn (nhanh hơn khi đọc nhưng
  4 đường mutate × nhiều cột thay đổi → dễ lệch, khó chứng minh SC-004); tính ở client (sai khi
  nhiều thành viên, không tin được).
- **Ghi chú**: entity model ghi `ACCOUNT.balance` là thuộc tính — hiện thực bằng view (giá trị suy
  ra), đúng constraint đã ghi. Khi dữ liệu lớn, chuyển sang materialized + reconcile là tối ưu sau.

## R17. Tài khoản mặc định của hộ (thu gọn BR-005)

- **Decision**: Bảng `accounts` (id, household_id, name, type ∈ CASH/BANK/EWALLET/CREDIT,
  initial_balance mặc định 0, created_by, timestamps). **Seed "Tiền mặt" (CASH)** cho mọi hộ hiện có
  (migration) và hộ tạo mới (trigger after insert on households). Form nhập: hộ chỉ có một tài
  khoản → chọn sẵn (FR-006).
- **Rationale**: FR-006/BR-TRK-006 cần tài khoản ngay nhưng BR-005 chưa triển khai — mirror đúng
  cách 001 xử lý tiền đề hộ (seed sẵn). `initial_balance` để BR-005 sau này đặt số dư ban đầu mà
  không phá công thức view.
- **Alternatives considered**: hoãn account_id sang BR-005 (bác — BR-TRK-006 In Scope); bảng
  accounts đầy đủ BR-005 luôn (bác — phình phạm vi: chuyển tiền, điều chỉnh số dư…).

## R18. Chặn ngày tương lai (FR-005)

- **Decision**: Chặn 2 lớp — client (date picker giới hạn `lastDate = now`) + **DB trigger**
  `enforce_txn_rules` mở rộng: `transaction_date > now() + interval '1 day'` → raise (nới 1 ngày
  cho lệch múi giờ thiết bị).
- **Rationale**: Hàng rào DB là chốt cuối như mọi bất biến khác của dự án; nới nhỏ tránh từ chối
  oan người dùng ở múi giờ sớm hơn server.
- **Alternatives considered**: CHECK constraint với `now()` (hợp lệ ở Postgres nhưng semantics gây
  ngạc nhiên khi restore/replicate — trigger rõ ràng hơn); chỉ chặn client (không đủ).

## R19. Sổ giao dịch — truy vấn, tên người nhập, phân trang

- **Decision**: Truy vấn `transactions` kèm **embed** `users(display_name)` (FK `created_by` →
  `users.id` cho phép join qua PostgREST) + `categories(name, icon)` + `accounts(name)`;
  `order=transaction_date.desc`, phân trang theo range (50 bản ghi/trang, infinite scroll).
  Realtime: tái dùng `subscribeHouseholdChanges` (đã subscribe bảng `transactions`) → invalidate
  provider sổ.
- **Rationale**: Một round-trip đủ dữ liệu hiển thị (FR-008: tên, không phải mã); range pagination
  đơn giản, đủ cho hàng nghìn bản ghi.
- **Alternatives considered**: view join sẵn phía DB (thêm artifact không cần thiết); tải toàn bộ
  rồi lọc client (không scale).

## R20. Sửa/xóa đồng thời (mở rộng R13 sang transactions)

- **Decision**: Thêm `transactions.updated_at` + trigger touch (touch function dùng chung
  `touch_updated_at()` đã có). Sửa: UPDATE kèm điều kiện mốc `updated_at` — 0 hàng khớp → phân biệt
  "đã đổi" (tải lại, báo `ConcurrencyConflict`) vs "đã xóa" (báo không còn tồn tại, về sổ — UC-TRK-04 E3).
  Xóa: DELETE theo id — nếu 0 hàng (người khác xóa trước) → thông báo nhẹ + làm tươi sổ (UC-TRK-05 E1).
- **Rationale**: Nhất quán cơ chế đã kiểm chứng e2e ở 001 (#16); UI/Failure (`ConcurrencyConflict`)
  tái dùng nguyên vẹn.
- **Alternatives considered**: khóa bi quan (kém UX mobile); version int (tương đương updated_at,
  thêm cột thừa).

## R21. Sửa giao dịch đổi loại Thu ⇄ Chi (FR-009)

- **Decision**: UI tái dùng mẫu 001: `CategorySelectField` coi danh mục khác loại là "chưa chọn"
  (effective-null) → buộc chọn lại; hàng rào DB `enforce_txn_rules` (type giao dịch = type danh mục,
  cùng hộ) đã có từ 001 chạy cho cả UPDATE.
- **Rationale**: Không cần cơ chế mới — bất biến loại đã được thực thi 2 lớp từ 001.
- **Alternatives considered**: cấm đổi loại khi sửa (bác — FR-009 cho phép sửa mọi trường).

## Tổng hợp

Toàn bộ điểm mở của Technical Context đã chốt (R15–R21), không còn NEEDS CLARIFICATION.
Điểm chờ nghiệp vụ (không chặn): baseline Success Metrics (BR-002 Open Question); danh mục tài
khoản mặc định cuối cùng thuộc BR-005. Migration hiện có: `0011_users.sql` (nền tảng); sẽ thêm
`0012_accounts.sql`, `0013_transactions_v2.sql` (xem data-model & contracts).
