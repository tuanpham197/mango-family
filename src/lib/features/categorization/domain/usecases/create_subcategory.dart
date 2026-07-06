import '../../../../core/error/failures.dart';
import '../entities/category.dart';
import '../repositories/category_repository.dart';

/// Tạo danh mục con dưới một danh mục gốc (UC-CAT-03). [T039]
/// - Con kế thừa `type` của cha — FR-011 (không nhận `type` từ ngoài).
/// - Chặn lồng quá một cấp: cha phải là danh mục gốc — FR-010, SC-008.
/// - Cảnh báo trùng tên trong cùng cha qua [DuplicateNameWarning] — FR-017.
class CreateSubcategory {
  final CategoryRepository _repo;
  const CreateSubcategory(this._repo);

  Future<Category> call({
    required String parentId,
    required String name,
    String? icon,
    bool allowDuplicate = false,
  }) {
    if (name.trim().isEmpty) throw const NameRequired();
    return _repo.createSubcategory(
      parentId: parentId,
      name: name,
      icon: icon,
      allowDuplicate: allowDuplicate,
    );
  }
}
