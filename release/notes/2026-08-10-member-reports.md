# 2026-08-10 · v0.8.0 — Báo cáo thu chi theo thành viên (feature 008)

- **Ngày deploy**: 2026-08-10
- **Version**: v0.8.0
- **Feature**: 008-member-reports · Nguồn: [BR-008](../../specs/business-requirements/BR-008.md) · [spec](../../specs/008-member-reports/spec.md)
- **Nhánh deploy**: `deploy/api` (Cloud Run) + `deploy/web` (Vercel)
- **Migration**: ❌ không · **Dependency mới**: ❌ không · **Đổi env**: ❌ không

## Tóm tắt
Mở rộng màn **Báo cáo** (`/reports`) với phần **theo thành viên**: tổng Thu/Chi/ròng của từng thành viên hộ theo khoảng thời gian chọn được, và drill-in danh sách giao dịch của một thành viên.

## Mới
- **Báo cáo theo thành viên**: mỗi thành viên hiện tại của hộ hiển thị tổng **Thu / Chi / ròng (Thu − Chi)** trong khoảng (tuần/tháng/quý/năm/tùy chỉnh). Thành viên chưa có giao dịch vẫn hiện **0/0/0**.
- **Đối soát**: tổng các dòng luôn khớp tổng toàn hộ; giao dịch của người **đã rời hộ** gộp một dòng **"Thành viên cũ / Đã rời hộ"** (FR-010) để không phá vỡ đối soát.
- **Drill-in**: chọn một thành viên → danh sách giao dịch (phân trang, có trạng thái trống), quay lại giữ nguyên khoảng.

## API (không phá vỡ tương thích)
- `GET /api/reports/members?from=&to=` — báo cáo theo thành viên + `totals`.
- `GET /api/reports/member/:id?from=&to=&page=&page_size=` — drill-in; `:id` = UUID thành viên (thuộc hộ) hoặc sentinel `former`; ngoài hộ → **404**.
- Quy thuộc theo `transactions.created_by`; cô lập theo `household_id`.

## Database
✅ **KHÔNG có migration mới.** Toàn bộ là giá trị suy ra (SQL `SUM ... GROUP BY created_by`) trên bảng hiện có. DB version giữ nguyên **9** (`00009_budget_alerts`). → **Bỏ qua bước migration khi deploy release này.**

## Thứ tự deploy
1. Không migration, không đổi env → an toàn.
2. Có **API mới** mà web dùng → **deploy API (`deploy/api`) trước, web (`deploy/web`) sau**.
3. Smoke test `/reports` phần theo thành viên trên domain Vercel.

## Kiểm chứng trước phát hành
- Go: `go vet` sạch · unit + integration (Postgres thật) xanh — đối soát Σ thành viên == tổng hộ.
- Web: `vue-tsc` sạch · **Vitest 94/94**.
- E2E: Playwright `member-report` **4/4** + regression 005 **5/5**.

## Rollback
- **API**: về revision Cloud Run trước — `gcloud run services update-traffic household-finance-api --region asia-southeast1 --to-revisions <REV_CŨ>=100`. Read-only → không cần rollback DB.
- **Web**: Vercel → Promote bản trước; hoặc reset `deploy/web` về commit cũ rồi push.

## Ghi chú vận hành
- Chỉ **read-only** — không đổi dữ liệu nguồn.
- Seed dev có thêm thành viên "Dave" (không giao dịch) + giao dịch mẫu Alice/Bob — **DEV-ONLY**, prod (`SEED_EMAIL_1`) không chạm.
- Commit: `e0a1ba6` (feature) · merge `472bfdd` vào `develop`.
