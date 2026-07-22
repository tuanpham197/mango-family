# Quickstart — Validate Màn Tổng Quan Đầy Đủ (feature 004, Go + Vue)

**Date**: 2026-07-14 · **Plan**: [plan.md](./plan.md) · **Contracts**: [contracts/overview-api.md](./contracts/overview-api.md) · **Data model**: [data-model.md](./data-model.md)

Tiền đề: feature 001 + 002 + 003 đã chạy (login, phạm vi hộ, danh mục, giao dịch + tài khoản/số dư, ngân sách, WS, màn Tổng quan). Feature 004 **chỉ đọc** dữ liệu để dựng màn Tổng quan đầy đủ theo [`dashboard.png`](../design/dashboard.png) và đặt Tổng quan làm trang mặc định. **Không migration mới.**

## Setup

```bash
# Không có migration mới cho 004 (chỉ đọc). Dùng DB 001–003 sẵn có.
cd src && make dev            # Postgres + migrations 00001–00009 (đã có)
make seed                     # Alice/Bob (hộ A) · Carol (hộ B); Password123!
make api                      # Go API :8080  (thêm route GET /api/overview)
make web                      # Vite; "/" nay là màn Tổng quan
```

## Kiểm chứng theo Acceptance Criteria

Mỗi kịch bản tương ứng một Playwright spec trong `src/web/e2e/`. Đa thành viên (Alice/Bob hộ A, Carol hộ B) dùng browser context riêng để kiểm realtime & cô lập hộ.

### US1 — Trang mặc định là Tổng quan

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 1 | Đăng nhập không kèm đích | Sau đăng nhập màn đầu tiên là **Tổng quan** (`/`) | FR-001, SC-004 |
| 2 | Mở app ở địa chỉ gốc khi đã đăng nhập | Hiển thị màn Tổng quan | FR-001 |
| 3 | Mở liên kết cụ thể (vd `/ledger`, `/budgets`) | Đích liên kết được tôn trọng (không ép về Tổng quan) | FR-001 |
| 4 | Điều hướng | Từ Tổng quan mở được Giao dịch (`/ledger`), Ngân sách, và (lối phụ) Danh mục | FR-009, D32 |

### US2 — Thu/Chi tháng này

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 5 | Thu/Chi tháng | Thẻ Thu nhập = Σ Thu tháng; Chi phí = Σ Chi tháng của hộ | FR-003, SC-001 |
| 6 | Thêm giao dịch Chi | Thẻ Chi phí cập nhật đúng (realtime/refetch) | FR-003, FR-010 |
| 7 | Sửa/xóa giao dịch quá khứ trong tháng | Thu/Chi tính lại đúng | FR-010, SC-001 |
| 8 | Vị trí | Hai thẻ nằm dưới Tổng tài sản ròng, trên Chi tiêu theo danh mục | FR-008 |

### US3 — Chi tiêu theo danh mục (trước Ngân sách)

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 9 | Liệt kê chi theo danh mục | Mỗi danh mục Chi có phát sinh: số tiền + % trên tổng chi, sắp giảm dần | FR-004, SC-002 |
| 10 | Danh mục con | Chi ở danh mục con gộp vào danh mục cha | FR-004, SC-002 |
| 11 | Danh mục không có chi | Không xuất hiện trong danh sách | FR-004 |
| 12 | Vị trí | Mục Chi tiêu theo danh mục nằm **ngay trước** phần Ngân sách | FR-004 |
| 13 | Danh mục ẩn có chi | Vẫn tính, kèm dấu hiệu đã ẩn | edge case |

### US4 — Tổng tài sản ròng

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 14 | Tài sản ròng | Thẻ đầu tiên (dưới lời chào) = Σ số dư mọi tài khoản của hộ | FR-002, SC-003 |
| 15 | % so tháng trước | Hiển thị chỉ số thay đổi (hoặc ẩn nếu không đủ dữ liệu tháng trước) | FR-002, D28 |
| 16 | Số dư đổi | Thêm giao dịch → tài sản ròng cập nhật | FR-002, FR-010 |

### US5 — Giao dịch gần đây

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 17 | Danh sách gần đây | Vài giao dịch mới nhất: mô tả/tên, danh mục · tài khoản, số tiền có dấu/màu | FR-006 |
| 18 | Giao dịch mới | Nhập giao dịch → xuất hiện đầu danh sách gần đây (realtime) | FR-006, FR-011 |
| 19 | Vị trí | Mục Giao dịch gần đây nằm cuối màn (dưới Ngân sách) | FR-008 |

### US6 — Bố cục & điều hướng khớp dashboard.png

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 20 | Lời chào | Trên cùng: lời chào theo buổi + tên thành viên + avatar | FR-007 |
| 21 | Thứ tự phần | Tài sản ròng → Thu/Chi → Chi tiêu theo danh mục → Ngân sách → Giao dịch gần đây | FR-008, SC-005 |
| 22 | Thanh nav | 5 mục: Tổng quan · Giao dịch · ＋ · Ngân sách · Báo cáo; ＋ nổi bật ở giữa | FR-009, SC-005 |
| 23 | "Báo cáo" placeholder | Bấm Báo cáo → màn "sắp có" (BR-004) | D32 |
| 24 | Ngân sách trên Tổng quan | Tiêu đề "Ngân sách tháng N" + "Xem tất cả ›" → danh sách ngân sách (giữ 003) | FR-005 |

### Đa thành viên & cô lập hộ

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 25 | Realtime giữa thành viên | Bob nhập giao dịch → màn Tổng quan Alice (Thu/Chi, chi theo danh mục, tài sản ròng, gần đây) cập nhật ≤ 5s | SC-006, D33 |
| 26 | Cô lập hộ | Carol (hộ B) mở Tổng quan → chỉ thấy số liệu hộ B; `GET /api/overview` chỉ trả dữ liệu hộ đang đăng nhập | FR-010 |

## Kiểm chứng phi UI (Go)

```bash
cd src && make test                      # Go unit (cửa sổ tháng, net worth cuối tháng trước, % danh mục, gộp cây) + Vitest
make test-api-integration                # storage overview trên Postgres thật (khớp SUM, cô lập hộ)
make test-e2e                            # Playwright (bố cục, trang mặc định, realtime đa thành viên)
```

**Điểm cần chứng minh bằng test**: Thu/Chi & chi-theo-danh-mục khớp tổng giao dịch tháng (SC-001/002); tài sản ròng khớp tổng số dư (SC-003); trang mặc định luôn là Tổng quan (SC-004); bố cục khớp thứ tự dashboard.png (SC-005); realtime ≤ 5s (SC-006).

## History

- v1 (2026-07-14): tạo mới — 26 kịch bản kiểm chứng US1–US6 + đa thành viên/cô lập hộ; không migration (chỉ đọc), thêm route `GET /api/overview`, đổi trang mặc định `/` = Tổng quan.
