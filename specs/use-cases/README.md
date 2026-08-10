# Use Cases — Chỉ mục

**Tác nhân chính**: Thành viên hộ gia đình (quyền ngang nhau, dùng chung sổ tài chính của hộ). Riêng UC-TRK-01: Người dùng (trước khi định danh thành viên).

**Sơ đồ use case** (PlantUML): [`specs/diagrams/use-cases.puml`](../diagrams/use-cases.puml) · ảnh render: [`use-cases.png`](../diagrams/use-cases.png)

> Mô hình chung (từ 2026-06-29): app dùng chung trong **hộ gia đình** — dữ liệu chia sẻ trong hộ, cô lập giữa các hộ; mọi thành viên **quyền ngang nhau**; bản ghi ghi rõ **người tạo** (từ 2026-07-06 tham chiếu **danh sách người dùng độc lập** — feature 002).

## Feature 001 — Phân Loại Giao Dịch ([spec](../001-transaction-categorization/spec.md) · [BR-001](../business-requirements/BR-001.md) — 🔁 re-implementation pending, re-platform Go+Vue 2026-07-10)

| Mã | Use Case | User Story | Truy vết FR (001) | Sơ đồ |
|----|----------|------------|-------------------|:-----:|
| [UC-CAT-01](./001-transaction-categorization/uc-cat-01-xem-loc-danh-muc.md) | Xem & lọc danh mục theo loại Thu/Chi | US1 | FR-002, FR-004, FR-014 | ✅ |
| [UC-CAT-02](./001-transaction-categorization/uc-cat-02-tao-danh-muc.md) | Tạo danh mục mới | US2 | FR-003, FR-004, FR-005, FR-017, FR-018 | ✅ |
| [UC-CAT-03](./001-transaction-categorization/uc-cat-03-tao-danh-muc-con.md) | Tạo danh mục con (một cấp) | US3 | FR-010, FR-011, FR-017 | ✅ |
| [UC-CAT-04](./001-transaction-categorization/uc-cat-04-chinh-sua-danh-muc.md) | Chỉnh sửa tên & biểu tượng danh mục | US2 | FR-006, FR-016, FR-017 | ✅ |
| [UC-CAT-05](./001-transaction-categorization/uc-cat-05-xoa-danh-muc.md) | Xóa danh mục (gán lại / xóa giao dịch) | US2, US3 | FR-007, FR-008, FR-009, FR-012 | ✅ |
| [UC-CAT-06](./001-transaction-categorization/uc-cat-06-an-bo-an-danh-muc.md) | Ẩn / bỏ ẩn danh mục | Edge case | FR-020 | ✅ |
| [UC-CAT-07](./001-transaction-categorization/uc-cat-07-gan-danh-muc-giao-dich.md) | Gán danh mục cho giao dịch | US1 | FR-013, FR-014, FR-019 | ✅ |
| [UC-CAT-08](./001-transaction-categorization/uc-cat-08-goi-y-danh-muc.md) | Gợi ý danh mục khi nhập giao dịch | US4 | FR-015 | ✅ |

## Feature 002 — Ghi Chép Thu Nhập & Chi Phí ([spec](../002-transaction-tracking/spec.md) · [BR-002](../business-requirements/BR-002.md) — 🚧 spec)

| Mã | Use Case | User Story | Truy vết FR (002) | Sơ đồ |
|----|----------|------------|-------------------|:-----:|
| [UC-TRK-01](./002-transaction-tracking/uc-trk-01-dang-nhap-dinh-danh.md) | Đăng nhập và định danh người dùng *(nền tảng — làm ĐẦU TIÊN)* | US5 | FR-015, FR-016 | ✅ |
| [UC-TRK-02](./002-transaction-tracking/uc-trk-02-nhap-giao-dich.md) | Nhập giao dịch thu/chi | US1 | FR-001…007, FR-011…013 | ✅ |
| [UC-TRK-03](./002-transaction-tracking/uc-trk-03-xem-so-giao-dich.md) | Xem sổ giao dịch chung của hộ | US2 | FR-008, FR-012, FR-013 | ✅ |
| [UC-TRK-04](./002-transaction-tracking/uc-trk-04-chinh-sua-giao-dich.md) | Chỉnh sửa giao dịch | US3 | FR-009, FR-011, FR-013, FR-014 | ✅ |
| [UC-TRK-05](./002-transaction-tracking/uc-trk-05-xoa-giao-dich.md) | Xóa giao dịch (có xác nhận) | US4 | FR-010, FR-011, FR-013, FR-014 | ✅ |

## Feature 003 — Thiết Lập Ngân Sách ([spec](../003-budgeting/spec.md) · [BR-003](../business-requirements/BR-003.md) — 🚧 spec, chờ plan)

| Mã | Use Case | User Story | Truy vết FR (003) | Sơ đồ |
|----|----------|------------|-------------------|:-----:|
| [UC-BGT-01](./003-budgeting/uc-bgt-01-tao-ngan-sach-danh-muc.md) | Tạo ngân sách theo danh mục | US1 | FR-001, FR-003, FR-004, FR-010, FR-011 | ✅ |
| [UC-BGT-02](./003-budgeting/uc-bgt-02-tao-ngan-sach-tong.md) | Tạo ngân sách tổng chi tiêu | US4 | FR-002, FR-003, FR-004, FR-010, FR-011 | ✅ |
| [UC-BGT-03](./003-budgeting/uc-bgt-03-theo-doi-tien-do.md) | Theo dõi tiến độ ngân sách | US2 | FR-005, FR-006 | ✅ |
| [UC-BGT-04](./003-budgeting/uc-bgt-04-canh-bao-ngan-sach.md) | Nhận cảnh báo ngưỡng & vượt ngân sách | US3 | FR-007, FR-008, FR-009 | ✅ |
| [UC-BGT-05](./003-budgeting/uc-bgt-05-chinh-sua-ngan-sach.md) | Chỉnh sửa ngân sách | US5 | FR-012, FR-013 | ✅ |
| [UC-BGT-06](./003-budgeting/uc-bgt-06-xoa-ngan-sach.md) | Xóa ngân sách (có xác nhận) | US5 | FR-012, FR-013 | ✅ |

## Feature 008 — Báo Cáo Thu Chi Theo Thành Viên ([BR-008](../business-requirements/BR-008.md) — 📝 draft, mở rộng báo cáo [005-reports](../005-reports/spec.md) / BR-006)

| Mã | Use Case | Truy vết BR (008) | Sơ đồ |
|----|----------|-------------------|:-----:|
| [UC-MBR-01](./008-member-reports/uc-mbr-01-xem-bao-cao-theo-thanh-vien.md) | Xem báo cáo thu/chi theo từng thành viên | BR-MBR-001…005, 008, 009 | ✅ |
| [UC-MBR-02](./008-member-reports/uc-mbr-02-xem-giao-dich-thanh-vien.md) | Xem danh sách giao dịch của một thành viên | BR-MBR-006, 007, 008, 009 | ✅ |

## Sơ đồ quan hệ (rút gọn)

![Sơ đồ use case](../diagrams/use-cases.png)

```
Người dùng ── UC-TRK-01  Đăng nhập & định danh  (tiền đề mọi UC khác)
   │
Thành viên hộ gia đình  (quyền ngang nhau — dữ liệu dùng chung trong hộ)
   │
   ├── Quản lý danh mục (001)
   │     ├── UC-CAT-01  Xem & lọc danh mục
   │     ├── UC-CAT-02  Tạo danh mục mới
   │     ├── UC-CAT-03  Tạo danh mục con      «extend» UC-CAT-02
   │     ├── UC-CAT-04  Chỉnh sửa danh mục
   │     ├── UC-CAT-05  Xóa danh mục          «include» gán lại giao dịch
   │     └── UC-CAT-06  Ẩn / bỏ ẩn danh mục
   │
   ├── Ghi chép thu chi (002)
   │     ├── UC-TRK-02  Nhập giao dịch        «include» UC-CAT-07 (gán danh mục)
   │     │                                     └─ «extend» UC-CAT-08 (gợi ý)
   │     ├── UC-TRK-03  Xem sổ giao dịch chung
   │     ├── UC-TRK-04  Chỉnh sửa giao dịch   «extend» UC-TRK-03 · «include» UC-CAT-07 khi đổi loại
   │     └── UC-TRK-05  Xóa giao dịch         «extend» UC-TRK-03
   │
   ├── Thiết lập ngân sách (003)
   │     ├── UC-BGT-01  Tạo ngân sách theo danh mục
   │     ├── UC-BGT-02  Tạo ngân sách tổng chi tiêu
   │     ├── UC-BGT-03  Theo dõi tiến độ ngân sách   ┄ phụ thuộc dữ liệu UC-TRK-02 (đọc giao dịch Chi; tiến độ suy ra)
   │     ├── UC-BGT-04  Nhận cảnh báo ngưỡng & vượt  «extend» UC-BGT-03
   │     ├── UC-BGT-05  Chỉnh sửa ngân sách          «extend» UC-BGT-03
   │     └── UC-BGT-06  Xóa ngân sách (có xác nhận)  «extend» UC-BGT-03
   │
   ├── Báo cáo theo thành viên (008 — mở rộng 005-reports)
   │     ├── UC-MBR-01  Xem báo cáo thu/chi theo từng thành viên  ┄ phụ thuộc dữ liệu UC-TRK-02 (đọc giao dịch theo created_by; số liệu suy ra)
   │     └── UC-MBR-02  Xem danh sách giao dịch của một thành viên  «extend» UC-MBR-01
   │
   └── Tiền đề: Quản lý hộ (feature riêng — BR đề xuất)
         ├── UC-HH-01  Tạo hộ gia đình
         ├── UC-HH-02  Mời thành viên
         └── UC-HH-03  Tham gia hộ
```

## Ghi chú phạm vi

- Feature 001 chỉ bao phủ phần **danh mục** của luồng nhập; feature 002 bao phủ luồng
  **nhập/sửa/xóa giao dịch** đầy đủ (số tiền, ngày giờ, tài khoản, sổ chung).
- **Tài khoản/Ví** (số dư, chuyển tiền) thuộc BR-005 — feature 002 chỉ cần "mỗi giao
  dịch gắn một tài khoản, số dư cập nhật đúng" (hộ có tài khoản mặc định).
- **Quản lý hộ** (tạo hộ, mời/tham gia) là tiền đề thuộc feature riêng (UC-HH-*).
- Mọi thành viên **quyền ngang nhau**; bản ghi ghi rõ **người tạo** — hiển thị **tên**
  (không phải mã định danh) trong sổ chung.
- Feature 003 (**ngân sách**) chỉ **đọc** giao dịch (002) và danh mục (001) để tính
  tiến độ suy ra; không sửa dữ liệu nguồn. Tiến độ & cảnh báo (80% / vượt 100%) đồng bộ
  realtime cho mọi thành viên; mỗi mức cảnh báo phát tối đa 1 lần/kỳ (phát lại nếu tụt
  dưới mức rồi vượt lại). Sửa/xóa ngân sách (UC-BGT-05/06) là vòng đời suy luận — chờ
  nghiệp vụ xác nhận (Assumptions spec 003).
