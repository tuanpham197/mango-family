# Quickstart — Validate Transaction Categorization (feature 001)

**Date**: 2026-06-29 · **Plan**: [plan.md](./plan.md) · **Contracts**: [contracts/](./contracts/) · **Data model**: [data-model.md](./data-model.md)

Hướng dẫn chạy & kiểm chứng tính năng end-to-end. Không chứa code cài đặt — chỉ lệnh chạy và kết quả mong đợi. Chi tiết schema/hành vi xem ở contracts và data-model.

## Prerequisites

- Flutter SDK 3.x (`flutter --version`), Dart 3.x
- Một project Supabase (URL + anon key) — local qua `supabase start`, hoặc cloud
- Supabase CLI (`supabase --version`) nếu chạy DB cục bộ

## Setup

```bash
# 1. Cài phụ thuộc
flutter pub get

# 2. Khởi tạo schema danh mục (áp dụng db-schema.sql qua thư mục migrations)
supabase db reset            # chạy supabase/migrations/* (gồm bản sao của contracts/db-schema.sql)

# 3. Cấu hình kết nối (SUPABASE_URL, SUPABASE_ANON_KEY) qua --dart-define hoặc .env

# 4. Chuẩn bị dữ liệu đa thành viên để kiểm chứng sổ chung:
#    - Tạo 2 tài khoản: Alice, Bob (cùng một hộ "Gia đình A")
#    - Tạo 1 tài khoản: Carol (hộ khác "Gia đình B") để kiểm cô lập
#    (Tạo hộ + thêm thành viên thuộc feature Quản lý hộ — dependency; với MVP có thể seed bằng SQL)

# 5. Chạy app
flutter run --dart-define=SUPABASE_URL=... --dart-define=SUPABASE_ANON_KEY=...
```

## Kiểm chứng theo Acceptance Criteria

Mỗi kịch bản dưới đây tương ứng test trong `integration_test/`. "Pass" = quan sát đúng kết quả mong đợi.

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 1 | Tạo & dùng danh mục mới (AC1) | Quản lý danh mục → Thêm → loại **Chi**, tên "Thú cưng" → Lưu; mở nhập giao dịch Chi | "Thú cưng" có trong nhóm Chi và chọn được khi nhập giao dịch Chi | UC-CAT-02, US2 #1 |
| 2 | Loại là bắt buộc (AC2) | Tạo danh mục, bỏ trống loại → Lưu | Bị chặn, yêu cầu chọn loại | FR-004 |
| 3 | Cảnh báo trùng tên (AC3) | Tạo danh mục trùng tên trong cùng loại & cấp cha | Hiện cảnh báo, vẫn cho xác nhận tiếp tục | FR-017 |
| 4 | Tạo nhanh kế thừa loại (AC4) | Trong luồng nhập giao dịch Chi → tạo nhanh danh mục | Danh mục mới mang loại Chi và được chọn ngay cho giao dịch | UC-CAT-07 |
| 5 | Lọc theo loại khi nhập (US1) | Nhập giao dịch chọn loại Chi → mở danh sách danh mục | Chỉ danh mục Chi hiển thị; không có danh mục Thu | FR-014, SC-003 |
| 6 | Chặn lưu nếu chưa chọn danh mục | Nhập đủ thông tin, chưa chọn danh mục → Lưu | Bị chặn, yêu cầu chọn danh mục | FR-013, SC-002 |
| 7 | Danh mục con một cấp (US3) | Tạo "Ăn ngoài" dưới "Ăn uống" (Chi); thử tạo cấp con thứ hai | Con kế thừa loại Chi; tạo cấp con thứ hai bị từ chối | FR-010, FR-011, SC-008 |
| 8 | Xóa an toàn — gán lại (US2) | Xóa danh mục đang có giao dịch → chọn gán lại sang danh mục cùng loại | Giao dịch chuyển sang danh mục mới; danh mục cũ (và con) bị xóa; không còn giao dịch mồ côi | FR-008, FR-009, FR-012, SC-007 |
| 9 | Gán lại sai loại bị chặn | Khi xóa, thử chọn danh mục đích khác loại | Hệ thống không cho chọn / từ chối | FR-009 |
| 10 | Ẩn / bỏ ẩn (UC-CAT-06) | Ẩn một danh mục → mở nhập giao dịch; rồi bỏ ẩn | Khi ẩn: không xuất hiện để chọn mới nhưng lịch sử giữ nguyên; bỏ ẩn: xuất hiện lại | FR-020 |
| 11 | Đổi tên phản ánh mọi nơi | Đổi tên danh mục đang có giao dịch lịch sử | Tên mới hiển thị ở giao dịch lịch sử và báo cáo (không cần gán lại) | FR-016 |
| 12 | Gợi ý danh mục (US4) | Nhập mô tả chứa "Grab" (loại Chi); dùng lịch sử chung hộ | Gợi ý "Di chuyển" cùng loại; chọn danh mục khác thì lựa chọn được lưu | FR-015, SC-004 |
| 13 | **Chia sẻ danh mục trong hộ** | Alice tạo danh mục "Thú cưng"; Bob mở app (cùng hộ) | Bob **thấy** "Thú cưng" và chọn được khi nhập giao dịch | R10, FR-018(đã đổi) |
| 14 | **Sổ chung giao dịch + authorship** | Alice nhập 1 giao dịch; Bob mở sổ | Bob thấy giao dịch đó, hiển thị **"do Alice nhập"** (`created_by`) | R10, R14 |
| 15 | **Cô lập giữa các hộ** | Carol (hộ B) đăng nhập | Carol **không** thấy danh mục/giao dịch của hộ A (RLS theo membership) | R12, FR-018(đã đổi) |
| 16 | **Đồng thời (concurrency)** | Alice & Bob cùng sửa/xóa một danh mục gần như đồng thời | Không ghi đè thầm lặng; người sau nhận thông báo bản ghi đã đổi/đã xóa; không phát sinh giao dịch mồ côi | R13, SC-007 |
| 17 | Quyền ngang nhau | Bob (không phải người tạo hộ) xóa/sửa danh mục do Alice tạo | Thành công — không có cổng quyền theo vai trò | R11 |

## Chạy test

```bash
flutter test                       # unit + widget (domain use cases, form, lọc theo loại)
flutter test integration_test      # end-to-end các kịch bản 1–17 ở trên (gồm đa thành viên 13–17)
```

**Done khi**: 17/17 kịch bản pass và `flutter test` xanh. Lưu ý: (1) danh sách **danh mục mặc định** chờ nghiệp vụ chốt (research R9); (2) việc **tạo hộ & mời thành viên** thuộc feature *Quản lý hộ* (dependency) — quickstart này dùng dữ liệu hộ đã có sẵn để kiểm chứng phần phân loại.
