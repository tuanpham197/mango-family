# Design: Re-platform từ Flutter/Supabase sang Golang API + Vue.js

**Ngày**: 2026-07-09 · **Trạng thái**: Đã duyệt (brainstorming với người dùng)
**Phạm vi**: Thay toàn bộ stack hiện thực (Flutter mobile + Supabase BaaS) bằng Go API + Vue.js web; cập nhật mọi tài liệu Spec Kit / AIUP bị ảnh hưởng. Code cũ trong `src/` đã bị xóa (khôi phục được qua git history nếu cần tham khảo).

## Bối cảnh & các quyết định đã chốt

| # | Câu hỏi | Quyết định |
|---|---------|-----------|
| 1 | Số phận Supabase | **Bỏ hẳn** — Go API + PostgreSQL tự quản; auth và phân quyền trong Go; realtime tự dựng |
| 2 | Feature 001 (đã từng implement bằng Flutter, code đã xóa) | **Re-plan cho stack mới** — giữ nguyên spec/BR/use case (tech-agnostic); viết lại plan/research/data-model/contracts/quickstart; reset tasks; 002 tiếp tục phụ thuộc 001 |
| 3 | Hình thái sản phẩm | **Web responsive, mobile-first** (thay iOS+Android+Web của Flutter) |
| 4 | Kiến trúc | **Gin + GORM + WebSocket** (người dùng chọn, thay khuyến nghị chi/pgx/SSE) |
| 5 | Cấu trúc Go API | Theo repo tham khảo **`tuanpham197/learn_go`** (module-first: model/biz/storage/transport) |

## §1. Kiến trúc hệ thống

Monorepo giữ `src/` làm gốc:

```
src/
├── api/                          # Go 1.22+ · Gin · GORM · layout theo learn_go
│   ├── main.go                   # entry: load env, kết nối Postgres (GORM), appctx, chạy router
│   ├── main_route.go             # đăng ký route các module
│   ├── .env.example · .air.toml · Dockerfile
│   ├── common/                   # app_error, app_response, paging, sql_model (base GORM), const, uid
│   ├── component/
│   │   ├── appctx/               # app context: db (GORM), secret, pubsub, ws hub
│   │   ├── tokenprovider/jwt/    # phát & xác thực JWT
│   │   ├── hasher/               # bcrypt (KHÁC repo mẫu: không dùng md5 cho mật khẩu)
│   │   ├── pubsub/               # local pubsub (theo mẫu learn_go)
│   │   ├── subscriber/           # subscriber đẩy event pubsub → WebSocket hub
│   │   └── wshub/                # WebSocket hub theo household (gorilla/websocket)
│   ├── middleware/               # authenticate (JWT), require_household_member, recover
│   └── module/
│       ├── user/                 # model/biz/storage/transport-ginuser: login, profile (FR-015/016)
│       ├── household/            # membership, scope theo hộ
│       ├── category/             # feature 001: danh mục + rules + gợi ý
│       ├── account/              # accounts + số dư (view account_balances)
│       └── transaction/          # feature 002: nhập/sổ/sửa/xóa giao dịch
├── web/                          # Vue 3 + Vite + TypeScript + Pinia + Vue Router · mobile-first
│   └── src/{views,components,stores,composables,api,router}
├── db/migrations/                # SQL thuần qua golang-migrate (KHÔNG dùng GORM AutoMigrate)
└── docker-compose.yml            # Postgres 16 cho dev (+ api dev qua air nếu muốn)
```

- Mỗi module theo đúng mẫu learn_go: `model/` (GORM entity + filter) · `biz/` (business logic, unit-testable) · `storage/` (GORM queries) · `transport/gin<module>/` (handlers). Map tự nhiên từ Clean Architecture cũ: biz ↔ usecases, storage ↔ datasources, transport ↔ presentation.
- **Migrations là SQL thuần** dù dùng GORM: view `account_balances` (mỏ neo SC-004) và các CHECK constraint (amount > 0, chặn ngày tương lai) không quản được bằng AutoMigrate. Schema tái chế từ `contracts/db-schema.sql` cũ, bỏ toàn bộ phần Supabase (schema `auth`, RLS policies, `current_user_id()`).
- Logic các trigger cũ (`set_created_by`, touch `updated_at`, enforce rules) chuyển vào tầng `biz`; DB giữ constraint làm hàng rào cuối.
- Khác biệt cố ý so với repo mẫu: **Postgres thay MySQL**, **bcrypt thay md5**, thêm `wshub` (repo mẫu không có WebSocket).

## §2. Auth & phân quyền (thay Supabase Auth + RLS)

- `users` thêm `password_hash` (bcrypt). Bảng users là nguồn định danh **kiêm** đăng nhập trực tiếp — cơ chế "đối chiếu qua email" cũ (workaround cho Supabase Auth) bị loại bỏ.
- Đăng nhập email + mật khẩu → JWT (tokenprovider) trong cookie HttpOnly, SameSite=Lax; middleware `authenticate` nạp user vào context.
- Phân quyền theo hộ: middleware/biz buộc mọi truy vấn filter theo `household_id` mà user là thành viên — **một tầng duy nhất thay RLS**; resource ngoài hộ trả 404.
- FR-014 (không ghi đè thầm lặng): optimistic locking bằng `updated_at` (`UPDATE ... WHERE id = ? AND updated_at = ?`); 0 hàng → phân biệt ConcurrencyConflict (bản ghi đã đổi) vs RecordGone (đã xóa).
- Seed dev: alice/bob/carol (hộ A/B) như quickstart cũ.
- Ngoài phạm vi (giữ nguyên như spec): tự đăng ký, quên mật khẩu, đổi email, xóa tài khoản.

## §3. Realtime (SC-006 ≤ 5s)

Mutation thành công → publish event lên **pubsub local** (mẫu learn_go) → subscriber broadcast qua **WebSocket hub theo household**: `{type: "transactions_changed" | "categories_changed" | "accounts_changed"}` → front-end refetch dữ liệu đang xem. Cùng ngữ nghĩa "invalidation" như Supabase Realtime cũ nên use case/quickstart không đổi nghiệp vụ. Fallback: refetch khi tab focus/visibility.

## §4. Bản đồ cập nhật tài liệu

Spec/BR/use case về cơ bản **giữ nguyên** (tech-agnostic) — chỉ tầng design đổi:

| Nhóm | File | Hành động |
|---|---|---|
| Gốc | `CLAUDE.md` | Stack mới, layout `src/api·web·db`, bỏ mô tả Supabase/email-mapping, 001 → "re-implementation pending" |
| BR | `BR-001` (nhắc Supabase) | Trung lập hóa tham chiếu + History line; `BR-002` không đổi |
| 001 | `plan/research/data-model/contracts/quickstart` | Viết lại cho Go+Vue; `contracts/` → REST API contract; `tasks.md` tái sinh, reset |
| 002 | `plan/research/data-model/contracts/quickstart` | Viết lại; `contracts/` → REST API + WS events; quickstart giữ 23 kịch bản nghiệp vụ, đổi Setup (docker compose + go run/air + npm run dev) |
| 002 | `spec.md` | Chỉ sửa chỗ "đối chiếu qua email" (Key Entities, Assumptions) thành trung lập; History v4 |
| UC | `uc-trk-01` | Generalize nếu có mô tả email-mapping |
| Entity | `entity-model.md` | Kiểm tra & sửa nếu dính cơ chế cũ (users độc lập vẫn đúng nguyên ý nghĩa) |
| Meta | `.specify/feature.json` | Trỏ theo feature đang re-plan (001 trước, rồi 002) |
| Cuối | Mọi file sửa | Cập nhật khối References/Truy vết + History; chạy `/speckit-analyze` |

## §5. Trình tự thực hiện

1. Design doc này (đã xong) → 2. Cập nhật `CLAUDE.md` + `BR-001` → 3. Re-plan **001** (plan/research/data-model/contracts/quickstart + tasks mới) → 4. Re-plan **002** (như trên) → 5. `/speckit-analyze` cả hai → 6. Implement 001 rồi 002 (`/speckit-implement`).

## §6. Testing

- **Go**: unit test tầng `biz` (validation FR-001…007, quyền theo hộ — theo mẫu `biz/*_test.go` của learn_go); integration test storage + optimistic lock với Postgres (docker); `httptest` cho transport.
- **Vue**: Vitest (unit/component) cho form validation, store.
- **E2E**: theo kịch bản quickstart của từng feature (Chrome), giữ nguyên nội dung nghiệp vụ (001) và 23 kịch bản (002).

## Rủi ro & lưu ý

- Tự viết auth: phạm vi nhỏ (login email+mật khẩu) nhưng phải đúng chuẩn — bcrypt cost mặc định, JWT secret qua env, không log mật khẩu.
- GORM + view: model `AccountBalance` chỉ đọc (map view), không migrate.
- Repo mẫu dùng md5 (`component/hasher`) — **không** bê nguyên cho password; chỉ mượn layout.
- Docs 001 đang ghi "✅ implemented & verified 53/53" — phải sửa để không đánh lừa người đọc rằng còn code chạy được.
