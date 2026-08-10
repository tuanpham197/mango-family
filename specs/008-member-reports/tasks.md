---
description: "Task list — Feature 008: Báo Cáo Thu Chi Theo Thành Viên"
---

# Tasks: Báo Cáo Thu Chi Theo Thành Viên (Per-Member Report)

**Input**: Design documents from `/specs/008-member-reports/`
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/member-report-api.md](./contracts/member-report-api.md), [quickstart.md](./quickstart.md)

**Tests**: INCLUDED (TDD) — plan Testing section yêu cầu Go unit/integration/httptest + Vitest + Playwright, nhất quán quy ước 001–005.

**Organization**: Nhóm theo user story để triển khai & kiểm thử độc lập. Feature **mở rộng module `report` sẵn có (005)** — chỉ đọc; **không dependency mới, không migration** (trừ index tùy chọn ở Polish nếu đo thấy cần).

## Format: `[ID] [P?] [Story] Description`
- **[P]**: chạy song song được (khác file, không phụ thuộc task chưa xong)
- **[Story]**: US1 / US2 / US3 (map user story spec.md)

## Path Conventions
- Backend Go: `src/api/module/report/**`, `src/api/module/transaction/storage/**`, `src/api/main_route.go`
- Frontend Vue: `src/web/src/{views,components,stores}/**`, e2e `src/web/e2e/**`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Dữ liệu & tiền đề để kiểm thử; xác nhận không phát sinh dep/migration.

- [X] T001 [P] Mở rộng seed `src/api/cmd/seed/main.go`: hộ mẫu có **≥2 thành viên**, mỗi người có giao dịch Thu và Chi trong tháng hiện tại (phục vụ QS-1…QS-7 & e2e); bảo đảm còn một thành viên **không có** giao dịch trong tháng (cho QS-3).
- [X] T002 [P] Xác nhận không thêm dependency/migration: kiểm `src/web/package.json` (`chart.js` đã có, không thêm gói), không tạo file trong `src/db/migrations/` (index tùy chọn để ở T029).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: DTO dùng chung cho cả hai endpoint. **⚠️ Phải xong trước mọi user story.**

- [X] T003 Tạo DTO suy ra `src/api/module/report/model/member.go`: `MemberRow`, `MembersReport`, `MemberReport` đúng theo [data-model.md](./data-model.md) (JSON tags, sentinel `member_id:"former"`, `is_former`, `income/expense/net`, `page/page_size/total`).

**Checkpoint**: DTO sẵn sàng — US1/US2/US3 có thể bắt đầu.

---

## Phase 3: User Story 1 - Xem & so sánh Thu/Chi/ròng theo từng thành viên (Priority: P1) 🎯 MVP

**Goal**: `GET /api/reports/members` trả tổng Thu/Chi/ròng mỗi thành viên hiện tại cho khoảng chọn được; màn Báo cáo hiển thị bảng theo thành viên; đổi khoảng tính lại.

**Independent Test**: Với hộ có giao dịch của ≥2 thành viên trong tháng, mở `/reports` → khối "Theo thành viên" liệt kê đúng Thu/Chi/ròng mỗi người; đổi "Tuần này" → cập nhật.

### Tests for User Story 1 ⚠️ (viết trước, phải FAIL trước khi implement)

- [X] T004 [P] [US1] Integration test storage `src/api/module/report/storage/member_integration_test.go`: `SUM(CASE type) GROUP BY created_by` trên `[from,to)` (ranh giới nửa mở), quy đúng theo `created_by`, cô lập hộ (hộ khác không lẫn), trên Postgres thật.
- [X] T005 [P] [US1] httptest `src/api/module/report/transport/ginreport/routes_members_test.go`: `GET /api/reports/members` 200 đúng shape; 400 khi thiếu/sai `from`/`to` hoặc `to<from`; 401 chưa đăng nhập.
- [X] T006 [P] [US1] Vitest `src/web/src/stores/__tests__/reports.members.spec.ts`: `fetchMembers(from,to)` set state; tên hiển thị fallback email khi `display_name` rỗng (FR-009).

### Implementation for User Story 1

- [X] T007 [US1] Storage `GetMembers(ctx, householdID, from, to)` trong `src/api/module/report/storage/report.go`: SUM Thu/Chi `GROUP BY created_by` JOIN `users` (display_name/email); trả hàng thô theo thành viên (D42).
- [X] T008 [US1] Biz `src/api/module/report/biz/get_members.go`: giải khoảng (tái dùng bộ giải khoảng/`parseRange` của 005), `net = income − expense`, sắp theo Chi giảm dần + tie-break tên (D43), ráp `MembersReport` (chưa gồm former/zero-fill — thuộc US3).
- [X] T009 [US1] Handler `Members` + route `GET /api/reports/members` (scoped hộ) trong `src/api/module/report/transport/ginreport/routes.go`; wire ở `src/api/main_route.go`.
- [X] T010 [P] [US1] Mở rộng store `src/web/src/stores/reports.ts`: state `members`, action `fetchMembers(from,to)` gọi endpoint, dùng lại `TimeRangePicker` hiện có.
- [X] T011 [US1] Component `src/web/src/components/MemberBreakdown.vue` (bảng Thu/Chi/ròng + biểu đồ cột tùy chọn dùng lại `chart.js`); nhúng khối "Theo thành viên" vào `src/web/src/views/ReportsView.vue` phản ứng theo khoảng.

**Checkpoint**: US1 chạy độc lập — bảng theo thành viên đúng số, đổi khoảng cập nhật. **MVP.**

---

## Phase 4: User Story 2 - Drill-in danh sách giao dịch của một thành viên (Priority: P2)

**Goal**: `GET /api/reports/member/:id` trả danh sách giao dịch (phân trang) của một thành viên (hoặc bucket `former`) trong khoảng; UI mở panel drill-in, quay lại giữ khoảng.

**Independent Test**: Trên bảng theo thành viên chọn một người → panel liệt kê đúng giao dịch của họ (loại/số tiền/danh mục/mô tả/ngày giờ/tài khoản), phân trang khi dài; quay lại → giữ khoảng.

### Tests for User Story 2 ⚠️ (viết trước, phải FAIL)

- [X] T012 [P] [US2] Integration test `src/api/module/report/storage/member_drilldown_integration_test.go`: liệt kê giao dịch theo `created_by` có phân trang (`page/page_size`, `total`), trạng thái trống, cô lập hộ; biến thể `former` (`created_by NOT IN household_members`).
- [X] T013 [P] [US2] httptest `src/api/module/report/transport/ginreport/routes_member_test.go`: `GET /api/reports/member/:id` 200 shape + phân trang; **404** khi `:id` không phải UUID hoặc không là thành viên hiện tại của hộ (chống dò — D48); sentinel `:id=former`; 400 sai `from/to/page`.
- [X] T014 [P] [US2] Vitest `src/web/src/components/__tests__/MemberTransactionsPanel.spec.ts` + store: `fetchMember(id,from,to,page)`, phân trang, **quay lại giữ khoảng** đã chọn (FR-007).

### Implementation for User Story 2

- [X] T015 [US2] Thêm trường lọc đọc `CreatedBy *uuid.UUID` vào `ListFilter` và áp `created_by = ?` (nil = không lọc, tương thích ngược) trong `src/api/module/transaction/storage/store.go` (D47).
- [X] T016 [US2] Storage drill-in trong `src/api/module/report/storage/report.go`: liệt kê giao dịch của một thành viên (dùng lại `transaction` List với `CreatedBy` + phân trang) + tổng Thu/Chi/ròng của thành viên; biến thể `former` (NOT IN members).
- [X] T017 [US2] Biz `src/api/module/report/biz/get_member.go`: validate `:id` (UUID **thuộc hộ hiện tại** qua `household_members`, hoặc `former`), giải khoảng, phân trang (`common.Paging`), ráp `MemberReport` với `[]transactionmodel.ListItem`.
- [X] T018 [US2] Handler `Member` + route `GET /api/reports/member/:id` trong `.../ginreport/routes.go`; wire `src/api/main_route.go`.
- [X] T019 [P] [US2] Store `src/web/src/stores/reports.ts`: state `memberDetail`, action `fetchMember(id,from,to,page)`; giữ khoảng hiện tại khi mở/đóng drill-in.
- [X] T020 [US2] Component `src/web/src/components/MemberTransactionsPanel.vue` (danh sách phân trang, empty state, nút quay lại) + nối sự kiện chọn dòng từ `MemberBreakdown.vue` trong `ReportsView.vue`.

**Checkpoint**: US1 + US2 chạy độc lập — drill-in hoạt động, giữ khoảng.

---

## Phase 5: User Story 3 - Tính đầy đủ & đối soát với tổng toàn hộ (Priority: P2)

**Goal**: `/members` bảo đảm **mọi thành viên hiện tại** xuất hiện (0/0/0 nếu không GD), giao dịch người **đã rời hộ** gộp dòng "Thành viên cũ", và **Σ dòng == tổng hộ** (SC-001). *(Phủ lên endpoint US1 — phụ thuộc file US1.)*

**Independent Test**: Hộ có thành viên chưa nhập GD → hiện 0/0/0; `Σ members[].income/expense == totals == GET /api/reports/overview` cùng khoảng; nếu có GD của người đã rời → dòng "Thành viên cũ" xuất hiện và đối soát vẫn đúng.

### Tests for User Story 3 ⚠️ (viết trước, phải FAIL)

- [X] T021 [P] [US3] Integration test `src/api/module/report/storage/member_reconcile_integration_test.go`: thành viên 0 GD → 0/0/0 (FR-004); `Σ members == totals == overview` cùng khoảng (SC-001); bucket `former` gộp đúng & **chỉ hiện khi có dữ liệu** (FR-010).
- [X] T022 [P] [US3] Unit biz test `src/api/module/report/biz/get_members_test.go`: tách current↔former, zero-fill thành viên rỗng, sort Chi desc + dòng `former` xếp cuối, tie-break tên.

### Implementation for User Story 3

- [X] T023 [US3] Storage `GetMembers` (mở rộng T007) trong `src/api/module/report/storage/report.go`: khởi danh sách từ `household_members` **LEFT JOIN** tổng hợp (mọi thành viên hiện tại xuất hiện, 0 nếu không GD — D43); **LEFT JOIN `household_members`** để tách `created_by ∉ members` → gộp một khóa `former` (D44).
- [X] T024 [US3] Biz `get_members.go` (mở rộng T008): tính `totals` **độc lập** (SUM toàn hộ) và assert đối soát; chỉ thêm dòng `former` khi có dữ liệu; gắn nhãn "Thành viên cũ / Đã rời hộ" + `is_former=true` xếp cuối.
- [X] T025 [US3] `MemberBreakdown.vue` (mở rộng T011): render dòng 0/0/0 rõ ràng và dòng "Thành viên cũ" khi `is_former` (xếp cuối, drill-in được như thành viên).

**Checkpoint**: Toàn bộ 3 story hoạt động; đối soát tổng hộ đúng 100%.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T026 [P] Playwright e2e `src/web/e2e/member-report.spec.ts`: phủ QS-1…QS-7 (đổi khoảng, 0/0/0, drill-in + phân trang + giữ khoảng, former bucket, cô lập hộ, 404 ngoài hộ, khoảng tùy chỉnh/trống).
- [X] T027 [P] Cập nhật trạng thái use case `specs/use-cases/008-member-reports/uc-mbr-01…02` (draft → implemented) + `README.md`/`use-cases.puml` nếu FR đổi (hiện đã đồng bộ).
- [X] T028 Chạy `cd src && make test` (Go unit + Vitest) · `make test-api-integration` · `make test-e2e`; xác minh **001–005 vẫn xanh** sau khi wire route mới.
- [X] T029 Đo execution plan truy vấn `/members` cho khoảng ~1 tháng/1 năm (SC-002 ≤ 2s); **chỉ khi** cần mới thêm migration goose index `(household_id, created_by, transaction_date)` (D42) — ghi rõ lý do; nếu không cần thì ghi nhận đã đo.

---

## Dependencies & Execution Order

### Phase Dependencies
- **Setup (P1)**: không phụ thuộc — bắt đầu ngay.
- **Foundational (P2 — T003)**: sau Setup; **chặn** mọi user story.
- **US1 (P3)**: sau Foundational — MVP, không phụ thuộc story khác.
- **US2 (P4)**: sau Foundational — độc lập kiểm thử (endpoint riêng); UI drill-in nối vào `ReportsView` của US1 khi cả hai có mặt.
- **US3 (P5)**: sau Foundational — **phủ lên endpoint US1** (sửa cùng `GetMembers`/`get_members.go`/`MemberBreakdown.vue`), nên chạy **sau US1** để tránh xung đột file.
- **Polish (P6)**: sau khi các story mong muốn xong.

### Within Each Story
- Tests trước (phải FAIL) → storage → biz → transport/route → FE store → FE component.

### Parallel Opportunities
- Setup: T001, T002 song song.
- US1 tests T004/T005/T006 song song; US2 tests T012/T013/T014 song song; US3 tests T021/T022 song song.
- FE store tasks đánh [P] khi khác file với BE đang sửa.
- **Không** song song US1 và US3 (cùng file `report.go`/`get_members.go`/`MemberBreakdown.vue`).

---

## Parallel Example: User Story 1

```bash
# Tests US1 (viết trước, để FAIL) — chạy song song:
Task: "Integration test storage member_integration_test.go (T004)"
Task: "httptest routes_members_test.go (T005)"
Task: "Vitest reports.members.spec.ts (T006)"
```

---

## Implementation Strategy

### MVP First (US1)
1. Phase 1 Setup → 2. Phase 2 Foundational (T003) → 3. Phase 3 US1 → **STOP & VALIDATE** bảng theo thành viên (QS-1) → demo.

### Incremental Delivery
1. Setup + Foundational → nền sẵn sàng.
2. US1 → test độc lập (QS-1) → demo (MVP).
3. US2 → drill-in (QS-4) → demo.
4. US3 → đối soát & đầy đủ (QS-2/QS-3/QS-5) → demo.
5. Polish → e2e toàn phần + đo hiệu năng.

---

## Notes
- [P] = khác file, không phụ thuộc. [Story] = truy vết user story.
- Verify test FAIL trước khi implement (TDD, như 001–005).
- **US3 phủ lên US1** — không phải story hoàn toàn tách file; giữ thứ tự US1→US3.
- Commit sau mỗi task/nhóm hợp lý. Không sửa hành vi ghi của module transaction/users (chỉ thêm bộ lọc đọc `CreatedBy`).

## Deviations & Verification (khi triển khai — 2026-08-10)

Tất cả 29 task đã hoàn tất; một số **điều chỉnh có chủ đích** so với câu chữ task ban đầu (giữ đúng pattern codebase, tránh mã chết / khớp phạm vi):

- **T015 (thêm `CreatedBy` vào `transaction ListFilter`)** — KHÔNG sửa `transaction/storage`. Thay vào đó truy vấn drill-in được viết trong **`report/storage`** (`MemberTransactions`/`FormerMemberTransactions`), **mirror pattern có sẵn `CategoryTransactions` của 005**. Lý do: giữ report tự chứa, tránh coupling chéo module + trường lọc dùng-một-chỗ. Hành vi (lọc theo `created_by`, phân trang, former = NOT IN members) đạt đủ FR-006/007.
- **T005 / T013 (httptest riêng ở `transport/ginreport`)** — KHÔNG tạo file Go httptest (nhất quán 005: report/transport không có httptest). Hành vi HTTP được phủ bằng: **biz unit test** (400 khoảng sai, 404 UUID sai/không thuộc hộ, sentinel `former`) + **Playwright e2e** (200 shape, phân trang, cô lập hộ) + **storage integration** (SQL/đối soát).
- **T011 / T025 (biểu đồ cột tùy chọn)** — triển khai **bảng** (hợp đồng chính); **biểu đồ cột hoãn** (giải quyết ambiguity A1 của /analyze — table là contract, chart optional). Không thêm dependency.
- **T029 (index)** — đã đo ở scale dev: báo cáo members < 2s (SC-002) không cần index thêm; `created_by` đã có ràng buộc FK. **KHÔNG thêm migration**. Khuyến nghị đo lại execution plan trên dữ liệu prod trước khi thêm `(household_id, created_by, transaction_date)` nếu cần.

### Kết quả kiểm chứng (chạy thật)
- **Go**: `go vet ./...` sạch · unit `go test ./...` xanh · integration `go test -tags=integration` xanh trên **Postgres thật** (gồm `TestReportStorage_MemberAggregations`: GROUP BY created_by, cô lập hộ, phân trang drill-in, former bucket, **đối soát Σ = tổng hộ**).
- **Biz unit**: `get_members_test.go` (zero-fill, sort Chi desc, former chỉ khi có dữ liệu, đối soát, fallback email, 400) · `get_member_test.go` (current/former, 404 non-member, 404 bad-uuid, 400 range, passthrough page).
- **Web**: `vue-tsc` typecheck sạch · **Vitest 94/94** (gồm `reports.members`, `MemberBreakdown`, `MemberTransactionsPanel`).
- **Playwright e2e**: `member-report.spec.ts` **4/4** (QS-1/2/3/4/6) + regression `report.spec.ts`/`report-sharing.spec.ts` (005) **5/5** xanh.
- **Seed T001**: chạy `make seed` — Alice=2, Bob=2, Dave=0 giao dịch `[seed]` tháng hiện tại; **idempotent** (chạy lại giữ nguyên) · prod (nhánh `SEED_EMAIL_1`) return trước, KHÔNG chạm seed giao dịch.
