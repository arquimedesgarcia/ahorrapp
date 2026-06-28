import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/receipt_models.dart';
import '../domain/receipt_repository.dart';

enum ReviewStatus { loading, ready, error }

class ReviewState {
  const ReviewState({
    this.status = ReviewStatus.loading,
    this.detail,
    this.errorMessage,
    this.store,
    this.purchaseDate,
    this.total,
    this.items = const [],
  });

  final ReviewStatus status;
  final ReceiptDetail? detail;
  final String? errorMessage;
  final StoreInfo? store;
  final String? purchaseDate;
  final double? total;
  final List<ReceiptItem> items;

  ReviewState copyWith({
    ReviewStatus? status,
    ReceiptDetail? detail,
    String? errorMessage,
    StoreInfo? store,
    String? purchaseDate,
    double? total,
    List<ReceiptItem>? items,
  }) {
    return ReviewState(
      status: status ?? this.status,
      detail: detail ?? this.detail,
      errorMessage: errorMessage,
      store: store ?? this.store,
      purchaseDate: purchaseDate ?? this.purchaseDate,
      total: total ?? this.total,
      items: items ?? this.items,
    );
  }
}

class ReviewNotifier extends StateNotifier<ReviewState> {
  ReviewNotifier(this._repository) : super(const ReviewState());

  final ReceiptRepository _repository;

  Future<void> loadReceipt(String id) async {
    state = const ReviewState(status: ReviewStatus.loading);
    try {
      final detail = await _repository.pollUntilReviewable(id);
      state = ReviewState(
        status: ReviewStatus.ready,
        detail: detail,
        store: detail.store,
        purchaseDate: detail.purchaseDate,
        total: detail.total,
        items: detail.items,
      );
    } on Exception catch (e) {
      state = ReviewState(
        status: ReviewStatus.error,
        errorMessage: e.toString(),
      );
    }
  }

  void updateStore(StoreInfo store) => state = state.copyWith(store: store);
  void updatePurchaseDate(String date) =>
      state = state.copyWith(purchaseDate: date);
  void updateTotal(double total) => state = state.copyWith(total: total);
  void updateItems(List<ReceiptItem> items) =>
      state = state.copyWith(items: items);
  void addItem() => state = state.copyWith(
    items: [
      ...state.items,
      const ReceiptItem(rawText: '', currency: 'USD'),
    ],
  );
  void removeItem(int index) {
    final items = List<ReceiptItem>.from(state.items);
    if (items.isNotEmpty && index >= 0 && index < items.length) {
      items.removeAt(index);
    }
    state = state.copyWith(items: items);
  }
}

final reviewNotifierProvider =
    StateNotifierProvider<ReviewNotifier, ReviewState>((ref) {
      return ReviewNotifier(ref.watch(receiptRepositoryProvider));
    });
