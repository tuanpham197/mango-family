/// Quy tắc gợi ý danh mục theo từ khóa + lịch sử chung của hộ (FR-015). [T018]
class CategorizationRule {
  final String id;
  final String householdId;
  final String keyword;
  final String categoryId;
  final int matchCount;

  const CategorizationRule({
    required this.id,
    required this.householdId,
    required this.keyword,
    required this.categoryId,
    this.matchCount = 0,
  });
}
