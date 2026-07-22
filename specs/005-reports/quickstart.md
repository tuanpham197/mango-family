# Quickstart — Validate Báo Cáo (feature 005, Go + Vue)

**Date**: 2026-07-15 · **Plan**: [plan.md](./plan.md) · **Contracts**: [contracts/report-api.md](./contracts/report-api.md) · **Data model**: [data-model.md](./data-model.md)

Tiền đề: feature 001 + 002 + 004 đã chạy (login, phạm vi hộ, danh mục, giao dịch, màn Tổng quan + mục nav "Báo cáo" placeholder). Feature 005 **chỉ đọc** để dựng màn Báo cáo (tổng quan + chi tiết theo danh mục, biểu đồ). **Không migration mới**; **thêm dependency FE `chart.js`**.

## Setup

```bash
cd src && make dev            # Postgres + migrations 00001–00009 (đã có, không thêm mới)
make seed                     # Alice/Bob (hộ A) · Carol (hộ B); Password123!
cd web && npm install         # cài chart.js (dependency FE mới)
cd .. && make api             # Go API :8080 (thêm route /api/reports/*)
make web                      # Vite; mở /reports (nay là màn Báo cáo thật)
```

## Kiểm chứng theo Acceptance Criteria

Mỗi kịch bản tương ứng một Playwright spec trong `src/web/e2e/`. Đa thành viên (Alice/Bob hộ A, Carol hộ B) cho cô lập hộ.

### US1 — Báo cáo tổng quan theo khoảng thời gian

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 1 | Mở Báo cáo, khoảng "Tháng này" | Hiển thị tổng thu, tổng chi, số dư ròng đúng tổng giao dịch trong tháng | FR-002, SC-001 |
| 2 | Đổi khoảng (tuần/quý/năm) | Các con số cập nhật theo khoảng, không rời màn | FR-001, SC-003 |
| 3 | Khoảng tùy chỉnh hợp lệ (end ≥ start) | Báo cáo tính đúng khoảng đã chọn | FR-001 |
| 4 | Khoảng tùy chỉnh end < start | Bị chặn kèm thông báo (400) | FR-008 |
| 5 | Khoảng trống (không giao dịch) | Thu/chi/ròng = 0, biểu đồ/empty state, không lỗi | FR-009, SC-004 |

### US2 — Phân bổ chi tiêu theo danh mục

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 6 | Biểu đồ phân bổ | Mỗi danh mục Chi kèm số tiền + %, tổng khớp tổng chi; sắp giảm dần | FR-003, SC-001 |
| 7 | Chi ở danh mục con | Gộp vào danh mục cha | FR-003 |
| 8 | Không có chi | Biểu đồ hiển thị trạng thái trống | FR-009 |

### US3 — Xu hướng thu/chi theo thời gian

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 9 | Biểu đồ đường thu/chi | Hai đường theo mốc thời gian; mỗi mốc khớp tổng giao dịch trong mốc | FR-004, SC-001 |
| 10 | Khoảng ngắn vs dài | Đơn vị gom nhóm phù hợp (ngày cho tuần/tháng, tháng cho quý/năm) | FR-004, D36 |

### US4 — Báo cáo chi tiết theo danh mục

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 11 | Chọn một danh mục | Tổng chi danh mục + danh sách mọi giao dịch (số tiền/mô tả/ngày), gồm con | FR-005, SC-005 |
| 12 | Đổi khoảng khi đang xem chi tiết | Tổng & danh sách cập nhật theo khoảng mới | FR-005 |
| 13 | Danh mục không giao dịch trong kỳ | Tổng = 0, danh sách trống | FR-009 |
| 14 | Gọi API `:id` danh mục hộ khác | 404 | FR-007 |

### US5 — Xu hướng chi tiêu của một danh mục

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 15 | Biểu đồ đường của danh mục | Chi của danh mục theo mốc thời gian; mỗi mốc khớp tổng chi danh mục trong mốc | FR-006, SC-001 |

### Cô lập hộ

| # | Kịch bản | Kết quả mong đợi | Truy vết |
|---|----------|------------------|----------|
| 16 | Carol (hộ B) mở Báo cáo | Chỉ thấy số liệu hộ B; API chỉ trả dữ liệu hộ đang đăng nhập | FR-007 |

## Kiểm chứng phi UI (Go)

```bash
cd src && make test                      # Go unit (giải khoảng, chọn đơn vị gom nhóm, gộp cây, net) + Vitest
make test-api-integration                # storage report trên Postgres thật (SUM/trend bucket khớp, cô lập hộ)
make test-e2e                            # Playwright (đổi khoảng, phân bổ/xu hướng, drill-in, empty state)
```

**Điểm cần chứng minh bằng test**: số liệu khớp giao dịch theo khoảng (SC-001); báo cáo 1 tháng ≤ 2s (SC-002); đổi khoảng cập nhật tại chỗ (SC-003); phân bổ/xu hướng đúng + empty state (SC-004); drill-in ≤ 2 thao tác (SC-005).

## History

- v1 (2026-07-15): tạo mới — 16 kịch bản kiểm chứng US1–US5 + cô lập hộ; không migration; thêm route `/api/reports/*` + dependency FE `chart.js`; màn `/reports` thay placeholder feature 004.
