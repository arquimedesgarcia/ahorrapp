import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/receipt_api_client.dart';
import '../data/receipt_models.dart';

class ReceiptRepository {
  ReceiptRepository(this._apiClient);

  final ReceiptApiClient _apiClient;

  Future<ReceiptUploadResponse> uploadReceipt(String imagePath) {
    return _apiClient.uploadReceipt(imagePath);
  }

  Future<ReceiptUploadResponse> uploadReceiptBytes(
    List<int> bytes, {
    String filename = 'receipt.jpg',
  }) {
    return _apiClient.uploadReceiptBytes(bytes, filename: filename);
  }

  Future<ReceiptDetail> getReceipt(String id) {
    return _apiClient.getReceipt(id);
  }

  Future<ReceiptDetail> pollUntilReviewable(
    String id, {
    Duration interval = const Duration(seconds: 5),
    Duration timeout = const Duration(minutes: 5),
  }) async {
    final deadline = DateTime.now().add(timeout);
    while (DateTime.now().isBefore(deadline)) {
      final receipt = await _apiClient.getReceipt(id);
      if (receipt.status != 'PENDING') return receipt;
      await Future<void>.delayed(interval);
    }
    return _apiClient.getReceipt(id);
  }

  Future<ConfirmReceiptResponse> confirmReceipt(
    String id,
    ConfirmReceiptRequest request,
  ) {
    return _apiClient.confirmReceipt(id, request);
  }

  bool validateCurrency(List<ReceiptItem> items) {
    return items.every(
      (item) => item.currency != null && item.currency!.isNotEmpty,
    );
  }
}

final receiptRepositoryProvider = Provider<ReceiptRepository>((ref) {
  return ReceiptRepository(ref.watch(receiptApiClientProvider));
});
