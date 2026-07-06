import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/error/failures.dart';
import '../../data/repositories/category_repository_impl.dart';
import '../../domain/entities/category.dart';
import '../../domain/repositories/category_repository.dart';
import '../../domain/usecases/assign_category_to_transaction.dart';
import '../widgets/category_select_field.dart';

/// Màn nhập giao dịch tối thiểu — đủ để kiểm chứng UC-CAT-07/08:
/// lọc danh mục theo loại, chặn lưu khi thiếu danh mục (FR-013), gợi ý theo
/// mô tả (FR-015), ghi `created_by` qua repo. Luồng giao dịch đầy đủ thuộc
/// BR-002 (feature riêng).
class TransactionEntryScreen extends ConsumerStatefulWidget {
  const TransactionEntryScreen({super.key});

  @override
  ConsumerState<TransactionEntryScreen> createState() =>
      _TransactionEntryScreenState();
}

class _TransactionEntryScreenState
    extends ConsumerState<TransactionEntryScreen> {
  final _amount = TextEditingController();
  final _desc = TextEditingController();
  CategoryType _type = CategoryType.expense;
  Category? _selected;
  bool _showError = false;
  bool _busy = false;

  @override
  void dispose() {
    _amount.dispose();
    _desc.dispose();
    super.dispose();
  }

  void _snack(String m) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(m)));
    }
  }

  Future<void> _save() async {
    final amount = double.tryParse(_amount.text.replaceAll(',', '.'));
    if (amount == null || amount <= 0) {
      _snack('Vui lòng nhập số tiền hợp lệ (> 0).');
      return;
    }
    // T029: danh mục khác loại hiện tại coi như chưa chọn.
    final selected =
        (_selected != null && _selected!.type == _type) ? _selected : null;
    if (selected == null) {
      setState(() => _showError = true); // FR-013: chặn lưu
      return;
    }
    setState(() => _busy = true);
    final repo = ref.read(categoryRepositoryProvider);
    try {
      await AssignCategoryToTransaction(repo).call(
        TransactionDraft(
          amount: amount,
          type: _type,
          description: _desc.text.trim().isEmpty ? null : _desc.text.trim(),
          date: DateTime.now(),
        ),
        selected.id,
      );
      _snack('Đã lưu giao dịch "${selected.name}".');
      setState(() {
        _amount.clear();
        _desc.clear();
        _selected = null;
        _showError = false;
      });
    } on Failure catch (e) {
      _snack(e.message);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Nhập giao dịch')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          SegmentedButton<CategoryType>(
            segments: const [
              ButtonSegment(value: CategoryType.expense, label: Text('Chi')),
              ButtonSegment(value: CategoryType.income, label: Text('Thu')),
            ],
            selected: {_type},
            // BR-002/T029: đổi loại → danh sách danh mục làm tươi theo loại.
            onSelectionChanged: (s) => setState(() => _type = s.first),
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _amount,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: const InputDecoration(
              labelText: 'Số tiền',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _desc,
            decoration: const InputDecoration(
              labelText: 'Mô tả (tùy chọn — dùng cho gợi ý)',
              border: OutlineInputBorder(),
            ),
            onChanged: (_) => setState(() {}), // cập nhật gợi ý theo mô tả
          ),
          const SizedBox(height: 16),
          CategorySelectField(
            type: _type,
            selected: _selected,
            showError: _showError,
            description: _desc.text,
            onChanged: (c) => setState(() {
              _selected = c;
              _showError = false;
            }),
          ),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: _busy ? null : _save,
            child: _busy
                ? const SizedBox(
                    height: 18,
                    width: 18,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('Lưu giao dịch'),
          ),
        ],
      ),
    );
  }
}
