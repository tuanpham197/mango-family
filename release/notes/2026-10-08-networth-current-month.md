# 2026-10-08 · v0.8.1 — Tổng tài sản ròng theo tháng hiện tại (feature 004)

- **Ngày deploy**: 2026-10-08
- **Version**: v0.8.1
- **Feature**: 004-monthly-income-expense-overview (điều chỉnh D28) · Nguồn: [research D28](../../specs/004-monthly-income-expense-overview/research.md) · [contract](../../specs/004-monthly-income-expense-overview/contracts/overview-api.md)
- **Nhánh deploy**: `deploy/api` (Cloud Run) + `deploy/web` (Vercel)
- **Migration**: ❌ không · **Dependency mới**: ❌ không · **Đổi env**: ❌ không

## Tóm tắt
Thẻ **Tổng tài sản ròng** trên màn Tổng quan (`/`) giờ hiển thị tài sản ròng của **tháng hiện tại** thay vì số dư cộng dồn từ trước đến nay.

## Thay đổi
- **Số tiền** = biến động ròng số dư trong tháng (số dư hiện tại − số dư đầu tháng). Không còn cộng dồn các tháng trước.
- **% thay đổi** = so với biến động ròng của **tháng trước**; vẫn ẩn (`null`) khi tháng trước bằng 0.
- **Nhãn thẻ**: "Tổng tài sản ròng" → "Tổng tài sản ròng tháng này".
- ⚠️ **Đổi ngữ nghĩa** trường `net_worth` / `net_worth_change_percent` của `GET /api/overview` (tên và kiểu trường giữ nguyên). Số hiển thị thường trùng `month.net` (Thu − Chi tháng); có thể khác nếu có số dư đầu kỳ/chuyển tiền.

## Database
✅ **KHÔNG có migration mới.** Giá trị suy ra từ `accounts.initial_balance` + giao dịch. DB version giữ nguyên **9**. → **Bỏ qua bước migration khi deploy.**

## Thứ tự deploy
1. Không migration, không đổi env → an toàn.
2. API đổi ngữ nghĩa trường mà web đọc → **deploy API (`deploy/api`) trước, web (`deploy/web`) sau** (nhãn mới chỉ khớp số liệu mới).
3. Smoke test `/` trên domain Vercel: thẻ ghi "tháng này", số khớp ô Thu/Chi tháng.

## Kiểm chứng trước phát hành
- Go: unit + integration (Postgres thật) xanh.
- Web: Vitest **94/94**.
- E2E: Playwright **80/80** (kịch bản #14 cập nhật: `net_worth` == `month.net`).

## Rollback
- **API**: về revision Cloud Run trước — `gcloud run services update-traffic household-finance-api --region asia-southeast1 --to-revisions <REV_CŨ>=100`. Read-only → không cần rollback DB.
- **Web**: Vercel → Promote bản trước; hoặc reset `deploy/web` về commit cũ rồi push.

## Ghi chú vận hành
- Chỉ **read-only** — không đổi dữ liệu nguồn.
- `vue-tsc` và `go vet` sạch.
