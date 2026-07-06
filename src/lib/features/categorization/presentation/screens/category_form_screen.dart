import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/error/failures.dart';
import '../../domain/entities/category.dart';
import '../controllers/category_manage_controller.dart';

/// Form tạo/sửa danh mục (UC-CAT-02, UC-CAT-04). [T035]
/// - `existing != null` → chế độ sửa (loại bị khóa — FR-005).
/// - `parent != null` → tạo danh mục con, loại kế thừa từ cha (US3). [T041]
/// - `initialType` → tạo nhanh từ luồng nhập giao dịch.
class CategoryFormScreen extends ConsumerStatefulWidget {
  final Category? existing;
  final CategoryType? initialType;
  final Category? parent;

  const CategoryFormScreen({
    super.key,
    this.existing,
    this.initialType,
    this.parent,
  });

  @override
  ConsumerState<CategoryFormScreen> createState() => _CategoryFormScreenState();
}

class _CategoryFormScreenState extends ConsumerState<CategoryFormScreen> {
  late final TextEditingController _name;
  late final TextEditingController _icon;
  CategoryType? _type;
  bool _saving = false;

  bool get _isEdit => widget.existing != null;
  bool get _typeLocked => _isEdit || widget.parent != null;

  @override
  void initState() {
    super.initState();
    _name = TextEditingController(text: widget.existing?.name ?? '');
    _icon = TextEditingController(text: widget.existing?.icon ?? '');
    // Danh mục con kế thừa loại của cha (FR-011).
    _type = widget.existing?.type ?? widget.parent?.type ?? widget.initialType;
  }

  @override
  void dispose() {
    _name.dispose();
    _icon.dispose();
    super.dispose();
  }

  void _snack(String m) {
    if (mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(m)));
    }
  }

  Future<void> _save({bool allowDuplicate = false}) async {
    final name = _name.text.trim();
    final type = _type;
    final icon = _icon.text.trim().isEmpty ? null : _icon.text.trim();
    if (type == null) {
      _snack('Vui lòng chọn loại Thu/Chi'); // FR-004
      return;
    }
    if (name.isEmpty) {
      _snack('Vui lòng nhập tên danh mục');
      return;
    }
    setState(() => _saving = true);
    final actions = ref.read(categoryActionsProvider);
    try {
      if (_isEdit) {
        await actions.rename(
          widget.existing!.id,
          name: name,
          icon: icon,
          // T050: phát hiện thành viên khác đã đổi/xóa trong lúc mình sửa.
          expectedUpdatedAt: widget.existing!.updatedAt,
        );
      } else if (widget.parent != null) {
        await actions.createSub(
          parentId: widget.parent!.id,
          name: name,
          icon: icon,
          allowDuplicate: allowDuplicate,
        );
      } else {
        await actions.create(
          type: type,
          name: name,
          icon: icon,
          allowDuplicate: allowDuplicate,
        );
      }
      if (mounted) Navigator.of(context).pop(true);
    } on DuplicateNameWarning {
      if (!mounted) return;
      final ok = await showDialog<bool>(
        context: context,
        builder: (_) => AlertDialog(
          title: const Text('Tên bị trùng'),
          content: const Text(
              'Đã có danh mục cùng tên trong loại & cấp này. Vẫn tiếp tục?'),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: const Text('Đổi tên')),
            FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: const Text('Tiếp tục')),
          ],
        ),
      );
      if (ok == true) {
        await _save(allowDuplicate: true);
        return;
      }
    } on Failure catch (e) {
      _snack(e.message);
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final title = _isEdit
        ? 'Sửa danh mục'
        : (widget.parent != null ? 'Thêm danh mục con' : 'Thêm danh mục');
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (widget.parent != null)
              Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: Text(
                  'Danh mục con của "${widget.parent!.name}" — '
                  'kế thừa loại ${widget.parent!.type.label}.',
                  style: const TextStyle(color: Colors.grey),
                ),
              ),
            SegmentedButton<CategoryType>(
              segments: const [
                ButtonSegment(value: CategoryType.expense, label: Text('Chi')),
                ButtonSegment(value: CategoryType.income, label: Text('Thu')),
              ],
              selected: _type == null ? <CategoryType>{} : {_type!},
              emptySelectionAllowed: true,
              onSelectionChanged: _typeLocked
                  ? null
                  : (s) => setState(() => _type = s.first),
            ),
            if (_typeLocked)
              const Padding(
                padding: EdgeInsets.only(top: 8),
                child: Text('Loại Thu/Chi cố định sau khi tạo (FR-005).',
                    style: TextStyle(fontSize: 12, color: Colors.grey)),
              ),
            const SizedBox(height: 16),
            TextField(
              controller: _name,
              decoration: const InputDecoration(
                labelText: 'Tên danh mục',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _icon,
              decoration: const InputDecoration(
                labelText: 'Biểu tượng (tùy chọn, ví dụ 🍜)',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: _saving ? null : () => _save(),
              child: _saving
                  ? const SizedBox(
                      height: 18,
                      width: 18,
                      child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Lưu'),
            ),
          ],
        ),
      ),
    );
  }
}
