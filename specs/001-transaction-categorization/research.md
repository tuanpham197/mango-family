# Phase 0 — Research: Phân Loại Giao Dịch

**Date**: 2026-06-29 · **Feature**: 001-transaction-categorization · **Plan**: [plan.md](./plan.md)

Giải quyết các điểm chưa chắc chắn về kỹ thuật trước khi thiết kế. Mỗi mục: Quyết định · Lý do · Phương án đã cân nhắc.

## R1. Nền tảng client

- **Decision**: Flutter 3.x (Dart 3.x), build iOS + Android từ một codebase.
- **Rationale**: Người dùng đã chọn; phù hợp app mobile tài chính cá nhân; hệ sinh thái đầy đủ cho form nhập liệu, danh sách phân nhóm, offline cache; khớp "đội Mobile" trong BR-001.
- **Alternatives considered**: React Native (ngang ngửa, nhưng team thống nhất Flutter); Native iOS (loại — bỏ Android); Web (loại — yêu cầu trải nghiệm mobile).

## R2. Backend / lưu trữ — chọn BaaS nào

- **Decision**: **Supabase** (PostgreSQL + Supabase Auth + Row-Level Security), truy cập qua `supabase_flutter`.
- **Rationale**:
  - Mô hình dữ liệu **quan hệ**: danh mục tự tham chiếu một cấp (`parent_id`), khóa ngoại tới người dùng và giao dịch → Postgres hỗ trợ trực tiếp ràng buộc & toàn vẹn.
  - **Roll-up (FR-019)**: tổng hợp giao dịch danh mục con vào cha dễ dàng bằng SQL aggregation; với Firestore (NoSQL) phải phi chuẩn hóa thủ công.
  - **Cô lập dữ liệu theo hộ (cập nhật 2026-06-29)**: ánh xạ sang **RLS policy theo membership** — truy cập được nếu `auth.uid()` là thành viên của hộ sở hữu bản ghi (xem R12). Trước đây là per-user (FR-018) nhưng đã đổi sang mô hình hộ gia đình.
  - **Ràng buộc loại (FR-014, SC-003)**: kiểm tra "loại giao dịch = loại danh mục" bằng CHECK/trigger phía DB làm hàng rào cuối.
- **Alternatives considered**:
  - **Firebase (Firestore)**: offline-first mạnh hơn, phổ biến với Flutter; nhưng NoSQL chống lại mô hình quan hệ + roll-up và ràng buộc khóa ngoại phải tự xử lý ở client → rủi ro toàn vẹn cao hơn. **Có thể đổi sang nếu** yêu cầu offline-write phức tạp được ưu tiên hơn tính toàn vẹn quan hệ.
  - **Backend tự xây (REST + Postgres)**: linh hoạt nhất nhưng tốn hạ tầng/vận hành, thừa cho MVP cá nhân.

## R3. Quản lý trạng thái Flutter

- **Decision**: Riverpod (`flutter_riverpod`) + controllers theo feature.
- **Rationale**: Testable, không phụ thuộc context, hợp Clean Architecture; dễ mock repository trong unit/widget test (hỗ trợ TDD).
- **Alternatives considered**: Bloc (nhiều boilerplate hơn cho phạm vi này); setState/Provider thuần (khó test & mở rộng).

## R4. Mô hình "danh mục con một cấp" (FR-010, FR-011)

- **Decision**: Một bảng `categories` **tự tham chiếu** qua `parent_id` (nullable). Ràng buộc một cấp: `parent_id` chỉ được trỏ tới danh mục có `parent_id IS NULL` (enforce bằng trigger/CHECK + validate ở domain).
- **Rationale**: Đơn giản, tránh bảng riêng cho subcategory; "danh mục con = danh mục có cha". Kế thừa loại: trigger gán `type` con = `type` cha và chặn lệch loại (FR-011).
- **Alternatives considered**: Bảng `subcategories` riêng (trùng lặp cấu trúc, khó truy vấn gộp cha+con); cột `depth` (thừa khi chỉ 1 cấp).

## R5. Bất biến "loại cố định sau khi tạo" (FR-005)

- **Decision**: Cột `type` không cho UPDATE (trigger từ chối thay đổi `type`); UI chỉnh sửa chỉ phơi bày tên & biểu tượng.
- **Rationale**: Giữ báo cáo/ngân sách nhất quán; hàng rào DB phòng lỗi client.
- **Alternatives considered**: Chỉ chặn ở client (rủi ro nếu gọi trực tiếp API) — loại.

## R6. Xóa danh mục an toàn (FR-007, FR-008, FR-009, FR-012, SC-007)

- **Decision**: Thực hiện qua một thao tác giao dịch (transaction/RPC Postgres): khi danh mục (và các con) còn giao dịch → bắt buộc **gán lại** sang danh mục **cùng loại** hoặc **xóa** giao dịch, rồi mới xóa danh mục. Không dùng `ON DELETE CASCADE` ngầm cho giao dịch.
- **Rationale**: Bảo đảm "không bao giờ để lại giao dịch mất danh mục"; xử lý cha kéo theo con trong cùng một bước nguyên tử.
- **Alternatives considered**: Cascade xóa giao dịch tự động (vi phạm yêu cầu phải hỏi gán lại/xóa); soft-delete danh mục (đó là tính năng Ẩn — UC-CAT-06, khác xóa).

## R7. Cơ chế gợi ý danh mục (FR-015, BR-CAT-007)

- **Decision**: Engine **rule-based + lịch sử**, KHÔNG AI/ML:
  1. Khớp từ khóa theo quy tắc trên mô tả (bảng `categorization_rules`: keyword → category), VÀ
  2. Tần suất lịch sử phân loại của chính người dùng (suy ra từ `transactions`: mô tả tương tự → danh mục hay dùng).
  Gợi ý phải cùng loại Thu/Chi đang chọn; luôn cho phép ghi đè.
- **Rationale**: Đáp ứng chốt Clarifications 2026-06-24; triển khai được hoàn toàn ở client hoặc bằng truy vấn Postgres, đo SC-004 (≥60% chấp nhận).
- **Alternatives considered**: ML on-device/cloud (ngoài phạm vi, Out of Scope BR-001).

## R8. Xác thực & danh tính người dùng

- **Decision**: Supabase Auth (email/mật khẩu) cho MVP; người dùng ánh xạ tới `auth.users`. Mỗi người dùng thuộc một **hộ** qua bảng membership (R10/R11).
- **Rationale**: Tính năng phân loại giả định "đã đăng nhập và là thành viên một hộ" (precondition mọi UC).
- **Open (không chặn)**: social login/biometric là cải tiến sau.

## R9. Bộ danh mục mặc định (FR-001) — Open Question của spec

- **Decision (giả định, cần nghiệp vụ chốt)**: Seed **khi tạo hộ** (không phải khi tạo từng tài khoản) — Chi: Ăn uống, Di chuyển, Hóa đơn, Mua sắm, Giải trí, Sức khỏe; Thu: Lương, Thưởng. Đối xử như danh mục tự tạo (sửa/ẩn/xóa được). Cả hộ dùng chung.
- **Rationale**: Bám ví dụ BR-001; cho phép US1 hoạt động ngay; tránh trùng lặp seed cho từng thành viên.
- **Alternatives considered**: Seed mỗi người dùng (sai — sẽ trùng lặp trong sổ chung); không seed (vi phạm FR-001/US1 #4).
- **⚠ NEEDS BUSINESS CONFIRMATION**: danh sách cuối cùng — không chặn thiết kế.

## R10. Mô hình chia sẻ — sổ chung hộ gia đình (cập nhật 2026-06-29)

- **Decision**: Thêm thực thể **`households`** và **`household_members`** (membership). `categories`, `transactions`, `categorization_rules` gắn **`household_id`** thay vì `user_id`. Cả danh mục VÀ giao dịch dùng chung trong hộ.
- **Rationale**: Người dùng chốt "sổ chung cả nhà". Đây là mẫu multi-tenant theo hộ; mọi truy vấn/RLS xoay quanh `household_id`.
- **Alternatives considered**: Chung danh mục nhưng riêng giao dịch (người dùng không chọn); giữ per-user + view tổng hợp (không phải "sổ chung").
- **Giả định**: mỗi người dùng thuộc **một hộ** ở MVP (bảng membership vẫn hỗ trợ nhiều, nhưng UI chọn 1 hộ hiện tại).

## R11. Phân quyền — mọi thành viên ngang quyền (cập nhật 2026-06-29)

- **Decision**: Không có vai trò quản trị; **mọi thành viên có quyền ngang nhau** (tạo/sửa/xóa danh mục, nhập/sửa/xóa giao dịch). `household_members` không cần cột `role` để gác quyền; (tùy chọn lưu `joined_at`, `created_by` cho audit).
- **Rationale**: Người dùng chốt "mọi người ngang quyền" → đơn giản hóa, không thêm actor Quản trị, không cổng quyền trong use case.
- **Hệ quả**: Không cần thêm tác nhân mới ngoài "Thành viên hộ" trong sơ đồ use case; chỉ đổi nhãn actor từ "Người dùng" → "Thành viên hộ gia đình".
- **Mời thành viên (dependency)**: vì ngang quyền, bất kỳ thành viên nào cũng mời được (mã/đường link mời); chi tiết thuộc feature Quản lý hộ.

## R12. RLS theo membership (cập nhật 2026-06-29)

- **Decision**: Policy: bản ghi truy cập được nếu `household_id IN (SELECT household_id FROM household_members WHERE user_id = auth.uid())`. Đóng gói bằng hàm `is_member(household uuid)` (SECURITY DEFINER, STABLE) để tránh lặp và đệ quy RLS.
- **Rationale**: Thay thế `user_id = auth.uid()` cũ; cho phép chia sẻ trong hộ nhưng cô lập **giữa các hộ**.
- **Alternatives considered**: Kiểm quyền ở client (rủi ro); JWT custom claim chứa household_id (nhanh hơn nhưng phải refresh khi đổi hộ) — để tối ưu sau.

## R13. Đồng thời nhiều thành viên (concurrency) (mới)

- **Decision**: Dùng cập nhật lạc quan dựa trên `updated_at`/version; với xóa danh mục dùng RPC nguyên tử (đã có) để tránh đua. Hiển thị thông báo nếu bản ghi đã bị thành viên khác đổi/xóa. Cân nhắc **Supabase Realtime** để làm tươi danh sách.
- **Rationale**: Sổ chung ⇒ hai thành viên có thể sửa/xóa cùng danh mục đồng thời; cần tránh ghi đè thầm lặng và giao dịch mồ côi.
- **Alternatives considered**: Khóa bi quan (phức tạp, kém UX mobile); bỏ qua (rủi ro mất dữ liệu).

## R14. Ghi nhận người nhập giao dịch (authorship) (mới)

- **Decision**: `transactions.created_by` (uuid → auth.users) ghi thành viên đã nhập; hiển thị "ai nhập" trong sổ chung. Không gác quyền sửa/xóa theo người nhập (mọi người ngang quyền).
- **Rationale**: Sổ chung cần minh bạch nguồn gốc giao dịch; phục vụ báo cáo theo thành viên sau này.
- **Alternatives considered**: Không lưu người nhập (mất minh bạch trong sổ chung).

## Tổng hợp

Các điểm kỹ thuật đã được quyết, gồm cả thay đổi **sổ chung hộ gia đình** (R10–R14). Mục chờ nghiệp vụ: danh sách danh mục mặc định (R9, seed theo hộ) và **feature Quản lý hộ** (tạo/mời/tham gia) là **dependency** ngoài phạm vi feature này. Lưu ý quan trọng: **spec (FR-018, Assumptions, BR-001 Out of Scope) cần cập nhật** để khớp mô hình chia sẻ — xem plan.md "Tác động tới spec".
