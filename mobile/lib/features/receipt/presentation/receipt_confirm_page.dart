import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme/spacing.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../data/receipt_models.dart';
import 'receipt_confirm_notifier.dart';
import 'receipt_review_notifier.dart';

class ReceiptConfirmPage extends ConsumerStatefulWidget {
  const ReceiptConfirmPage({super.key, required this.receiptId});

  final String receiptId;

  @override
  ConsumerState<ReceiptConfirmPage> createState() => _ReceiptConfirmPageState();
}

class _ReceiptConfirmPageState extends ConsumerState<ReceiptConfirmPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _submitConfirmation();
    });
  }

  Future<void> _submitConfirmation() async {
    final reviewState = ref.read(reviewNotifierProvider);
    final request = ConfirmReceiptRequest(
      store: reviewState.store ?? const StoreInfo(),
      purchaseDate: reviewState.purchaseDate ?? '',
      total: reviewState.total ?? 0.0,
      items: reviewState.items,
    );
    await ref
        .read(confirmNotifierProvider.notifier)
        .confirm(widget.receiptId, request);
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(confirmNotifierProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Confirm Receipt')),
      body: SafeArea(
        child: switch (state.status) {
          ConfirmStatus.confirming => const LoadingIndicator(
            message: 'Confirming...',
          ),
          ConfirmStatus.confirmed => _buildSuccess(
            context,
            state.pointsEarned!,
          ),
          ConfirmStatus.error => Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(state.errorMessage ?? 'Confirmation failed'),
                const SizedBox(height: AppSpacing.lg),
                OutlinedButton(
                  onPressed: _submitConfirmation,
                  child: const Text('Retry'),
                ),
              ],
            ),
          ),
          ConfirmStatus.idle => const SizedBox(),
        },
      ),
    );
  }

  Widget _buildSuccess(BuildContext context, int points) {
    final theme = Theme.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.celebration, size: 80, color: theme.colorScheme.primary),
            const SizedBox(height: AppSpacing.lg),
            Text('Receipt Confirmed!', style: theme.textTheme.headlineMedium),
            const SizedBox(height: AppSpacing.xl),
            Container(
              padding: const EdgeInsets.all(AppSpacing.xl),
              decoration: BoxDecoration(
                color: theme.colorScheme.primaryContainer,
                shape: BoxShape.circle,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.stars,
                    size: 48,
                    color: theme.colorScheme.onPrimaryContainer,
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  Text(
                    '+$points',
                    style: theme.textTheme.headlineMedium?.copyWith(
                      color: theme.colorScheme.onPrimaryContainer,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            const Text('Points Earned'),
            const SizedBox(height: AppSpacing.xxl),
            FilledButton(
              onPressed: () => context.go('/home'),
              child: const Text('Done'),
            ),
          ],
        ),
      ),
    );
  }
}
