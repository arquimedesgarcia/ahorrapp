import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../../shared/widgets/empty_state.dart';

class ReceiptListPage extends StatelessWidget {
  const ReceiptListPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Receipts')),
      body: SafeArea(
        child: Center(
          child: EmptyState(
            icon: Icons.receipt_outlined,
            title: 'No receipts yet',
            subtitle: 'Scan your first receipt to start saving.',
            actionLabel: 'Scan Receipt',
            onAction: () => context.go('/home'),
          ),
        ),
      ),
    );
  }
}
