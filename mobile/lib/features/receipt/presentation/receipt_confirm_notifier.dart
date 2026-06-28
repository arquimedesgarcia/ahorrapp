import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/receipt_models.dart';
import '../domain/receipt_repository.dart';

enum ConfirmStatus { idle, confirming, confirmed, error }

class ConfirmState {
  const ConfirmState({
    this.status = ConfirmStatus.idle,
    this.pointsEarned,
    this.errorMessage,
  });

  final ConfirmStatus status;
  final int? pointsEarned;
  final String? errorMessage;
}

class ConfirmNotifier extends StateNotifier<ConfirmState> {
  ConfirmNotifier(this._repository) : super(const ConfirmState());

  final ReceiptRepository _repository;

  Future<void> confirm(String receiptId, ConfirmReceiptRequest request) async {
    state = const ConfirmState(status: ConfirmStatus.confirming);
    try {
      final response = await _repository.confirmReceipt(receiptId, request);
      state = ConfirmState(
        status: ConfirmStatus.confirmed,
        pointsEarned: response.pointsEarned,
      );
    } on Exception catch (e) {
      state = ConfirmState(
        status: ConfirmStatus.error,
        errorMessage: e.toString(),
      );
    }
  }
}

final confirmNotifierProvider =
    StateNotifierProvider<ConfirmNotifier, ConfirmState>((ref) {
      return ConfirmNotifier(ref.watch(receiptRepositoryProvider));
    });
