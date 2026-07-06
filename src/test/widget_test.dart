import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:specs_app/features/categorization/presentation/screens/transaction_entry_screen.dart';
import 'package:specs_app/main.dart';

void main() {
  testWidgets('App khởi động và hiển thị màn hình quản lý danh mục',
      (WidgetTester tester) async {
    // Không cấu hình Supabase trong test → repo demo in-memory được dùng.
    await tester.pumpWidget(const ProviderScope(child: SpecsApp()));
    await tester.pumpAndSettle();

    expect(find.text('Quản lý danh mục'), findsOneWidget);
    // Danh mục mặc định từ seed demo (research R9).
    expect(find.textContaining('Ăn uống'), findsWidgets);
  });

  testWidgets('Quickstart #6 — chặn lưu giao dịch khi chưa chọn danh mục',
      (WidgetTester tester) async {
    await tester.pumpWidget(const ProviderScope(
        child: MaterialApp(home: TransactionEntryScreen())));
    await tester.pumpAndSettle();

    await tester.enterText(find.widgetWithText(TextField, 'Số tiền'), '50000');
    await tester.tap(find.text('Lưu giao dịch'));
    await tester.pump();

    expect(find.text('Vui lòng chọn một danh mục'), findsOneWidget); // FR-013
  });
}
