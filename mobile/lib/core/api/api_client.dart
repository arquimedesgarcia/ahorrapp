import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../storage/secure_storage.dart';
import 'auth_interceptor.dart';
import 'retry_interceptor.dart';

class ApiClient {
  ApiClient({required this.dio, required this.secureStorage});

  final Dio dio;
  final SecureStorage secureStorage;

  /// Base URL of the backend API.
  ///
  /// Override at build time with:
  ///   flutter run --dart-define=API_BASE_URL=http://192.168.x.x:8080
  ///   flutter build apk --dart-define=API_BASE_URL=http://192.168.x.x:8080
  ///
  /// Defaults:
  ///   - Android emulator: http://10.0.2.2:8080 (alias to host's localhost)
  ///   - Other platforms:   http://localhost:8080
  ///
  /// For a physical device on the same Wi-Fi as the dev machine, pass the
  /// machine's LAN IP (e.g. http://192.168.1.50:8080) via --dart-define.
  static String get _defaultBaseUrl {
    const defined = String.fromEnvironment('API_BASE_URL', defaultValue: '');
    if (defined.isNotEmpty) return defined;
    if (Platform.isAndroid) {
      return 'http://10.0.2.2:8080';
    }
    return 'http://localhost:8080';
  }

  static ApiClient create({String? baseUrl, SecureStorage? secureStorage}) {
    final storage = secureStorage ?? SecureStorage();
    final dio = Dio(
      BaseOptions(
        baseUrl: baseUrl ?? _defaultBaseUrl,
        connectTimeout: const Duration(seconds: 10),
        receiveTimeout: const Duration(seconds: 30),
        sendTimeout: const Duration(seconds: 30),
        headers: {'Content-Type': 'application/json'},
      ),
    );
    dio.interceptors.add(AuthInterceptor(storage));
    dio.interceptors.add(RetryInterceptor(dio: dio));
    if (kDebugMode) {
      dio.interceptors.add(
        LogInterceptor(requestBody: true, responseBody: true, error: true),
      );
    }
    return ApiClient(dio: dio, secureStorage: storage);
  }
}

final secureStorageProvider = Provider<SecureStorage>((ref) {
  return SecureStorage();
});

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient.create(secureStorage: ref.watch(secureStorageProvider));
});
