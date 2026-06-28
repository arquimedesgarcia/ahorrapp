import 'package:flutter/material.dart';

import '../../../../app/theme/spacing.dart';
import '../../../../shared/widgets/currency_selector.dart';
import '../../data/receipt_models.dart';

class ReceiptItemForm extends StatelessWidget {
  const ReceiptItemForm({
    super.key,
    required this.item,
    required this.onChanged,
    required this.onDelete,
  });

  final ReceiptItem item;
  final ValueChanged<ReceiptItem> onChanged;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(item.rawText, style: theme.textTheme.titleSmall),
                ),
                IconButton(
                  icon: const Icon(Icons.delete_outline, size: 20),
                  onPressed: onDelete,
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.sm),
            TextFormField(
              initialValue: item.rawText,
              decoration: const InputDecoration(
                labelText: 'Product',
                isDense: true,
              ),
              onChanged: (value) => onChanged(item.copyWith(rawText: value)),
            ),
            const SizedBox(height: AppSpacing.sm),
            Row(
              children: [
                Expanded(
                  flex: 2,
                  child: TextFormField(
                    initialValue: item.quantity?.toString() ?? '',
                    decoration: const InputDecoration(
                      labelText: 'Qty',
                      isDense: true,
                    ),
                    keyboardType: TextInputType.number,
                    onChanged: (value) {
                      final qty = int.tryParse(value);
                      onChanged(item.copyWith(quantity: qty));
                    },
                  ),
                ),
                const SizedBox(width: AppSpacing.sm),
                Expanded(
                  flex: 3,
                  child: TextFormField(
                    initialValue: item.unitPrice?.toString() ?? '',
                    decoration: const InputDecoration(
                      labelText: 'Price',
                      isDense: true,
                    ),
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
                    onChanged: (value) {
                      final price = double.tryParse(value);
                      onChanged(item.copyWith(unitPrice: price));
                    },
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.sm),
            CurrencySelector(
              value: item.currency ?? 'USD',
              onChanged: (value) => onChanged(item.copyWith(currency: value)),
            ),
          ],
        ),
      ),
    );
  }
}
