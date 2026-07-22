# Quickstart — Validate Budgeting (feature 003, Go + Vue)

**Date**: 2026-07-13 · **Plan**: [plan.md](./plan.md) · **Contracts**: [contracts/](./contracts/) ·
**Data model**: [data-model.md](./data-model.md)

**34 kịch bản** (v2 2026-07-13: +#34 tóm tắt ngân sách trên màn Tổng quan — FR-014).

Hướng dẫn chạy & kiểm chứng tính năng ngân sách end-to-end trên stack Go + Vue. Tiền đề: feature 001 +
002 (Go+Vue) đã dựng nền tảng (login, phạm vi hộ, danh mục, **giao dịch đầy đủ + tài khoản**, WS) — xem
[quickstart 002](../002-transaction-tracking/quickstart.md). Ngân sách chỉ **đọc** giao dịch/danh mục.

## Prerequisites

- Feature 001 + 002 chạy được (quickstart 001/002 pass) — cùng monorepo `src/`
- Go 1.22+ · Node 20+ · Docker (Postgres 16) · goose CLI

## Setup

```bash
# 1. Postgres + migrations (thêm 00008/00009 trên nền 00001–00007 của 001/002)
docker compose up -d postgres
cd src/db && goose postgres "$DATABASE_URL" up     # 00008_budgets · 00009_budget_alerts

# 2. Seed dev (nếu chưa chạy): Alice/Bob (hộ A) · Carol (hộ B) · mật khẩu Password123!
#    Có sẵn danh mục Chi "Ăn uống" + vài giao dịch để tiến độ ngân sách khác 0
cd ../api && go run ./cmd/seed

# 3. Chạy API + web
go run .                                            # hoặc air — cổng 8080
cd ../web && npm install && npm run dev             # Vite proxy /api + /ws
```

> **Tắt lối**: dùng `src/Makefile` — `make dev && make seed`, rồi `make api` + `make web`; test: `make test`, `make test-e2e`. Xem `make help`.

## Kiểm chứng theo Acceptance Criteria

Mỗi kịch bản tương ứng một Playwright spec trong `src/web/e2e/`. "Pass" = quan sát đúng kết quả mong đợi.
Đa thành viên (Alice/Bob cùng hộ A, Carol hộ B) dùng browser context riêng để kiểm realtime & cô lập hộ.

### US1 — Tạo ngân sách theo danh mục (UC-BGT-01)

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 1 | Tạo ngân sách danh mục hợp lệ | Chọn "Ăn uống", giới hạn 5.000.000, kỳ hàng tháng → Lưu | Xuất hiện trong danh sách; tiến độ = chi "Ăn uống" từ đầu tháng | UC-BGT-01 AC-1, FR-001 |
| 2 | Chặn giới hạn ≤ 0 | Nhập 0 / số âm → Lưu | Chặn: "giới hạn phải lớn hơn 0" (`LIMIT_INVALID`) | UC-BGT-01 AC-2, FR-003 |
| 3 | Bắt buộc chọn danh mục Chi | Bỏ trống danh mục → Lưu | Chặn (`CATEGORY_REQUIRED`); danh sách chọn chỉ gồm danh mục **Chi** | UC-BGT-01 AC-3, FR-001 |
| 4 | Chặn ngân sách trùng danh mục/kỳ | "Ăn uống" đã có ngân sách tháng → tạo thêm ngân sách tháng | Chặn (`BUDGET_DUPLICATE`), chỉ tới ngân sách hiện có | UC-BGT-01 AC-4, FR-010 |
| 5 | Kỳ một lần ngày sai | ONE_TIME với ngày kết thúc < bắt đầu → Lưu | Chặn (`PERIOD_INVALID`) | UC-BGT-01 E4, FR-004 |
| 6 | Dùng chung trong hộ, cô lập giữa hộ | Alice tạo; Bob mở danh sách; Carol mở | Bob thấy (≤ 5s); Carol (hộ B) không thấy; gọi API bằng id hộ A → 404 | UC-BGT-01 AC-5, FR-011 |

### US2 — Theo dõi tiến độ ngân sách (UC-BGT-03)

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 7 | Hiển thị tiến độ đúng định dạng | Ngân sách 5.000.000, đã chi 3.500.000 | Hiển thị "3.500.000/5.000.000 (70%)" | UC-BGT-03 AC-1, FR-005 |
| 8 | Cập nhật khi thêm giao dịch | Nhập giao dịch Chi "Ăn uống" 500.000 | Tiến độ → 4.000.000 (80%) | UC-BGT-03 AC-2, FR-006 |
| 9 | Tính lại khi xóa giao dịch quá khứ | Xóa giao dịch 500.000 vừa thêm | Tiến độ quay về 3.500.000 (70%) | UC-BGT-03 AC-3, SC-003 |
| 10 | Đổi danh mục cập nhật cả hai ngân sách | Đổi giao dịch từ danh mục A (có NS) sang B (có NS) | Tiến độ cả hai ngân sách cập nhật đúng | UC-BGT-03 AC-4 |
| 11 | Đồng bộ realtime giữa thành viên | Alice mở màn ngân sách; Bob nhập giao dịch ảnh hưởng | Tiến độ trên màn Alice cập nhật ≤ 5s | UC-BGT-03 AC-5, SC-006 |
| 12 | Danh mục con tính vào ngân sách cha | Nhập giao dịch Chi thuộc danh mục con của danh mục có NS | Tính vào tiến độ ngân sách cha | UC-BGT-03 AC-6, FR-006 |
| 13 | Giao dịch đổi Chi → Thu | Đổi loại một giao dịch đang tính vào NS sang Thu | Không còn tính vào NS; tiến độ giảm tương ứng | UC-BGT-03 E1, edge case |
| 14 | Kỳ chưa có giao dịch | Tạo ngân sách kỳ mới chưa có chi | Tiến độ 0%, không cảnh báo | UC-BGT-03 E3, edge case |

### US3 — Cảnh báo ngưỡng & vượt (UC-BGT-04)

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 15 | Cảnh báo đạt 80% | NS 1.000.000, chi 750.000 → nhập thêm 100.000 (85%) | Cảnh báo "đạt 80%" hiển thị in-app cho mọi thành viên | UC-BGT-04 AC-1, FR-007 |
| 16 | Cảnh báo vượt + số tiền vượt | Nhập thêm 200.000 (tổng 1.050.000, 105%) | Cảnh báo "vượt ngân sách" kèm đúng "vượt 50.000" | UC-BGT-04 AC-2, FR-008, SC-005 |
| 17 | Không phát lại cùng mức trong kỳ | Nhập thêm giao dịch khi vẫn trên 80% | Không lặp lại cảnh báo 80% | UC-BGT-04 AC-3, FR-009 |
| 18 | Phát lại sau khi tụt dưới mức | Xóa giao dịch để tụt < 80% rồi nhập lại vượt 80% | Cảnh báo 80% **phát lại** | UC-BGT-04 AC-4, US3 #4 |
| 19 | Cảnh báo tới mọi thành viên | Bob gây giao dịch vượt ngưỡng; Alice đang mở app | Alice cũng thấy cảnh báo ≤ 5s | UC-BGT-04 AC-5, SC-006 |

### US4 — Ngân sách tổng chi tiêu (UC-BGT-02)

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 20 | Tiến độ = tổng chi của hộ | Tạo ngân sách tổng 20.000.000/tháng | Tiến độ = tổng mọi giao dịch Chi trong tháng của hộ | UC-BGT-02 AC-1, FR-002 |
| 21 | Chặn ngân sách tổng trùng kỳ | Đã có NS tổng tháng → tạo thêm NS tổng tháng | Chặn (`BUDGET_DUPLICATE`) | UC-BGT-02 AC-2, FR-010 |
| 22 | Song song & độc lập với NS danh mục | Cùng tồn tại NS tổng + NS danh mục; nhập 1 giao dịch Chi | Cả hai tiến độ cập nhật; cảnh báo từng NS độc lập | UC-BGT-02 AC-3 |

### US5 — Sửa & xóa ngân sách (UC-BGT-05, UC-BGT-06)

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 23 | Sửa giới hạn → tiến độ & cảnh báo tính lại | NS 5.000.000 đã chi 3.500.000 (70%) → sửa giới hạn 4.000.000 | Tiến độ 87,5%; cảnh báo 80% phát theo trạng thái mới | UC-BGT-05 AC-1, FR-012 |
| 24 | Chặn khi sửa giới hạn không hợp lệ | Sửa giới hạn thành 0 / âm → Lưu | Chặn (`LIMIT_INVALID`) — cùng bộ xác thực như tạo | UC-BGT-05 AC-4, FR-012 |
| 25 | Ngang quyền sửa ngân sách người khác | Bob sửa NS do Alice tạo → Lưu hợp lệ | Thành công (mọi thành viên ngang quyền) | UC-BGT-05 AC-2, FR-012 |
| 26 | Không ghi đè thầm lặng | Alice & Bob cùng sửa; người sau lưu | Người sau nhận `CONCURRENCY_CONFLICT`, thấy dữ liệu mới | UC-BGT-05 AC-3, FR-013 |
| 27 | Xóa có xác nhận | Bấm Xóa → hộp xác nhận → xác nhận | NS + cảnh báo liên quan biến mất khỏi danh sách mọi thành viên | UC-BGT-06 AC-1, FR-012 |
| 28 | Hủy xóa không đổi gì | Bấm Xóa → Hủy | NS vẫn còn nguyên | UC-BGT-06 AC-2 |
| 29 | Xóa NS đã bị người khác xóa | Bob xóa; Alice (form đang mở) bấm Lưu/Xóa | Thông báo `RECORD_GONE`, làm tươi danh sách | UC-BGT-05 E4 / UC-BGT-06 E1, D25 |

### Edge cases khác

| # | Kịch bản | Thao tác | Kết quả mong đợi | Truy vết |
|---|----------|----------|------------------|----------|
| 30 | Danh mục có ngân sách bị ẩn | Ẩn danh mục "Ăn uống" (feature 001) | NS vẫn theo dõi bình thường; danh sách kèm trạng thái danh mục (đã ẩn) | edge case, `category_hidden` |
| 31 | Ngân sách một lần kết thúc kỳ | ONE_TIME qua `end_date` | Chuyển ENDED; giữ để xem lại; không phát cảnh báo nữa | edge case, D20 |
| 32 | Sang kỳ mới (tháng/tuần) | Bước sang tháng mới | NS lặp lại cùng giới hạn; tiến độ & cảnh báo reset 0 | edge case, D20 |
| 33 | Tạo giữa kỳ tính từ đầu kỳ | Tạo NS giữa tháng khi đã có chi | Tiến độ gồm toàn bộ chi từ đầu kỳ (không chỉ từ lúc tạo) | edge case, D21 |
| 34 | Tóm tắt ngân sách trên màn Tổng quan | Mở màn Tổng quan khi hộ có ngân sách hoạt động | Thấy widget tóm tắt (tổng nếu có + vài danh mục nổi bật) với mã màu ngưỡng (green/amber/đỏ); bấm "Xem tất cả" mở danh sách ngân sách; cập nhật realtime khi có giao dịch mới | US2 #7, FR-014, wireframe màn 1 |

## Kiểm chứng phi UI (Go)

```bash
cd src && make test                      # Go unit (cửa sổ kỳ, subtree, %, máy trạng thái cảnh báo) + Vitest
make test-api-integration                # storage + evaluator trên Postgres thật (idempotent recompute)
make test-e2e                            # Playwright 34 kịch bản trên (đa thành viên, cảnh báo realtime, Dashboard)
```

**Điểm cần chứng minh bằng test**: tiến độ khớp 100% tổng giao dịch EXPENSE liên quan sau mọi thao tác
(SC-003); mỗi mức cảnh báo phát ≤ 1 lần/kỳ và phát lại đúng sau khi tụt dưới (SC-004); `over_amount`
đúng (SC-005); realtime ≤ 5s (SC-006); recompute cảnh báo idempotent (D24).

## History

- v1 (2026-07-13, claude): tạo mới cho stack Go+Vue — 33 kịch bản kiểm chứng US1–US5 + edge cases,
  ánh xạ Acceptance Criteria của UC-BGT-01…06; setup goose 00008/00009 trên nền 001/002.
- v2 (2026-07-13, claude): +#34 tóm tắt ngân sách trên màn Tổng quan/Dashboard (FR-014, wireframe màn 1) → 34 kịch bản.
