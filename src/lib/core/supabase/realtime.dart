import 'package:supabase_flutter/supabase_flutter.dart';

import 'supabase_client.dart';

/// Đồng bộ gần thời gian thực giữa các thành viên trong hộ (plan, R13). [T049]
///
/// Đăng ký nhận thay đổi Postgres trên các bảng dùng chung của một hộ;
/// gọi [onChange] mỗi khi thành viên khác thêm/sửa/xóa bản ghi.
/// Người gọi chịu trách nhiệm hủy kênh qua [unsubscribeHouseholdChanges].
RealtimeChannel subscribeHouseholdChanges(
  String householdId,
  void Function() onChange,
) {
  var channel = supabase.channel('household-$householdId');
  for (final table in const [
    'categories',
    'transactions',
    'categorization_rules',
  ]) {
    channel = channel.onPostgresChanges(
      event: PostgresChangeEvent.all,
      schema: 'public',
      table: table,
      filter: PostgresChangeFilter(
        type: PostgresChangeFilterType.eq,
        column: 'household_id',
        value: householdId,
      ),
      callback: (_) => onChange(),
    );
  }
  return channel.subscribe();
}

Future<void> unsubscribeHouseholdChanges(RealtimeChannel channel) =>
    supabase.removeChannel(channel);
