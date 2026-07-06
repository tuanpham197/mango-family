import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/error/failures.dart';
import '../../../../core/household/current_household.dart';
import '../../../../core/supabase/env.dart';
import '../../../../core/supabase/supabase_client.dart';
import '../../domain/entities/category.dart';
import '../../domain/repositories/category_repository.dart';
import '../datasources/category_local_cache.dart';
import '../datasources/category_remote_datasource.dart';
import '../datasources/rule_remote_datasource.dart';
import '../models/category_model.dart';
import 'in_memory_category_repository.dart';

/// Hiện thực [CategoryRepository] trên Supabase, gắn ngữ cảnh hộ hiện tại. [T022]
class CategoryRepositoryImpl implements CategoryRepository {
  final CategoryRemoteDataSource _remote;
  final RuleRemoteDataSource _rules;
  final CategoryLocalCache? _cache;
  final Future<String> Function() _householdId;
  final String? Function() _userId;

  CategoryRepositoryImpl(
      this._remote, this._rules, this._cache, this._householdId, this._userId);

  // ---- US1 ----
  @override
  Future<List<Category>> listCategories({
    CategoryType? type,
    bool includeHidden = false,
  }) async {
    final hh = await _householdId();
    final List<Category> cats;
    try {
      cats =
          await _remote.list(hh, type: type?.db, includeHidden: includeHidden);
      // T048: làm mới cache đọc offline (chỉ khi tải đủ, không lọc).
      if (type == null && includeHidden) {
        try {
          await _cache?.save(hh, cats);
        } catch (_) {/* cache là tối ưu, không chặn luồng chính */}
      }
    } on Exception {
      // Mất mạng/lỗi server → đọc từ cache offline nếu có (T048).
      final cached = await _cache?.read(hh,
          type: type?.db, includeHidden: includeHidden);
      if (cached == null) rethrow;
      return _nestParentChild(cached);
    }
    return _nestParentChild(cats);
  }

  /// Sắp xếp lồng cha→con: gốc theo tên, mỗi con ngay sau cha của nó. [T040]
  static List<Category> _nestParentChild(List<Category> cats) {
    final byId = {for (final c in cats) c.id: c};
    int rootRank(Category c) {
      // Con mà cha bị lọc khỏi danh sách (ẩn/khác loại) xếp như gốc.
      final root = c.parentId != null ? byId[c.parentId] ?? c : c;
      return cats.indexOf(root);
    }

    final sorted = [...cats]..sort((a, b) {
        final r = rootRank(a).compareTo(rootRank(b));
        if (r != 0) return r;
        // Cùng nhóm gốc: cha đứng trước, các con theo tên.
        if (a.parentId == null) return -1;
        if (b.parentId == null) return 1;
        return a.name.toLowerCase().compareTo(b.name.toLowerCase());
      });
    return sorted;
  }

  @override
  Future<void> assignCategoryToTransaction(
    TransactionDraft draft,
    String categoryId,
  ) async {
    final hh = await _householdId();
    final uid = _userId();
    if (uid == null) throw const NoHousehold();

    // Xác thực danh mục cùng loại & cùng hộ (hàng rào DB cũng có trigger).
    final cats = await _remote.list(hh, includeHidden: true);
    Category? cat;
    for (final c in cats) {
      if (c.id == categoryId) {
        cat = c;
        break;
      }
    }
    if (cat == null) throw const CategoryRequired();
    if (cat.type != draft.type) throw const TypeMismatch();

    await _remote.insertTransaction({
      'household_id': hh,
      'created_by': uid,
      'amount': draft.amount,
      'type': draft.type.db,
      'category_id': categoryId,
      'description': draft.description,
      'transaction_date': draft.date.toIso8601String(),
    });

    // T047: học từ xác nhận (kể cả khi ghi đè gợi ý) — best-effort,
    // lỗi ghi quy tắc không được làm hỏng việc lưu giao dịch.
    final desc = draft.description?.trim();
    if (desc != null && desc.isNotEmpty) {
      try {
        await _rules.learn(hh, desc, categoryId);
      } catch (_) {/* bỏ qua: gợi ý chỉ mang tính tham khảo */}
    }
  }

  // ---- US2 (T030–T034): hiện thực ở phase US2 ----
  @override
  Future<Category> createCategory({
    required CategoryType type,
    required String name,
    String? icon,
    String? parentId,
    bool allowDuplicate = false,
  }) async {
    final hh = await _householdId();
    final trimmed = name.trim();
    if (!allowDuplicate) {
      final existing = await _remote.list(hh, type: type.db, includeHidden: true);
      final dup = existing.any((c) =>
          c.parentId == parentId &&
          c.name.toLowerCase() == trimmed.toLowerCase());
      if (dup) throw const DuplicateNameWarning(); // FR-017
    }
    return _remote.insert(CategoryModel.toInsert(
      householdId: hh,
      name: trimmed,
      type: type,
      icon: icon,
      parentId: parentId,
      createdBy: _userId(),
    ));
  }

  // ---- US3 (T040) ----
  @override
  Future<Category> createSubcategory({
    required String parentId,
    required String name,
    String? icon,
    bool allowDuplicate = false,
  }) async {
    final hh = await _householdId();
    final all = await _remote.list(hh, includeHidden: true);
    Category? parent;
    for (final c in all) {
      if (c.id == parentId) {
        parent = c;
        break;
      }
    }
    if (parent == null) throw const ParentNotFound();
    if (parent.parentId != null) throw const NestingTooDeep(); // FR-010

    final trimmed = name.trim();
    if (!allowDuplicate) {
      final dup = all.any((c) =>
          c.parentId == parentId &&
          c.name.toLowerCase() == trimmed.toLowerCase());
      if (dup) throw const DuplicateNameWarning(); // FR-017
    }
    // DB trigger `trg_inherit_type` là hàng rào cuối cho kế thừa loại (FR-011).
    return _remote.insert(CategoryModel.toInsert(
      householdId: hh,
      name: trimmed,
      type: parent.type,
      icon: icon,
      parentId: parentId,
      createdBy: _userId(),
    ));
  }

  @override
  Future<Category> renameCategory(
    String id, {
    String? name,
    String? icon,
    DateTime? expectedUpdatedAt,
  }) async {
    final newName = name?.trim();
    if (newName != null && newName.isNotEmpty) {
      final hh = await _householdId();
      final all = await _remote.list(hh, includeHidden: true);
      Category? current;
      for (final c in all) {
        if (c.id == id) {
          current = c;
          break;
        }
      }
      if (current != null) {
        final cur = current;
        final dup = all.any((c) =>
            c.id != id &&
            c.parentId == cur.parentId &&
            c.type == cur.type &&
            c.name.toLowerCase() == newName.toLowerCase());
        if (dup) throw const DuplicateNameWarning(); // FR-017
      }
    }
    // T050: ghi có điều kiện — 0 hàng khớp nghĩa là thành viên khác đã
    // đổi/xóa bản ghi (R13); báo xung đột thay vì ghi đè thầm lặng.
    final updated = await _remote.update(
      id,
      {
        if (newName != null && newName.isNotEmpty) 'name': newName,
        if (icon != null) 'icon': icon,
      },
      ifUnmodifiedSince: expectedUpdatedAt,
    );
    if (updated == null) throw const ConcurrencyConflict();
    return updated;
  }

  @override
  Future<void> deleteCategory(
    String id, {
    required DeleteAction action,
    String? targetId,
  }) =>
      _remote.deleteViaRpc(
        id,
        action == DeleteAction.reassign ? 'REASSIGN' : 'DELETE',
        targetId,
      );

  @override
  Future<void> setHidden(String id, bool hidden,
      {DateTime? expectedUpdatedAt}) async {
    final updated = await _remote.update(
      id,
      {'is_hidden': hidden},
      ifUnmodifiedSince: expectedUpdatedAt,
    );
    if (updated == null) throw const ConcurrencyConflict(); // T050, R13
  }

  // ---- US4 (T044/T045): quy tắc từ khóa + lịch sử chung của hộ ----
  @override
  Future<Category?> suggestCategory({
    required String description,
    required CategoryType type,
  }) async {
    final hh = await _householdId();
    final tokens = RuleRemoteDataSource.extractKeywords(description).toSet();
    if (tokens.isEmpty) return null;

    // Danh mục ứng viên: cùng loại, không ẩn (gợi ý để chọn khi nhập mới).
    final cats = await _remote.list(hh, type: type.db);
    if (cats.isEmpty) return null;
    final byId = {for (final c in cats) c.id: c};

    // 1) Quy tắc từ khóa — đã sắp theo match_count giảm dần (FR-015).
    final rules = await _rules.listRules(hh);
    for (final r in rules) {
      if (tokens.contains(r.keyword.toLowerCase())) {
        final cat = byId[r.categoryId];
        if (cat != null) return cat; // byId đã lọc cùng loại + không ẩn
      }
    }

    // 2) Lịch sử chung của hộ: danh mục hay dùng nhất cho mô tả tương tự.
    final history = await _rules.recentHistory(hh, type.db);
    final scores = <String, int>{};
    for (final h in history) {
      final hTokens =
          RuleRemoteDataSource.extractKeywords(h.description).toSet();
      if (hTokens.intersection(tokens).isNotEmpty) {
        scores.update(h.categoryId, (v) => v + 1, ifAbsent: () => 1);
      }
    }
    String? bestId;
    var best = 0;
    scores.forEach((id, score) {
      if (score > best && byId.containsKey(id)) {
        best = score;
        bestId = id;
      }
    });
    return bestId != null ? byId[bestId] : null;
  }
}

/// Provider repository. Khi chưa cấu hình Supabase → dùng repo demo in-memory
/// để chạy thử UI; ngược lại dùng Supabase thật.
final categoryRepositoryProvider = Provider<CategoryRepository>((ref) {
  if (!Env.isConfigured) {
    return InMemoryCategoryRepository();
  }
  return CategoryRepositoryImpl(
    CategoryRemoteDataSource(),
    RuleRemoteDataSource(),
    CategoryLocalCache(),
    () => ref.read(currentHouseholdProvider.future),
    () => supabase.auth.currentUser?.id,
  );
});
