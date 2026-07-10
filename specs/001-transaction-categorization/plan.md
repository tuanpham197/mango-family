# Implementation Plan: Phân Loại Giao Dịch (Transaction Categorization)

**Branch**: `001-transaction-categorization` | **Date**: 2026-06-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-transaction-categorization/spec.md`

> ⚠️ **LEGACY STACK (2026-07-10)**: Plan này mô tả kiến trúc **Flutter + Supabase đã gỡ bỏ** (re-platform Go + Vue — xem [design](../../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md)). Nghiệp vụ & bất biến vẫn đúng; phần kỹ thuật chờ viết lại bằng `/speckit-plan`.

**User focus**: `@specs/use-cases/001-transaction-categorization/uc-cat-02-tao-danh-muc.md` (Tạo danh mục mới). Kế hoạch bao phủ toàn bộ tính năng (UC-CAT-01…08).

> ⚠️ **Thay đổi phạm vi (2026-06-29)**: App là **dùng chung trong gia đình, nhiều thành viên**. Điều này **đảo ngược FR-018** của spec ("danh mục riêng từng người dùng, không chia sẻ") và mục **Out of Scope của BR-001** ("chia sẻ/đồng bộ giữa nhiều người dùng"). Plan dưới đây đã cập nhật theo mô hình **sổ chung hộ gia đình** với **quyền ngang nhau**. **Spec cần được cập nhật tương ứng** (xem mục "Tác động tới spec" cuối file) — nên chạy `/speckit-clarify` hoặc `/speckit-specify`.

## Summary

Cho phép **các thành viên trong một hộ gia đình dùng chung một sổ tài chính**: chia sẻ chung cả **hệ thống danh mục Thu/Chi** lẫn **giao dịch**. Mọi thành viên có **quyền ngang nhau** (ai cũng xem/tạo/sửa/xóa danh mục và nhập giao dịch); mỗi giao dịch ghi rõ **thành viên nào đã nhập**. Vẫn giữ các bất biến cốt lõi: mọi giao dịch được gán đúng một danh mục cùng loại; danh mục con một cấp; loại Thu/Chi cố định; xóa danh mục không để lại giao dịch mồ côi; gợi ý theo quy tắc/lịch sử (không AI/ML) dùng lịch sử **chung của hộ**.

**Technical approach**: Flutter (iOS+Android), Clean Architecture + Riverpod; backend **Supabase (PostgreSQL + Auth + RLS)**. Dữ liệu danh mục/giao dịch/quy tắc gắn với **`household_id`** thay vì người dùng. Cô lập dữ liệu chuyển từ "theo người dùng" sang **"theo thành viên của hộ"**: RLS cho phép truy cập nếu `auth.uid()` là thành viên của hộ sở hữu bản ghi. Việc tạo hộ & mời thành viên là **tiền đề** (quản lý hộ — xem Dependency), tính năng này giả định người dùng đã là thành viên của một hộ.

## Technical Context

**Language/Version**: Dart 3.x trên Flutter 3.x (stable)

**Primary Dependencies**: `flutter`, `supabase_flutter` (Auth + Postgres + Realtime), `flutter_riverpod`, `go_router`, `freezed` + `json_serializable`, `sqflite`/`drift` (cache đọc); test: `flutter_test`, `integration_test`, `mocktail`

**Storage**: Supabase PostgreSQL; dữ liệu gắn `household_id`; cô lập bằng **Row-Level Security theo membership** (thành viên của hộ). Cân nhắc **Supabase Realtime** để đồng bộ thay đổi danh mục/giao dịch giữa các thiết bị thành viên gần thời gian thực

**Testing**: `flutter_test` (unit + widget), `integration_test` (end-to-end gồm kịch bản đa thành viên & cô lập giữa các hộ), `mocktail`

**Target Platform**: iOS 13+ và Android 8+ (Flutter mobile)

**Project Type**: Mobile app (Flutter) + BaaS, **multi-user chia sẻ theo hộ gia đình**

**Performance Goals**: Chọn danh mục < 10s (SC-006); tạo danh mục ≤ 3 bước & < 30s (SC-005); danh sách mượt ~60fps; thay đổi của một thành viên hiển thị cho thành viên khác trong vài giây (đồng bộ)

**Constraints**: Đọc offline cơ bản (dữ liệu tài chính cá nhân/hộ — tham chiếu Nghị định 13/2023/NĐ-CP); một loại tiền tệ; một cấp danh mục con; loại cố định; **không để lại giao dịch mồ côi**; **đồng thời (concurrency)**: nhiều thành viên có thể sửa/xóa cùng danh mục → cần xử lý xung đột (xem research R13)

**Scale/Scope**: Mỗi hộ vài thành viên (~2–8); mỗi hộ vài chục danh mục và hàng nghìn giao dịch dùng chung; phạm vi tính năng = 8 use case danh mục (UC-CAT-01…08) trên dữ liệu chung của hộ

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` vẫn là **bản mẫu chưa phê chuẩn** (toàn placeholder). Không có nguyên tắc ràng buộc để kiểm tra cổng.

- **Kết luận**: PASS (không có gate ràng buộc). Không có vi phạm cần ghi vào Complexity Tracking.
- **Khuyến nghị (không chặn)**: chạy `/speckit-constitution` trước `/speckit-tasks`.

**Re-check sau Phase 1**: Vẫn PASS. Mô hình hộ gia đình thêm 2 thực thể (household, membership) nhưng là cấu trúc tiêu chuẩn cho multi-tenant theo hộ, không phát sinh độ phức tạp cần biện minh riêng.

## Project Structure

### Documentation (this feature)

```text
specs/001-transaction-categorization/
├── plan.md              # This file (/speckit-plan)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── db-schema.sql            # households, household_members, categories, transactions, rules + RLS theo membership + seed
│   └── category-repository.md   # Hợp đồng thao tác danh mục (trong ngữ cảnh hộ hiện tại)
└── tasks.md             # Phase 2 (/speckit-tasks — KHÔNG tạo ở bước này)
```

### Source Code (dưới `src/`)

```text
src/
├── pubspec.yaml
├── analysis_options.yaml
├── lib/
│   ├── main.dart
│   ├── core/
│   │   ├── supabase/             # client + helper RLS/household context
│   │   ├── household/            # "hộ hiện tại" (current household provider) — tiền đề chia sẻ
│   │   ├── router/
│   │   └── error/
│   └── features/
│       └── categorization/
│           ├── domain/           # entities (Category, CategoryType, CategorizationRule), repositories, usecases
│           ├── data/             # models (DTO), supabase datasources, repo impl, cache
│           └── presentation/     # screens (CategoryList, CategoryForm, DeleteReassign), widgets, controllers (Riverpod)
├── supabase/
│   └── migrations/               # bản thi hành của contracts/db-schema.sql (gồm households + RLS theo membership)
├── test/
│   ├── unit/                     # domain use cases + rule gợi ý (lịch sử chung hộ)
│   └── widget/                   # form tạo/sửa, lọc theo loại
└── integration_test/             # end-to-end: đa thành viên thấy chung dữ liệu, cô lập giữa các hộ, authorship
```

**Structure Decision**: Toàn bộ mã nguồn ứng dụng đặt dưới **`src/`** (project root của Flutter là `src/`); repo root giữ gọn cho `specs/` + tài liệu. Giữ feature-first cho `categorization`. Bổ sung **`src/lib/core/household`** giữ ngữ cảnh "hộ hiện tại" (household_id) mà các repository dùng để gắn/lọc dữ liệu. Lược đồ Postgres + RLS + seed nằm trong `src/supabase/migrations`. Quản lý hộ (tạo hộ, mời/loại thành viên) là **dependency** tách riêng (đề xuất BR/feature mới), tính năng này tiêu thụ "hộ hiện tại".

## Dependency — Quản lý hộ gia đình (ngoài phạm vi feature này)

Tính năng phân loại giả định người dùng **đã thuộc một hộ**. Cần một feature/BR riêng cho: tạo hộ, mời thành viên (link/mã mời), tham gia, rời/loại thành viên. Vì **quyền ngang nhau**, không có vai trò quản trị — bất kỳ thành viên nào cũng có thể mời (chính sách mời: xem research R11). Đề xuất tạo `BR-00x: Quản lý hộ gia đình` và use case tương ứng.

## Tác động tới spec (cần cập nhật)

| Vị trí trong spec | Hiện tại | Cần sửa thành |
|-------------------|----------|----------------|
| FR-018 | Danh mục riêng từng người dùng, không chia sẻ | Danh mục **dùng chung trong hộ**; cô lập **giữa các hộ** |
| Assumptions "Danh mục theo từng người dùng" | Không chia sẻ giữa người dùng | Chia sẻ trong hộ; mỗi người dùng thuộc một hộ |
| BR-001 Out of Scope "chia sẻ/đồng bộ giữa nhiều người dùng" | Out of Scope | **In Scope** (chia sẻ trong hộ); vẫn Out of Scope: chia sẻ giữa các hộ khác nhau |
| Key Entities | Danh mục thuộc một người dùng | Danh mục/giao dịch thuộc một **hộ**; giao dịch có **người nhập** |
| Actor (use case diagram) | Người dùng | **Thành viên hộ gia đình** (quyền ngang nhau) |

→ Khuyến nghị `/speckit-clarify` hoặc `/speckit-specify` để chính thức hóa; sau đó cập nhật `specs/diagrams/use-cases.puml` (đổi nhãn actor) và `specs/entities/entity-model.md` (thêm Household/Membership).

## Complexity Tracking

> Không áp dụng — Constitution Check PASS, không có vi phạm cần biện minh.
