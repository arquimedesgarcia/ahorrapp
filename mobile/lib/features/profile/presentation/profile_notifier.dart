import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/profile_models.dart';
import '../domain/profile_repository.dart';

enum ProfileStatus { loading, ready, error }

class ProfileState {
  const ProfileState({
    this.status = ProfileStatus.loading,
    this.totalPoints = 0,
    this.level = 'Bronce',
    this.contributor = const ContributorStats(),
    this.recentTransactions = const [],
    this.errorMessage,
  });

  final ProfileStatus status;
  final int totalPoints;
  final String level;
  final ContributorStats contributor;
  final List<PointsTransaction> recentTransactions;
  final String? errorMessage;
}

class ProfileNotifier extends StateNotifier<ProfileState> {
  ProfileNotifier(this._repository) : super(const ProfileState());

  final ProfileRepository _repository;

  Future<void> loadPoints() async {
    state = const ProfileState(status: ProfileStatus.loading);
    try {
      final response = await _repository.getPoints();
      state = ProfileState(
        status: ProfileStatus.ready,
        totalPoints: response.totalPoints,
        level: response.level,
        contributor: response.contributor,
        recentTransactions: response.recentTransactions,
      );
    } on Exception catch (e) {
      state = ProfileState(
        status: ProfileStatus.error,
        errorMessage: e.toString(),
      );
    }
  }
}

final profileNotifierProvider =
    StateNotifierProvider<ProfileNotifier, ProfileState>((ref) {
      return ProfileNotifier(ref.watch(profileRepositoryProvider));
    });
