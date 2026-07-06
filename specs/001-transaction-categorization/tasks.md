---
description: "Task list — Phân loại giao dịch (sổ chung hộ gia đình)"
---

# Tasks: Phân Loại Giao Dịch (Transaction Categorization)

**Input**: Design documents from `/specs/001-transaction-categorization/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Stack**: Flutter (Dart 3.x) · Riverpod · Supabase (PostgreSQL + Auth + RLS). Mô hình **sổ chung hộ gia đình**, nhiều thành viên, quyền ngang nhau.

**Tests**: Không có yêu cầu TDD trong spec → **không sinh phase test riêng**. Kiểm chứng end-to-end theo các kịch bản trong [quickstart.md](./quickstart.md) (tham chiếu ở "Independent Test" mỗi phase).

**Path conventions** (mobile, theo plan.md): mã nguồn trong `src/lib/features/categorization/{domain,data,presentation}`, hạ tầng dùng chung trong `src/lib/core/`, migration trong `src/supabase/migrations/`.

> **Lưu ý theo yêu cầu**: Migration tạo **hộ gia đình mặc định + thêm 2 user vào hộ** (T012) nằm ở Phase 2 Foundational và **phải hoàn tất trước mọi chức năng danh mục** (Phase 3+).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Khởi tạo dự án Flutter và cấu trúc thư mục.

- [X] T001 Tạo cấu trúc thư mục theo plan.md: `src/lib/core/{supabase,household,router,error}/`, `src/lib/features/categorization/{domain/{entities,repositories,usecases},data/{models,datasources,repositories},presentation/{screens,widgets,controllers}}/`, `src/supabase/migrations/`, `src/test/{unit,widget}/`, `src/integration_test/`
- [X] T002 Khởi tạo Flutter project và khai báo dependencies trong `pubspec.yaml` (`flutter_riverpod`, `supabase_flutter`, `go_router`, `freezed`, `json_serializable`, `build_runner`, `sqflite`, `mocktail`)
- [X] T003 [P] Cấu hình lint/format trong `analysis_options.yaml`
- [X] T004 [P] Cấu hình kết nối Supabase qua `--dart-define` (SUPABASE_URL/SUPABASE_ANON_KEY) trong `src/lib/core/supabase/env.dart`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Lược đồ CSDL, RLS theo hộ, hạ tầng domain/data dùng chung — và **seed hộ gia đình mặc định + 2 thành viên** (yêu cầu của bạn).

**⚠️ CRITICAL**: Không bắt đầu bất kỳ chức năng danh mục (Phase 3+) nào trước khi phase này xong — đặc biệt T012 (seed hộ + 2 user) phải có trước.

- [X] T005 Migration: bảng `households` + `household_members` trong `src/supabase/migrations/0001_households.sql` (data-model §HOUSEHOLD/HOUSEHOLD_MEMBER)
- [X] T006 Migration: hàm `is_member(uuid)` + bật RLS các bảng trong `src/supabase/migrations/0002_rls_helpers.sql` (research R12)
- [X] T007 Migration: bảng `categories` + trigger `enforce_one_level`/`inherit_type`/`type_immutable` trong `src/supabase/migrations/0003_categories.sql` (FR-005, FR-010, FR-011)
- [X] T008 Migration: bảng `transactions` + trigger `enforce_txn_rules` (cùng loại & cùng hộ) trong `src/supabase/migrations/0004_transactions.sql` (FR-013, FR-014)
- [X] T009 [P] Migration: bảng `categorization_rules` trong `src/supabase/migrations/0005_categorization_rules.sql` (FR-015)
- [X] T010 Migration: RPC `delete_category(p_category, p_action, p_target)` (gán lại/xóa, kiểm `is_member`) trong `src/supabase/migrations/0006_delete_category.sql` (FR-008, FR-009, FR-012)
- [X] T011 Migration: RLS policies `member_*` cho tất cả bảng trong `src/supabase/migrations/0007_rls_policies.sql` (FR-018 — cô lập giữa các hộ)
- [X] T012 **Migration seed: tạo HỘ GIA ĐÌNH MẶC ĐỊNH + thêm 2 user làm thành viên** trong `src/supabase/migrations/0008_seed_default_household.sql` *(tiền đề bắt buộc — mọi chức năng danh mục phụ thuộc)*
- [X] T013 Migration seed: bộ danh mục mặc định (FR-001) cho hộ mặc định trong `src/supabase/migrations/0009_seed_default_categories.sql` (phụ thuộc T012; danh sách theo research R9)
- [X] T014 [P] Khởi tạo Supabase client trong `src/lib/core/supabase/supabase_client.dart`
- [X] T015 [P] Provider "hộ hiện tại" (resolve `household_id` của thành viên đăng nhập) trong `src/lib/core/household/current_household.dart`
- [X] T016 [P] Cấu hình router (go_router) trong `src/lib/core/router/app_router.dart`
- [X] T017 [P] Định nghĩa kiểu lỗi/Failure + Result trong `src/lib/core/error/failures.dart`
- [X] T018 [P] Tạo domain entities `Category`, `CategoryType`, `CategorizationRule` trong `src/lib/features/categorization/domain/entities/`
- [X] T019 Định nghĩa interface `CategoryRepository` (theo `contracts/category-repository.md`) trong `src/lib/features/categorization/domain/repositories/category_repository.dart`
- [X] T020 [P] DTO models + mappers trong `src/lib/features/categorization/data/models/category_model.dart`
- [X] T021 Supabase data source (CRUD + gọi RPC, gắn `household_id` từ hộ hiện tại) trong `src/lib/features/categorization/data/datasources/category_remote_datasource.dart` (phụ thuộc T014, T015)
- [X] T022 `CategoryRepositoryImpl` ghép data source + hộ hiện tại trong `src/lib/features/categorization/data/repositories/category_repository_impl.dart` (phụ thuộc T019, T021)

**Checkpoint**: Hộ mặc định tồn tại với 2 thành viên; lược đồ + RLS sẵn sàng → bắt đầu các user story.

---

## Phase 3: User Story 1 - Bắt buộc phân loại đúng loại Thu/Chi (Priority: P1) 🎯 MVP

**Goal**: Mọi giao dịch được gán đúng một danh mục cùng loại; chỉ hiển thị danh mục cùng loại; chặn lưu nếu chưa chọn. (UC-CAT-01, UC-CAT-07)

**Independent Test**: quickstart kịch bản #1, #5, #6 (lọc theo loại; chặn lưu khi thiếu danh mục; danh mục mặc định dùng được).

- [X] T023 [P] [US1] Use case `ListCategories` (nhóm/lọc theo loại, loại trừ ẩn) trong `src/lib/features/categorization/domain/usecases/list_categories.dart`
- [X] T024 [P] [US1] Use case `AssignCategoryToTransaction` (bắt buộc chọn, trùng loại, cùng hộ, gán `created_by`) trong `src/lib/features/categorization/domain/usecases/assign_category_to_transaction.dart`
- [X] T025 [US1] Bổ sung repo methods `listCategories` + `assignCategoryToTransaction` trong `src/lib/features/categorization/data/repositories/category_repository_impl.dart` (phụ thuộc T022)
- [X] T026 [P] [US1] Controller (Riverpod) danh sách/chọn danh mục trong `src/lib/features/categorization/presentation/controllers/category_list_controller.dart`
- [X] T027 [US1] Màn hình chọn danh mục nhóm & lọc theo loại trong `src/lib/features/categorization/presentation/screens/category_picker_screen.dart`
- [X] T028 [US1] Widget chọn danh mục khi nhập giao dịch + chặn lưu nếu rỗng trong `src/lib/features/categorization/presentation/widgets/category_select_field.dart`
- [X] T029 [US1] Ghép vào luồng nhập giao dịch: làm tươi danh sách khi đổi loại Thu/Chi (BR-002) trong `src/lib/features/categorization/presentation/widgets/category_select_field.dart`

**Checkpoint**: US1 hoạt động độc lập — MVP có thể demo.

---

## Phase 4: User Story 2 - Tự tạo & quản lý danh mục (Priority: P2)

**Goal**: Tạo/sửa/xóa danh mục và ẩn/bỏ ẩn; xóa an toàn (gán lại cùng loại hoặc xóa giao dịch). (UC-CAT-02, UC-CAT-04, UC-CAT-05, UC-CAT-06)

**Independent Test**: quickstart kịch bản #2, #3, #8, #9, #10, #11 (loại bắt buộc; cảnh báo trùng tên; xóa-gán lại; chặn gán lại sai loại; ẩn/bỏ ẩn; đổi tên phản ánh mọi nơi).

- [X] T030 [P] [US2] Use case `CreateCategory` (loại bắt buộc; cảnh báo trùng tên; tạo nhanh kế thừa loại giao dịch) trong `src/lib/features/categorization/domain/usecases/create_category.dart`
- [X] T031 [P] [US2] Use case `RenameCategory` (tên/biểu tượng; chặn đổi loại) trong `src/lib/features/categorization/domain/usecases/rename_category.dart`
- [X] T032 [P] [US2] Use case `DeleteCategory` (gán lại cùng loại / xóa qua RPC) trong `src/lib/features/categorization/domain/usecases/delete_category.dart`
- [X] T033 [P] [US2] Use case `SetHidden` (ẩn/bỏ ẩn) trong `src/lib/features/categorization/domain/usecases/set_hidden.dart`
- [X] T034 [US2] Bổ sung repo methods create/rename/delete/setHidden trong `src/lib/features/categorization/data/repositories/category_repository_impl.dart` (phụ thuộc T022)
- [X] T035 [US2] Màn hình form danh mục (tạo/sửa: loại, tên, biểu tượng) trong `src/lib/features/categorization/presentation/screens/category_form_screen.dart`
- [X] T036 [US2] Màn hình xóa-gán lại (chọn gán lại cùng loại hoặc xóa giao dịch) trong `src/lib/features/categorization/presentation/screens/delete_reassign_screen.dart`
- [X] T037 [US2] Toggle ẩn/bỏ ẩn + nhãn "đã ẩn" trong màn hình quản lý trong `src/lib/features/categorization/presentation/widgets/hide_toggle.dart`
- [X] T038 [US2] Controller quản lý danh mục trong `src/lib/features/categorization/presentation/controllers/category_manage_controller.dart`

**Checkpoint**: US1 + US2 hoạt động độc lập.

---

## Phase 5: User Story 3 - Tổ chức bằng danh mục con (Priority: P3)

**Goal**: Tạo danh mục con một cấp, kế thừa loại của cha; gán giao dịch vào cha hoặc con; báo cáo roll-up. (UC-CAT-03, mở rộng UC-CAT-05/07)

**Independent Test**: quickstart kịch bản #7 (con kế thừa loại; chặn cấp con thứ hai) + gán cha/con.

- [X] T039 [P] [US3] Use case `CreateSubcategory` (kế thừa loại cha; chặn quá một cấp) trong `src/lib/features/categorization/domain/usecases/create_subcategory.dart`
- [X] T040 [US3] Repo method `createSubcategory` + truy vấn lồng cha→con trong `src/lib/features/categorization/data/repositories/category_repository_impl.dart` (phụ thuộc T034)
- [X] T041 [US3] UI "Thêm danh mục con" dưới danh mục cha + hiển thị lồng một cấp trong `src/lib/features/categorization/presentation/screens/category_form_screen.dart`
- [X] T042 [US3] Cho phép gán giao dịch vào cha HOẶC con trong picker + ghi chú roll-up trong `src/lib/features/categorization/presentation/screens/category_picker_screen.dart`
- [X] T043 [US3] Đảm bảo xóa danh mục cha xử lý cả con (kiểm đường RPC `delete_category`) trong `src/lib/features/categorization/presentation/screens/delete_reassign_screen.dart`

**Checkpoint**: US1 + US2 + US3 hoạt động độc lập.

---

## Phase 6: User Story 4 - Gợi ý danh mục (Priority: P4)

**Goal**: Gợi ý danh mục cùng loại theo từ khóa + lịch sử chung của hộ; tham khảo, ghi đè được. (UC-CAT-08)

**Independent Test**: quickstart kịch bản #12 (mô tả "Grab" → gợi ý "Di chuyển"; chọn khác → lựa chọn người dùng được lưu).

- [X] T044 [P] [US4] Use case `SuggestCategory` (khớp từ khóa + lịch sử hộ, cùng loại) trong `src/lib/features/categorization/domain/usecases/suggest_category.dart`
- [X] T045 [US4] Data source/repo cho `categorization_rules` (đọc quy tắc, cập nhật `match_count`) trong `src/lib/features/categorization/data/datasources/rule_remote_datasource.dart`
- [X] T046 [US4] Hiển thị gợi ý trong luồng nhập giao dịch (chấp nhận/ghi đè) trong `src/lib/features/categorization/presentation/widgets/suggestion_chip.dart`
- [X] T047 [US4] Học từ lịch sử chung hộ (ghi nhận/tăng trọng số khi xác nhận) trong `src/lib/features/categorization/data/datasources/rule_remote_datasource.dart`

**Checkpoint**: Cả 4 user story hoạt động độc lập.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Hoàn thiện xuyên suốt các story (đa thành viên, offline, hiệu năng).

- [X] T048 [P] Cache đọc danh mục offline (sqflite) trong `src/lib/features/categorization/data/datasources/category_local_cache.dart`
- [X] T049 [P] Đồng bộ realtime danh mục/giao dịch giữa thành viên (Supabase Realtime) trong `src/lib/core/supabase/realtime.dart`
- [X] T050 Xử lý đồng thời nhiều thành viên (cập nhật lạc quan + thông báo xung đột) per research R13 trong `src/lib/features/categorization/data/repositories/category_repository_impl.dart` (migration `0010_updated_at.sql` + `expectedUpdatedAt` xuyên suốt domain→UI)
- [X] T051 [P] Tối ưu hiệu năng đạt SC-005 (≤3 bước/<30s) và SC-006 (<10s) — rà soát luồng tạo & chọn danh mục *(đã rà soát: tạo = FAB → loại+tên → Lưu = 3 bước; chọn = 1 chạm mở picker đã lọc + 1 chạm chọn; chip gợi ý = 1 chạm; provider cache + realtime invalidation, không N+1)*
- [X] T052 Chạy kiểm chứng `quickstart.md` (17 kịch bản, gồm đa thành viên #13–#17) — *(2026-07-06)* unit test (#2, #3, #5, #7, #10, #12) + widget test (#6) pass 10/10; **e2e trên Supabase thật (project boufzdnh…, schema + seed qua `src/supabase/setup_dev.sql`) pass toàn bộ qua API**: #13 Bob thấy danh mục Alice tạo; #14 sổ chung + `created_by`; #15 Carol (hộ B) cô lập; #16 ghi mốc `updated_at` cũ → 0 hàng (không ghi đè thầm lặng); #17 Bob sửa/xóa danh mục Alice tạo (ngang quyền); #7 con kế thừa loại + chặn cấp 2; #8 xóa-gán lại kèm con, không mồ côi (SC-007); #9 gán lại khác loại bị chặn; FR-014 trigger chặn giao dịch sai loại. Kiểm UI thủ công: app chạy Chrome với login dev alice/bob/carol
- [X] T053 [P] Cập nhật tài liệu & khối References của các artifact (theo quy tắc lan truyền trong `CLAUDE.md`) *(đã lan truyền `updated_at`/R13: data-model.md, contracts/db-schema.sql, contracts/category-repository.md (+ `createSubcategory`, `ConcurrencyConflict`), specs/entities/entity-model.md + History)*

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: không phụ thuộc — bắt đầu ngay.
- **Foundational (Phase 2)**: phụ thuộc Setup. **BLOCKS mọi user story.** Trong phase: T005 → T006 → (T007, T008, T009) → T010 → T011 → **T012 (seed hộ + 2 user)** → T013. Hạ tầng Dart T014–T022 chạy song song nhánh migration nhưng T021/T022 phụ thuộc T014/T015/T019.
- **User Stories (Phase 3–6)**: đều phụ thuộc Phase 2 hoàn tất. Sau đó có thể song song hoặc theo thứ tự P1→P2→P3→P4.
- **Polish (Phase 7)**: sau khi các story mong muốn đã xong.

### User Story Dependencies

- **US1 (P1)**: chỉ phụ thuộc Foundational — độc lập, là MVP.
- **US2 (P2)**: phụ thuộc Foundational; độc lập với US1 (dùng chung repo nhưng method khác).
- **US3 (P3)**: phụ thuộc Foundational; tái dùng create/delete của US2 (T040 phụ thuộc T034) — nên làm sau US2.
- **US4 (P4)**: phụ thuộc Foundational; độc lập, gắn vào luồng nhập của US1.

### Within Each User Story

- Use case (domain) → repo method (data) → controller → screen/widget.
- Không có phụ thuộc chéo story phá vỡ tính độc lập (trừ US3 tái dùng US2 ở mức repo).

### Parallel Opportunities

- Setup: T003, T004 song song.
- Foundational: T009, T014–T018, T020 (đánh [P]) song song; nhánh migration tuần tự tới T013.
- Mỗi story: các use case [P] (vd T023/T024; T030–T033; ) song song trước khi ghép repo.
- Sau Foundational, US1/US2/US4 có thể giao cho nhiều người làm song song.

---

## Parallel Example: User Story 2

```bash
# Các use case độc lập của US2 (khác file) chạy song song:
Task: "CreateCategory use case in src/lib/features/categorization/domain/usecases/create_category.dart"
Task: "RenameCategory use case in src/lib/features/categorization/domain/usecases/rename_category.dart"
Task: "DeleteCategory use case in src/lib/features/categorization/domain/usecases/delete_category.dart"
Task: "SetHidden use case in src/lib/features/categorization/domain/usecases/set_hidden.dart"
# Sau đó T034 (repo) ghép chung — KHÔNG song song với nhau (cùng file).
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Phase 1 Setup → 2. Phase 2 Foundational (gồm **T012 seed hộ + 2 user**, T013 seed danh mục) → 3. Phase 3 US1 → 4. **DỪNG & kiểm chứng** quickstart #1/#5/#6 → demo MVP.

### Incremental Delivery

Foundational → US1 (MVP, demo) → US2 (quản lý danh mục, gồm UC-CAT-02 trọng tâm) → US3 (danh mục con) → US4 (gợi ý) → Polish (đa thành viên/offline/hiệu năng). Mỗi story test độc lập trước khi sang story sau.

### Parallel Team Strategy

Sau Foundational: Dev A → US1; Dev B → US2; Dev C → US4 (US3 sau khi US2 xong vì tái dùng repo create/delete).

---

## Notes

- [P] = khác file, không phụ thuộc nhau.
- Nhãn [US#] gắn task với user story để truy vết.
- **T012 là tiền đề bắt buộc** (hộ mặc định + 2 user) — không chạy chức năng danh mục trước nó.
- Quản lý hộ đầy đủ (tạo/mời/tham gia — UC-HH-*) thuộc **feature riêng**; ở đây chỉ seed sẵn để chạy được phần phân loại.
- Commit sau mỗi task hoặc nhóm logic; dừng ở mỗi Checkpoint để kiểm chứng story độc lập.
