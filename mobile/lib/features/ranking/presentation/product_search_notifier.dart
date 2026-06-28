import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/ranking_models.dart';
import '../domain/ranking_repository.dart';

enum SearchStatus { idle, loading, results, empty, error }

class SearchState {
  const SearchState({
    this.status = SearchStatus.idle,
    this.results = const [],
    this.lastQuery = '',
    this.errorMessage,
  });

  final SearchStatus status;
  final List<ProductSearchResult> results;
  final String lastQuery;
  final String? errorMessage;
}

class SearchNotifier extends StateNotifier<SearchState> {
  SearchNotifier(this._repository) : super(const SearchState());

  final RankingRepository _repository;

  Future<void> search(String query) async {
    if (query.trim().isEmpty) {
      state = const SearchState(status: SearchStatus.idle);
      return;
    }
    state = SearchState(
      status: SearchStatus.loading,
      lastQuery: query,
      results: state.results,
    );
    try {
      final response = await _repository.search(query);
      state = SearchState(
        status: response.results.isEmpty
            ? SearchStatus.empty
            : SearchStatus.results,
        results: response.results,
        lastQuery: query,
      );
    } on Exception catch (e) {
      state = SearchState(
        status: SearchStatus.error,
        errorMessage: e.toString(),
        lastQuery: query,
      );
    }
  }

  void reset() {
    state = const SearchState();
  }
}

final searchNotifierProvider =
    StateNotifierProvider<SearchNotifier, SearchState>((ref) {
      return SearchNotifier(ref.watch(rankingRepositoryProvider));
    });
