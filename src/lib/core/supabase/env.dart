/// Cấu hình kết nối Supabase, nạp qua `--dart-define` lúc build/run.
///
/// Cách chạy khuyến nghị (giá trị đặt trong `.env.json` — file đã gitignore):
///   flutter run --dart-define-from-file=.env.json
///
/// KHÔNG hardcode URL/key làm defaultValue: sẽ commit key vào git và làm
/// chế độ demo in-memory (Env chưa cấu hình) + widget test ngừng hoạt động.
class Env {
  const Env._();

  static const String supabaseUrl =
      String.fromEnvironment('SUPABASE_URL', defaultValue: '');
  static const String supabaseAnonKey =
      String.fromEnvironment('SUPABASE_ANON_KEY', defaultValue: '');

  static bool get isConfigured =>
      supabaseUrl.isNotEmpty && supabaseAnonKey.isNotEmpty;
}
