import 'package:path/path.dart' as p;
import 'package:sqflite/sqflite.dart';

import '../../domain/entities/category.dart';

/// Cache đọc danh mục offline (sqflite) — dữ liệu tài chính hộ chỉ lưu cục bộ
/// bản sao để xem khi mất mạng; nguồn sự thật vẫn là Supabase. [T048]
class CategoryLocalCache {
  Database? _db;

  Future<Database> _open() async {
    return _db ??= await openDatabase(
      p.join(await getDatabasesPath(), 'categorization_cache.db'),
      version: 1,
      onCreate: (db, _) => db.execute('''
        CREATE TABLE categories_cache (
          id           TEXT PRIMARY KEY,
          household_id TEXT NOT NULL,
          name         TEXT NOT NULL,
          type         TEXT NOT NULL,
          icon         TEXT,
          parent_id    TEXT,
          is_default   INTEGER NOT NULL DEFAULT 0,
          is_hidden    INTEGER NOT NULL DEFAULT 0,
          created_by   TEXT
        )
      '''),
    );
  }

  /// Ghi đè toàn bộ cache của hộ bằng ảnh chụp mới nhất từ server.
  Future<void> save(String householdId, List<Category> cats) async {
    final db = await _open();
    await db.transaction((txn) async {
      await txn.delete('categories_cache',
          where: 'household_id = ?', whereArgs: [householdId]);
      final batch = txn.batch();
      for (final c in cats) {
        batch.insert('categories_cache', {
          'id': c.id,
          'household_id': c.householdId,
          'name': c.name,
          'type': c.type.db,
          'icon': c.icon,
          'parent_id': c.parentId,
          'is_default': c.isDefault ? 1 : 0,
          'is_hidden': c.isHidden ? 1 : 0,
          'created_by': c.createdBy,
        });
      }
      await batch.commit(noResult: true);
    });
  }

  /// Đọc cache của hộ; trả `null` nếu chưa từng cache (phân biệt với hộ
  /// thật sự chưa có danh mục).
  Future<List<Category>?> read(
    String householdId, {
    String? type,
    bool includeHidden = false,
  }) async {
    final db = await _open();
    final any = Sqflite.firstIntValue(await db.rawQuery(
        'SELECT COUNT(*) FROM categories_cache WHERE household_id = ?',
        [householdId]));
    if (any == null || any == 0) return null;

    final where = StringBuffer('household_id = ?');
    final args = <Object?>[householdId];
    if (type != null) {
      where.write(' AND type = ?');
      args.add(type);
    }
    if (!includeHidden) where.write(' AND is_hidden = 0');
    final rows = await db.query('categories_cache',
        where: where.toString(), whereArgs: args, orderBy: 'name');
    return [
      for (final r in rows)
        Category(
          id: r['id'] as String,
          householdId: r['household_id'] as String,
          name: r['name'] as String,
          type: CategoryType.fromDb(r['type'] as String),
          icon: r['icon'] as String?,
          parentId: r['parent_id'] as String?,
          isDefault: (r['is_default'] as int) == 1,
          isHidden: (r['is_hidden'] as int) == 1,
          createdBy: r['created_by'] as String?,
        ),
    ];
  }
}
