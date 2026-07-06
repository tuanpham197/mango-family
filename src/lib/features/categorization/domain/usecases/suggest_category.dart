import '../entities/category.dart';
import '../repositories/category_repository.dart';

/// Gợi ý danh mục theo từ khóa + lịch sử chung của hộ (UC-CAT-08). [T044]
/// - Chỉ gợi ý danh mục **cùng loại** với giao dịch đang nhập — FR-015.
/// - Mang tính tham khảo: người dùng luôn ghi đè được (SC-004).
class SuggestCategory {
  final CategoryRepository _repo;
  const SuggestCategory(this._repo);

  Future<Category?> call({
    required String description,
    required CategoryType type,
  }) {
    final d = description.trim();
    if (d.isEmpty) return Future.value(null);
    return _repo.suggestCategory(description: d, type: type);
  }
}
