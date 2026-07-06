/// Lỗi nghiệp vụ của tầng domain/data. [T017]
sealed class Failure implements Exception {
  final String message;
  const Failure(this.message);
  @override
  String toString() => message;
}

class NoHousehold extends Failure {
  const NoHousehold() : super('Tài khoản chưa thuộc hộ gia đình nào.');
}

class CategoryRequired extends Failure {
  const CategoryRequired() : super('Vui lòng chọn một danh mục.');
}

class TypeRequired extends Failure {
  const TypeRequired() : super('Vui lòng chọn loại Thu/Chi.');
}

class TypeMismatch extends Failure {
  const TypeMismatch()
      : super('Danh mục phải cùng loại Thu/Chi với giao dịch.');
}

class TypeImmutable extends Failure {
  const TypeImmutable() : super('Không thể đổi loại danh mục sau khi tạo.');
}

class NestingTooDeep extends Failure {
  const NestingTooDeep() : super('Chỉ hỗ trợ một cấp danh mục con.');
}

class ParentNotFound extends Failure {
  const ParentNotFound() : super('Danh mục cha không tồn tại.');
}

class ReassignTypeMismatch extends Failure {
  const ReassignTypeMismatch() : super('Danh mục đích phải cùng loại Thu/Chi.');
}

class ReassignTargetRequired extends Failure {
  const ReassignTargetRequired()
      : super('Hãy chọn danh mục đích để gán lại, hoặc chọn xóa giao dịch.');
}

class NameRequired extends Failure {
  const NameRequired() : super('Vui lòng nhập tên danh mục.');
}

class DuplicateNameWarning extends Failure {
  const DuplicateNameWarning()
      : super('Tên danh mục bị trùng trong cùng loại và cùng cấp cha.');
}

class HouseholdMismatch extends Failure {
  const HouseholdMismatch()
      : super('Danh mục và giao dịch phải thuộc cùng một hộ.');
}

class ConcurrencyConflict extends Failure {
  const ConcurrencyConflict()
      : super('Danh mục vừa được thành viên khác thay đổi hoặc xóa. '
            'Danh sách đã được tải lại — vui lòng thử lại.');
}

class RemoteFailure extends Failure {
  const RemoteFailure(super.message);
}
