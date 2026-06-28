import 'package:flutter/material.dart';

import '../../../../core/utils/currency_utils.dart';

class PriceDisplay extends StatelessWidget {
  const PriceDisplay({
    super.key,
    required this.price,
    required this.currency,
    this.fontSize = 16,
  });

  final double price;
  final String currency;
  final double fontSize;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          CurrencyUtils.formatPrice(price, currency),
          style: TextStyle(
            fontSize: fontSize,
            fontWeight: FontWeight.bold,
            color: CurrencyUtils.currencyColor(currency),
          ),
        ),
        const SizedBox(width: 6),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
          decoration: BoxDecoration(
            color: CurrencyUtils.currencyColor(
              currency,
            ).withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(4),
          ),
          child: Text(
            currency,
            style: TextStyle(
              fontSize: fontSize * 0.7,
              color: CurrencyUtils.currencyColor(currency),
            ),
          ),
        ),
      ],
    );
  }
}
