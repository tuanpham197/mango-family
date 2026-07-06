import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/category_repository_impl.dart';
import '../../domain/entities/category.dart';

/// Tải danh mục theo loại (null = tất cả), loại trừ danh mục ẩn. [T026]
final categoriesProvider =
    FutureProvider.family<List<Category>, CategoryType?>((ref, type) async {
  final repo = ref.watch(categoryRepositoryProvider);
  return repo.listCategories(type: type);
});
