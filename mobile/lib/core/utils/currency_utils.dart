import 'package:flutter/material.dart';

import '../../app/theme/colors.dart';

class CurrencyUtils {
  const CurrencyUtils._();

  static String currencySymbol(String currency) {
    switch (currency) {
      case 'USD':
        return '\$';
      case 'Bs.':
      case 'Bs':
        return 'Bs.';
      default:
        return currency;
    }
  }

  static String formatPrice(double price, String currency) {
    final symbol = currencySymbol(currency);
    final formatted = price.toStringAsFixed(2);
    return '$symbol $formatted';
  }

  static Color currencyColor(String currency) {
    switch (currency) {
      case 'USD':
        return CurrencyColors.usd;
      case 'Bs.':
      case 'Bs':
        return CurrencyColors.bs;
      default:
        return Colors.grey;
    }
  }
}
