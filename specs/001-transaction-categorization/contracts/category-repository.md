# Contract — CategoryRepository (domain interface)

**Feature**: 001-transaction-categorization · **Plan**: [../plan.md](../plan.md)

Hợp đồng tầng domain mà `presentation` phụ thuộc và `data` (Supabase) hiện thực. Mỗi thao tác ghi rõ tiền điều kiện, hậu điều kiện và lỗi nghiệp vụ, truy vết về use case / FR. Đây là **hợp đồng hành vi** — không phải code cài đặt.

> **Ngữ cảnh hộ gia đình (cập nhật 2026-06-29)**: Mọi thao tác chạy trong ngữ cảnh **hộ hiện tại** (`householdId` lấy từ `core/household`). Dữ liệu **dùng chung trong hộ**, cô lập **giữa các hộ** bằng RLS theo membership (`is_member(household_id)` — FR-018 đã đổi). **Mọi thành viên ngang quyền**: không có cổng quyền theo vai trò; bất kỳ thành viên nào cũng thực hiện được mọi thao tác dưới đây.
>
> Kiểu dữ liệu: `Category { id, householdId, name, type: INCOME|EXPENSE, icon?, parentId?, isDefault, isHidden, createdBy?, updatedAt? }`. Tiền điều kiện chung: **đã đăng nhập và là thành viên của hộ hiện tại**. `updatedAt` là mốc cập nhật lạc quan (R13): thao tác ghi nhận `expectedUpdatedAt` tùy chọn — lệch mốc (thành viên khác đã đổi/xóa) → `ConcurrencyConflict`, không ghi đè thầm lặng.

| Thao tác | Mô tả | Tiền điều kiện | Hậu điều kiện / Lỗi | Truy vết |
|----------|-------|----------------|----------------------|----------|
| `listCategories({type?, includeHidden})` | Lấy danh mục, nhóm cha→con, lọc theo loại | Đã đăng nhập | Trả danh sách đúng loại; khi nhập giao dịch `includeHidden=false` (ẩn không hiện) | UC-CAT-01, FR-002, FR-014, FR-020 |
| `createCategory({type, name, icon?, parentId?})` | Tạo danh mục (hoặc con nếu có `parentId`) | Đã đăng nhập; nếu có `parentId` thì cha tồn tại & là cấp gốc | Tạo danh mục mới của người dùng; con kế thừa `type` của cha. **Lỗi**: thiếu `type` (chưa có `parentId`) → `TypeRequired`; cha đã là con → `NestingTooDeep`; trùng tên cùng (type,parent) → cảnh báo `DuplicateNameWarning` (không chặn) | UC-CAT-02, UC-CAT-03, FR-003, FR-010, FR-011, FR-017 |
| `createSubcategory({parentId, name, icon?})` | Tạo danh mục con dưới một danh mục gốc | Đã đăng nhập; cha tồn tại (`ParentNotFound` nếu không) & là cấp gốc | Con kế thừa `type` của cha; **Lỗi**: cha đã là con → `NestingTooDeep`; trùng tên cùng cha → `DuplicateNameWarning` | UC-CAT-03, FR-010, FR-011, SC-008 |
| `renameCategory(id, {name?, icon?, expectedUpdatedAt?})` | Đổi tên/biểu tượng | Đã đăng nhập; danh mục tồn tại | Tên/biểu tượng cập nhật, phản ánh mọi nơi tham chiếu (giao dịch tham chiếu theo id). **Không** đổi được `type` → `TypeImmutable`; lệch mốc `expectedUpdatedAt` → `ConcurrencyConflict` (R13) | UC-CAT-04, FR-006, FR-016, FR-005 |
| `deleteCategory(id, {action: REASSIGN\|DELETE, targetId?})` | Xóa an toàn (kèm con) | Đã đăng nhập; danh mục tồn tại; nếu `REASSIGN` thì `targetId` cùng loại | Khi còn giao dịch: gán-lại (cùng loại) hoặc xóa giao dịch, xử lý cả con, rồi xóa danh mục. Không còn giao dịch → xóa ngay. **Lỗi**: target khác loại → `ReassignTypeMismatch`; thiếu lựa chọn khi còn giao dịch → `ReassignOrDeleteRequired` | UC-CAT-05, FR-007, FR-008, FR-009, FR-012, SC-007 |
| `setHidden(id, hidden, {expectedUpdatedAt?})` | Ẩn / bỏ ẩn | Đã đăng nhập; danh mục tồn tại | `is_hidden` cập nhật; ẩn không xóa dữ liệu, đảo ngược được; giữ nguyên lịch sử/báo cáo; lệch mốc → `ConcurrencyConflict` (R13) | UC-CAT-06, FR-020 |
| `assignCategoryToTransaction(txnDraft, categoryId)` | Gán danh mục khi nhập giao dịch | `categoryId` cùng loại & cùng hộ với giao dịch | Giao dịch lưu kèm danh mục và `createdBy` = thành viên hiện tại. **Lỗi**: chưa chọn → `CategoryRequired` (chặn lưu); khác loại → `TypeMismatch`; khác hộ → `HouseholdMismatch` | UC-CAT-07, FR-013, FR-014, FR-019, SC-002, SC-003 |
| `suggestCategory({description, type})` | Gợi ý danh mục | (trong hộ) | Trả danh mục đề xuất cùng loại theo từ khóa + **lịch sử chung của hộ**, hoặc rỗng; luôn ghi đè được | UC-CAT-08, FR-015, SC-004 |

## Ánh xạ lỗi nghiệp vụ ↔ ràng buộc DB

| Lỗi domain | Hàng rào DB (db-schema.sql) |
|------------|-----------------------------|
| `TypeMismatch` / luôn cùng loại | `trg_type_match` trên `transactions` |
| `TypeImmutable` | `trg_type_immutable` trên `categories` |
| `NestingTooDeep` / kế thừa loại con | `enforce_one_level()` |
| `CategoryRequired` (không giao dịch mồ côi) | `transactions.category_id NOT NULL` |
| `ReassignTypeMismatch` / xóa an toàn | RPC `delete_category(...)` (kiểm `is_member`) |
| `HouseholdMismatch` (danh mục & giao dịch cùng hộ) | `enforce_txn_rules()` |
| `ConcurrencyConflict` (không ghi đè thầm lặng — R13) | `categories.updated_at` + trigger `trg_touch_updated_at`; client UPDATE kèm điều kiện `updated_at` |
| Dùng chung trong hộ, cô lập giữa các hộ | RLS `member_*` policies + `is_member(household_id)` |
