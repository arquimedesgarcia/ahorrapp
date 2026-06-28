import 'package:flutter/material.dart';

import 'colors.dart';
import 'spacing.dart';
import 'typography.dart';

ThemeData appTheme(Brightness brightness) {
  final isLight = brightness == Brightness.light;
  final primary = isLight ? AppColors.primary : AppDarkColors.primary;
  final onPrimary = isLight ? AppColors.onPrimary : AppDarkColors.onPrimary;
  final primaryContainer = isLight
      ? AppColors.primaryContainer
      : AppDarkColors.primaryContainer;
  final onPrimaryContainer = isLight
      ? AppColors.onPrimaryContainer
      : AppDarkColors.onPrimaryContainer;
  final secondary = isLight ? AppColors.secondary : AppDarkColors.secondary;
  final onSecondary = isLight
      ? AppColors.onSecondary
      : AppDarkColors.onSecondary;
  final tertiary = isLight ? AppColors.tertiary : AppDarkColors.tertiary;
  final onTertiary = isLight ? AppColors.onTertiary : AppDarkColors.onTertiary;
  final error = isLight ? AppColors.error : AppDarkColors.error;
  final onError = isLight ? AppColors.onError : AppDarkColors.onError;
  final errorContainer = isLight
      ? AppColors.errorContainer
      : AppDarkColors.errorContainer;
  final onErrorContainer = isLight
      ? AppColors.onErrorContainer
      : AppDarkColors.onErrorContainer;
  final surface = isLight ? AppColors.surface : AppDarkColors.surface;
  final onSurface = isLight ? AppColors.onSurface : AppDarkColors.onSurface;
  final surfaceVariant = isLight
      ? AppColors.surfaceVariant
      : AppDarkColors.surfaceVariant;
  final onSurfaceVariant = isLight
      ? AppColors.onSurfaceVariant
      : AppDarkColors.onSurfaceVariant;
  final outline = isLight ? AppColors.outline : AppDarkColors.outline;
  final outlineVariant = isLight
      ? AppColors.outlineVariant
      : AppDarkColors.outlineVariant;

  final colorScheme = ColorScheme(
    brightness: brightness,
    primary: primary,
    onPrimary: onPrimary,
    primaryContainer: primaryContainer,
    onPrimaryContainer: onPrimaryContainer,
    secondary: secondary,
    onSecondary: onSecondary,
    tertiary: tertiary,
    onTertiary: onTertiary,
    error: error,
    onError: onError,
    errorContainer: errorContainer,
    onErrorContainer: onErrorContainer,
    surface: surface,
    onSurface: onSurface,
    surfaceContainerHighest: surfaceVariant,
    onSurfaceVariant: onSurfaceVariant,
    outline: outline,
    outlineVariant: outlineVariant,
  );

  return ThemeData(
    useMaterial3: true,
    colorScheme: colorScheme,
    textTheme: const TextTheme(
      displayLarge: AppTypography.displayLarge,
      displayMedium: AppTypography.displayMedium,
      displaySmall: AppTypography.displaySmall,
      headlineLarge: AppTypography.headlineLarge,
      headlineMedium: AppTypography.headlineMedium,
      headlineSmall: AppTypography.headlineSmall,
      titleLarge: AppTypography.titleLarge,
      titleMedium: AppTypography.titleMedium,
      titleSmall: AppTypography.titleSmall,
      bodyLarge: AppTypography.bodyLarge,
      bodyMedium: AppTypography.bodyMedium,
      bodySmall: AppTypography.bodySmall,
      labelLarge: AppTypography.labelLarge,
      labelMedium: AppTypography.labelMedium,
      labelSmall: AppTypography.labelSmall,
    ),
    cardTheme: CardThemeData(
      elevation: AppElevation.level1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppRadius.lg),
      ),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        minimumSize: const Size(double.infinity, 56),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(AppRadius.xl),
        ),
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        minimumSize: const Size(double.infinity, 56),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(AppRadius.xl),
        ),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppRadius.sm),
      ),
      contentPadding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.md,
        vertical: AppSpacing.md,
      ),
    ),
    appBarTheme: AppBarTheme(
      centerTitle: true,
      backgroundColor: surface,
      foregroundColor: onSurface,
      elevation: AppElevation.level0,
    ),
  );
}
