import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/error/failures.dart';
import '../../domain/entities/category.dart';
import '../../domain/repositories/category_repository.dart';
import '../controllers/category_manage_controller.dart';

/// Xóa danh mục: gán lại (cùng loại) hoặc xóa giao dịch (UC-CAT-05). [T036]
class DeleteReassignScreen extends ConsumerStatefulWidget {
  final Category category;
  const DeleteReassignScreen({super.key, required this.category});

  @override
  ConsumerState<DeleteReassignScreen> createState() =>
      _DeleteReassignScreenState();
}

class _DeleteReassignScreenState extends ConsumerState<DeleteReassignScreen> {
  DeleteAction _action = DeleteAction.reassign;
  String? _targetId;
  bool _busy = false;

  Future<void> _confirm() async {
    setState(() => _busy = true);
    try {
      await ref.read(categoryActionsProvider).delete(
            widget.category.id,
            action: _action,
            targetId: _action == DeleteAction.reassign ? _targetId : null,
          );
      if (mounted) Navigator.of(context).pop(true);
    } on Failure catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(e.message)));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final cat = widget.category;
    final async = ref.watch(manageCategoriesProvider);
    return Scaffold(
      appBar: AppBar(title: Text('Xóa "${cat.name}"')),
      body: async.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Lỗi: $e')),
        data: (all) {
          // Danh mục đích hợp lệ: cùng loại, khác chính nó và các con của nó
          // (con sẽ bị xóa cùng cha qua RPC `delete_category` — T043, FR-012).
          final targets = all
              .where((c) =>
                  c.type == cat.type &&
                  c.id != cat.id &&
                  c.parentId != cat.id)
              .toList();
          final children = all.where((c) => c.parentId == cat.id).toList();
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              const Text(
                  'Danh mục có thể đang chứa giao dịch. Chọn cách xử lý trước khi xóa:'),
              if (children.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Text(
                    'Sẽ xóa cả ${children.length} danh mục con: '
                    '${children.map((c) => c.name).join(', ')}. '
                    'Giao dịch của các danh mục con cũng được gán lại hoặc xóa theo lựa chọn dưới đây.',
                    style: TextStyle(
                        color: Theme.of(context).colorScheme.error),
                  ),
                ),
              const SizedBox(height: 8),
              RadioGroup<DeleteAction>(
                groupValue: _action,
                onChanged: (v) => setState(() => _action = v!),
                child: Column(
                  children: [
                    const RadioListTile<DeleteAction>(
                      value: DeleteAction.reassign,
                      title: Text(
                          'Gán lại giao dịch sang danh mục khác (cùng loại)'),
                    ),
                    if (_action == DeleteAction.reassign)
                      Padding(
                        padding: const EdgeInsets.only(left: 16, bottom: 8),
                        child: DropdownButtonFormField<String>(
                          initialValue: _targetId,
                          decoration: const InputDecoration(
                            labelText: 'Danh mục đích',
                            border: OutlineInputBorder(),
                          ),
                          items: [
                            for (final t in targets)
                              DropdownMenuItem(
                                  value: t.id, child: Text(t.name)),
                          ],
                          onChanged: (v) => setState(() => _targetId = v),
                        ),
                      ),
                    const RadioListTile<DeleteAction>(
                      value: DeleteAction.delete,
                      title: Text('Xóa luôn các giao dịch của danh mục'),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _busy ? null : _confirm,
                style: FilledButton.styleFrom(backgroundColor: Colors.red),
                child: _busy
                    ? const SizedBox(
                        height: 18,
                        width: 18,
                        child: CircularProgressIndicator(strokeWidth: 2))
                    : const Text('Xóa danh mục'),
              ),
            ],
          );
        },
      ),
    );
  }
}
