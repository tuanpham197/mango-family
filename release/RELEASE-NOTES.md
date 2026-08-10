# Release Notes — Mango Family (Sổ tài chính hộ gia đình)

**Chỉ mục** ghi chú phát hành. Mỗi lần deploy = **một file riêng** trong [`release/notes/`](./notes/), đặt tên theo **ngày deploy**: `notes/YYYY-MM-DD-<slug>.md`. Danh sách mới nhất trước.

Nền tảng: Go API (Cloud Run) · Vue SPA (Vercel) · Postgres (Supabase). Quy trình & lệnh deploy: [`DEPLOY.md`](./DEPLOY.md) · hạ tầng gốc: [`../docs/deploy-cloud-run.md`](../docs/deploy-cloud-run.md).

| Ngày deploy | Version | Feature | Migration | File |
|-------------|---------|---------|:---------:|------|
| 2026-08-10 | v0.8.0 | 008 — Báo cáo theo thành viên | ❌ | [notes/2026-08-10-member-reports.md](./notes/2026-08-10-member-reports.md) |

## Quy ước tạo release note mới

Khi deploy một release:
1. Tạo `release/notes/YYYY-MM-DD-<slug>.md` (dùng [`notes/2026-08-10-member-reports.md`](./notes/2026-08-10-member-reports.md) làm mẫu): ngày deploy, version, feature, **có migration/dep/env đổi không**, tóm tắt, API, thứ tự deploy, kiểm chứng, rollback.
2. Thêm một dòng vào bảng chỉ mục phía trên (mới nhất trước).

---

## Lịch sử trước v0.8.0 (chưa tách theo ngày deploy)

> Các release 001–005 được phát hành **theo tính năng** trước khi áp dụng quy ước "một file/ngày deploy"; không lưu mốc deploy chính xác nên gộp tóm tắt ở đây. Chi tiết từng feature: `CLAUDE.md` (Feature status) và `specs/<feature>/`.

- **v0.5.0 — 005 Báo cáo**: màn `/reports` (tổng quan + phân bổ danh mục + xu hướng + drill-in danh mục). Thêm dep FE `chart.js`. Không migration.
- **v0.4.0 — 004 Tổng quan tháng**: màn Tổng quan `/`, `GET /api/overview`. Không migration.
- **v0.3.0 — 003 Ngân sách**: module `budget`, migration `00008_budgets` · `00009_budget_alerts`.
- **v0.2.0 — 002 Ghi chép thu chi**: accounts + vòng đời giao dịch, migration `00006_accounts` · `00007_transactions_v2`.
- **v0.1.0 — 001 Phân loại giao dịch**: nền tảng users/JWT/household + danh mục, migration `00001…00005`.
