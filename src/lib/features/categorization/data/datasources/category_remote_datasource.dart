import '../../../../core/supabase/supabase_client.dart';
import '../../domain/entities/category.dart';
import '../models/category_model.dart';

/// Gọi Supabase cho danh mục/giao dịch (gắn `household_id`). [T021]
/// Nhắm tới `supabase_flutter` ^2.5.
class CategoryRemoteDataSource {
  Future<List<Category>> list(
    String householdId, {
    String? type,
    bool includeHidden = false,
  }) async {
    var query = supabase
        .from('categories')
        .select()
        .eq('household_id', householdId);
    if (type != null) query = query.eq('type', type);
    if (!includeHidden) query = query.eq('is_hidden', false);
    final data = await query.order('name');
    return (data as List)
        .map((e) => CategoryModel.fromMap(e as Map<String, dynamic>))
        .toList();
  }

  Future<Category> insert(Map<String, dynamic> values) async {
    final row =
        await supabase.from('categories').insert(values).select().single();
    return CategoryModel.fromMap(row);
  }

  /// Cập nhật có kiểm tra lạc quan (R13, T050): khi truyền [ifUnmodifiedSince],
  /// chỉ ghi nếu `updated_at` chưa đổi. Trả `null` nếu bản ghi đã bị thành
  /// viên khác thay đổi hoặc xóa (0 hàng khớp) — không ghi đè thầm lặng.
  Future<Category?> update(
    String id,
    Map<String, dynamic> values, {
    DateTime? ifUnmodifiedSince,
  }) async {
    var query = supabase.from('categories').update(values).eq('id', id);
    if (ifUnmodifiedSince != null) {
      query = query.eq('updated_at', ifUnmodifiedSince.toIso8601String());
    }
    final rows = await query.select();
    if (rows.isEmpty) return null;
    return CategoryModel.fromMap(rows.first);
  }

  /// Xóa an toàn qua RPC (gán lại/xóa giao dịch, xử lý con). [FR-008]
  Future<void> deleteViaRpc(String id, String action, String? target) async {
    await supabase.rpc('delete_category', params: {
      'p_category': id,
      'p_action': action,
      'p_target': target,
    });
  }

  Future<void> insertTransaction(Map<String, dynamic> values) async {
    await supabase.from('transactions').insert(values);
  }
}
