import 'package:supabase_flutter/supabase_flutter.dart';

import 'env.dart';

/// Khởi tạo Supabase (gọi một lần ở `main`). [T014]
Future<void> initSupabase() async {
  await Supabase.initialize(
    url: Env.supabaseUrl,
    // Giá trị từ SUPABASE_ANON_KEY (publishable key) — anonKey đã deprecated.
    publishableKey: Env.supabaseAnonKey,
  );
}

/// Client dùng chung cho tầng data.
SupabaseClient get supabase => Supabase.instance.client;
