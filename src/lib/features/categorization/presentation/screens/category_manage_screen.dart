import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/supabase/env.dart';
import '../../../../core/supabase/supabase_client.dart';
import '../../domain/entities/category.dart';
import '../controllers/category_manage_controller.dart';
import '../widgets/hide_toggle.dart';
import 'category_form_screen.dart';
import 'delete_reassign_screen.dart';

/// Màn hình quản lý danh mục: liệt kê (gồm ẩn), tạo/sửa/xóa/ẩn (US2).
class CategoryManageScreen extends ConsumerWidget {
  const CategoryManageScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    ref.watch(categoryRealtimeSyncProvider); // T049: đồng bộ giữa thành viên
    final async = ref.watch(manageCategoriesProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Quản lý danh mục'),
        actions: [
          IconButton(
            icon: const Icon(Icons.receipt_long),
            tooltip: 'Nhập giao dịch',
            onPressed: () => context.push('/txn'),
          ),
          if (Env.isConfigured)
            IconButton(
              icon: const Icon(Icons.logout),
              tooltip: 'Đăng xuất (đổi thành viên)',
              onPressed: () => supabase.auth.signOut(),
            ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        tooltip: 'Thêm danh mục',
        onPressed: () => Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => const CategoryFormScreen()),
        ),
        child: const Icon(Icons.add),
      ),
      body: async.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Lỗi: $e')),
        data: (cats) {
          final parents = cats.where((c) => c.parentId == null).toList();
          if (parents.isEmpty) {
            return const Center(child: Text('Chưa có danh mục — bấm + để tạo.'));
          }
          return ListView(
            children: [
              for (final parent in parents) ...[
                _CategoryRow(category: parent, cats: cats),
                for (final child in cats.where((c) => c.parentId == parent.id))
                  Padding(
                    padding: const EdgeInsets.only(left: 24),
                    child: _CategoryRow(category: child, cats: cats),
                  ),
              ],
            ],
          );
        },
      ),
    );
  }
}

class _CategoryRow extends StatelessWidget {
  final Category category;
  final List<Category> cats;
  const _CategoryRow({required this.category, required this.cats});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Text(category.icon ?? (category.isSubcategory ? '↳' : '📂'),
          style: const TextStyle(fontSize: 18)),
      title: Text(
        category.name + (category.isHidden ? '  (đã ẩn)' : ''),
        style: TextStyle(
          color: category.isHidden ? Colors.grey : null,
          fontStyle: category.isHidden ? FontStyle.italic : null,
        ),
      ),
      subtitle: Text(category.type.label),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (!category.isSubcategory)
            IconButton(
              icon: const Icon(Icons.subdirectory_arrow_right),
              tooltip: 'Thêm danh mục con',
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute(
                    builder: (_) => CategoryFormScreen(parent: category)),
              ),
            ),
          HideToggle(category: category),
          IconButton(
            icon: const Icon(Icons.edit),
            tooltip: 'Sửa',
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                  builder: (_) => CategoryFormScreen(existing: category)),
            ),
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline),
            tooltip: 'Xóa',
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                  builder: (_) => DeleteReassignScreen(category: category)),
            ),
          ),
        ],
      ),
    );
  }
}
