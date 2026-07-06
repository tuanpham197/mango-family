import '../entities/category.dart';
import '../repositories/category_repository.dart';

/// Đổi tên/biểu tượng danh mục (UC-CAT-04). [T031]
/// Không cho đổi loại (FR-005, chặn ở DB trigger); cảnh báo trùng tên (FR-017, ở repo).
/// [expectedUpdatedAt]: cập nhật lạc quan giữa các thành viên (R13, T050).
class RenameCategory {
  final CategoryRepository _repo;
  const RenameCategory(this._repo);

  Future<Category> call(
    String id, {
    String? name,
    String? icon,
    DateTime? expectedUpdatedAt,
  }) {
    final newName = name?.trim();
    return _repo.renameCategory(
      id,
      name: (newName != null && newName.isNotEmpty) ? newName : null,
      icon: icon,
      expectedUpdatedAt: expectedUpdatedAt,
    );
  }
}
