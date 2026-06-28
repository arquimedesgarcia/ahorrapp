import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/secure_storage.dart';
import '../data/auth_api_client.dart';
import '../data/auth_models.dart';

class AuthRepository {
  AuthRepository(this._apiClient, this._secureStorage);

  final AuthApiClient _apiClient;
  final SecureStorage _secureStorage;

  Future<AuthResponse> register(RegisterRequest request) async {
    final response = await _apiClient.register(request);
    await _secureStorage.writeToken(response.token);
    return response;
  }

  Future<AuthResponse> login(LoginRequest request) async {
    final response = await _apiClient.login(request);
    await _secureStorage.writeToken(response.token);
    return response;
  }

  Future<UserProfile?> restoreSession() async {
    final token = await _secureStorage.readToken();
    if (token == null || token.isEmpty) return null;
    return _apiClient.getMe();
  }

  Future<void> logout() async {
    await _secureStorage.deleteToken();
  }

  Future<bool> isAuthenticated() async {
    final token = await _secureStorage.readToken();
    return token != null && token.isNotEmpty;
  }
}

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(
    ref.watch(authApiClientProvider),
    ref.watch(secureStorageProvider),
  );
});
