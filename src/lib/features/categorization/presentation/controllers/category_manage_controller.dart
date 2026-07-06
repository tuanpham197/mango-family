import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/household/current_household.dart';
import '../../../../core/supabase/env.dart';
import '../../../../core/supabase/realtime.dart';
import '../../data/repositories/category_repository_impl.dart';
import '../../domain/entities/category.dart';
import '../../domain/repositories/category_repository.dart';
import '../../domain/usecases/create_category.dart';
import '../../domain/usecases/create_subcategory.dart';
import '../../domain/usecases/delete_category.dart';
import '../../domain/usecases/rename_category.dart';
import '../../domain/usecases/set_hidden.dart';
import 'category_list_controller.dart';

/// Toàn bộ danh mục (gồm cả danh mục ẩn) cho màn hình quản lý. [T038]
final manageCategoriesProvider = FutureProvider<List<Category>>((ref) async {
  final repo = ref.watch(categoryRepositoryProvider);
  return repo.listCategories(includeHidden: true);
});

/// Tập hợp thao tác quản lý danh mục; tự làm tươi danh sách sau mỗi thay đổi.
class CategoryActions {
  final Ref _ref;
  CategoryActions(this._ref);

  CategoryRepository get _repo => _ref.read(categoryRepositoryProvider);

  Future<Category> create({
    required CategoryType type,
    required String name,
    String? icon,
    String? parentId,
    bool allowDuplicate = false,
  }) async {
    final c = await CreateCategory(_repo).call(
      type: type,
      name: name,
      icon: icon,
      parentId: parentId,
      allowDuplicate: allowDuplicate,
    );
    _refresh();
    return c;
  }

  /// Tạo danh mục con dưới [parentId] — loại kế thừa từ cha (US3).
  Future<Category> createSub({
    required String parentId,
    required String name,
    String? icon,
    bool allowDuplicate = false,
  }) async {
    final c = await CreateSubcategory(_repo).call(
      parentId: parentId,
      name: name,
      icon: icon,
      allowDuplicate: allowDuplicate,
    );
    _refresh();
    return c;
  }

  /// [expectedUpdatedAt]: mốc bản ghi UI đang thấy — phát hiện xung đột
  /// khi thành viên khác đã đổi/xóa (R13, T050). Khi xung đột, danh sách
  /// vẫn được làm tươi (finally) để UI hiển thị dữ liệu mới nhất.
  Future<void> rename(String id,
      {String? name, String? icon, DateTime? expectedUpdatedAt}) async {
    try {
      await RenameCategory(_repo).call(id,
          name: name, icon: icon, expectedUpdatedAt: expectedUpdatedAt);
    } finally {
      _refresh();
    }
  }

  Future<void> delete(String id,
      {required DeleteAction action, String? targetId}) async {
    await DeleteCategory(_repo).call(id, action: action, targetId: targetId);
    _refresh();
  }

  Future<void> setHidden(String id,
      {required bool hidden, DateTime? expectedUpdatedAt}) async {
    try {
      await SetHidden(_repo)
          .call(id, hidden: hidden, expectedUpdatedAt: expectedUpdatedAt);
    } finally {
      _refresh();
    }
  }

  void _refresh() {
    _ref.invalidate(manageCategoriesProvider);
    _ref.invalidate(categoriesProvider); // làm tươi cả picker (US1)
  }
}

final categoryActionsProvider = Provider((ref) => CategoryActions(ref));

/// Đồng bộ realtime danh mục giữa các thành viên trong hộ (T049):
/// thành viên khác thêm/sửa/xóa → làm tươi danh sách trong vài giây.
/// Watch provider này ở các màn hình hiển thị danh mục.
final categoryRealtimeSyncProvider = Provider<void>((ref) {
  if (!Env.isConfigured) return; // demo in-memory: không có realtime
  ref.watch(currentHouseholdProvider).whenData((hh) {
    final channel = subscribeHouseholdChanges(hh, () {
      ref.invalidate(manageCategoriesProvider);
      ref.invalidate(categoriesProvider);
    });
    ref.onDispose(() => unsubscribeHouseholdChanges(channel));
  });
});
