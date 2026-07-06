import '../../../../core/error/failures.dart';
import '../../domain/entities/category.dart';
import '../../domain/repositories/category_repository.dart';

/// Repository demo (in-memory) dùng khi CHƯA cấu hình Supabase — để chạy thử UI.
/// Seed sẵn bộ danh mục mặc định + một danh mục con; không bền vững sau khi tắt app.
class InMemoryCategoryRepository implements CategoryRepository {
  static const _hh = 'demo-household';
  final List<Category> _cats = [];
  // keyword (chữ thường) → categoryId → số lần xác nhận (học demo — US4).
  final Map<String, Map<String, int>> _learned = {};
  int _seq = 0;

  InMemoryCategoryRepository() {
    _seed();
  }

  String _id() => 'c${++_seq}';

  void _seed() {
    for (final n in const [
      'Ăn uống', 'Di chuyển', 'Hóa đơn', 'Mua sắm', 'Giải trí', 'Sức khỏe'
    ]) {
      _cats.add(Category(
          id: _id(),
          householdId: _hh,
          name: n,
          type: CategoryType.expense,
          isDefault: true));
    }
    for (final n in const ['Lương', 'Thưởng']) {
      _cats.add(Category(
          id: _id(),
          householdId: _hh,
          name: n,
          type: CategoryType.income,
          isDefault: true));
    }
    final parent = _cats.firstWhere((c) => c.name == 'Ăn uống');
    _cats.add(Category(
        id: _id(),
        householdId: _hh,
        name: 'Ăn ngoài',
        type: CategoryType.expense,
        parentId: parent.id));
  }

  @override
  Future<List<Category>> listCategories({
    CategoryType? type,
    bool includeHidden = false,
  }) async {
    return _cats
        .where((c) =>
            (type == null || c.type == type) &&
            (includeHidden || !c.isHidden))
        .toList();
  }

  @override
  Future<Category> createCategory({
    required CategoryType type,
    required String name,
    String? icon,
    String? parentId,
    bool allowDuplicate = false,
  }) async {
    final trimmed = name.trim();
    final effType = parentId != null
        ? _cats.firstWhere((c) => c.id == parentId).type
        : type;
    if (!allowDuplicate &&
        _cats.any((c) =>
            c.parentId == parentId &&
            c.type == effType &&
            c.name.toLowerCase() == trimmed.toLowerCase())) {
      throw const DuplicateNameWarning();
    }
    final c = Category(
      id: _id(),
      householdId: _hh,
      name: trimmed,
      type: effType,
      icon: icon,
      parentId: parentId,
    );
    _cats.add(c);
    return c;
  }

  @override
  Future<Category> createSubcategory({
    required String parentId,
    required String name,
    String? icon,
    bool allowDuplicate = false,
  }) async {
    final i = _cats.indexWhere((c) => c.id == parentId);
    if (i < 0) throw const ParentNotFound();
    final parent = _cats[i];
    if (parent.parentId != null) throw const NestingTooDeep(); // FR-010
    return createCategory(
      type: parent.type, // kế thừa loại của cha (FR-011)
      name: name,
      icon: icon,
      parentId: parentId,
      allowDuplicate: allowDuplicate,
    );
  }

  @override
  Future<Category> renameCategory(
    String id, {
    String? name,
    String? icon,
    DateTime? expectedUpdatedAt, // demo một người dùng: không có xung đột
  }) async {
    final i = _cats.indexWhere((c) => c.id == id);
    if (i < 0) throw const ConcurrencyConflict(); // đã bị xóa (T050)
    final cur = _cats[i];
    final newName = name?.trim();
    if (newName != null &&
        newName.isNotEmpty &&
        _cats.any((c) =>
            c.id != id &&
            c.parentId == cur.parentId &&
            c.type == cur.type &&
            c.name.toLowerCase() == newName.toLowerCase())) {
      throw const DuplicateNameWarning();
    }
    final updated = cur.copyWith(
      name: (newName != null && newName.isNotEmpty) ? newName : cur.name,
      icon: icon ?? cur.icon,
    );
    _cats[i] = updated;
    return updated;
  }

  @override
  Future<void> deleteCategory(
    String id, {
    required DeleteAction action,
    String? targetId,
  }) async {
    _cats.removeWhere((c) => c.id == id || c.parentId == id);
  }

  @override
  Future<void> setHidden(String id, bool hidden,
      {DateTime? expectedUpdatedAt}) async {
    final i = _cats.indexWhere((c) => c.id == id);
    if (i < 0) throw const ConcurrencyConflict(); // đã bị xóa (T050)
    _cats[i] = _cats[i].copyWith(isHidden: hidden);
  }

  @override
  Future<void> assignCategoryToTransaction(
    TransactionDraft draft,
    String categoryId,
  ) async {
    final cat = _cats.firstWhere((c) => c.id == categoryId);
    if (cat.type != draft.type) throw const TypeMismatch();
    // demo: không lưu giao dịch, chỉ học từ khóa → danh mục (US4).
    for (final w in _tokens(draft.description ?? '')) {
      final perCat = _learned.putIfAbsent(w, () => {});
      perCat.update(categoryId, (v) => v + 1, ifAbsent: () => 1);
    }
  }

  @override
  Future<Category?> suggestCategory({
    required String description,
    required CategoryType type,
  }) async {
    final visible = _cats.where((c) => c.type == type && !c.isHidden);
    // 1) Quy tắc đã học (chọn danh mục được xác nhận nhiều nhất).
    String? bestId;
    var best = 0;
    for (final w in _tokens(description)) {
      _learned[w]?.forEach((id, count) {
        if (count > best && visible.any((c) => c.id == id)) {
          best = count;
          bestId = id;
        }
      });
    }
    if (bestId != null) return visible.firstWhere((c) => c.id == bestId);
    // 2) Fallback: mô tả chứa tên danh mục.
    final d = description.toLowerCase();
    for (final c in visible) {
      if (d.contains(c.name.toLowerCase())) return c;
    }
    return null;
  }

  static Iterable<String> _tokens(String text) => text
      .toLowerCase()
      .split(RegExp(r'\s+'))
      .where((w) => w.length >= 3);
}
