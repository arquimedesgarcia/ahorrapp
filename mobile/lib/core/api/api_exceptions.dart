class AppException implements Exception {
  const AppException(this.message, {this.statusCode});

  final String message;
  final int? statusCode;

  @override
  String toString() => message;
}

/// Diagnostic helper that exposes the underlying cause for debugging UI.
String describeError(Object e) {
  if (e is AppException) return e.message;
  return '${e.runtimeType}: $e';
}

class NetworkException extends AppException {
  const NetworkException(super.message);
}

class AuthException extends AppException {
  const AuthException(super.message, {super.statusCode = 401});
}

class ServerException extends AppException {
  const ServerException(super.message, {super.statusCode = 500});
}

class ValidationException extends AppException {
  const ValidationException(super.message, {super.statusCode = 400});

  factory ValidationException.fromJson(Map<String, dynamic> json) {
    return ValidationException(json['error'] as String? ?? 'Validation error');
  }
}

class NotFoundException extends AppException {
  const NotFoundException(super.message, {super.statusCode = 404});
}
