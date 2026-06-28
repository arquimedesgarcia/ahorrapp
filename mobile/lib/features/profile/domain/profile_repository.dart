import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/profile_api_client.dart';
import '../data/profile_models.dart';

class ProfileRepository {
  ProfileRepository(this._apiClient);

  final ProfileApiClient _apiClient;

  Future<PointsResponse> getPoints() {
    return _apiClient.getPoints();
  }
}

final profileRepositoryProvider = Provider<ProfileRepository>((ref) {
  return ProfileRepository(ref.watch(profileApiClientProvider));
});
