# Quickstart — Validate Income & Expense Tracking (feature 002)

**Date**: 2026-07-06 (updated 2026-07-09) · **Plan**: [plan.md](./plan.md) · **Contracts**: [contracts/](./contracts/) ·
**Data model**: [data-model.md](./data-model.md)

Hướng dẫn chạy & kiểm chứng tính năng end-to-end. Không chứa code cài đặt.

## Prerequisites

- Flutter SDK 3.x; project Supabase đã dùng cho feature 001 (URL/key trong `src/.env.json`).
- **Schema nền tảng đã áp**: `src/supabase/setup_dev.sql` (gồm 0001–0011: schema 001 + `users` độc lập
  + hộ A/B + 3 user dev). Sau khi implement 002, script này gộp thêm 0012 (accounts + view) và
  0013 (transactions v2) — dán lại một lần trong SQL Editor.
- Tài khoản dev: `alice@dev.local` / `bob@dev.local` (hộ A) · `carol@dev.local` (hộ B) — mật khẩu
  `Password123!`.

## Setup

```bash
cd src && flutter pub get
# Dán setup_dev.sql (bản mới nhất) vào Supabase SQL Editor → Run
flutter run -d chrome --dart-define-from-file=.env.json
```

## Kiểm chứng theo Acceptance Criteria

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 1 | Đăng nhập & định danh | Đăng nhập Alice; xem hồ sơ | Vào sổ hộ A; tên "Alice" hiển thị | UC-TRK-01 AC-1 |
| 2 | Chặn khi chưa đăng nhập | Mở app chưa đăng nhập | Về màn đăng nhập; không thấy dữ liệu | UC-TRK-01 AC-2 |
| 3 | Sai mật khẩu | Đăng nhập sai mật khẩu | Từ chối, thông báo an toàn | UC-TRK-01 AC-3 |
| 4 | Nhập giao dịch hợp lệ | Chi 50.000, danh mục "Ăn uống", lưu | Xuất hiện trong sổ; số dư "Tiền mặt" −50.000 | UC-TRK-02 AC-1 |
| 5 | Số tiền không hợp lệ | Nhập 0 / số âm → Lưu | Chặn: "số tiền phải lớn hơn 0" | UC-TRK-02 AC-2 |
| 6 | Thiếu danh mục | Bỏ trống danh mục → Lưu | Chặn, yêu cầu chọn (nhất quán 001) | UC-TRK-02 AC-3 |
| 7 | Mô tả > 255 ký tự | Dán mô tả dài → Lưu | Chặn/giới hạn kèm thông báo | UC-TRK-02 AC-4 |
| 8 | Ngày mặc định & quá khứ | Không chỉnh ngày; rồi chỉnh về hôm qua | Ghi hiện tại; ghi theo ngày chọn | UC-TRK-02 AC-5 |
| 9 | Chặn ngày tương lai | Chọn ngày mai → Lưu | Bị từ chối kèm thông báo | FR-005, R18 |
| 10 | Lọc danh mục theo loại | Chọn Thu → mở danh mục | Chỉ danh mục Thu | UC-TRK-02 AC-6 |
| 11 | Sổ chung + tên người nhập | Alice nhập; Bob mở sổ | Bob thấy giao dịch, "do **Alice** nhập" (tên, không phải mã) | UC-TRK-03 AC-1 |
| 12 | Thứ tự & phân trang | Nhập nhiều giao dịch | Mới nhất trước; cuộn tải tiếp | UC-TRK-03 AC-2 |
| 13 | Cô lập hộ | Carol mở sổ | Không thấy giao dịch hộ A | UC-TRK-03 AC-3 |
| 14 | Sửa số tiền → số dư khớp | Sửa 50.000 → 80.000 | Số dư điều chỉnh đúng 30.000 chênh lệch | UC-TRK-04 AC-1 |
| 15 | Đổi loại buộc chọn lại danh mục | Đổi Chi → Thu khi sửa | Danh mục cũ mất hiệu lực; buộc chọn danh mục Thu | UC-TRK-04 AC-2 |
| 16 | Ngang quyền sửa | Bob sửa giao dịch Alice nhập | Thành công | UC-TRK-04 AC-3 |
| 17 | Không ghi đè thầm lặng | Alice & Bob cùng sửa; người sau lưu | Người sau nhận thông báo bản ghi đã đổi, thấy dữ liệu mới | UC-TRK-04 AC-4 |
| 18 | Sửa trong lúc bị xóa | Bob xóa; Alice đang mở form sửa → lưu | Báo "không còn tồn tại", về sổ | UC-TRK-04 E3 |
| 19 | Xóa có xác nhận | Bấm Xóa | Dialog cảnh báo trước khi xóa vĩnh viễn | UC-TRK-05 AC-1 |
| 20 | Hủy xóa | Hủy tại dialog | Không gì thay đổi | UC-TRK-05 AC-2 |
| 21 | Xóa xong số dư hoàn tác | Xác nhận xóa giao dịch Chi 50.000 | Biến mất khỏi sổ mọi thành viên; số dư +50.000 | UC-TRK-05 AC-3/AC-4 |
| 22 | Số dư luôn khớp (đối chiếu) | Sau chuỗi thêm/sửa/xóa bất kỳ | `account_balances` = initial + tổng bút toán (truy vấn đối chiếu) | SC-004, R16 |
| 23 | Mất kết nối khi lưu — không trùng | Ngắt mạng → Lưu (lỗi) → bật mạng → thử lại Lưu | Thông báo lỗi rõ ràng khi mất kết nối; sau thử lại chỉ có **một** giao dịch được tạo | UC-TRK-02 E5, Edge case spec |

## Chạy test

```bash
cd src
flutter test          # unit (validate/usecase) + widget (form chặn lưu, sổ hiển thị tên)
```

**Done khi**: 23/23 kịch bản pass (đa thành viên #11, #13, #16–#18, #21 kiểm bằng 2 phiên
Alice/Bob + Carol như cách e2e của feature 001) và `flutter test` xanh.
