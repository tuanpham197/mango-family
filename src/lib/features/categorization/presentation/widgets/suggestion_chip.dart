import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/category_repository_impl.dart';
import '../../domain/entities/category.dart';
import '../../domain/usecases/suggest_category.dart';

/// Gợi ý theo (mô tả, loại) — autoDispose để tính lại khi mô tả/loại đổi.
final categorySuggestionProvider = FutureProvider.autoDispose
    .family<Category?, ({String description, CategoryType type})>(
        (ref, args) async {
  final repo = ref.watch(categoryRepositoryProvider);
  return SuggestCategory(repo)
      .call(description: args.description, type: args.type);
});

/// Chip gợi ý danh mục trong luồng nhập giao dịch (UC-CAT-08). [T046]
/// - Chỉ mang tính tham khảo: bấm để chấp nhận, hoặc bỏ qua và chọn tay
///   (ghi đè) — FR-015, SC-004.
/// - Tự ẩn khi không có gợi ý, hoặc gợi ý trùng danh mục đã chọn.
class SuggestionChip extends ConsumerWidget {
  final String description;
  final CategoryType type;
  final Category? selected;
  final ValueChanged<Category> onAccepted;

  const SuggestionChip({
    super.key,
    required this.description,
    required this.type,
    required this.onAccepted,
    this.selected,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (description.trim().isEmpty) return const SizedBox.shrink();
    final async = ref.watch(
        categorySuggestionProvider((description: description, type: type)));
    final suggestion = async.valueOrNull;
    if (suggestion == null || suggestion.id == selected?.id) {
      return const SizedBox.shrink();
    }
    return Align(
      alignment: Alignment.centerLeft,
      child: Padding(
        padding: const EdgeInsets.only(top: 8),
        child: ActionChip(
          avatar: const Icon(Icons.lightbulb_outline, size: 18),
          label: Text('Gợi ý: ${suggestion.name}'),
          tooltip: 'Bấm để dùng gợi ý — vẫn có thể chọn danh mục khác',
          onPressed: () => onAccepted(suggestion),
        ),
      ),
    );
  }
}
