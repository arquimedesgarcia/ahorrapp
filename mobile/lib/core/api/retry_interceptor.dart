import 'dart:async';

import 'package:dio/dio.dart';

/// Retries failed requests with exponential backoff (1s, 2s, 4s).
///
/// Only retries on transient failures: network errors and 5xx responses.
/// Reuses the same [Dio] instance that owns the interceptor chain so the
/// original [BaseOptions.baseUrl], timeouts and other interceptors are
/// preserved across retries.
class RetryInterceptor extends Interceptor {
  RetryInterceptor({required this.dio, this.maxRetries = 3});

  final Dio dio;
  final int maxRetries;

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final retryCount = err.requestOptions.extra['retryCount'] as int? ?? 0;

    if (retryCount >= maxRetries || !_shouldRetry(err)) {
      handler.next(err);
      return;
    }

    final backoff = Duration(seconds: 1 << retryCount);
    await Future<void>.delayed(backoff);

    err.requestOptions.extra['retryCount'] = retryCount + 1;
    try {
      final response = await dio.fetch(err.requestOptions);
      handler.resolve(response);
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  bool _shouldRetry(DioException err) {
    switch (err.type) {
      case DioExceptionType.connectionError:
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.receiveTimeout:
      case DioExceptionType.sendTimeout:
        return true;
      case DioExceptionType.badResponse:
        final status = err.response?.statusCode ?? 0;
        return status >= 500 && status < 600;
      case DioExceptionType.cancel:
      case DioExceptionType.badCertificate:
      case DioExceptionType.unknown:
        return false;
    }
  }
}
