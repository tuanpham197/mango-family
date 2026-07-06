import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../error/failures.dart';
import '../supabase/supabase_client.dart';

/// Giải quyết `household_id` của thành viên đang đăng nhập. [T015]
///
/// MVP: mỗi người dùng thuộc một hộ → lấy bản ghi membership đầu tiên.
final currentHouseholdProvider = FutureProvider<String>((ref) async {
  final user = supabase.auth.currentUser;
  if (user == null) throw const NoHousehold();
  final rows = await supabase
      .from('household_members')
      .select('household_id')
      .eq('user_id', user.id)
      .limit(1);
  if (rows.isEmpty) throw const NoHousehold();
  return rows.first['household_id'] as String;
});
