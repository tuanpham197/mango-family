# Quy tắc artifact & lan truyền thay đổi (AIUP + Spec Kit)

Khi sửa MỘT artifact, phải rà soát & cập nhật các artifact dẫn xuất phía dưới rồi kiểm tra lại tính nhất quán. Mỗi file chứa sẵn khối liên kết **Nguồn / Truy vết / References** — luôn giữ chúng cập nhật để dễ lan truyền thay đổi.

## Bản đồ artifact (nguồn → dẫn xuất)

| Artifact | Vị trí |
|----------|--------|
| Business Requirement (BR) | `specs/business-requirements/BR-*.md` |
| Feature spec | `specs/<feature>/spec.md` (dẫn xuất từ một BR) |
| Use case | `specs/use-cases/<feature>/uc-*.md` — gộp theo folder feature, vd `002-transaction-tracking/uc-trk-01-*.md` (+ `README.md` chỉ mục chung) |
| Use case diagram | `specs/diagrams/use-cases.puml` (+ ảnh `use-cases.png`) |
| Entity model | `specs/entities/entity-model.md` |
| Plan & design | `specs/<feature>/plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md` |

## Quy tắc lan truyền (sửa X → rà soát/cập nhật Y)

- **BR** → `spec.md` (phạm vi/FR), use case, entity model, use case diagram, plan + design artifacts.
- **spec.md (FR / clarification)** → BR (nếu đổi phạm vi), use case liên quan, actor & quan hệ trong diagram, mục Key Entities của entity model, `plan.md`/`research.md`/`data-model.md`/`contracts/`/`quickstart.md`.
- **Use case** → use case diagram, `use-cases/README.md` (chỉ mục), Acceptance Criteria trong spec, kịch bản trong `quickstart.md`.
- **Entity model** → `data-model.md`, `contracts/db-schema.sql`, mục Key Entities của spec.
- **Use case diagram (.puml)** → render lại PNG: `plantuml -tpng specs/diagrams/use-cases.puml` rồi đổi tên kết quả về `use-cases.png`; cập nhật `use-cases/README.md`.

## Luôn làm sau khi sửa

- Cập nhật khối **Nguồn/Truy vết/References** trong mỗi file đã sửa để trỏ đúng nguồn hiện tại.
- Ghi một dòng vào History/Changelog của artifact (vd BR có `## History`).
- Chạy `/speckit-analyze` để phát hiện lệch nhau giữa spec ↔ plan ↔ tasks sau khi cập nhật.

<!-- SPECKIT START -->
For technologies, project structure, and other important context, read the current plan:
`specs/005-reports/plan.md` (Go+Vue, 2026-07-15) — read-only Reports (overview: income/expense/net +
category-distribution + trend charts; per-category detail) over a selectable date range, on the
implemented 001–004 foundation; **adds the first new FE dependency `chart.js`** (thin in-house Vue
wrapper). Base stack/structure context in `specs/001-transaction-categorization/plan.md`. Stack
decisions are authoritative in the re-platform design:
`docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md`

> ⚠️ **Re-platform (2026-07-09/10)**: Flutter + Supabase have been REMOVED (code in `src/` deleted;
> recoverable via git history). New stack: **Go API (Gin + GORM + WebSocket, migrations via goose)
> + Vue 3 (Vite/TS/Pinia) + self-managed PostgreSQL**; responsive web, mobile-first. Go API layout
> follows `github.com/tuanpham197/learn_go` (module-first `model/biz/storage/transport`); e2e =
> Playwright. Source lives at `src/api` (Go) · `src/web` (Vue) · `src/db/migrations` (goose) —
> scaffolded & 001 implemented. Dev workflow: `cd src && make dev && make seed`, then `make api` +
> `make web`; tests `make test` (Go unit + Vitest) · `make test-api-integration` · `make test-e2e`.

Feature status:
- **001-transaction-categorization**: ✅ **implemented (Go + Vue)** — 35/35 task xong & kiểm chứng
  (Go unit + integration Postgres thật, Vitest 11/11, Playwright e2e 18/18 kịch bản #0–#17). Gồm nền
  tảng định danh tối thiểu (users + JWT cookie + household scope), US1–US4 (bắt buộc phân loại cùng
  loại, CRUD danh mục + concurrency `expected_updated_at`, danh mục con 1 cấp, gợi ý học lịch sử),
  đồng bộ realtime qua WebSocket. Còn mở: bộ danh mục mặc định chờ nghiệp vụ; Quản lý hộ dùng dev seed.
- **002-transaction-tracking**: ✅ **implemented (Go + Vue)** — 28/28 task xong & kiểm chứng
  (Go unit + integration, Vitest 20/20, Playwright e2e 33/33 = 18 của 001 + 15 kịch bản 002). Mở
  rộng nền tảng 001: `module/account` (đọc + view số dư `account_balances`), vòng đời giao dịch
  đầy đủ (tài khoản bắt buộc, chặn ngày tương lai, mốc lạc quan `updated_at` cho sửa/xóa), sổ chung
  phân trang + realtime, tài khoản mặc định "Tiền mặt" seed theo hộ. Migrations goose 00006 (accounts)
  · 00007 (transactions.account_id + updated_at + view `account_balances`). Còn mở: quản lý tài khoản
  đầy đủ (CRUD/chuyển tiền) thuộc BR-005.
- **003-budgeting**: ✅ **implemented (Go + Vue)** — 34/34 task xong & kiểm chứng (Go unit biz+model,
  integration storage+evaluator trên Postgres thật, Vitest 40/40, Playwright e2e 22 spec/003 = toàn bộ
  34 kịch bản quickstart). Module `budget` (model/storage/biz/transport/**eval**) trên nền 001/002 (chỉ
  ĐỌC giao dịch/danh mục); migrations goose `00008_budgets` · `00009_budget_alerts`. Tiến độ = giá trị
  **suy ra** ở biz (cửa sổ kỳ từ `now()`+`period_type`; danh mục con 1 cấp; chỉ EXPENSE — không view,
  không cột spent); cảnh báo 80%/vượt 100% qua máy trạng thái `budget_alerts` (chống trùng + phát lại),
  đánh giá bởi **subscriber pubsub `module/budget/eval`** trên `transactions_changed` + recompute inline
  khi CRUD ngân sách → publish `budgets_changed` (realtime ≤ 5s). Sửa/xóa ngang quyền, mốc lạc quan
  `updated_at` (PATCH merge trường vắng). **FR-014**: widget tóm tắt ngân sách + "Xem tất cả" trên màn
  Tổng quan (feature 004 chuyển màn này lên `/`). Còn mở: ngưỡng cấu hình + push/email (mở rộng sau);
  sửa/xóa là vòng đời suy luận chờ nghiệp vụ xác nhận.
- **004-monthly-income-expense-overview**: ✅ **implemented (Go + Vue)** — 24/24 task xong & kiểm chứng
  (Go unit biz + integration storage + httptest, Vitest 52/52, Playwright e2e 12 spec/004 + toàn bộ
  001/002/003 xanh sau khi lan truyền route). Màn **Tổng quan đầy đủ** khớp `specs/design/dashboard.png`:
  lời chào → **Tổng tài sản ròng** (+% so tháng trước) → **Thu/Chi tháng** → **Chi tiêu theo danh mục**
  (MỚI, trước ngân sách, gộp con) → **Ngân sách** (widget 003) → **Giao dịch gần đây**. Module **read-only**
  `overview` phơi một endpoint tổng hợp `GET /api/overview` (giá trị suy ra; **không bảng/migration**);
  ngân sách vẫn dùng `GET /api/budgets`. **Đổi trang mặc định (D31)**: `/` = Tổng quan (trang chủ), sổ giao
  dịch → **`/ledger`**, `/overview`→`/` redirect; nav 5 mục **Tổng quan · Giao dịch · ＋ · Ngân sách · Báo cáo**
  (Báo cáo = placeholder BR-004; **Danh mục** thành lối phụ trên header Tổng quan). Realtime: store `overview`
  refetch theo transactions/accounts/budgets/categories_changed. Còn mở: Báo cáo (BR-004); có bỏ hẳn Danh mục
  khỏi nav không (chờ nghiệp vụ). Xem [[budget-dashboard-route]].
- **005-reports**: ✅ **implemented (Go + Vue)** — 23/23 task xong & kiểm chứng (Go unit biz+model +
  integration storage + httptest, Vitest 63/63, Playwright e2e 5 spec/005 + toàn bộ 001–004 xanh). Màn
  **Báo cáo** (`/reports`, thay placeholder của 004): báo cáo tổng quan theo khoảng chọn được (tuần/tháng/
  quý/năm/tùy chỉnh) — tổng Thu/Chi/ròng + **phân bổ chi theo danh mục** (donut) + **xu hướng thu/chi**
  (line); **báo cáo chi tiết theo danh mục** (drill-in: tổng + danh sách giao dịch + xu hướng danh mục).
  Module **read-only** `report` với 2 endpoint `GET /api/reports/overview` + `/api/reports/category/:id`
  (tổng hợp bằng SQL `SUM`/`date_trunc`; gộp danh mục con; **không bảng/migration**); FE quy preset →
  from/to, server chọn đơn vị gom nhóm (day/week/month) + điền mốc trống. **Dependency FE mới đầu tiên:
  `chart.js`** (bọc wrapper Vue mỏng `components/charts/`). Còn mở: xuất PDF/Excel, forecasting, benchmark
  (Out of Scope BR); **số hiệu BR lệch**: file `specs/business-requirements/BR-006.md` mang ID nội bộ
  "BR-004" (không có BR-004.md) — cần nghiệp vụ thống nhất.

Model (unchanged): **shared family household** — multiple members, EQUAL permissions; data shared
per `household_id`, isolated across households. **Users are an INDEPENDENT directory** — the single
identity source referenced by household membership and every `created_by`; household-membership
authorization is enforced in the API layer (replaces the old RLS). Foundational order for 002:
**users FIRST** → accounts (default "Tiền mặt" per household; balance = derived view
`account_balances`) → transactions (`account_id`, `updated_at`, no future dates).
Design artifacts: research.md · data-model.md · contracts/ · quickstart.md in each feature's spec dir.
<!-- SPECKIT END -->
