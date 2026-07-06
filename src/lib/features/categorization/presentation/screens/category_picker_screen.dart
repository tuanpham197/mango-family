import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/category.dart';
import '../controllers/category_list_controller.dart';
import '../controllers/category_manage_controller.dart';

/// Màn hình xem/chọn danh mục, nhóm cha→con và lọc theo loại. [T027]
/// UC-CAT-01. Khi `onSelected` != null → chế độ chọn (dùng khi nhập giao dịch).
class CategoryPickerScreen extends ConsumerWidget {
  final CategoryType? type;
  final ValueChanged<Category>? onSelected;

  const CategoryPickerScreen({super.key, this.type, this.onSelected});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    ref.watch(categoryRealtimeSyncProvider); // T049: đồng bộ giữa thành viên
    final async = ref.watch(categoriesProvider(type));
    final title = type == null ? 'Danh mục' : 'Danh mục ${type!.label}';
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: async.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Lỗi: $e')),
        data: (cats) => cats.isEmpty
            ? const _EmptyState()
            : _CategoryList(cats: cats, onSelected: onSelected),
      ),
    );
  }
}

class _CategoryList extends StatelessWidget {
  final List<Category> cats;
  final ValueChanged<Category>? onSelected;
  const _CategoryList({required this.cats, this.onSelected});

  @override
  Widget build(BuildContext context) {
    final parents = cats.where((c) => c.parentId == null).toList();
    final hasChildren = cats.any((c) => c.parentId != null);
    return ListView(
      children: [
        // T042: gán được vào cha HOẶC con; báo cáo cộng dồn con vào cha (FR-019).
        if (onSelected != null && hasChildren)
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 12, 16, 4),
            child: Text(
              'Có thể chọn danh mục cha hoặc danh mục con. '
              'Báo cáo sẽ cộng dồn danh mục con vào danh mục cha.',
              style: TextStyle(fontSize: 12, color: Colors.grey),
            ),
          ),
        for (final parent in parents) ...[
          ListTile(
            leading: Text(parent.icon ?? '📂',
                style: const TextStyle(fontSize: 20)),
            title: Text(parent.name),
            subtitle: Text(parent.type.label),
            onTap: onSelected == null ? null : () => onSelected!(parent),
          ),
          for (final child in cats.where((c) => c.parentId == parent.id))
            Padding(
              padding: const EdgeInsets.only(left: 24),
              child: ListTile(
                leading: Text(child.icon ?? '↳'),
                title: Text(child.name),
                onTap: onSelected == null ? null : () => onSelected!(child),
              ),
            ),
        ],
      ],
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState();
  @override
  Widget build(BuildContext context) => const Center(
        child: Text('Chưa có danh mục — hãy tạo danh mục mới.'),
      );
}
