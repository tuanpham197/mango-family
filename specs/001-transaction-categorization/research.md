# Phase 0 — Research: Phân Loại Giao Dịch (Go + Vue re-plan)

**Date**: 2026-07-10 · **Feature**: 001-transaction-categorization · **Plan**: [plan.md](./plan.md)

> ♻️ Thay thế bản research Flutter/Supabase 2026-06-29 (R1…R16 — xem git history). Các quyết định
> NGHIỆP VỤ của bản cũ được giữ nguyên và đánh dấu *(carry-over)*; quyết định STACK lấy từ
> [design re-platform](../../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md) (user đã chốt).

## D1. Backend framework & layout

- **Decision**: Go 1.22+ · Gin · GORM; layout module-first theo `github.com/tuanpham197/learn_go`: mỗi aggregate một `module/<x>/{model,biz,storage,transport}`; `common/` (app_error, app_response, paging, sql_model); `component/` (appctx, tokenprovider/jwt, hasher, pubsub, wshub); `middleware/`.
- **Rationale**: User chọn rõ ràng (design §1, quyết định #4/#5); map tự nhiên từ Clean Architecture cũ (biz ↔ usecases, storage ↔ datasources, transport ↔ presentation).
- **Alternatives**: chi/pgx/SSE (khuyến nghị ban đầu — user không chọn); Echo (tương đương, không chọn).

## D2. Migrations

- **Decision**: **goose** (`pressly/goose/v3`), SQL thuần tại `src/db/migrations/` (`00001_users.sql`…); KHÔNG dùng GORM AutoMigrate. Statement nhiều dòng bọc `-- +goose StatementBegin/StatementEnd`. Chạy qua CLI goose hoặc nhúng embed.FS.
- **Rationale**: User chốt 2026-07-10; SQL thuần giữ được CHECK/UNIQUE/index mà AutoMigrate không quản nổi.
- **Alternatives**: golang-migrate (quyết định ban đầu, bị thay); GORM AutoMigrate (mất kiểm soát schema).

## D3. Nơi thực thi bất biến nghiệp vụ (thay trigger cũ)

- **Decision**: Bất biến thực thi ở tầng **biz trong DB transaction** (GORM `Transaction`): loại bất biến sau tạo, con 1 cấp kế thừa loại, type giao dịch = type danh mục, delete-reassign nguyên tử. PostgreSQL giữ **hàng rào cuối**: CHECK (`type IN (...)`, `amount > 0`), FK, NOT NULL, UNIQUE.
- **Rationale**: API là writer duy nhất (khác thời Supabase client-direct); logic ở Go dễ test (biz unit test) và trả lỗi có ngữ nghĩa; trigger phức tạp (inherit type, type-match) khó bảo trì khi không còn bắt buộc.
- **Alternatives**: giữ nguyên bộ trigger cũ (thừa khi có API trung gian; khó debug); constraint-only (không diễn đạt được quy tắc xuyên bảng).

## D4. Auth & phiên đăng nhập *(nền tảng — thay Supabase Auth)*

- **Decision**: `users` kiêm credentials (`password_hash` bcrypt, cost mặc định); `POST /api/auth/login` → JWT (HS256, hạn 7 ngày, secret qua env) trong **cookie HttpOnly SameSite=Lax**; `POST /api/auth/logout` xóa cookie; `GET /api/me` trả hồ sơ + hộ. Không refresh token, không tự đăng ký (ngoài phạm vi — dev seed).
- **Rationale**: Design §2 đã chốt; phạm vi FR-016/002 chỉ cần login email+mật khẩu; bỏ hẳn cơ chế "đối chiếu qua email" (workaround Supabase cũ).
- **Alternatives**: session server-side (thêm bảng/state không cần thiết ở MVP); Authorization header (cookie HttpOnly an toàn XSS hơn cho web).

## D5. Phạm vi hộ (thay RLS) *(carry-over R10/R12 về ngữ nghĩa)*

- **Decision**: Middleware `authenticate` (JWT → user) + `RequireHousehold` (tra `household_members`, gắn `household_id` vào context); MỌI query storage filter theo `household_id`; bản ghi ngoài hộ trả **404** (không lộ tồn tại). Không có vai trò — ngang quyền (carry R11/FR-021).
- **Rationale**: Một tầng phân quyền duy nhất, dễ kiểm thử (biz/integration test); giữ đúng ngữ nghĩa cô lập của RLS cũ (FR-018).
- **Alternatives**: RLS trên Postgres tự quản (hai tầng, thừa khi API là cổng duy nhất).

## D6. Chống ghi đè thầm lặng *(carry-over R13)*

- **Decision**: Giữ `categories.updated_at` làm mốc lạc quan: client gửi `expected_updated_at` khi sửa/xóa; storage chạy `UPDATE/DELETE ... WHERE id = ? AND updated_at = ?`; 0 hàng → phân biệt `ErrConcurrencyConflict` (bản ghi đã đổi) vs `ErrRecordGone` (đã xóa). `updated_at` do biz/GORM hook cập nhật (bỏ trigger touch cũ).
- **Rationale**: Ngữ nghĩa giữ nguyên từ bản cũ (quickstart #16); GORM hỗ trợ conditional update tự nhiên.
- **Alternatives**: cột version int (tương đương, thêm cột mới vô cớ); khóa bi quan (quá nặng cho hộ 2–8 người).

## D7. Gợi ý danh mục *(carry-over — FR-015, không AI/ML)*

- **Decision**: Giữ mô hình `categorization_rules` (household_id, keyword, category_id, match_count): `GET /api/categories/suggest?description=&type=` — normalize (lowercase, bỏ dấu tùy chọn), match keyword chứa-trong-mô-tả cùng hộ + cùng loại, ưu tiên `match_count` cao nhất; khi lưu giao dịch có mô tả, biz **upsert** rule (keyword ← mô tả chuẩn hóa, tăng `match_count`) — "học" từ lịch sử chung của hộ.
- **Rationale**: Đúng Clarifications 2026-06-24 (rule-based + lịch sử, ghi đè được); chuyển từ RPC/SQL cũ sang biz Go thuần — dễ unit test.
- **Alternatives**: full-text search Postgres (quá cỡ cho keyword match MVP); AI/ML (Out of Scope BR-001).

## D8. Đồng bộ realtime giữa thành viên *(thay Supabase Realtime)*

- **Decision**: Local **pubsub** (mẫu learn_go) → subscriber đẩy sang **wshub** theo household; mutation danh mục/quy tắc/giao dịch publish `{type: "categories_changed"|"transactions_changed"}`; web nhận qua `WS /ws` (auth bằng cookie khi handshake) và **refetch** store liên quan. Fallback: refetch khi tab focus.
- **Rationale**: Design §3 đã chốt; giữ ngữ nghĩa invalidation của bản cũ nên use case/quickstart không đổi; mục tiêu ≤ 5s thoải mái với broadcast trực tiếp.
- **Alternatives**: SSE (đủ dùng nhưng user chọn WebSocket); polling (sát ngưỡng, tốn request).

## D9. Danh mục mặc định theo hộ *(carry-over R9)*

- **Decision**: Bộ danh mục mặc định (Chi: Ăn uống, Di chuyển, Hóa đơn, Mua sắm, Giải trí, Sức khỏe, Khác; Thu: Lương, Thưởng, Khác — chờ nghiệp vụ chốt) được seed **theo hộ** bằng logic app trong `module/household` (hàm `SeedDefaultCategories` gọi khi hộ được tạo); `cmd/seed` (dev-only) tạo users Alice/Bob/Carol + hộ A/B và gọi hàm này. KHÔNG seed qua migration.
- **Rationale**: Carry đúng R9 (seed theo hộ, `is_default=true`); tách seed dev khỏi migrations (design §2); hộ mới trong tương lai (feature Quản lý hộ) tái dùng hàm này.
- **Alternatives**: seed trong migration (trộn dữ liệu dev vào schema — đã loại ở design).

## D10. Front-end

- **Decision**: Vue 3 + Vite + TypeScript + **Pinia** (stores: auth, categories, transactions) + Vue Router (guard chưa đăng nhập → /login); mobile-first; gọi API qua fetch wrapper (`src/web/src/api/`), lỗi theo format `app_response`; components chính: `CategoryPicker` (lọc theo loại, nhóm cha/con, ẩn danh mục hidden), `SuggestionChip`, `DeleteReassignDialog`; Vite dev proxy `/api` + `/ws` → Go (cookie same-origin).
- **Rationale**: Design §1 đã chốt; Pinia store + WS invalidation tái tạo pattern Riverpod provider + realtime cũ.
- **Alternatives**: Nuxt (SSR không cần); axios (fetch đủ).

## D11. Testing

- **Decision**: Go — unit `biz` (validate, suggest, delete-reassign, conflict) với storage mock (interface theo learn_go), integration `storage` với Postgres docker (goose up trong test setup); `httptest` cho transport + middleware. Web — Vitest cho stores/components. E2E — **Playwright** tại `src/web/e2e/`: 17 kịch bản quickstart, multi-context cho #13–#17 (Alice/Bob cùng hộ + Carol hộ B).
- **Rationale**: Design §6 + quyết định Playwright 2026-07-10; multi-context thay cách "2 phiên e2e" cũ.
- **Alternatives**: Cypress (multi-session yếu hơn); testcontainers-go (dùng compose sẵn có đủ).

## D12. Xóa danh mục an toàn *(carry-over — thay RPC `delete_category`)*

- **Decision**: `DELETE /api/categories/{id}` body `{mode: "reassign"|"delete_transactions", target_category_id?, expected_updated_at}` — biz chạy MỘT DB transaction: (1) kiểm đích cùng loại + cùng hộ (FR-009), (2) xử lý danh mục con theo cùng cơ chế (FR-012), (3) gán lại hoặc xóa giao dịch liên quan, (4) xóa danh mục; 0 giao dịch liên quan → xóa ngay không cần mode.
- **Rationale**: Nguyên tử như RPC cũ nhưng nằm ở biz (test được); không cascade ngầm — không bao giờ để giao dịch mồ côi (SC-007).
- **Alternatives**: ON DELETE CASCADE (mất dữ liệu ngầm — cấm); soft-delete danh mục (ẩn đã có FR-020 riêng).

## Unknowns còn lại

- Danh sách danh mục mặc định cuối cùng: chờ nghiệp vụ (Open Question BR-001) — seed dev dùng danh sách đề xuất D9.
- Không còn NEEDS CLARIFICATION kỹ thuật nào.
