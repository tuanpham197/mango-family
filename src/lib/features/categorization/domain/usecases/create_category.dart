import '../../../../core/error/failures.dart';
import '../entities/category.dart';
import '../repositories/category_repository.dart';

/// Tạo danh mục mới (UC-CAT-02). [T030]
/// - Loại bắt buộc (kiểu non-null) — FR-004.
/// - Tên bắt buộc; cảnh báo trùng tên qua [DuplicateNameWarning] khi `allowDuplicate=false` — FR-017.
/// - Tạo nhanh từ luồng nhập giao dịch: truyền `type` theo loại giao dịch (A3).
class CreateCategory {
  final CategoryRepository _repo;
  const CreateCategory(this._repo);

  Future<Category> call({
    required CategoryType type,
    required String name,
    String? icon,
    String? parentId,
    bool allowDuplicate = false,
  }) {
    if (name.trim().isEmpty) throw const NameRequired();
    return _repo.createCategory(
      type: type,
      name: name,
      icon: icon,
      parentId: parentId,
      allowDuplicate: allowDuplicate,
    );
  }
}
