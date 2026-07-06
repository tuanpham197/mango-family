import 'package:flutter/material.dart';

import '../../domain/entities/category.dart';
import '../screens/category_picker_screen.dart';
import 'suggestion_chip.dart';

/// Ô chọn danh mục trong luồng nhập giao dịch. [T028, T029]
///
/// - Mở picker đã lọc theo [type] (chỉ danh mục cùng loại — FR-014).
/// - Báo lỗi khi [showError] và chưa chọn (chặn lưu — FR-013).
/// - Khi loại Thu/Chi đổi, parent nên xóa [selected] khác loại rồi truyền lại
///   (xem chú thích bên dưới — phần làm tươi danh sách theo loại).
/// - Truyền [description] để hiển thị gợi ý danh mục dưới ô chọn (US4 — T046);
///   gợi ý chỉ tham khảo, bấm chip để chấp nhận hoặc bỏ qua và chọn tay.
class CategorySelectField extends StatelessWidget {
  final CategoryType type;
  final Category? selected;
  final ValueChanged<Category> onChanged;
  final bool showError;
  final String? description;

  const CategorySelectField({
    super.key,
    required this.type,
    required this.selected,
    required this.onChanged,
    this.showError = false,
    this.description,
  });

  Future<void> _pick(BuildContext context) async {
    final picked = await Navigator.of(context).push<Category>(
      MaterialPageRoute(
        builder: (_) => CategoryPickerScreen(
          type: type, // chỉ hiển thị danh mục cùng loại (FR-014)
          onSelected: (c) => Navigator.of(context).pop(c),
        ),
      ),
    );
    if (picked != null) onChanged(picked);
  }

  @override
  Widget build(BuildContext context) {
    // T029: nếu danh mục đã chọn khác loại hiện tại thì coi như chưa chọn.
    final effective = (selected != null && selected!.type == type) ? selected : null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        InkWell(
          onTap: () => _pick(context),
          child: InputDecorator(
            decoration: InputDecoration(
              labelText: 'Danh mục',
              border: const OutlineInputBorder(),
              errorText: showError && effective == null
                  ? 'Vui lòng chọn một danh mục'
                  : null,
            ),
            child: Text(effective?.name ?? 'Chọn danh mục'),
          ),
        ),
        if (description != null)
          SuggestionChip(
            description: description!,
            type: type,
            selected: effective,
            onAccepted: onChanged,
          ),
      ],
    );
  }
}
