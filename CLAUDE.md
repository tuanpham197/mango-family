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
`specs/001-transaction-categorization/plan.md` (re-plan Go+Vue, 2026-07-10)
Stack decisions are authoritative in the re-platform design:
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
- **003-budgeting**: spec v1 complete (tech-agnostic), awaiting plan.

Model (unchanged): **shared family household** — multiple members, EQUAL permissions; data shared
per `household_id`, isolated across households. **Users are an INDEPENDENT directory** — the single
identity source referenced by household membership and every `created_by`; household-membership
authorization is enforced in the API layer (replaces the old RLS). Foundational order for 002:
**users FIRST** → accounts (default "Tiền mặt" per household; balance = derived view
`account_balances`) → transactions (`account_id`, `updated_at`, no future dates).
Design artifacts: research.md · data-model.md · contracts/ · quickstart.md in each feature's spec dir.
<!-- SPECKIT END -->
