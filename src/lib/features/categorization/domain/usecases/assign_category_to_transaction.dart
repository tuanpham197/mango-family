import '../../../../core/error/failures.dart';
import '../repositories/category_repository.dart';

/// Gán danh mục cho giao dịch; chặn lưu nếu chưa chọn. [T024]
/// UC-CAT-07 · FR-013, FR-014, SC-002, SC-003.
class AssignCategoryToTransaction {
  final CategoryRepository _repo;
  const AssignCategoryToTransaction(this._repo);

  Future<void> call(TransactionDraft draft, String? categoryId) {
    if (categoryId == null || categoryId.isEmpty) {
      throw const CategoryRequired(); // FR-013: chặn lưu khi chưa chọn
    }
    return _repo.assignCategoryToTransaction(draft, categoryId);
  }
}
