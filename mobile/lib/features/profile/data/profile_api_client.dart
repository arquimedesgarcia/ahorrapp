import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exceptions.dart';
import 'profile_models.dart';

class ProfileApiClient {
  ProfileApiClient(this._apiClient);

  final ApiClient _apiClient;

  Future<PointsResponse> getPoints() async {
    try {
      final response = await _apiClient.dio.get(
        '/api/v1/users/me/points',
        options: Options(
          receiveTimeout: const Duration(seconds: 6),
          sendTimeout: const Duration(seconds: 4),
        ),
      );
      return PointsResponse.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      final message = e.response?.data?['error'] as String? ?? 'Request failed';
      if (status == 401) {
        throw AuthException(message, statusCode: status);
      } else if (status != null && status >= 500) {
        throw ServerException(message, statusCode: status);
      }
      throw NetworkException(message);
    }
  }
}

final profileApiClientProvider = Provider<ProfileApiClient>((ref) {
  return ProfileApiClient(ref.watch(apiClientProvider));
});
