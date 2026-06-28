class ReceiptUploadResponse {
  const ReceiptUploadResponse({
    required this.receiptId,
    required this.status,
    required this.duplicate,
  });

  final String receiptId;
  final String status;
  final bool duplicate;

  factory ReceiptUploadResponse.fromJson(Map<String, dynamic> json) {
    return ReceiptUploadResponse(
      receiptId: json['receipt_id'] as String,
      status: json['status'] as String,
      duplicate: json['duplicate'] as bool? ?? false,
    );
  }
}

class StoreInfo {
  const StoreInfo({this.name, this.branch, this.address});

  final String? name;
  final String? branch;
  final String? address;

  factory StoreInfo.fromJson(Map<String, dynamic> json) {
    return StoreInfo(
      name: json['name'] as String?,
      branch: json['branch'] as String?,
      address: json['address'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
    'name': name,
    'branch': branch,
    'address': address,
  };
}

class ReceiptItem {
  const ReceiptItem({
    required this.rawText,
    this.quantity,
    this.unitPrice,
    this.currency,
  });

  final String rawText;
  final int? quantity;
  final double? unitPrice;
  final String? currency;

  factory ReceiptItem.fromJson(Map<String, dynamic> json) {
    return ReceiptItem(
      rawText: json['raw_text'] as String? ?? '',
      quantity: json['quantity'] as int?,
      unitPrice: (json['unit_price'] as num?)?.toDouble(),
      currency: json['currency'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
    'raw_text': rawText,
    'quantity': quantity,
    'unit_price': unitPrice,
    'currency': currency,
  };

  ReceiptItem copyWith({
    String? rawText,
    int? quantity,
    double? unitPrice,
    String? currency,
  }) {
    return ReceiptItem(
      rawText: rawText ?? this.rawText,
      quantity: quantity ?? this.quantity,
      unitPrice: unitPrice ?? this.unitPrice,
      currency: currency ?? this.currency,
    );
  }
}

class ReceiptDetail {
  const ReceiptDetail({
    required this.receiptId,
    required this.status,
    this.store,
    this.purchaseDate,
    this.total,
    this.items = const [],
  });

  final String receiptId;
  final String status;
  final StoreInfo? store;
  final String? purchaseDate;
  final double? total;
  final List<ReceiptItem> items;

  factory ReceiptDetail.fromJson(Map<String, dynamic> json) {
    return ReceiptDetail(
      receiptId: json['receipt_id'] as String,
      status: json['status'] as String,
      store: json['store'] != null
          ? StoreInfo.fromJson(json['store'] as Map<String, dynamic>)
          : null,
      purchaseDate: json['purchase_date'] as String?,
      total: (json['total'] as num?)?.toDouble(),
      items:
          (json['items'] as List<dynamic>?)
              ?.map((e) => ReceiptItem.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class ConfirmReceiptRequest {
  const ConfirmReceiptRequest({
    required this.store,
    required this.purchaseDate,
    required this.total,
    required this.items,
  });

  final StoreInfo store;
  final String purchaseDate;
  final double total;
  final List<ReceiptItem> items;

  Map<String, dynamic> toJson() => {
    'store': store.toJson(),
    'purchase_date': purchaseDate,
    'total': total,
    'items': items.map((e) => e.toJson()).toList(),
  };
}

class ConfirmReceiptResponse {
  const ConfirmReceiptResponse({required this.pointsEarned});

  final int pointsEarned;

  factory ConfirmReceiptResponse.fromJson(Map<String, dynamic> json) {
    return ConfirmReceiptResponse(
      pointsEarned: json['points_earned'] as int? ?? 0,
    );
  }
}
