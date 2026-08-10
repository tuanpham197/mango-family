# Quy trình phát triển tính năng mới (AIUP + Spec Kit) — Guide & Rules cho Claude

Tài liệu chuẩn cho **một vòng đời tính năng**: từ Business Requirement → spec → plan → tasks → review → implement → verify → commit → release → deploy. Dùng cho người **và** cho Claude (mục [Rules cho Claude](#rules-cho-claude) ở cuối là bản rút gọn để tuân thủ).

> Nguyên tắc nền: mỗi artifact có khối **Nguồn / Truy vết / References** — sửa X thì **lan truyền** xuống Y (xem bản đồ artifact trong [`CLAUDE.md`](../CLAUDE.md)). Bằng chứng trước khẳng định: **không tick "xong" khi chưa chạy lệnh kiểm chứng.**

---

## Bản đồ nhanh (thứ tự & đầu ra)

| # | Giai đoạn | Command chính | Đầu ra | Cổng review |
|---|-----------|---------------|--------|-------------|
| 0 | Business Requirement | (viết tay) | `specs/business-requirements/BR-XXX.md` | rà soát scope |
| 1 | Use case + sơ đồ | `aiup-core:use-case-spec`, `aiup-core:use-case-diagram` | `specs/use-cases/<feat>/`, `diagrams/use-cases.puml` (+png) | — |
| 2 | Spec | `/speckit-specify` | `specs/<NNN-feat>/spec.md` + `checklists/requirements.md` | checklist PASS |
| 3 | Làm rõ (nếu cần) | `/speckit-clarify` | spec cập nhật, hết `[NEEDS CLARIFICATION]` | — |
| 4 | Plan & design | `/speckit-plan` | `plan.md` · `research.md` · `data-model.md` · `contracts/` · `quickstart.md` | Constitution check |
| 5 | Tasks | `/speckit-tasks` | `tasks.md` | — |
| 6 | **Phân tích chéo** | `/speckit-analyze` | báo cáo lệch spec↔plan↔tasks | **gate: hết CRITICAL** |
| 7 | Implement (TDD) | `/speckit-implement` | code + test, tasks `[X]` | test FAIL→PASS |
| 8 | Verify | `make test` · `make test-api-integration` · `make test-e2e` | tất cả xanh | **gate** |
| 9 | Review code | `/code-review` · `/security-review` | sửa theo góp ý | — |
| 10 | Commit + merge | `git` | commit trên feature branch → merge `develop` | — |
| 11 | Release note | (viết tay) | `release/notes/YYYY-MM-DD-<slug>.md` + cập nhật index | — |
| 12 | Deploy | `deploy/api` + `deploy/web` → `./deploy/deploy.sh` / Vercel | API Cloud Run + Web Vercel | smoke test |

---

## 0. Business Requirement (BR)

- Tạo/sửa `specs/business-requirements/BR-XXX.md` (mẫu: `template.md`). Nêu **Goal, Success Metrics, In/Out of Scope, Constraints, Open Questions, References, History**.
- Mã con quy ước theo tên phân hệ, ví dụ BR-008 dùng `BR-MBR-001…` (member reports).
- **Review BR trước khi đi tiếp**: mâu thuẫn nội bộ? Open Question nào chặn scope? (ví dụ 008: đối soát tổng vs thành viên đã rời hộ).

## 1. Use case + sơ đồ (AIUP)

- `aiup-core:use-case-spec` → `specs/use-cases/<NNN-feature>/uc-<prefix>-NN-*.md` (**theo template repo**: Actor/Trigger/Preconditions/Main Flow/Alternative/Exceptions/Postconditions/Acceptance Criteria/Dependencies/Notes/History).
- `aiup-core:use-case-diagram` → cập nhật `specs/diagrams/use-cases.puml`, **render lại**:
  ```bash
  plantuml -tpng specs/diagrams/use-cases.puml
  # ⚠️ @startuml có tên chứa dấu cách → plantuml xuất file theo TÊN đó; ĐỔI TÊN về use-cases.png
  ```
- Cập nhật `specs/use-cases/README.md` (bảng chỉ mục + cây quan hệ).

## 2. Spec — `/speckit-specify`

- Mô tả tính năng (hoặc trỏ `@specs/business-requirements/BR-XXX.md`).
- Hook `before_specify` (`speckit.git.feature`) **tạo feature branch**. Đặt số **khớp BR** để tránh lệch số hiệu (dùng `GIT_BRANCH_NAME=<NNN>-<slug>` nếu cần ép số).
- Đầu ra: `specs/<NNN-feature>/spec.md` (User Stories P1/P2/P3 + FR + SC + Assumptions) và `checklists/requirements.md`.
- **Gate**: checklist PASS; nếu còn `[NEEDS CLARIFICATION]` → sang bước 3.

## 3. Làm rõ — `/speckit-clarify` (khi cần)

- Chỉ khi spec còn điểm chặn scope/an toàn. Trả lời tối đa vài câu, encode lại vào spec. Cập nhật FR/SC/Assumptions + History.

## 4. Plan & design — `/speckit-plan`

- Đầu ra: `plan.md` (Technical Context, Constitution Check, Project Structure), `research.md` (quyết định **DXX**, kế thừa số cũ), `data-model.md`, `contracts/`, `quickstart.md`.
- **Ground bằng code thật** trước khi viết: đọc module liên quan để plan khớp pattern hiện có (đừng bịa cấu trúc).
- Ghi rõ **có migration/dependency mới không** (đây là yếu tố quyết định bước deploy).
- Cập nhật con trỏ plan trong `CLAUDE.md` (khối `<!-- SPECKIT START/END -->`).

## 5. Tasks — `/speckit-tasks`

- Đầu ra `tasks.md`: Phase 1 Setup → Phase 2 Foundational → Phase 3+ theo **User Story** (Pn) → Polish.
- Mỗi task đúng format: `- [ ] T001 [P?] [USn?] Mô tả + đường dẫn file`. Có test task (TDD) nếu spec/plan yêu cầu.

## 6. Phân tích chéo — `/speckit-analyze` (GATE)

- Kiểm tra lệch **spec ↔ plan ↔ tasks** + coverage FR/SC + Constitution.
- **Không được** có CRITICAL trước khi implement. Ghi lại các quyết định (ví dụ "biểu đồ cột optional → hoãn").

## 7. Implement — `/speckit-implement` (TDD)

Thứ tự trong mỗi story: **test trước (phải FAIL)** → storage → biz → transport/route → FE store → FE component.
- Bám pattern repo (layout `model/biz/storage/transport`; FE Pinia store + component + view).
- Mark task `[X]` **chỉ khi** đã chạy và xanh. Ghi **Deviations** nếu làm khác câu chữ task (tránh mã chết / cross-module coupling).
- Backend cần DB: `cd src && make up && make migrate-up` (Postgres dev qua Docker).

## 8. Verify (GATE — bằng chứng thật)

```bash
cd src
go vet ./...                         # trong src/api
make test                            # Go unit + Vitest
make test-api-integration            # integration storage (Postgres thật)
make test-e2e                        # Playwright (cần API :8080 + web :5173)
# FE riêng: cd web && npx vue-tsc -b && npx vitest run
```
- Chạy seed dev khi cần dữ liệu e2e: `make seed` (DEV-ONLY; prod dùng `SEED_EMAIL_1` — không seed dữ liệu mẫu).
- **Không tuyên bố "xong" nếu chưa thấy output xanh.** Nếu integration cần Postgres mà chưa có → `make up` rồi chạy; nêu rõ nếu không chạy được.

## 9. Review code

- `/speckit-analyze` (lại, sau khi implement nếu spec đổi).
- `/code-review` (bug + dọn dẹp) · `/security-review` (nếu chạm auth/dữ liệu/nhập liệu) · `/simplify` (chất lượng).
- Nhận góp ý theo `superpowers:receiving-code-review`: xác minh kỹ thuật, không đồng ý cho có.

## 10. Commit + merge `develop`

```bash
git add -A                                   # KHÔNG add build artifact (vd src/api/seed → .gitignore)
git commit -m "feat(<scope>): ... (BR-XXX / feature NNN)"   # kết bằng dòng Co-Authored-By
git switch develop && git merge --no-ff <NNN-feature>
git push origin develop
```
- Commit **chỉ khi user yêu cầu**. Không gộp thay đổi không liên quan; nếu buộc kèm WIP sẵn có thì nói rõ trong body.
- Nhánh chính để mở PR/merge: **`develop`**.

## 11. Release note (một file / một lần deploy)

- Tạo `release/notes/YYYY-MM-DD-<slug>.md` (mẫu: `release/notes/2026-08-10-member-reports.md`): ngày deploy, version, feature, **cờ migration/dep/env**, tóm tắt, API, thứ tự deploy, kiểm chứng, rollback.
- Thêm dòng vào bảng chỉ mục `release/RELEASE-NOTES.md`. Chi tiết quy trình: `release/DEPLOY.md`.

## 12. Deploy (theo `release/DEPLOY.md`)

Nhánh deploy tách biệt: **`deploy/api`** (Cloud Run) · **`deploy/web`** (Vercel).
```bash
# API trước (nếu web phụ thuộc API mới)
git switch deploy/api && git merge --ff-only develop && git push origin deploy/api
# ⚠️ CHỈ chạy migration khi release CÓ migration mới. Feature read-only/không đổi schema → BỎ QUA.
#   (kiểm: so số file trong src/db/migrations vs DB version; deploy.sh KHÔNG tự migrate)
cd src && ./deploy/deploy.sh                 # cần gcloud auth đúng account/project
curl -s <cloud-run-url>/api/me               # 401 = API sống

# Web sau
git switch deploy/web && git merge --ff-only develop && git push origin deploy/web
# Vercel auto-build nếu Production Branch = deploy/web; hoặc: cd src/web && vercel --prod
```
- Ràng buộc: Cloud Run **`--max-instances 1`** (realtime in-memory) — đã set trong `deploy.sh`, đừng tăng.
- Đổi account gcloud: **user tự** `! gcloud auth login` (tương tác) — Claude không login hộ.
- Smoke test cuối: mở `https://<app>.vercel.app`, đăng nhập, kiểm màn của tính năng; DevTools Network thấy CORS + cookie đúng.

---

## Rules cho Claude

**Thứ tự bắt buộc**: BR → (use case) → `/speckit-specify` → (`/speckit-clarify`) → `/speckit-plan` → `/speckit-tasks` → `/speckit-analyze` → `/speckit-implement` → verify → review → commit/merge → release note → deploy. Không nhảy bước; `/speckit-tasks` cần `plan.md`, `/speckit-analyze`/`/speckit-implement` cần `tasks.md`.

**Bằng chứng trước khẳng định**: chỉ tick task/`✅` sau khi chạy lệnh và thấy output xanh (`go vet`, `make test`, `vitest`, `playwright`, integration trên Postgres thật). Nêu rõ cái gì chưa chạy được và vì sao.

**Lan truyền artifact**: sửa BR/spec/use case/entity → cập nhật các artifact dẫn xuất + khối Nguồn/Truy vết/References + một dòng History. Render lại `use-cases.png` và **đổi tên** về `use-cases.png`.

**Đánh số**: feature dir + UC folder + branch đặt **khớp số BR** để tránh lệch số hiệu (bài học BR-006↔005).

**Git**: chỉ commit/push/merge khi user yêu cầu; nhánh chính `develop`; commit message kết bằng
`Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`. Không commit build artifact; không gộp thay đổi lạc.

**Deploy**: đọc `release/DEPLOY.md`; **không chạy migration** trừ khi release có migration mới; `deploy.sh` không tự migrate nên an toàn; user tự làm bước `gcloud auth login`; Vercel cần Production Branch/CLI để build web (Claude không truy cập dashboard).

**An toàn**: hành động ra ngoài (deploy, push) cần user xác nhận; xác nhận ở ngữ cảnh này không tự động áp cho lần sau.
