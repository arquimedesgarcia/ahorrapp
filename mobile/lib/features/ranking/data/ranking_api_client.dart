import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exceptions.dart';
import 'ranking_models.dart';

class RankingApiClient {
  RankingApiClient(this._apiClient);

  final ApiClient _apiClient;

  Future<ProductSearchResponse> search(String query) async {
    try {
      final response = await _apiClient.dio.get(
        '/api/v1/ranking/products/search',
        queryParameters: {'q': query},
      );
      return ProductSearchResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      final message = e.response?.data?['error'] as String? ?? 'Search failed';
      if (status == 401) {
        throw AuthException(message, statusCode: status);
      } else if (status != null && status >= 500) {
        throw ServerException(message, statusCode: status);
      }
      throw NetworkException(message);
    }
  }
}

final rankingApiClientProvider = Provider<RankingApiClient>((ref) {
  return RankingApiClient(ref.watch(apiClientProvider));
});
