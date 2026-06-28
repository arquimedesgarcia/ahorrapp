import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme/spacing.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../../auth/presentation/auth_notifier.dart';
import 'profile_notifier.dart';

class ProfilePage extends ConsumerStatefulWidget {
  const ProfilePage({super.key});

  @override
  ConsumerState<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends ConsumerState<ProfilePage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(profileNotifierProvider.notifier).loadPoints();
    });
  }

  Future<void> _logout() async {
    await ref.read(authNotifierProvider.notifier).logout();
    if (mounted) {
      context.go('/login');
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final authState = ref.watch(authNotifierProvider);
    final profileState = ref.watch(profileNotifierProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Profile')),
      body: SafeArea(
        child: switch (profileState.status) {
          ProfileStatus.loading => const LoadingIndicator(),
          ProfileStatus.error => Center(
            child: Text(profileState.errorMessage ?? 'Error loading profile'),
          ),
          ProfileStatus.ready => SingleChildScrollView(
            padding: const EdgeInsets.all(AppSpacing.md),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SizedBox(height: AppSpacing.xl),
                CircleAvatar(
                  radius: 48,
                  backgroundColor: theme.colorScheme.primaryContainer,
                  child: Icon(
                    Icons.person,
                    size: 48,
                    color: theme.colorScheme.onPrimaryContainer,
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                Text(
                  authState.user?.displayName ?? 'User',
                  style: theme.textTheme.headlineSmall,
                  textAlign: TextAlign.center,
                ),
                Text(
                  authState.user?.email ?? '',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: AppSpacing.xl),
                Container(
                  padding: const EdgeInsets.all(AppSpacing.xl),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(AppRadius.lg),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.stars,
                        size: 36,
                        color: theme.colorScheme.onPrimaryContainer,
                      ),
                      const SizedBox(width: AppSpacing.sm),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            profileState.totalPoints.toString(),
                            style: theme.textTheme.headlineMedium?.copyWith(
                              color: theme.colorScheme.onPrimaryContainer,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          Text(
                            'Total Points',
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onPrimaryContainer,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.xl),
                if (profileState.recentTransactions.isNotEmpty) ...[
                  Text('Recent Activity', style: theme.textTheme.titleMedium),
                  const SizedBox(height: AppSpacing.sm),
                  ...profileState.recentTransactions.map(
                    (tx) => ListTile(
                      leading: Icon(
                        tx.points >= 0 ? Icons.add_circle : Icons.remove_circle,
                        color: tx.points >= 0 ? Colors.green : Colors.red,
                      ),
                      title: Text(tx.reason),
                      trailing: Text(
                        '${tx.points >= 0 ? '+' : ''}${tx.points}',
                        style: theme.textTheme.titleMedium,
                      ),
                    ),
                  ),
                ],
                const SizedBox(height: AppSpacing.xl),
                OutlinedButton.icon(
                  onPressed: _logout,
                  icon: const Icon(Icons.logout, color: Colors.red),
                  label: const Text(
                    'Logout',
                    style: TextStyle(color: Colors.red),
                  ),
                ),
              ],
            ),
          ),
        },
      ),
    );
  }
}
