# UC-CAT-08: Gợi ý danh mục khi nhập giao dịch

| Thuộc tính | Giá trị |
|------------|---------|
| **Mã** | UC-CAT-08 |
| **Tính năng** | Phân loại giao dịch ([spec.md](../../001-transaction-categorization/spec.md)) |
| **User Story** | US4 (P4) |
| **Tác nhân chính** | Người dùng |
| **Mức độ ưu tiên** | Thấp (tiện ích) |
| **Quan hệ** | «extend» UC-CAT-07 |

## Mục tiêu

Tăng tốc độ nhập liệu và nâng tỷ lệ phân loại đúng bằng cách đề xuất sẵn một danh mục
phù hợp, dựa trên mô tả giao dịch và lịch sử phân loại của chính người dùng.

## Tiền điều kiện

- Người dùng đã đăng nhập và đang ở luồng nhập giao dịch.
- Đã có danh mục và (tùy chọn) lịch sử phân loại trước đó.

## Kích hoạt (Trigger)

Người dùng nhập mô tả/loại cho giao dịch trong lúc nhập.

## Luồng sự kiện chính

1. Người dùng nhập thông tin giao dịch (gồm mô tả và loại Thu/Chi).
2. Hệ thống đối chiếu **KẾT HỢP**:
   - khớp **từ khóa theo quy tắc** trên mô tả giao dịch, VÀ
   - **lịch sử phân loại** của chính người dùng.
3. Khi nhận diện được, hệ thống **đề xuất sẵn** một danh mục cùng loại Thu/Chi đang chọn
   (ví dụ mô tả chứa "Grab" → gợi ý "Di chuyển").
4. Người dùng **chấp nhận** gợi ý → danh mục được điền sẵn cho giao dịch.

## Luồng thay thế / ngoại lệ

- **3a. Không khớp quy tắc/lịch sử nào**: Hệ thống không hiển thị gợi ý; người dùng tự chọn danh mục (UC-CAT-07).
- **4a. Người dùng ghi đè**: Người dùng chọn một danh mục khác → lựa chọn của người
  dùng được ưu tiên và lưu cùng giao dịch (gợi ý bị ghi đè).

## Hậu điều kiện

- Nếu chấp nhận, giao dịch được điền sẵn danh mục gợi ý (vẫn phải qua UC-CAT-07 để lưu).
- Gợi ý luôn mang tính tham khảo, không bao giờ ép buộc.

## Quy tắc nghiệp vụ

- Cơ chế gợi ý KẾT HỢP khớp từ khóa theo quy tắc VÀ lịch sử phân loại của người dùng;
  **không dùng AI/ML** (FR-015, chốt tại Clarifications 2026-06-24).
- Người dùng luôn có thể **chấp nhận hoặc ghi đè** gợi ý (FR-015).
- Gợi ý phải cùng loại Thu/Chi với giao dịch đang chọn.

## Truy vết

- **FR**: FR-015
- **Acceptance**: US4 #1, #2, #3
- **Success Criteria**: SC-004 (≥ 60% gợi ý được chấp nhận khi bật tính năng)
- **BR**: BR-CAT-007
