import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../domain/receipt_repository.dart';

enum UploadStatus { idle, uploading, success, duplicate, error }

class UploadState {
  const UploadState({
    this.status = UploadStatus.idle,
    this.progress = 0.0,
    this.receiptId,
    this.errorMessage,
  });

  final UploadStatus status;
  final double progress;
  final String? receiptId;
  final String? errorMessage;

  UploadState copyWith({
    UploadStatus? status,
    double? progress,
    String? receiptId,
    String? errorMessage,
  }) {
    return UploadState(
      status: status ?? this.status,
      progress: progress ?? this.progress,
      receiptId: receiptId ?? this.receiptId,
      errorMessage: errorMessage,
    );
  }
}

class UploadNotifier extends StateNotifier<UploadState> {
  UploadNotifier(this._repository) : super(const UploadState());

  final ReceiptRepository _repository;

  Future<void> upload(String imagePath) async {
    state = state.copyWith(status: UploadStatus.uploading, progress: 0.0);
    try {
      final response = await _repository.uploadReceipt(imagePath);
      if (response.duplicate) {
        state = UploadState(
          status: UploadStatus.duplicate,
          receiptId: response.receiptId,
        );
      } else {
        state = UploadState(
          status: UploadStatus.success,
          receiptId: response.receiptId,
        );
      }
    } on Object catch (e) {
      state = UploadState(
        status: UploadStatus.error,
        errorMessage: e.toString(),
      );
    }
  }

  Future<void> uploadBytes(
    List<int> bytes, {
    String filename = 'receipt.jpg',
  }) async {
    state = state.copyWith(status: UploadStatus.uploading, progress: 0.0);
    try {
      final response = await _repository.uploadReceiptBytes(
        bytes,
        filename: filename,
      );
      if (response.duplicate) {
        state = UploadState(
          status: UploadStatus.duplicate,
          receiptId: response.receiptId,
        );
      } else {
        state = UploadState(
          status: UploadStatus.success,
          receiptId: response.receiptId,
        );
      }
    } on Object catch (e) {
      state = UploadState(
        status: UploadStatus.error,
        errorMessage: e.toString(),
      );
    }
  }

  void reset() {
    state = const UploadState();
  }
}

final uploadNotifierProvider =
    StateNotifierProvider<UploadNotifier, UploadState>((ref) {
      return UploadNotifier(ref.watch(receiptRepositoryProvider));
    });
