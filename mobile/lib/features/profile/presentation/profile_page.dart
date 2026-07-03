import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme/spacing.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../../auth/presentation/auth_notifier.dart';
import '../data/profile_models.dart';
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

  Color _levelColor(String level, BuildContext context) {
    switch (level) {
      case 'Platino':
        return const Color(0xFFB0BEC5);
      case 'Oro':
        return const Color(0xFFFFC107);
      case 'Plata':
        return const Color(0xFF9E9E9E);
      default:
        return const Color(0xFFB87333);
    }
  }

  IconData _levelIcon(String level) {
    switch (level) {
      case 'Platino':
      case 'Oro':
        return Icons.workspace_premium;
      case 'Plata':
        return Icons.emoji_events;
      default:
        return Icons.military_tech;
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final authState = ref.watch(authNotifierProvider);
    final profileState = ref.watch(profileNotifierProvider);
    final levelColor = _levelColor(profileState.level, context);

    return Scaffold(
      appBar: AppBar(title: const Text('Profile')),
      body: SafeArea(
        child: switch (profileState.status) {
          ProfileStatus.loading => const LoadingIndicator(),
          ProfileStatus.error => Center(
            child: Text(profileState.errorMessage ?? 'Error loading profile'),
          ),
          ProfileStatus.ready => RefreshIndicator(
            onRefresh: () =>
                ref.read(profileNotifierProvider.notifier).loadPoints(),
            child: SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
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
                        const Spacer(),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: AppSpacing.md,
                            vertical: AppSpacing.sm,
                          ),
                          decoration: BoxDecoration(
                            color: levelColor.withValues(alpha: 0.18),
                            borderRadius: BorderRadius.circular(AppRadius.lg),
                            border: Border.all(color: levelColor, width: 1.5),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                _levelIcon(profileState.level),
                                size: 18,
                                color: levelColor,
                              ),
                              const SizedBox(width: AppSpacing.xs),
                              Text(
                                profileState.level,
                                style: theme.textTheme.labelLarge?.copyWith(
                                  color: levelColor,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: AppSpacing.xl),
                  _ContributorCard(stats: profileState.contributor),
                  const SizedBox(height: AppSpacing.xl),
                  if (profileState.recentTransactions.isNotEmpty) ...[
                    Text('Recent Activity', style: theme.textTheme.titleMedium),
                    const SizedBox(height: AppSpacing.sm),
                    ...profileState.recentTransactions.map(
                      (tx) => ListTile(
                        leading: Icon(
                          tx.points >= 0
                              ? Icons.add_circle
                              : Icons.remove_circle,
                          color: tx.points >= 0 ? Colors.green : Colors.red,
                        ),
                        title: Text(_humanizeReason(tx.reason)),
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
          ),
        },
      ),
    );
  }

  String _humanizeReason(String reason) {
    return reason
        .split(';')
        .map((part) {
          switch (part) {
            case 'receipt_confirmed':
              return 'Recibo confirmado';
            case 'first_observation_product':
              return 'Primer producto registrado';
            case 'first_observation_store':
              return 'Nuevo establecimiento';
            case 'data_completion':
              return 'Datos completos';
            case 'daily_limit_reached':
              return 'Límite diario';
            default:
              return part;
          }
        })
        .join(' + ');
  }
}

class _ContributorCard extends StatelessWidget {
  const _ContributorCard({required this.stats});

  final ContributorStats stats;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(AppRadius.lg),
        border: Border.all(
          color: theme.colorScheme.outline.withValues(alpha: 0.2),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.insights, size: 20, color: theme.colorScheme.primary),
              const SizedBox(width: AppSpacing.sm),
              Text(
                'Tu contribución',
                style: theme.textTheme.titleSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          Row(
            children: [
              Expanded(
                child: _Stat(
                  icon: Icons.receipt_long,
                  value: stats.receiptsConfirmed,
                  label: 'Recibos',
                ),
              ),
              Expanded(
                child: _Stat(
                  icon: Icons.price_change,
                  value: stats.priceObservations,
                  label: 'Precios',
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          Row(
            children: [
              Expanded(
                child: _Stat(
                  icon: Icons.store,
                  value: stats.uniqueStores,
                  label: 'Tiendas',
                ),
              ),
              Expanded(
                child: _Stat(
                  icon: Icons.shopping_basket,
                  value: stats.uniqueProducts,
                  label: 'Productos',
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({required this.icon, required this.value, required this.label});

  final IconData icon;
  final int value;
  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      children: [
        Icon(icon, color: theme.colorScheme.primary, size: 28),
        const SizedBox(height: AppSpacing.xs),
        Text(
          value.toString(),
          style: theme.textTheme.headlineSmall?.copyWith(
            fontWeight: FontWeight.bold,
          ),
        ),
        Text(
          label,
          style: theme.textTheme.bodySmall?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      ],
    );
  }
}
