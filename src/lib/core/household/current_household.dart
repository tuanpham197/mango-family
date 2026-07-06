import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../error/failures.dart';
import '../supabase/supabase_client.dart';

/// Giải quyết `household_id` của thành viên đang đăng nhập. [T015]
///
/// `users` là bảng độc lập (không tham chiếu auth) — phiên đăng nhập được nối
/// với hồ sơ người dùng qua EMAIL, rồi lấy membership theo `users.id`
/// (feature 002 · FR-015). MVP: mỗi người dùng thuộc một hộ.
final currentHouseholdProvider = FutureProvider<String>((ref) async {
  final user = supabase.auth.currentUser;
  final email = user?.email;
  if (user == null || email == null) throw const NoHousehold();
  final rows = await supabase
      .from('users')
      .select('household_members(household_id)')
      .eq('email', email)
      .limit(1);
  if (rows.isEmpty) throw const NoHousehold();
  final members = rows.first['household_members'] as List?;
  if (members == null || members.isEmpty) throw const NoHousehold();
  return (members.first as Map)['household_id'] as String;
});
