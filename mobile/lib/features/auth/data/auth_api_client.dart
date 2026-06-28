import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exceptions.dart';
import 'auth_models.dart';

class AuthApiClient {
  AuthApiClient(this._apiClient);

  final ApiClient _apiClient;

  Future<AuthResponse> register(RegisterRequest request) async {
    try {
      final response = await _apiClient.dio.post(
        '/api/v1/auth/register',
        data: request.toJson(),
      );
      return AuthResponse.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _mapException(e);
    }
  }

  Future<AuthResponse> login(LoginRequest request) async {
    try {
      final response = await _apiClient.dio.post(
        '/api/v1/auth/login',
        data: request.toJson(),
      );
      return AuthResponse.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _mapException(e);
    }
  }

  Future<UserProfile> getMe() async {
    try {
      final response = await _apiClient.dio.get('/api/v1/auth/me');
      return UserProfile.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _mapException(e);
    }
  }

  /// Sends a GET to /health on the configured base URL. Used by the diagnostic
  /// screen on the login page so we can tell connectivity failures from auth
  /// failures without needing logcat.
  Future<String> pingHealth() async {
    try {
      final response = await _apiClient.dio.get(
        '/api/v1/health',
        options: Options(extra: {'retryCount': 99}),
      );
      return 'OK ${response.statusCode}';
    } on DioException catch (e) {
      throw _mapException(e);
    }
  }

  Exception _mapException(DioException e) {
    final status = e.response?.statusCode;
    final serverError = e.response?.data?['error'] as String?;
    // Build a verbose description so the user can see right on the screen
    // what kind of failure we got (network? HTTP status? bad URL?).
    final base =
        serverError ??
        '${e.type.name}: ${e.message ?? 'no message'} '
            '-> ${e.requestOptions.baseUrl}${e.requestOptions.path}';
    if (status == 401) {
      return AuthException('Unauthorized: $base', statusCode: status);
    } else if (status == 404) {
      return NotFoundException('Not Found: $base', statusCode: status);
    } else if (status == 409) {
      return ValidationException('Conflict: $base', statusCode: status);
    } else if (status != null && status >= 500) {
      return ServerException('Server ($status): $base', statusCode: status);
    }
    // No HTTP response: real network problem (DNS, connection refused,
    // cleartext blocked, timeout, etc.).
    return NetworkException(
      'Network: ${e.type.name} -> '
      '${e.requestOptions.baseUrl}${e.requestOptions.path}'
      '${e.message == null ? '' : ' - ${e.message}'}',
    );
  }
}

final authApiClientProvider = Provider<AuthApiClient>((ref) {
  return AuthApiClient(ref.watch(apiClientProvider));
});
