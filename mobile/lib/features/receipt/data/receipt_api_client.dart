import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exceptions.dart';
import 'receipt_models.dart';

class ReceiptApiClient {
  ReceiptApiClient(this._apiClient);

  final ApiClient _apiClient;

  Future<ReceiptUploadResponse> uploadReceipt(String imagePath) async {
    try {
      final normalizedPath = _normalizeImagePath(imagePath);
      final filename = p.basename(normalizedPath);
      final formData = FormData.fromMap({
        'image': await MultipartFile.fromFile(
          normalizedPath,
          filename: filename,
        ),
      });
      final response = await _apiClient.dio.post(
        '/api/v1/receipts',
        data: formData,
        options: Options(contentType: 'multipart/form-data'),
      );
      return ReceiptUploadResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (e) {
      throw _mapException(e);
    } on Object catch (e) {
      throw NetworkException('Unexpected upload error: $e');
    }
  }

  Future<ReceiptUploadResponse> uploadReceiptBytes(
    List<int> bytes, {
    String filename = 'receipt.jpg',
  }) async {
    try {
      final formData = FormData.fromMap({
        'image': MultipartFile.fromBytes(bytes, filename: filename),
      });
      final response = await _apiClient.dio.post(
        '/api/v1/receipts',
        data: formData,
        options: Options(contentType: 'multipart/form-data'),
      );
      return ReceiptUploadResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (e) {
      throw _mapException(e);
    } on Object catch (e) {
      throw NetworkException('Unexpected upload-bytes error: $e');
    }
  }

  Future<ReceiptDetail> getReceipt(String id) async {
    try {
      final response = await _apiClient.dio.get('/api/v1/receipts/$id');
      return ReceiptDetail.fromJson(response.data as Map<String, dynamic>);
    } on DioException catch (e) {
      throw _mapException(e);
    }
  }

  Future<ReceiptListResponse> listReceipts({
    int limit = 50,
    int offset = 0,
  }) async {
    try {
      final response = await _apiClient.dio.get(
        '/api/v1/receipts',
        queryParameters: {'limit': limit, 'offset': offset},
        options: Options(
          receiveTimeout: const Duration(seconds: 8),
          sendTimeout: const Duration(seconds: 8),
        ),
      );
      return ReceiptListResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (e) {
      throw _mapException(e);
    }
  }

  Future<ConfirmReceiptResponse> confirmReceipt(
    String id,
    ConfirmReceiptRequest request,
  ) async {
    try {
      final response = await _apiClient.dio.post(
        '/api/v1/receipts/$id/confirm',
        data: request.toJson(),
      );
      if (response.statusCode == 204) {
        return const ConfirmReceiptResponse(pointsEarned: 0);
      }
      return ConfirmReceiptResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (e) {
      throw _mapException(e);
    }
  }

  Exception _mapException(DioException e) {
    final status = e.response?.statusCode;
    final data = e.response?.data;
    String message = 'Request failed';
    if (data is Map<String, dynamic>) {
      message = data['error'] as String? ?? message;
    } else if (data is String && data.trim().isNotEmpty) {
      message = data.trim();
    } else if (e.message != null && e.message!.trim().isNotEmpty) {
      message = e.message!.trim();
    }
    if (status == 401) {
      return AuthException(message, statusCode: status);
    } else if (status == 400) {
      return ValidationException(message);
    } else if (status == 404) {
      return NotFoundException(message, statusCode: status);
    } else if (status != null && status >= 500) {
      return ServerException(message, statusCode: status);
    }
    return NetworkException(message);
  }

  String _normalizeImagePath(String rawPath) {
    var path = rawPath.trim();
    // Defensive normalization for paths copied from PowerShell command format:
    // & 'd:\\path with spaces\\image.jpg'
    if (path.startsWith("& '") && path.endsWith("'")) {
      path = path.substring(3, path.length - 1);
    } else if (path.startsWith("'") && path.endsWith("'")) {
      path = path.substring(1, path.length - 1);
    } else if (path.startsWith('"') && path.endsWith('"')) {
      path = path.substring(1, path.length - 1);
    }
    return path;
  }
}

final receiptApiClientProvider = Provider<ReceiptApiClient>((ref) {
  return ReceiptApiClient(ref.watch(apiClientProvider));
});
