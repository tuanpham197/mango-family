import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/error/failures.dart';
import '../../domain/entities/category.dart';
import '../controllers/category_manage_controller.dart';

/// Nút ẩn / bỏ ẩn một danh mục (UC-CAT-06). [T037]
class HideToggle extends ConsumerWidget {
  final Category category;
  const HideToggle({super.key, required this.category});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return IconButton(
      icon: Icon(category.isHidden ? Icons.visibility_off : Icons.visibility),
      tooltip: category.isHidden ? 'Bỏ ẩn' : 'Ẩn',
      onPressed: () async {
        try {
          await ref.read(categoryActionsProvider).setHidden(
                category.id,
                hidden: !category.isHidden,
                // T050: báo xung đột nếu thành viên khác vừa đổi/xóa.
                expectedUpdatedAt: category.updatedAt,
              );
        } on Failure catch (e) {
          if (context.mounted) {
            ScaffoldMessenger.of(context)
                .showSnackBar(SnackBar(content: Text(e.message)));
          }
        }
      },
    );
  }
}
