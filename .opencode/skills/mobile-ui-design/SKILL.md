---
name: mobile-ui-design
description: Mobile UI/UX design system for AhorraApp Flutter app — Material Design 3, theming, layout patterns, component guidelines, accessibility, and receipt-flow-specific UI patterns
license: MIT
compatibility: opencode
metadata:
  audience: developers
  category: design
  framework: flutter
  project: ahorrapp
  system: material-design-3
---

## What I do

- Apply Material Design 3 (Material You) principles to Flutter widgets
- Generate and maintain a consistent design token system (colors, typography, spacing, elevation)
- Create responsive layouts that adapt across phone sizes and orientations
- Design accessible UI components meeting WCAG 2.1 AA standards
- Implement proper loading, empty, error, and success states for all screens
- Build receipt-specific UI patterns (upload, review form, confirmation, price display)
- Ensure dark mode support with proper contrast ratios
- Guide animation and transition choices for delightful UX
- Enforce consistent form patterns for data entry and validation feedback
- Create reusable component patterns (cards, sheets, dialogs, lists, chips)

## When to use me

Use this skill when:
- Designing or building Flutter UI widgets or screens
- Creating themes, color schemes, or typography scales
- Building layouts for mobile screens (receipt upload, review, confirm, list, detail)
- Implementing form fields with validation states
- Designing loading/empty/error state widgets
- Creating reusable UI components or design system widgets
- Setting up dark mode or dynamic theming
- Building bottom sheets, dialogs, snackbars, or navigation patterns
- Designing price/currency display components
- Creating accessibility-compliant UI

## Design Token System

### Color tokens (Material Design 3)

```
// lib/app/theme/colors.dart
class AppColors {
  // Primary brand
  static const primary = Color(0xFF006495);
  static const onPrimary = Color(0xFFFFFFFF);
  static const primaryContainer = Color(0xFFC8E6FF);
  static const onPrimaryContainer = Color(0xFF001E2E);

  // Secondary (accents)
  static const secondary = Color(0xFF4F626E);
  static const onSecondary = Color(0xFFFFFFFF);

  // Tertiary (highlights, currency badges)
  static const tertiary = Color(0xFF5B6300);
  static const onTertiary = Color(0xFFFFFFFF);

  // Error
  static const error = Color(0xFFBA1A1A);
  static const onError = Color(0xFFFFFFFF);
  static const errorContainer = Color(0xFFFFDAD6);

  // Surface
  static const surface = Color(0xFFF5FAFD);
  static const onSurface = Color(0xFF06171F);
  static const surfaceVariant = Color(0xFFDCE3E9);
  static const onSurfaceVariant = Color(0xFF40484C);

  // Currency colors (Article V)
  static const usd = Color(0xFF2E7D32);  // Green for USD
  static const bs = Color(0xFF1565C0);   // Blue for Bs.
}
```

### Dark mode color tokens

```
class AppDarkColors {
  static const primary = Color(0xFF87CEFF);
  static const onPrimary = Color(0xFF00344D);
  static const surface = Color(0xFF06171F);
  static const onSurface = Color(0xFFDCE3E9);
  // ... mirrors light scheme with inverted luminance
}
```

### Typography scale (Material Design 3)

```
// lib/app/theme/typography.dart
class AppTypography {
  // Display
  static const displayLarge = TextStyle(fontSize: 57, fontWeight: FontWeight.w400);
  static const displayMedium = TextStyle(fontSize: 45, fontWeight: FontWeight.w400);
  static const displaySmall = TextStyle(fontSize: 36, fontWeight: FontWeight.w400);

  // Headline
  static const headlineLarge = TextStyle(fontSize: 32, fontWeight: FontWeight.w400);
  static const headlineMedium = TextStyle(fontSize: 28, fontWeight: FontWeight.w400);
  static const headlineSmall = TextStyle(fontSize: 24, fontWeight: FontWeight.w400);

  // Title
  static const titleLarge = TextStyle(fontSize: 22, fontWeight: FontWeight.w500);
  static const titleMedium = TextStyle(fontSize: 16, fontWeight: FontWeight.w500);
  static const titleSmall = TextStyle(fontSize: 14, fontWeight: FontWeight.w500);

  // Body
  static const bodyLarge = TextStyle(fontSize: 16, fontWeight: FontWeight.w400);
  static const bodyMedium = TextStyle(fontSize: 14, fontWeight: FontWeight.w400);
  static const bodySmall = TextStyle(fontSize: 12, fontWeight: FontWeight.w400);

  // Label
  static const labelLarge = TextStyle(fontSize: 14, fontWeight: FontWeight.w500);
  static const labelMedium = TextStyle(fontSize: 12, fontWeight: FontWeight.w500);
  static const labelSmall = TextStyle(fontSize: 11, fontWeight: FontWeight.w500);
}
```

### Spacing system (8px grid)

```
// lib/app/theme/spacing.dart
class AppSpacing {
  static const xs = 4.0;
  static const sm = 8.0;
  static const md = 16.0;
  static const lg = 24.0;
  static const xl = 32.0;
  static const xxl = 48.0;
  static const xxxl = 64.0;
}
```

### Elevation tokens

```
class AppElevation {
  static const level0 = 0.0;  // Flat surfaces
  static const level1 = 1.0;  // Cards, list items
  static const level2 = 3.0;  // Raised cards, FAB
  static const level3 = 6.0;  // Dialogs, bottom sheets
  static const level4 = 8.0;  // Nav drawer, modal sheets
}
```

### Border radius tokens

```
class AppRadius {
  static const xs = 4.0;
  static const sm = 8.0;
  static const md = 12.0;
  static const lg = 16.0;
  static const xl = 28.0;   // M3 large components
  static const full = 9999.0;  // Pills, circular
}
```

## Screen patterns

### Receipt upload screen

```
Layout:
┌─────────────────────────────┐
│  AppBar: "Subir recibo"      │
├─────────────────────────────┤
│                              │
│  ┌─────────────────────┐    │
│  │                     │    │
│  │   [Camera Icon]     │    │  ← Drag-drop zone / camera trigger
│  │   "Toma foto o      │    │
│  │    selecciona"       │    │
│  │                     │    │
│  └─────────────────────┘    │
│                              │
│  [Subir recibo] (FilledBtn)  │
│                              │
│  Recent uploads:             │
│  ┌─── Card ───┐ ┌─── Card ──┐│
│  │ 🧾 PENDING  │ │ ✅ DONE   ││
│  │ Store name  │ │ Store name││
│  └─────────────┘ └───────────┘│
└─────────────────────────────┘
```

States:
- **Idle**: Upload zone visible, button enabled
- **Uploading**: Progress indicator on button, zone disabled
- **Success**: Snackbar "Recibo subido", navigate to review
- **Duplicate**: Info dialog "Recibo ya subido", offer to view existing
- **Error**: Snackbar with retry action

### Receipt review screen (NEEDS_REVIEW)

```
Layout:
┌─────────────────────────────┐
│  AppBar: "Revisar recibo"    │
├─────────────────────────────┤
│  [Receipt thumbnail]         │
│                              │
│  Store:                      │
│  [Text field: store name]    │
│                              │
│  Date:                       │
│  [Date picker field]         │
│                              │
│  Total:                      │
│  [Text field: amount]        │
│  [Currency selector: USD/Bs]│
│                              │
│  Items:                      │
│  ┌─────────────────────────┐│
│  │ Item 1                  ││
│  │ [name] [qty] [price]    ││
│  │ [currency: USD/Bs]      ││
│  │                  [delete]│
│  └─────────────────────────┘│
│  [+ Add item]                │
│                              │
│  [Confirmar] [Rechazar]      │
└─────────────────────────────┘
```

Key UI rules:
- Every item row MUST have currency selector (Constitution Article V)
- Currency selector uses segmented button (USD | Bs.)
- Price field shows currency symbol as prefix
- Delete item uses icon button with confirmation
- Confirm button is filled (primary action)
- Reject button is outlined (secondary action)

### Receipt list screen

```
Layout:
┌─────────────────────────────┐
│  AppBar: "Mis recibos"       │
│  [Search] [Filter]           │
├─────────────────────────────┤
│  ┌─── Filter chips ────────┐│
│  │ [Todos] [Pendiente] [OK]││
│  └─────────────────────────┘│
│                              │
│  ┌─── List item ───────────┐│
│  │ 🧾  Central Market       ││
│  │     2026-06-24  $42.50  ││
│  │     [Status chip]        ││
│  └─────────────────────────┘│
│  ┌─── List item ───────────┐│
│  │ 🧾  Unknown Store        ││
│  │     2026-06-23  Bs. 150 ││
│  │     [Status chip]        ││
│  └─────────────────────────┘│
│                              │
│  FAB: [+] Upload new         │
└─────────────────────────────┘
```

Status chips:
- `PENDING`: amber chip with clock icon
- `NEEDS_REVIEW`: orange chip with edit icon
- `CONFIRMED`: green chip with check icon
- `REJECTED`: red chip with x icon

### Price display component

Every price display in the app MUST include:
1. Numeric value with 2 decimal places
2. Currency code badge (USD or Bs.)
3. Color-coded by currency (green for USD, blue for Bs.)

```
┌──────────────────┐
│  $42.50 [USD]    │   ← green badge
│  Bs. 150.00 [Bs] │   ← blue badge
└──────────────────┘
```

## Component guidelines

### Buttons

| Type         | Usage                          | Widget                       |
|--------------|--------------------------------|------------------------------|
| Filled       | Primary action (upload, confirm)| `FilledButton`              |
| Outlined     | Secondary action (reject)       | `OutlinedButton`            |
| Text         | Tertiary action (cancel)        | `TextButton`                |
| FAB          | Create/new action               | `FloatingActionButton`      |
| Icon         | Quick actions (delete, edit)    | `IconButton`                |

### Cards

- Use `Card` with `elevation: AppElevation.level1`
- Border radius: `AppRadius.lg`
- Padding: `AppSpacing.md` on all sides
- Content: title + subtitle + optional trailing widget

### Bottom sheets

- Use `showModalBottomSheet` with `useSafeArea: true`
- Border radius: `AppRadius.xl` on top corners
- Max height: 80% of screen
- Drag handle visible by default

### Dialogs

- Confirmation dialogs: `AlertDialog` with title, content, two actions
- Use `barrierDismissible: false` for destructive actions
- Action order: dismiss (left), confirm (right)

### Snackbars

- Success: green background, check icon
- Error: red background, error icon, retry action
- Info: surface color, info icon
- Duration: 4 seconds for info, 6 seconds for error

### Form fields

- Use `TextFormField` with `InputDecoration`
- Error state: red border + error text below
- Focused state: primary color border
- Disabled state: grey border + reduced opacity
- Currency field: prefix text ($ or Bs.) + suffix currency badge
- Date field: tap opens `showDatePicker`

### Loading states

- Full screen: `CircularProgressIndicator` centered
- Inline (buttons): `SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))`
- List: `Shimmer` placeholder cards
- Image: `Container` with `BoxDecoration(color: surfaceVariant)` + fade in

### Empty states

```
┌─────────────────────────────┐
│                              │
│        [Illustration]        │
│                              │
│   "No tienes recibos aún"    │
│                              │
│   "Sube tu primer recibo     │
│    para empezar a ahorrar"   │
│                              │
│   [Subir recibo]             │
│                              │
└─────────────────────────────┘
```

## Accessibility requirements

- All interactive elements minimum 48x48 dp touch target
- Text contrast ratio minimum 4.5:1 (WCAG AA)
- Large text contrast ratio minimum 3:1
- Semantic labels on all icons and images
- Screen reader support via `Semantics` widget
- Focus order follows visual order
- Focus indicator visible on all interactive elements
- Reduce motion support: check `MediaQuery.disableAnimations`

## Animation guidelines

- Page transitions: 300ms ease-in-out
- Bottom sheet open: 250ms ease-out
- Snackbar appear: 150ms ease-in
- Card tap feedback: 100ms ease-out scale(0.98)
- Loading spinner: 1000ms linear infinite
- Use `AnimatedSwitcher` for state changes
- Use `Hero` animation for receipt thumbnail → detail

## Navigation patterns

- Bottom navigation bar with 3 tabs: "Inicio", "Recibos", "Perfil"
- Nested navigation with `GoRouter` or `NavigationStack`
- Back button always available except on root tabs
- Deep linking support for receipt detail (`/receipts/{id}`)

## Theme setup

```
// lib/app/theme/app_theme.dart
ThemeData appTheme(Brightness brightness) {
  final colors = brightness == Brightness.light ? AppColors : AppDarkColors;
  final colorScheme = ColorScheme(
    brightness: brightness,
    primary: colors.primary,
    onPrimary: colors.onPrimary,
    secondary: colors.secondary,
    onSecondary: colors.onSecondary,
    tertiary: colors.tertiary,
    onTertiary: colors.onTertiary,
    error: colors.error,
    onError: colors.onError,
    surface: colors.surface,
    onSurface: colors.onSurface,
  );

  return ThemeData(
    useMaterial3: true,
    colorScheme: colorScheme,
    textTheme: TextTheme(
      displayLarge: AppTypography.displayLarge,
      // ... all text styles
    ),
    cardTheme: CardTheme(
      elevation: AppElevation.level1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadius.lg)),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        minimumSize: const Size(double.infinity, 56),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadius.full)),
      ),
    ),
  );
}
```
