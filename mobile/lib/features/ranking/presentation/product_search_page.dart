import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/theme/spacing.dart';
import '../../../shared/widgets/empty_state.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../data/ranking_models.dart';
import 'product_search_notifier.dart';
import 'widgets/ranking_list.dart';

class ProductSearchPage extends ConsumerStatefulWidget {
  const ProductSearchPage({super.key});

  @override
  ConsumerState<ProductSearchPage> createState() => _ProductSearchPageState();
}

class _ProductSearchPageState extends ConsumerState<ProductSearchPage> {
  final _searchController = TextEditingController();

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _search() async {
    await ref
        .read(searchNotifierProvider.notifier)
        .search(_searchController.text);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final state = ref.watch(searchNotifierProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Search Products')),
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(AppSpacing.md),
              child: TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  hintText: 'Search for a product...',
                  prefixIcon: const Icon(Icons.search),
                  suffixIcon: IconButton(
                    icon: const Icon(Icons.send),
                    onPressed: _search,
                  ),
                ),
                onSubmitted: (_) => _search(),
              ),
            ),
            Expanded(
              child: switch (state.status) {
                SearchStatus.idle => Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.search,
                        size: 64,
                        color: theme.colorScheme.outline,
                      ),
                      const SizedBox(height: AppSpacing.md),
                      const Text('Search for products to find the best prices'),
                    ],
                  ),
                ),
                SearchStatus.loading => const LoadingIndicator(),
                SearchStatus.empty => EmptyState(
                  icon: Icons.search_off,
                  title: 'No results found',
                  subtitle: "No results found for '${state.lastQuery}'",
                ),
                SearchStatus.error => Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [Text(state.errorMessage ?? 'Search failed')],
                  ),
                ),
                SearchStatus.results => ListView.builder(
                  padding: const EdgeInsets.symmetric(
                    horizontal: AppSpacing.md,
                    vertical: AppSpacing.sm,
                  ),
                  itemCount: state.results.length,
                  itemBuilder: (context, index) {
                    final result = state.results[index];
                    return _buildResultCard(context, result);
                  },
                ),
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildResultCard(BuildContext context, ProductSearchResult result) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
      child: ExpansionTile(
        title: Text(result.productName),
        subtitle: result.unit != null ? Text('Unit: ${result.unit}') : null,
        children: [RankingList(stores: result.stores)],
      ),
    );
  }
}
