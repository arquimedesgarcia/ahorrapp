import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/ranking_api_client.dart';
import '../data/ranking_models.dart';

class RankingRepository {
  RankingRepository(this._apiClient);

  final RankingApiClient _apiClient;

  Future<ProductSearchResponse> search(String query) {
    return _apiClient.search(query);
  }
}

final rankingRepositoryProvider = Provider<RankingRepository>((ref) {
  return RankingRepository(ref.watch(rankingApiClientProvider));
});
