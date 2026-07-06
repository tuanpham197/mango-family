import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:go_router/go_router.dart';

import '../../features/categorization/presentation/screens/category_manage_screen.dart';
import '../../features/categorization/presentation/screens/transaction_entry_screen.dart';
import '../auth/login_screen.dart';
import '../supabase/env.dart';
import '../supabase/supabase_client.dart';

/// Điều hướng ứng dụng. [T016]
/// Home = quản lý danh mục (US2); `/txn` = nhập giao dịch (US1/US4);
/// `/login` = đăng nhập dev. Khi CHƯA cấu hình Supabase (demo in-memory)
/// thì không yêu cầu đăng nhập.
final appRouter = GoRouter(
  initialLocation: '/',
  refreshListenable:
      Env.isConfigured ? _AuthRefresh(supabase.auth.onAuthStateChange) : null,
  redirect: (context, state) {
    if (!Env.isConfigured) return null; // demo: không có auth
    final loggedIn = supabase.auth.currentUser != null;
    final atLogin = state.matchedLocation == '/login';
    if (!loggedIn && !atLogin) return '/login';
    if (loggedIn && atLogin) return '/';
    return null;
  },
  routes: [
    GoRoute(
      path: '/',
      builder: (context, state) => const CategoryManageScreen(),
    ),
    GoRoute(
      path: '/txn',
      builder: (context, state) => const TransactionEntryScreen(),
    ),
    GoRoute(
      path: '/login',
      builder: (context, state) => const LoginScreen(),
    ),
  ],
);

/// Ép router tính lại `redirect` mỗi khi trạng thái đăng nhập đổi.
class _AuthRefresh extends ChangeNotifier {
  _AuthRefresh(Stream<dynamic> stream) {
    _sub = stream.listen((_) => notifyListeners());
  }

  late final StreamSubscription<dynamic> _sub;

  @override
  void dispose() {
    _sub.cancel();
    super.dispose();
  }
}
