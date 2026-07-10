# Quickstart — Validate Income & Expense Tracking (feature 002, Go + Vue)

**Date**: 2026-07-06 (rewritten 2026-07-10) · **Plan**: [plan.md](./plan.md) · **Contracts**: [contracts/](./contracts/) ·
**Data model**: [data-model.md](./data-model.md)

Hướng dẫn chạy & kiểm chứng tính năng end-to-end trên stack Go + Vue. 23 kịch bản nghiệp vụ
**giữ nguyên** từ bản cũ. Tiền đề: feature 001 (Go+Vue) đã dựng xong nền tảng (login, phạm vi hộ,
danh mục, WS) — xem [quickstart 001](../001-transaction-categorization/quickstart.md).

## Prerequisites

- Feature 001 chạy được (quickstart 001 pass) — cùng monorepo `src/`
- Go 1.22+ · Node 20+ · Docker (Postgres 16) · goose CLI

## Setup

```bash
# 1. Postgres + migrations (thêm 00006/00007 trên nền 00001–00005 của 001)
docker compose up -d postgres
cd src/db && goose postgres "$DATABASE_URL" up     # 00006_accounts · 00007_transactions_v2

# 2. Seed dev (nếu chưa chạy ở 001): Alice/Bob (hộ A) · Carol (hộ B) · mật khẩu Password123!
#    Seed đảm bảo mỗi hộ có tài khoản mặc định "Tiền mặt" (D13) → account_balances trả 0
cd ../api && go run ./cmd/seed

# 3. Chạy API + web
go run .                                            # hoặc air — cổng 8080
cd ../web && npm install && npm run dev             # Vite proxy /api + /ws
```

## Kiểm chứng theo Acceptance Criteria

Mỗi kịch bản tương ứng một Playwright spec trong `src/web/e2e/`. "Pass" = quan sát đúng kết quả mong đợi.

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 1 | Đăng nhập & định danh | Đăng nhập Alice; xem hồ sơ | Vào sổ hộ A; tên "Alice" hiển thị | UC-TRK-01 AC-1 |
| 2 | Chặn khi chưa đăng nhập | Mở app chưa đăng nhập | Về màn đăng nhập; không thấy dữ liệu | UC-TRK-01 AC-2 |
| 3 | Sai mật khẩu | Đăng nhập sai mật khẩu | Từ chối, thông báo an toàn | UC-TRK-01 AC-3 |
| 4 | Nhập giao dịch hợp lệ | Chi 50.000, danh mục "Ăn uống", lưu | Xuất hiện trong sổ; số dư "Tiền mặt" −50.000 | UC-TRK-02 AC-1 |
| 5 | Số tiền không hợp lệ | Nhập 0 / số âm → Lưu | Chặn: "số tiền phải lớn hơn 0" (`AMOUNT_INVALID`) | UC-TRK-02 AC-2 |
| 6 | Thiếu danh mục | Bỏ trống danh mục → Lưu | Chặn, yêu cầu chọn (`CATEGORY_REQUIRED`) | UC-TRK-02 AC-3 |
| 7 | Mô tả > 255 ký tự | Dán mô tả dài → Lưu | Chặn/giới hạn kèm thông báo (`DESCRIPTION_TOO_LONG`) | UC-TRK-02 AC-4 |
| 8 | Ngày mặc định & quá khứ | Không chỉnh ngày; rồi chỉnh về hôm qua | Ghi hiện tại; ghi theo ngày chọn | UC-TRK-02 AC-5 |
| 9 | Chặn ngày tương lai | Chọn ngày mai → Lưu | Bị từ chối kèm thông báo (`FUTURE_DATE_NOT_ALLOWED`) | FR-005, D15 |
| 10 | Lọc danh mục theo loại | Chọn Thu → mở danh mục | Chỉ danh mục Thu | UC-TRK-02 AC-6 |
| 11 | Sổ chung + tên người nhập | Alice nhập; Bob mở sổ (context riêng) | Bob thấy giao dịch trong ≤ 5s, "do **Alice** nhập" (tên, không phải mã) | UC-TRK-03 AC-1, SC-006 |
| 12 | Thứ tự & phân trang | Nhập nhiều giao dịch | Mới nhất trước; cuộn tải tiếp (50/trang) | UC-TRK-03 AC-2, D16 |
| 13 | Cô lập hộ | Carol mở sổ | Không thấy giao dịch hộ A; gọi API bằng id hộ A → 404 | UC-TRK-03 AC-3, D5 |
| 14 | Sửa số tiền → số dư khớp | Sửa 50.000 → 80.000 | Số dư điều chỉnh đúng 30.000 chênh lệch | UC-TRK-04 AC-1 |
| 15 | Đổi loại buộc chọn lại danh mục | Đổi Chi → Thu khi sửa | Danh mục cũ mất hiệu lực; buộc chọn danh mục Thu | UC-TRK-04 AC-2, D18 |
| 16 | Ngang quyền sửa | Bob sửa giao dịch Alice nhập | Thành công | UC-TRK-04 AC-3 |
| 17 | Không ghi đè thầm lặng | Alice & Bob cùng sửa; người sau lưu | Người sau nhận `CONCURRENCY_CONFLICT`, thấy dữ liệu mới | UC-TRK-04 AC-4, D17 |
| 18 | Sửa trong lúc bị xóa | Bob xóa; Alice đang mở form sửa → lưu | Báo "không còn tồn tại" (`RECORD_GONE`), về sổ | UC-TRK-04 E3, D17 |
| 19 | Xóa có xác nhận | Bấm Xóa | Dialog cảnh báo trước khi xóa vĩnh viễn | UC-TRK-05 AC-1 |
| 20 | Hủy xóa | Hủy tại dialog | Không gì thay đổi | UC-TRK-05 AC-2 |
| 21 | Xóa xong số dư hoàn tác | Xác nhận xóa giao dịch Chi 50.000 | Biến mất khỏi sổ mọi thành viên; số dư +50.000 | UC-TRK-05 AC-3/AC-4 |
| 22 | Số dư luôn khớp (đối chiếu) | Sau chuỗi thêm/sửa/xóa bất kỳ | `account_balances` = initial + tổng bút toán (truy vấn đối chiếu qua GET /api/accounts) | SC-004, D14 |
| 23 | Mất kết nối khi lưu — không trùng | Ngắt mạng → Lưu (lỗi) → bật mạng → thử lại Lưu | Thông báo lỗi rõ ràng khi mất kết nối; sau thử lại chỉ có **một** giao dịch được tạo | UC-TRK-02 E5, Edge case spec |

## Chạy test

```bash
cd src/api && go test ./...                        # unit biz (validate, concurrency, future-date)
DATABASE_URL=... go test ./... -tags=integration   # storage + view balances với Postgres docker
cd ../web && npm run test:unit                     # Vitest
npx playwright test                                # e2e 1–23 (multi-context #11/#13/#16–#18/#21)
```

**Done khi**: 23/23 kịch bản pass (đa thành viên #11, #13, #16–#18, #21 dùng Playwright multi-context
Alice/Bob + Carol; #23 kiểm chống trùng khi thử lại; đối chiếu số dư #22 qua API) và toàn bộ test xanh.

## History

- 2026-07-10: Viết lại cho stack Go + Vue (re-platform) — giữ nguyên 23 kịch bản nghiệp vụ; setup đổi sang goose 00006/00007 + seed app-logic; truy vết R16/R18/R20 → D14/D15/D17; e2e Playwright multi-context.
