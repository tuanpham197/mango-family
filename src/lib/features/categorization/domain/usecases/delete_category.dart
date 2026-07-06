import '../../../../core/error/failures.dart';
import '../repositories/category_repository.dart';

/// Xóa danh mục an toàn (UC-CAT-05). [T032]
/// Còn giao dịch → gán lại (cùng loại) hoặc xóa giao dịch; xử lý cả con (FR-008, FR-009, FR-012).
class DeleteCategory {
  final CategoryRepository _repo;
  const DeleteCategory(this._repo);

  Future<void> call(
    String id, {
    required DeleteAction action,
    String? targetId,
  }) {
    if (action == DeleteAction.reassign &&
        (targetId == null || targetId.isEmpty)) {
      throw const ReassignTargetRequired();
    }
    return _repo.deleteCategory(id, action: action, targetId: targetId);
  }
}
