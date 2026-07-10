# Quickstart — Validate Transaction Categorization (feature 001, Go + Vue)

**Date**: 2026-07-10 · **Plan**: [plan.md](./plan.md) · **Contracts**: [contracts/](./contracts/) · **Data model**: [data-model.md](./data-model.md)

Hướng dẫn chạy & kiểm chứng tính năng end-to-end trên stack mới. Không chứa code cài đặt —
chỉ lệnh chạy và kết quả mong đợi. 17 kịch bản nghiệp vụ **giữ nguyên** từ bản cũ.

## Prerequisites

- Go 1.22+ (`go version`) · Node 20+ (`node -v`) · Docker (Postgres dev)
- goose CLI (`go install github.com/pressly/goose/v3/cmd/goose@latest`) — hoặc dùng lệnh migrate nhúng của api

## Setup

```bash
# 1. Postgres dev
docker compose up -d postgres          # src/docker-compose.yml

# 2. Migrations (goose)
cd src/db && goose postgres "$DATABASE_URL" up      # 00001_users … 00005_categorization_rules

# 3. Seed dev (users + hộ + danh mục mặc định — KHÔNG phải migration, research D9)
cd ../api && go run ./cmd/seed
#    → Alice/Bob (hộ "Gia đình A"), Carol (hộ "Gia đình B"), mật khẩu dev: Password123!
#    → mỗi hộ có bộ danh mục mặc định (is_default=true)

# 4. Chạy API (Gin, cổng 8080) — .env theo .env.example (DATABASE_URL, JWT_SECRET)
go run .                               # hoặc air (hot reload)

# 5. Chạy web (Vite dev server, proxy /api + /ws → :8080)
cd ../web && npm install && npm run dev
```

> **Tắt lối**: `src/Makefile` gói sẵn toàn bộ — `make tools && make dev && make seed`, rồi `make api` + `make web` (2 terminal); test: `make test`, `make test-e2e`. Xem `make help`.

## Kiểm chứng theo Acceptance Criteria

Mỗi kịch bản tương ứng một Playwright spec trong `src/web/e2e/`. "Pass" = quan sát đúng kết quả mong đợi.

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 0 | Đăng nhập nền tảng | Đăng nhập Alice đúng/sai mật khẩu; mở app chưa đăng nhập | Đúng → vào hộ A, thấy tên "Alice"; sai → từ chối an toàn; chưa đăng nhập → về /login | UC-TRK-01, D4 |
| 1 | Tạo & dùng danh mục mới (AC1) | Quản lý danh mục → Thêm → loại **Chi**, tên "Thú cưng" → Lưu; mở nhập giao dịch Chi | "Thú cưng" có trong nhóm Chi và chọn được khi nhập giao dịch Chi | UC-CAT-02, US2 #1 |
| 2 | Loại là bắt buộc (AC2) | Tạo danh mục, bỏ trống loại → Lưu | Bị chặn, yêu cầu chọn loại | FR-004 |
| 3 | Cảnh báo trùng tên (AC3) | Tạo danh mục trùng tên trong cùng loại & cấp cha | Cảnh báo (`NAME_DUPLICATE_WARNING`), vẫn cho xác nhận tiếp tục | FR-017 |
| 4 | Tạo nhanh kế thừa loại (AC4) | Trong luồng nhập giao dịch Chi → tạo nhanh danh mục | Danh mục mới mang loại Chi và được chọn ngay cho giao dịch | UC-CAT-07 |
| 5 | Lọc theo loại khi nhập (US1) | Nhập giao dịch chọn loại Chi → mở danh sách danh mục | Chỉ danh mục Chi hiển thị; không có danh mục Thu | FR-014, SC-003 |
| 6 | Chặn lưu nếu chưa chọn danh mục | Nhập đủ thông tin, chưa chọn danh mục → Lưu | Bị chặn (`CATEGORY_REQUIRED`) | FR-013, SC-002 |
| 7 | Danh mục con một cấp (US3) | Tạo "Ăn ngoài" dưới "Ăn uống" (Chi); thử tạo cấp con thứ hai | Con kế thừa loại Chi; cấp con thứ hai bị từ chối (`NESTING_TOO_DEEP`) | FR-010/011, SC-008 |
| 8 | Xóa an toàn — gán lại (US2) | Xóa danh mục đang có giao dịch → chọn gán lại sang danh mục cùng loại | Giao dịch chuyển sang danh mục mới; danh mục cũ (và con) bị xóa trong MỘT giao dịch DB; không còn giao dịch mồ côi | FR-008/009/012, SC-007, D12 |
| 9 | Gán lại sai loại bị chặn | Khi xóa, thử chọn danh mục đích khác loại | Từ chối (`REASSIGN_TYPE_MISMATCH`) | FR-009 |
| 10 | Ẩn / bỏ ẩn (UC-CAT-06) | Ẩn một danh mục → mở nhập giao dịch; rồi bỏ ẩn | Khi ẩn: không xuất hiện để chọn mới nhưng lịch sử giữ nguyên; bỏ ẩn: xuất hiện lại | FR-020 |
| 11 | Đổi tên phản ánh mọi nơi | Đổi tên danh mục đang có giao dịch lịch sử | Tên mới hiển thị ở giao dịch lịch sử (tham chiếu id, không sao chép tên) | FR-016 |
| 12 | Gợi ý danh mục (US4) | Nhập mô tả chứa "Grab" (loại Chi); dùng lịch sử chung hộ | Gợi ý "Di chuyển" cùng loại; chọn danh mục khác thì lựa chọn được lưu và rule học thêm | FR-015, SC-004, D7 |
| 13 | **Chia sẻ danh mục trong hộ** | Alice tạo danh mục "Thú cưng"; Bob mở app (cùng hộ, context Playwright riêng) | Bob **thấy** "Thú cưng" trong ≤ 5s (WS invalidation) và chọn được khi nhập | FR-018, D8 |
| 14 | **Sổ chung + authorship** | Alice nhập 1 giao dịch; Bob mở sổ | Bob thấy giao dịch, hiển thị **"do Alice nhập"** (`created_by_name`) | FR-022 |
| 15 | **Cô lập giữa các hộ** | Carol (hộ B) đăng nhập ở context riêng | Carol **không** thấy danh mục/giao dịch hộ A; gọi thẳng API bằng id của hộ A → **404** | FR-018, D5 |
| 16 | **Đồng thời (concurrency)** | Alice & Bob cùng sửa/xóa một danh mục gần như đồng thời | Không ghi đè thầm lặng: người sau nhận `CONCURRENCY_CONFLICT`/`RECORD_GONE`; không phát sinh giao dịch mồ côi | D6, SC-007 |
| 17 | Quyền ngang nhau | Bob (không phải người tạo hộ) xóa/sửa danh mục do Alice tạo | Thành công — không có cổng quyền theo vai trò | FR-021 |

## Chạy test

```bash
cd src/api && go test ./...                    # unit biz (validate, suggest, delete-reassign, conflict)
DATABASE_URL=... go test ./... -tags=integration   # storage với Postgres docker
cd ../web && npm run test:unit                 # Vitest: stores/components
npx playwright test                            # e2e 0–17 (multi-context cho 13–17)
```

**Done khi**: 18/18 kịch bản (0–17) pass, `go test` + Vitest xanh. Lưu ý: (1) danh sách **danh mục
mặc định** chờ nghiệp vụ chốt (research D9); (2) **tạo hộ & mời thành viên** thuộc feature *Quản lý hộ*
(dependency) — quickstart dùng dữ liệu seed dev.

## History

- 2026-07-10: Viết lại cho stack Go + Vue (re-platform) — giữ nguyên 17 kịch bản nghiệp vụ, thêm #0 (đăng nhập nền tảng); setup đổi sang docker compose + goose + seed + Vite; e2e chuyển sang Playwright multi-context.
- 2026-07-10 (b): thêm ghi chú `src/Makefile` (up/migrate/seed/api/web/test) làm tắt lối cho Setup.
