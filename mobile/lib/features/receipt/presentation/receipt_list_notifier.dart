import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/receipt_api_client.dart';
import '../data/receipt_models.dart';

enum ReceiptListStatus { initial, loading, ready, error }

class ReceiptListState {
  const ReceiptListState({
    this.status = ReceiptListStatus.initial,
    this.items = const [],
    this.errorMessage,
  });

  final ReceiptListStatus status;
  final List<ReceiptListItem> items;
  final String? errorMessage;

  ReceiptListState copyWith({
    ReceiptListStatus? status,
    List<ReceiptListItem>? items,
    String? errorMessage,
  }) {
    return ReceiptListState(
      status: status ?? this.status,
      items: items ?? this.items,
      errorMessage: errorMessage,
    );
  }
}

class ReceiptListNotifier extends StateNotifier<ReceiptListState> {
  ReceiptListNotifier(this._api) : super(const ReceiptListState());

  final ReceiptApiClient _api;

  Future<void> load() async {
    state = state.copyWith(status: ReceiptListStatus.loading);
    try {
      final resp = await _api.listReceipts(limit: 50);
      state = state.copyWith(
        status: ReceiptListStatus.ready,
        items: resp.items,
      );
    } on Exception catch (e) {
      state = state.copyWith(
        status: ReceiptListStatus.error,
        errorMessage: e.toString(),
      );
    }
  }
}

final receiptListProvider =
    StateNotifierProvider<ReceiptListNotifier, ReceiptListState>((ref) {
      return ReceiptListNotifier(ref.watch(receiptApiClientProvider));
    });
