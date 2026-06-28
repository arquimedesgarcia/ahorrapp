import 'package:flutter/material.dart';

import '../../../../app/theme/spacing.dart';
import '../../data/ranking_models.dart';
import 'store_price_card.dart';

/// Renders the ranked list of store prices for a single product, ordered from
/// cheapest to most expensive.
///
/// The API contract already returns stores sorted ascending by price, but this
/// widget defensively re-sorts to guarantee cheapest-first ordering regardless
/// of the source.
class RankingList extends StatelessWidget {
  const RankingList({super.key, required this.stores});

  final List<StorePriceEntry> stores;

  @override
  Widget build(BuildContext context) {
    if (stores.isEmpty) {
      return const Padding(
        padding: EdgeInsets.all(AppSpacing.md),
        child: Text('No price data available'),
      );
    }

    final ranked = List<StorePriceEntry>.of(stores)
      ..sort((a, b) => a.averagePrice.compareTo(b.averagePrice));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (int i = 0; i < ranked.length; i++)
          Padding(
            padding: const EdgeInsets.symmetric(
              horizontal: AppSpacing.md,
              vertical: AppSpacing.xs,
            ),
            child: StorePriceCard(store: ranked[i], rank: i),
          ),
      ],
    );
  }
}
