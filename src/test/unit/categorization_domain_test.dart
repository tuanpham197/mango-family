import 'package:flutter_test/flutter_test.dart';

import 'package:specs_app/core/error/failures.dart';
import 'package:specs_app/features/categorization/data/repositories/in_memory_category_repository.dart';
import 'package:specs_app/features/categorization/domain/entities/category.dart';
import 'package:specs_app/features/categorization/domain/repositories/category_repository.dart';
import 'package:specs_app/features/categorization/domain/usecases/create_category.dart';
import 'package:specs_app/features/categorization/domain/usecases/create_subcategory.dart';
import 'package:specs_app/features/categorization/domain/usecases/suggest_category.dart';

/// Kiểm chứng mức logic các kịch bản quickstart tự động hóa được
/// (#2, #3, #5, #7, #10, #12) trên repo in-memory. E2E đầy đủ (đa thành
/// viên #13–#17) cần Supabase — xem quickstart.md.
void main() {
  late InMemoryCategoryRepository repo;

  setUp(() => repo = InMemoryCategoryRepository());

  Future<Category> anUong() async =>
      (await repo.listCategories()).firstWhere((c) => c.name == 'Ăn uống');

  group('Quickstart #2/#3 — tạo danh mục', () {
    test('tên rỗng bị chặn (NameRequired)', () {
      expect(
        () => CreateCategory(repo)
            .call(type: CategoryType.expense, name: '   '),
        throwsA(isA<NameRequired>()),
      );
    });

    test('trùng tên cùng loại & cấp → cảnh báo, xác nhận thì vẫn tạo được',
        () async {
      expect(
        () => CreateCategory(repo)
            .call(type: CategoryType.expense, name: 'ăn uống'),
        throwsA(isA<DuplicateNameWarning>()), // FR-017
      );
      final c = await CreateCategory(repo).call(
          type: CategoryType.expense, name: 'ăn uống', allowDuplicate: true);
      expect(c.name, 'ăn uống');
    });
  });

  group('Quickstart #5 — lọc theo loại (FR-014)', () {
    test('chỉ trả danh mục cùng loại', () async {
      final chi = await repo.listCategories(type: CategoryType.expense);
      expect(chi, isNotEmpty);
      expect(chi.every((c) => c.type == CategoryType.expense), isTrue);
      expect(chi.any((c) => c.name == 'Lương'), isFalse);
    });

    test('gán danh mục khác loại bị chặn (TypeMismatch)', () async {
      final luong =
          (await repo.listCategories(type: CategoryType.income)).first;
      final draft = TransactionDraft(
        amount: 100,
        type: CategoryType.expense,
        date: DateTime(2026, 7, 1),
      );
      expect(() => repo.assignCategoryToTransaction(draft, luong.id),
          throwsA(isA<TypeMismatch>())); // SC-003
    });
  });

  group('Quickstart #7 — danh mục con một cấp (FR-010, FR-011)', () {
    test('con kế thừa loại của cha', () async {
      final parent = await anUong();
      final sub = await CreateSubcategory(repo)
          .call(parentId: parent.id, name: 'Cà phê');
      expect(sub.type, parent.type); // SC-008
      expect(sub.parentId, parent.id);
    });

    test('tạo cấp con thứ hai bị từ chối (NestingTooDeep)', () async {
      final parent = await anUong();
      final sub = await CreateSubcategory(repo)
          .call(parentId: parent.id, name: 'Cà phê');
      expect(
        () => CreateSubcategory(repo).call(parentId: sub.id, name: 'Espresso'),
        throwsA(isA<NestingTooDeep>()),
      );
    });
  });

  group('Quickstart #10 — ẩn / bỏ ẩn (FR-020)', () {
    test('ẩn: không hiện khi chọn mới; bỏ ẩn: hiện lại; không mất dữ liệu',
        () async {
      final cat = await anUong();
      await repo.setHidden(cat.id, true);
      var visible = await repo.listCategories(type: CategoryType.expense);
      expect(visible.any((c) => c.id == cat.id), isFalse);
      // Vẫn còn trong danh sách quản lý (gồm ẩn) — giữ lịch sử/báo cáo.
      final all = await repo.listCategories(
          type: CategoryType.expense, includeHidden: true);
      expect(all.any((c) => c.id == cat.id), isTrue);

      await repo.setHidden(cat.id, false);
      visible = await repo.listCategories(type: CategoryType.expense);
      expect(visible.any((c) => c.id == cat.id), isTrue);
    });
  });

  group('Quickstart #12 — gợi ý & học từ xác nhận (FR-015)', () {
    test('học từ xác nhận rồi gợi ý cùng loại; ghi đè được', () async {
      final diChuyen = (await repo.listCategories(type: CategoryType.expense))
          .firstWhere((c) => c.name == 'Di chuyển');
      // Thành viên xác nhận "Grab đi làm" → Di chuyển (học).
      await repo.assignCategoryToTransaction(
        TransactionDraft(
          amount: 50,
          type: CategoryType.expense,
          description: 'Grab đi làm',
          date: DateTime(2026, 7, 1),
        ),
        diChuyen.id,
      );
      final suggested = await SuggestCategory(repo).call(
          description: 'grab về nhà', type: CategoryType.expense);
      expect(suggested?.id, diChuyen.id);

      // Gợi ý chỉ cùng loại: mô tả đó với giao dịch Thu → không trả Di chuyển.
      final incomeSuggest = await SuggestCategory(repo)
          .call(description: 'grab về nhà', type: CategoryType.income);
      expect(incomeSuggest?.id, isNot(diChuyen.id));

      // Ghi đè: người dùng chọn danh mục khác → lựa chọn được học tiếp.
      final anUongCat = await anUong();
      await repo.assignCategoryToTransaction(
        TransactionDraft(
          amount: 80,
          type: CategoryType.expense,
          description: 'grab food trưa',
          date: DateTime(2026, 7, 2),
        ),
        anUongCat.id,
      );
      // Không ném lỗi — lựa chọn người dùng luôn thắng (SC-004).
    });
  });
}
