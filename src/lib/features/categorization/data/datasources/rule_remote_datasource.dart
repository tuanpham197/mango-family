import '../../../../core/supabase/supabase_client.dart';
import '../../domain/entities/categorization_rule.dart';

/// Gọi Supabase cho quy tắc gợi ý (`categorization_rules`) + lịch sử chung
/// của hộ. [T045, T047]
class RuleRemoteDataSource {
  /// Quy tắc của hộ, ưu tiên quy tắc khớp nhiều nhất (FR-015).
  Future<List<CategorizationRule>> listRules(String householdId) async {
    final data = await supabase
        .from('categorization_rules')
        .select()
        .eq('household_id', householdId)
        .order('match_count', ascending: false);
    return (data as List)
        .map((e) => _fromMap(e as Map<String, dynamic>))
        .toList();
  }

  /// Giao dịch gần đây của hộ (mô tả + danh mục) làm lịch sử gợi ý.
  Future<List<({String description, String categoryId})>> recentHistory(
    String householdId,
    String type, {
    int limit = 200,
  }) async {
    final data = await supabase
        .from('transactions')
        .select('description, category_id')
        .eq('household_id', householdId)
        .eq('type', type)
        .not('description', 'is', null)
        .order('transaction_date', ascending: false)
        .limit(limit);
    return [
      for (final row in data as List)
        if ((row['description'] as String?)?.trim().isNotEmpty ?? false)
          (
            description: row['description'] as String,
            categoryId: row['category_id'] as String,
          ),
    ];
  }

  /// Học từ xác nhận của thành viên: ghi nhận / tăng trọng số từ khóa →
  /// danh mục đã chọn (lịch sử chung của hộ). [T047]
  Future<void> learn(
    String householdId,
    String description,
    String categoryId,
  ) async {
    for (final keyword in extractKeywords(description)) {
      final existing = await supabase
          .from('categorization_rules')
          .select('id, match_count')
          .eq('household_id', householdId)
          .eq('keyword', keyword)
          .eq('category_id', categoryId)
          .limit(1);
      if (existing.isNotEmpty) {
        final row = existing.first;
        await supabase
            .from('categorization_rules')
            .update({'match_count': (row['match_count'] as int) + 1})
            .eq('id', row['id'] as String);
      } else {
        await supabase.from('categorization_rules').insert({
          'household_id': householdId,
          'keyword': keyword,
          'category_id': categoryId,
          'match_count': 1,
        });
      }
    }
  }

  /// Chuẩn hóa mô tả thành từ khóa: chữ thường, bỏ ký tự lạ, ≥ 3 ký tự,
  /// tối đa 5 từ (tránh phình bảng quy tắc).
  static List<String> extractKeywords(String description) {
    final words = description
        .toLowerCase()
        .replaceAll(RegExp(r'[^\p{L}\p{N}\s]', unicode: true), ' ')
        .split(RegExp(r'\s+'))
        .where((w) => w.length >= 3)
        .toSet()
        .take(5)
        .toList();
    return words;
  }

  static CategorizationRule _fromMap(Map<String, dynamic> m) =>
      CategorizationRule(
        id: m['id'] as String,
        householdId: m['household_id'] as String,
        keyword: m['keyword'] as String,
        categoryId: m['category_id'] as String,
        matchCount: (m['match_count'] as int?) ?? 0,
      );
}
