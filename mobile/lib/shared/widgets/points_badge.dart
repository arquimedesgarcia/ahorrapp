import 'package:flutter/material.dart';

import '../../app/theme/spacing.dart';

class PointsBadge extends StatelessWidget {
  const PointsBadge({
    super.key,
    required this.points,
    this.size = PointsBadgeSize.medium,
  });

  final int points;
  final PointsBadgeSize size;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final fontSize = size == PointsBadgeSize.large ? 24.0 : 16.0;
    final iconSize = size == PointsBadgeSize.large ? 28.0 : 20.0;

    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.md,
        vertical: AppSpacing.sm,
      ),
      decoration: BoxDecoration(
        color: theme.colorScheme.primaryContainer,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            Icons.stars,
            size: iconSize,
            color: theme.colorScheme.onPrimaryContainer,
          ),
          const SizedBox(width: AppSpacing.xs),
          Text(
            points.toString(),
            style: theme.textTheme.labelLarge?.copyWith(
              fontSize: fontSize,
              fontWeight: FontWeight.bold,
              color: theme.colorScheme.onPrimaryContainer,
            ),
          ),
        ],
      ),
    );
  }
}

enum PointsBadgeSize { medium, large }
