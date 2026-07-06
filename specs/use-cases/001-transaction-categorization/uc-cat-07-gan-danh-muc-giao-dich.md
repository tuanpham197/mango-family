# UC-CAT-07: Gán danh mục cho giao dịch

| Thuộc tính | Giá trị |
|------------|---------|
| **Mã** | UC-CAT-07 |
| **Tính năng** | Phân loại giao dịch ([spec.md](../../001-transaction-categorization/spec.md)) |
| **User Story** | US1 (P1) — giá trị cốt lõi (MVP) |
| **Tác nhân chính** | Người dùng |
| **Mức độ ưu tiên** | Cao nhất |
| **Quan hệ** | «include» UC-CAT-01 (lọc theo loại); «extend» UC-CAT-08 (gợi ý) |

## Mục tiêu

Bảo đảm **mọi giao dịch đều được gán đúng một danh mục cùng loại Thu/Chi** ngay khi
nhập, làm nền tảng dữ liệu cho ngân sách và báo cáo.

## Tiền điều kiện

- Người dùng đã đăng nhập và đang ở luồng nhập giao dịch (BR-002).
- Đã có sẵn bộ danh mục mặc định phân theo Thu/Chi (FR-001).

## Kích hoạt (Trigger)

Người dùng nhập một giao dịch và đến bước chọn danh mục.

## Luồng sự kiện chính

1. Người dùng chọn **loại** giao dịch (Thu hoặc Chi).
2. Người dùng mở danh sách danh mục.
3. Hệ thống **chỉ hiển thị các danh mục cùng loại** với loại giao dịch đang chọn.
4. Người dùng chọn một danh mục — có thể là **danh mục cha trực tiếp** hoặc **một danh mục con**.
5. Người dùng lưu giao dịch.
6. Hệ thống lưu giao dịch với danh mục đã gán; báo cáo **gộp (roll-up)** giao dịch của
   danh mục con vào danh mục cha.

## Luồng thay thế / ngoại lệ

- **5a. Chưa chọn danh mục**: Khi người dùng bấm Lưu mà chưa chọn danh mục, hệ thống
  **chặn lưu** và yêu cầu chọn một danh mục (không có cơ chế gán "Chưa phân loại" tự động).
- **3a. Loại chưa có danh mục phù hợp**: Nếu loại đang chọn chưa có danh mục nào, hệ
  thống hướng người dùng **tạo nhanh** một danh mục cùng loại (UC-CAT-02) trước khi lưu.
- **Có gợi ý danh mục**: Hệ thống có thể đề xuất sẵn một danh mục (UC-CAT-08); người
  dùng chấp nhận hoặc chọn danh mục khác.
- **Đổi loại giao dịch sau khi đã chọn danh mục**: Nếu người dùng đổi loại Thu/Chi, hệ
  thống làm mới danh sách để chỉ còn danh mục cùng loại mới; lựa chọn khác loại bị loại bỏ.

## Hậu điều kiện

- Giao dịch được lưu, tham chiếu **đúng một** danh mục có loại trùng với loại giao dịch.
- Không tồn tại giao dịch nào được lưu mà thiếu danh mục.

## Quy tắc nghiệp vụ

- Bắt buộc chọn danh mục; chặn lưu nếu chưa chọn (FR-013).
- Chỉ hiển thị danh mục cùng loại Thu/Chi với giao dịch (FR-014).
- Cho phép gán trực tiếp vào danh mục cha HOẶC một danh mục con; báo cáo roll-up con vào cha (FR-019).
- Loại Thu/Chi của giao dịch phải trùng loại của danh mục được gán (Key Entities — Giao dịch).

## Truy vết

- **FR**: FR-013, FR-014, FR-019
- **Acceptance**: US1 #1, #2, #3; US3 #4
- **Success Criteria**: SC-001, SC-002, SC-003, SC-006
- **BR**: BR-CAT-006, BR-CAT-008 (liên kết BR-TRK-002 → BR-TRK-003)
