class PointsTransaction {
  const PointsTransaction({
    required this.id,
    required this.points,
    required this.reason,
    required this.createdAt,
  });

  final String id;
  final int points;
  final String reason;
  final String createdAt;

  factory PointsTransaction.fromJson(Map<String, dynamic> json) {
    return PointsTransaction(
      id: json['id'] as String,
      points: json['points'] as int,
      reason: json['reason'] as String,
      createdAt: json['created_at'] as String,
    );
  }
}

class ContributorStats {
  const ContributorStats({
    this.receiptsConfirmed = 0,
    this.priceObservations = 0,
    this.uniqueStores = 0,
    this.uniqueProducts = 0,
  });

  final int receiptsConfirmed;
  final int priceObservations;
  final int uniqueStores;
  final int uniqueProducts;

  factory ContributorStats.fromJson(Map<String, dynamic> json) {
    return ContributorStats(
      receiptsConfirmed: json['receipts_confirmed'] as int? ?? 0,
      priceObservations: json['price_observations'] as int? ?? 0,
      uniqueStores: json['unique_stores'] as int? ?? 0,
      uniqueProducts: json['unique_products'] as int? ?? 0,
    );
  }
}

class PointsResponse {
  const PointsResponse({
    required this.totalPoints,
    required this.level,
    required this.contributor,
    this.recentTransactions = const [],
  });

  final int totalPoints;
  final String level;
  final ContributorStats contributor;
  final List<PointsTransaction> recentTransactions;

  factory PointsResponse.fromJson(Map<String, dynamic> json) {
    return PointsResponse(
      totalPoints: json['total_points'] as int? ?? 0,
      level: json['level'] as String? ?? 'Bronce',
      contributor: ContributorStats.fromJson(
        (json['contributor'] as Map<String, dynamic>?) ?? const {},
      ),
      recentTransactions:
          (json['recent_transactions'] as List<dynamic>?)
              ?.map(
                (e) => PointsTransaction.fromJson(e as Map<String, dynamic>),
              )
              .toList() ??
          [],
    );
  }
}
