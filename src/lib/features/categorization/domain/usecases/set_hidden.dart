import '../repositories/category_repository.dart';

/// Ẩn / bỏ ẩn danh mục (UC-CAT-06). [T033]
/// Thao tác không phá hủy, đảo ngược được (FR-020).
/// [expectedUpdatedAt]: cập nhật lạc quan giữa các thành viên (R13, T050).
class SetHidden {
  final CategoryRepository _repo;
  const SetHidden(this._repo);

  Future<void> call(String id,
          {required bool hidden, DateTime? expectedUpdatedAt}) =>
      _repo.setHidden(id, hidden, expectedUpdatedAt: expectedUpdatedAt);
}
