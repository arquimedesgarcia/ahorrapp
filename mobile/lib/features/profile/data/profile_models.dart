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

class PointsResponse {
  const PointsResponse({
    required this.totalPoints,
    this.recentTransactions = const [],
  });

  final int totalPoints;
  final List<PointsTransaction> recentTransactions;

  factory PointsResponse.fromJson(Map<String, dynamic> json) {
    return PointsResponse(
      totalPoints: json['total_points'] as int? ?? 0,
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
