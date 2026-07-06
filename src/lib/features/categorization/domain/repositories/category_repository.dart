import '../entities/category.dart';

/// Hành động khi xóa danh mục còn giao dịch (FR-008).
enum DeleteAction { reassign, delete }

/// Dữ liệu giao dịch tối thiểu cho bước gán danh mục (đầy đủ thuộc BR-002).
class TransactionDraft {
  final double amount;
  final CategoryType type;
  final String? description;
  final DateTime date;

  const TransactionDraft({
    required this.amount,
    required this.type,
    this.description,
    required this.date,
  });
}

/// Hợp đồng tầng domain cho danh mục (xem `contracts/category-repository.md`). [T019]
abstract interface class CategoryRepository {
  Future<List<Category>> listCategories({
    CategoryType? type,
    bool includeHidden = false,
  });

  Future<Category> createCategory({
    required CategoryType type,
    required String name,
    String? icon,
    String? parentId,
    bool allowDuplicate = false,
  });

  /// Tạo danh mục con dưới [parentId] (cha phải là danh mục gốc);
  /// con kế thừa `type` của cha (FR-010, FR-011).
  Future<Category> createSubcategory({
    required String parentId,
    required String name,
    String? icon,
    bool allowDuplicate = false,
  });

  /// [expectedUpdatedAt]: mốc `updatedAt` đã thấy — cập nhật lạc quan giữa
  /// các thành viên; lệch mốc → `ConcurrencyConflict` (R13).
  Future<Category> renameCategory(
    String id, {
    String? name,
    String? icon,
    DateTime? expectedUpdatedAt,
  });

  Future<void> deleteCategory(
    String id, {
    required DeleteAction action,
    String? targetId,
  });

  Future<void> setHidden(String id, bool hidden,
      {DateTime? expectedUpdatedAt});

  Future<void> assignCategoryToTransaction(
    TransactionDraft draft,
    String categoryId,
  );

  Future<Category?> suggestCategory({
    required String description,
    required CategoryType type,
  });
}
