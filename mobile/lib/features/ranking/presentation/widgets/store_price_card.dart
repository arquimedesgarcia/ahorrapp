import 'package:flutter/material.dart';

import '../../../../app/theme/spacing.dart';
import '../../../../core/utils/currency_utils.dart';
import '../../data/ranking_models.dart';

class StorePriceCard extends StatelessWidget {
  const StorePriceCard({super.key, required this.store, required this.rank});

  final StorePriceEntry store;
  final int rank;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isCheapest = rank == 0;

    return Card(
      color: isCheapest ? theme.colorScheme.primaryContainer : null,
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Row(
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                color: isCheapest
                    ? theme.colorScheme.primary
                    : theme.colorScheme.surfaceContainerHighest,
                shape: BoxShape.circle,
              ),
              child: Center(
                child: Text(
                  '${rank + 1}',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: isCheapest
                        ? theme.colorScheme.onPrimary
                        : theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
            ),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(store.storeName, style: theme.textTheme.titleSmall),
                  if (store.branch != null) ...[
                    const SizedBox(height: 2),
                    Text(store.branch!, style: theme.textTheme.bodySmall),
                  ],
                ],
              ),
            ),
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  CurrencyUtils.formatPrice(store.averagePrice, store.currency),
                  style: theme.textTheme.titleMedium?.copyWith(
                    color: CurrencyUtils.currencyColor(store.currency),
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 2),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 6,
                    vertical: 2,
                  ),
                  decoration: BoxDecoration(
                    color: CurrencyUtils.currencyColor(
                      store.currency,
                    ).withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    store.currency,
                    style: TextStyle(
                      fontSize: 11,
                      color: CurrencyUtils.currencyColor(store.currency),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
