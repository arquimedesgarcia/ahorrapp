class StorePriceEntry {
  const StorePriceEntry({
    required this.storeId,
    required this.storeName,
    this.branch,
    required this.averagePrice,
    required this.currency,
    required this.sampleCount,
  });

  final String storeId;
  final String storeName;
  final String? branch;
  final double averagePrice;
  final String currency;
  final int sampleCount;

  factory StorePriceEntry.fromJson(Map<String, dynamic> json) {
    return StorePriceEntry(
      storeId: json['store_id'] as String,
      storeName: json['store_name'] as String,
      branch: json['branch'] as String?,
      averagePrice: (json['average_price'] as num).toDouble(),
      currency: json['currency'] as String,
      sampleCount: json['sample_count'] as int,
    );
  }
}

class ProductSearchResult {
  const ProductSearchResult({
    required this.productId,
    required this.productName,
    this.unit,
    required this.stores,
  });

  final String productId;
  final String productName;
  final String? unit;
  final List<StorePriceEntry> stores;

  factory ProductSearchResult.fromJson(Map<String, dynamic> json) {
    return ProductSearchResult(
      productId: json['product_id'] as String,
      productName: json['product_name'] as String,
      unit: json['unit'] as String?,
      stores: (json['stores'] as List<dynamic>)
          .map((e) => StorePriceEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

class ProductSearchResponse {
  const ProductSearchResponse({required this.results});

  final List<ProductSearchResult> results;

  factory ProductSearchResponse.fromJson(Map<String, dynamic> json) {
    return ProductSearchResponse(
      results: (json['results'] as List<dynamic>)
          .map((e) => ProductSearchResult.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
