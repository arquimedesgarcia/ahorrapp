import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme/spacing.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../data/receipt_models.dart';
import '../domain/receipt_repository.dart';
import 'receipt_review_notifier.dart';
import 'widgets/receipt_item_form.dart';

class ReceiptReviewPage extends ConsumerStatefulWidget {
  const ReceiptReviewPage({super.key, required this.receiptId});

  final String receiptId;

  @override
  ConsumerState<ReceiptReviewPage> createState() => _ReceiptReviewPageState();
}

class _ReceiptReviewPageState extends ConsumerState<ReceiptReviewPage> {
  final _storeNameController = TextEditingController();
  final _branchController = TextEditingController();
  final _dateController = TextEditingController();
  final _totalController = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(reviewNotifierProvider.notifier).loadReceipt(widget.receiptId);
    });
  }

  @override
  void dispose() {
    _storeNameController.dispose();
    _branchController.dispose();
    _dateController.dispose();
    _totalController.dispose();
    super.dispose();
  }

  void _syncControllers(ReviewState state) {
    _storeNameController.text = state.store?.name ?? '';
    _branchController.text = state.store?.branch ?? '';
    _dateController.text = state.purchaseDate ?? '';
    _totalController.text = state.total?.toString() ?? '';
  }

  Future<void> _confirm() async {
    final state = ref.read(reviewNotifierProvider);
    if (!ref.read(receiptRepositoryProvider).validateCurrency(state.items)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Currency is required for all items')),
      );
      return;
    }
    context.go('/receipts/${widget.receiptId}/confirm');
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(reviewNotifierProvider);

    if (state.status == ReviewStatus.loading) {
      return Scaffold(
        appBar: AppBar(title: const Text('Review Receipt')),
        body: const LoadingIndicator(message: 'Processing receipt...'),
      );
    }

    if (state.status == ReviewStatus.error) {
      return Scaffold(
        appBar: AppBar(title: const Text('Review Receipt')),
        body: Center(child: Text(state.errorMessage ?? 'Error loading')),
      );
    }

    if (state.detail != null &&
        _storeNameController.text.isEmpty &&
        state.store?.name != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _syncControllers(state);
      });
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Review Receipt')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Store',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              TextFormField(
                controller: _storeNameController,
                decoration: const InputDecoration(labelText: 'Store Name'),
                onChanged: (value) => ref
                    .read(reviewNotifierProvider.notifier)
                    .updateStore(
                      StoreInfo(
                        name: value,
                        branch: _branchController.text,
                        address: state.store?.address,
                      ),
                    ),
              ),
              const SizedBox(height: AppSpacing.sm),
              TextFormField(
                controller: _branchController,
                decoration: const InputDecoration(labelText: 'Branch'),
                onChanged: (value) => ref
                    .read(reviewNotifierProvider.notifier)
                    .updateStore(
                      StoreInfo(
                        name: _storeNameController.text,
                        branch: value,
                        address: state.store?.address,
                      ),
                    ),
              ),
              const SizedBox(height: AppSpacing.md),
              TextFormField(
                controller: _dateController,
                decoration: const InputDecoration(labelText: 'Purchase Date'),
                onChanged: (value) => ref
                    .read(reviewNotifierProvider.notifier)
                    .updatePurchaseDate(value),
              ),
              const SizedBox(height: AppSpacing.sm),
              TextFormField(
                controller: _totalController,
                decoration: const InputDecoration(labelText: 'Total'),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                onChanged: (value) {
                  final total = double.tryParse(value);
                  if (total != null) {
                    ref
                        .read(reviewNotifierProvider.notifier)
                        .updateTotal(total);
                  }
                },
              ),
              const SizedBox(height: AppSpacing.lg),
              Row(
                children: [
                  const Text(
                    'Items',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  const Spacer(),
                  TextButton.icon(
                    onPressed: () =>
                        ref.read(reviewNotifierProvider.notifier).addItem(),
                    icon: const Icon(Icons.add),
                    label: const Text('Add Item'),
                  ),
                ],
              ),
              ...state.items.asMap().entries.map((entry) {
                final index = entry.key;
                final item = entry.value;
                return ReceiptItemForm(
                  item: item,
                  onChanged: (updated) {
                    final items = List<ReceiptItem>.from(state.items);
                    items[index] = updated;
                    ref
                        .read(reviewNotifierProvider.notifier)
                        .updateItems(items);
                  },
                  onDelete: () => ref
                      .read(reviewNotifierProvider.notifier)
                      .removeItem(index),
                );
              }),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(onPressed: _confirm, child: const Text('Confirm')),
            ],
          ),
        ),
      ),
    );
  }
}
