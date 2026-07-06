import '../../domain/entities/category.dart';

/// Ánh xạ giữa hàng Postgres và entity `Category`. [T020]
class CategoryModel {
  const CategoryModel._();

  static Category fromMap(Map<String, dynamic> m) => Category(
        id: m['id'] as String,
        householdId: m['household_id'] as String,
        name: m['name'] as String,
        type: CategoryType.fromDb(m['type'] as String),
        icon: m['icon'] as String?,
        parentId: m['parent_id'] as String?,
        isDefault: (m['is_default'] as bool?) ?? false,
        isHidden: (m['is_hidden'] as bool?) ?? false,
        createdBy: m['created_by'] as String?,
        updatedAt: m['updated_at'] != null
            ? DateTime.parse(m['updated_at'] as String)
            : null,
      );

  /// Map cho INSERT (không gồm `id`, `created_at` — DB tự sinh).
  static Map<String, dynamic> toInsert({
    required String householdId,
    required String name,
    required CategoryType type,
    String? icon,
    String? parentId,
    bool isDefault = false,
    String? createdBy,
  }) =>
      {
        'household_id': householdId,
        'name': name,
        'type': type.db,
        'icon': icon,
        'parent_id': parentId,
        'is_default': isDefault,
        'created_by': createdBy,
      };
}
