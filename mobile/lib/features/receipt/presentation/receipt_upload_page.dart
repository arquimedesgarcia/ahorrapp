import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/theme/spacing.dart';
import 'receipt_camera_page.dart';
import 'receipt_upload_notifier.dart';

class ReceiptUploadPage extends ConsumerStatefulWidget {
  const ReceiptUploadPage({super.key});

  @override
  ConsumerState<ReceiptUploadPage> createState() => _ReceiptUploadPageState();
}

class _ReceiptUploadPageState extends ConsumerState<ReceiptUploadPage> {
  Future<void> _openCamera() async {
    await Navigator.of(context).push<void>(
      PageRouteBuilder<void>(
        pageBuilder: (context, animation, secondaryAnimation) =>
            const ReceiptCameraPage(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final state = ref.watch(uploadNotifierProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('AhorraApp')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Spacer(),
              Icon(
                Icons.receipt_long,
                size: 100,
                color: theme.colorScheme.primary,
              ),
              const SizedBox(height: AppSpacing.lg),
              Text(
                'Scan a receipt to start',
                style: theme.textTheme.headlineSmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppSpacing.sm),
              Text(
                'Take a photo and we will extract the details',
                style: theme.textTheme.bodyLarge?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
              const Spacer(),
              if (state.status == UploadStatus.uploading) ...[
                const LinearProgressIndicator(),
                const SizedBox(height: AppSpacing.md),
                const Text('Uploading...'),
              ] else ...[
                FilledButton.icon(
                  onPressed: _openCamera,
                  icon: const Icon(Icons.camera_alt),
                  label: const Text('Scan Receipt'),
                ),
              ],
              const SizedBox(height: AppSpacing.xl),
            ],
          ),
        ),
      ),
    );
  }
}
