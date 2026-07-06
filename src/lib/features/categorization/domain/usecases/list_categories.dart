import '../entities/category.dart';
import '../repositories/category_repository.dart';

/// Liệt kê danh mục, lọc theo loại, loại trừ danh mục ẩn khi nhập giao dịch. [T023]
/// UC-CAT-01 · FR-002, FR-014, FR-020.
class ListCategories {
  final CategoryRepository _repo;
  const ListCategories(this._repo);

  Future<List<Category>> call({
    CategoryType? type,
    bool includeHidden = false,
  }) =>
      _repo.listCategories(type: type, includeHidden: includeHidden);
}
