# Quy tắc artifact & lan truyền thay đổi (AIUP + Spec Kit)

Khi sửa MỘT artifact, phải rà soát & cập nhật các artifact dẫn xuất phía dưới rồi kiểm tra lại tính nhất quán. Mỗi file chứa sẵn khối liên kết **Nguồn / Truy vết / References** — luôn giữ chúng cập nhật để dễ lan truyền thay đổi.

## Bản đồ artifact (nguồn → dẫn xuất)

| Artifact | Vị trí |
|----------|--------|
| Business Requirement (BR) | `specs/business-requirements/BR-*.md` |
| Feature spec | `specs/<feature>/spec.md` (dẫn xuất từ một BR) |
| Use case | `specs/use-cases/uc-*.md` (+ `README.md` chỉ mục) |
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
`specs/001-transaction-categorization/plan.md`

Active feature: **001-transaction-categorization** (Phân loại giao dịch).
Model: **shared family household** — multiple members, EQUAL permissions; categories + transactions
are shared per `household_id` (RLS by membership, isolated across households). Transactions record
`created_by` (which member entered them). (Reversed original FR-018 per-user — spec.md/BR-001 already reconciled.)
Stack: Flutter (Dart 3.x, iOS+Android) · Riverpod · Supabase (PostgreSQL + Auth + RLS + Realtime).
Source lives under **`src/`** (Flutter project root = `src/`): feature-first / Clean Architecture under `src/lib/features/categorization`; `src/lib/core/household` holds current-household context; migrations in `src/supabase/migrations`.
Design artifacts: research.md · data-model.md · contracts/ · quickstart.md in the same spec dir.
<!-- SPECKIT END -->
