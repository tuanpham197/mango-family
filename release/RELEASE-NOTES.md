# Release Notes — Mango Family (Sổ tài chính hộ gia đình)

Ghi chú phát hành, **mới nhất trước**. Mỗi mục nêu: phạm vi, thay đổi API/DB, và **có cần migration không** (quyết định bước deploy). Quy trình & lệnh deploy: [`release/DEPLOY.md`](./DEPLOY.md).

Nền tảng: Go API (Cloud Run) · Vue SPA (Vercel) · Postgres (Supabase). Kiến trúc tách origin — xem [`docs/deploy-cloud-run.md`](../docs/deploy-cloud-run.md).

---

## v0.8.0 — 2026-08-10 · Feature 008: Báo cáo thu chi theo thành viên

**Tóm tắt**: Mở rộng màn **Báo cáo** (`/reports`) với phần **theo thành viên** — tổng Thu/Chi/ròng của từng thành viên hộ theo khoảng thời gian chọn được, và drill-in danh sách giao dịch của một thành viên. Nguồn: [BR-008](../specs/business-requirements/BR-008.md) · [spec](../specs/008-member-reports/spec.md).

### Mới
- **Báo cáo theo thành viên**: mỗi thành viên hiện tại của hộ hiển thị tổng **Thu / Chi / ròng (Thu − Chi)** trong khoảng (tuần/tháng/quý/năm/tùy chỉnh). Thành viên chưa có giao dịch vẫn hiện **0/0/0**.
- **Đối soát**: tổng các dòng luôn khớp tổng toàn hộ; giao dịch của người **đã rời hộ** gộp một dòng **"Thành viên cũ / Đã rời hộ"** (FR-010) để không phá vỡ đối soát.
- **Drill-in**: chọn một thành viên → danh sách giao dịch (phân trang, có trạng thái trống), quay lại giữ nguyên khoảng.

### API (không phá vỡ tương thích)
- `GET /api/reports/members?from=&to=` — báo cáo theo thành viên + `totals`.
- `GET /api/reports/member/:id?from=&to=&page=&page_size=` — drill-in; `:id` = UUID thành viên (thuộc hộ) hoặc sentinel `former`; ngoài hộ → **404**.
- Quy thuộc theo `transactions.created_by`; cô lập theo `household_id`.

### Database
- ✅ **KHÔNG có migration mới.** Toàn bộ là giá trị suy ra (SQL `SUM ... GROUP BY created_by`) trên bảng hiện có. DB version giữ nguyên **9** (`00009_budget_alerts`). → **Bỏ qua bước migration khi deploy release này.**

### Dependency
- ✅ **Không thêm dependency** (BE/FE). Tái dùng `chart.js` (đã có từ 005), `ListItem`, `Paging`.

### Kiểm chứng trước phát hành
- Go: `go vet` sạch · unit + integration (Postgres thật) xanh — đối soát Σ thành viên == tổng hộ.
- Web: `vue-tsc` sạch · **Vitest 94/94**.
- E2E: Playwright `member-report` **4/4** + regression 005 **5/5**.

### Ghi chú vận hành
- Chỉ **read-only** — không đổi dữ liệu nguồn, an toàn rollback (chỉ cần deploy lại bản trước).
- Seed dev có thêm thành viên "Dave" (không giao dịch) + giao dịch mẫu Alice/Bob — **DEV-ONLY**, prod (`SEED_EMAIL_1`) không chạm.

---

## Trước đó (đã triển khai — tóm tắt)

> Chi tiết từng feature xem `CLAUDE.md` (mục Feature status) và `specs/<feature>/`.

- **v0.5.0 — 005 Báo cáo**: màn `/reports` (tổng quan + phân bổ danh mục + xu hướng + drill-in danh mục). Thêm dep FE `chart.js`. Không migration.
- **v0.4.0 — 004 Tổng quan tháng**: màn Tổng quan `/`, `GET /api/overview`. Không migration.
- **v0.3.0 — 003 Ngân sách**: module `budget`, migration `00008_budgets` · `00009_budget_alerts`.
- **v0.2.0 — 002 Ghi chép thu chi**: accounts + vòng đời giao dịch, migration `00006_accounts` · `00007_transactions_v2`.
- **v0.1.0 — 001 Phân loại giao dịch**: nền tảng users/JWT/household + danh mục, migration `00001…00005`.
