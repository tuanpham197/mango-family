/// Loại danh mục — đúng một trong hai, cố định sau khi tạo (FR-004, FR-005).
enum CategoryType {
  income,
  expense;

  String get db => this == CategoryType.income ? 'INCOME' : 'EXPENSE';
  String get label => this == CategoryType.income ? 'Thu' : 'Chi';

  static CategoryType fromDb(String value) =>
      value == 'INCOME' ? CategoryType.income : CategoryType.expense;
}

/// Danh mục dùng chung trong hộ; có thể lồng một cấp dưới một danh mục cha. [T018]
class Category {
  final String id;
  final String householdId;
  final String name;
  final CategoryType type;
  final String? icon;
  final String? parentId;
  final bool isDefault;
  final bool isHidden;
  final String? createdBy;

  /// Mốc sửa đổi cuối — dùng cho cập nhật lạc quan giữa các thành viên
  /// (R13, T050). `null` với dữ liệu cũ/demo chưa có mốc.
  final DateTime? updatedAt;

  const Category({
    required this.id,
    required this.householdId,
    required this.name,
    required this.type,
    this.icon,
    this.parentId,
    this.isDefault = false,
    this.isHidden = false,
    this.createdBy,
    this.updatedAt,
  });

  bool get isSubcategory => parentId != null;

  Category copyWith({String? name, String? icon, bool? isHidden}) => Category(
        id: id,
        householdId: householdId,
        name: name ?? this.name,
        type: type, // bất biến (FR-005)
        icon: icon ?? this.icon,
        parentId: parentId,
        isDefault: isDefault,
        isHidden: isHidden ?? this.isHidden,
        createdBy: createdBy,
        updatedAt: updatedAt,
      );
}
