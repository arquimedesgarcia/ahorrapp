import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../app/theme/spacing.dart';
import '../../../shared/widgets/empty_state.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../data/receipt_models.dart';
import 'receipt_list_notifier.dart';

class ReceiptListPage extends ConsumerStatefulWidget {
  const ReceiptListPage({super.key});

  @override
  ConsumerState<ReceiptListPage> createState() => _ReceiptListPageState();
}

class _ReceiptListPageState extends ConsumerState<ReceiptListPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(receiptListProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(receiptListProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Receipts'),
        actions: [
          IconButton(
            tooltip: 'Reload',
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(receiptListProvider.notifier).load(),
          ),
        ],
      ),
      body: SafeArea(child: _buildBody(context, state)),
    );
  }

  Widget _buildBody(BuildContext context, ReceiptListState state) {
    if (state.items.isEmpty) {
      if (state.status == ReceiptListStatus.loading) {
        return const LoadingIndicator();
      }
      if (state.status == ReceiptListStatus.error) {
        return _ErrorView(
          message: state.errorMessage ?? 'Request failed',
          onRetry: () => ref.read(receiptListProvider.notifier).load(),
        );
      }
    }
    return RefreshIndicator(
      onRefresh: () => ref.read(receiptListProvider.notifier).load(),
      child: state.items.isEmpty
          ? ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              children: [
                SizedBox(
                  height: MediaQuery.of(context).size.height * 0.6,
                  child: EmptyState(
                    icon: Icons.receipt_outlined,
                    title: 'No receipts yet',
                    subtitle: 'Scan your first receipt to start saving.',
                    actionLabel: 'Scan Receipt',
                    onAction: () => context.go('/home'),
                  ),
                ),
              ],
            )
          : ListView.separated(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.all(AppSpacing.md),
              itemCount: state.items.length,
              separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
              itemBuilder: (_, i) =>
                  _ReceiptCard(item: state.items[i], theme: Theme.of(context)),
            ),
    );
  }
}

class _ReceiptCard extends StatelessWidget {
  const _ReceiptCard({required this.item, required this.theme});

  final ReceiptListItem item;
  final ThemeData theme;

  @override
  Widget build(BuildContext context) {
    final dateStr =
        _formatDate(item.purchaseDate) ??
        DateFormat('dd MMM yyyy').format(item.createdAt.toLocal());
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppRadius.lg),
        side: BorderSide(
          color: theme.colorScheme.outline.withValues(alpha: 0.2),
        ),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadius.lg),
        onTap: () => context.push('/receipts/${item.id}'),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Row(
            children: [
              _StatusDot(status: item.status),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      item.storeName.isEmpty ? 'Sin tienda' : item.storeName,
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 2),
                    Text(
                      dateStr,
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        Icon(
                          Icons.shopping_basket_outlined,
                          size: 14,
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          '${item.itemCount} ${item.itemCount == 1 ? 'item' : 'items'}',
                          style: theme.textTheme.bodySmall,
                        ),
                        if (item.total != null) ...[
                          const SizedBox(width: AppSpacing.md),
                          Icon(
                            Icons.attach_money,
                            size: 14,
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                          Text(
                            item.total!.toStringAsFixed(2),
                            style: theme.textTheme.bodySmall,
                          ),
                        ],
                      ],
                    ),
                  ],
                ),
              ),
              Icon(
                Icons.chevron_right,
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ],
          ),
        ),
      ),
    );
  }

  String? _formatDate(String? iso) {
    if (iso == null || iso.isEmpty) return null;
    final d = DateTime.tryParse(iso);
    if (d == null) return null;
    return DateFormat('dd MMM yyyy').format(d.toLocal());
  }
}

class _StatusDot extends StatelessWidget {
  const _StatusDot({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final (color, label) = switch (status) {
      'CONFIRMED' => (Colors.green, 'OK'),
      'NEEDS_REVIEW' => (Colors.orange, 'Rev'),
      'PENDING' => (Colors.blueGrey, '...'),
      'REJECTED' => (Colors.red, 'X'),
      _ => (Colors.grey, '?'),
    };
    return Container(
      width: 40,
      height: 40,
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        shape: BoxShape.circle,
      ),
      alignment: Alignment.center,
      child: Text(
        label,
        style: TextStyle(
          color: color,
          fontWeight: FontWeight.bold,
          fontSize: 12,
        ),
      ),
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.error_outline, size: 48),
            const SizedBox(height: AppSpacing.md),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: AppSpacing.lg),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}
