# Implementation Plan: Flutter Mobile Application MVP

**Branch**: `004-flutter-mobile-app` | **Date**: 2026-06-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-flutter-mobile-app/spec.md`

## Summary

Deliver the AhorraApp Flutter mobile application (iOS & Android) consuming the existing `/api/v1` REST API. The MVP includes onboarding + auth, receipt photo capture & upload with offline-tolerant retry, editable OCR summary review & confirmation with points feedback, product search ranked by cheapest store, and user profile with accumulated points. Architecture follows feature-first Clean Architecture in the Flutter layer mirroring the backend's separation of concerns, with Riverpod for state management, Dio for HTTP with retry interceptors, and Material Design 3 theming.

## Technical Context

**Language/Version**: Dart 3.5+ / Flutter 3.24+

**Primary Dependencies**:
- State management: flutter_riverpod, riverpod_annotation
- HTTP client: dio (retry, interceptors, multipart upload)
- Routing: go_router
- Camera: camera (device camera), image_picker (gallery fallback)
- Secure storage: flutter_secure_storage (JWT persistence)
- Code generation: json_serializable, freezed, riverpod_generator, build_runner
- Testing: flutter_test, mocktail, golden_toolkit

**Storage**: N/A (client-only; backend manages PostgreSQL, Redis, S3)

**Testing**:
- Unit tests: API client models, repositories, state notifiers
- Widget tests: key screens (onboarding, upload, review, confirm, search, profile)
- Golden tests: critical UI components (price display, status chips)
- Integration tests: end-to-end user journeys

**Target Platform**: iOS 16+ / Android 8.0 (API 26+)

**Project Type**: Mobile application (Flutter cross-platform)

**Performance Goals**:
- 60 fps UI rendering on mid-range devices
- Screen transitions under 300ms
- Image upload p95 under 10s on stable network
- Product search results displayed under 3s

**Constraints**:
- Consume only published `/api/v1` contracts (Constitution Article IV.3)
- Every price display MUST include currency (Constitution Article V.1)
- JWT token-based auth with secure local storage (Constitution Article VII.1)
- Offline-tolerant upload: retry on transient network failures
- Material Design 3 theming with light/dark mode support
- English for all code, identifiers, and documentation (Constitution Article IX)

**Scale/Scope**: 7 screens, ~20 custom widgets, ~10 API endpoints consumed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Article I — Clean Architecture in the Go backend**
  - Not directly applicable to Flutter client, but the Flutter app mirrors Clean Architecture principles: data layer (API clients, models), domain layer (entities, repositories), presentation layer (screens, widgets, state).
  - No backend infrastructure leaks into Flutter code.
  - **Status**: PASS.

- **Article IV — Explicit, versioned contracts**
  - Flutter app consumes ONLY endpoints documented in `specs/*/contracts/receipt-api-contract.md` plus auth/ranking contracts defined in this plan.
  - Every API response model matches contract JSON shapes exactly.
  - Undocumented fields are never consumed (Constitution Article IV.3).
  - **Status**: PASS.

- **Article V — Data, currency, and normalization**
  - Every price observation in Flutter UI displays mandatory currency (USD or Bs.).
  - Currency selector present in every item edit form.
  - Confirmation rejects when any currency is missing (validated client-side + server-side).
  - **Status**: PASS.

- **Article VII — Minimal MVP security**
  - JWT token stored in platform secure storage (flutter_secure_storage).
  - Token attached to all authenticated requests via `Authorization: Bearer <token>`.
  - Expired tokens trigger redirect to login.
  - **Status**: PASS.

- **Article VIII — Ready to grow, without over-engineering**
  - Feature-first architecture allows adding supermarket API screens and subscription flows without touching core receipt/ranking modules.
  - No over-engineering: simple state management, no unnecessary abstractions.
  - **Status**: PASS.

- **Article IX — Working language**
  - All Dart code, identifiers, comments, and documentation in English.
  - **Status**: PASS.

**Gate Result**: PASS — no constitutional deviations identified.

## Project Structure

### Documentation (this feature)

```text
specs/004-flutter-mobile-app/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── auth-api-contract.md
│   ├── receipt-api-contract.md
│   ├── ranking-api-contract.md
│   └── profile-api-contract.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
mobile/
├── lib/
│   ├── main.dart                    # App entry point
│   ├── app/
│   │   ├── app.dart                 # MaterialApp with theme/routing
│   │   ├── router.dart              # GoRouter configuration
│   │   └── theme/
│   │       ├── app_theme.dart       # Light/dark ThemeData
│   │       ├── colors.dart          # MD3 color tokens
│   │       ├── typography.dart      # Text styles
│   │       └── spacing.dart         # Spacing/elevation/radius tokens
│   ├── core/
│   │   ├── api/
│   │   │   ├── api_client.dart      # Dio instance with interceptors
│   │   │   ├── api_exceptions.dart  # Typed API error classes
│   │   │   └── retry_interceptor.dart
│   │   ├── storage/
│   │   │   └── secure_storage.dart  # JWT persistence
│   │   └── utils/
│   │       └── currency_utils.dart  # Currency display helpers
│   ├── features/
│   │   ├── auth/
│   │   │   ├── data/
│   │   │   │   ├── auth_api_client.dart
│   │   │   │   └── auth_models.dart
│   │   │   ├── domain/
│   │   │   │   ├── auth_entity.dart
│   │   │   │   └── auth_repository.dart
│   │   │   └── presentation/
│   │   │       ├── onboarding_page.dart
│   │   │       ├── login_page.dart
│   │   │       ├── register_page.dart
│   │   │       └── auth_notifier.dart
│   │   ├── receipt/
│   │   │   ├── data/
│   │   │   │   ├── receipt_api_client.dart
│   │   │   │   └── receipt_models.dart
│   │   │   ├── domain/
│   │   │   │   ├── receipt_entity.dart
│   │   │   │   └── receipt_repository.dart
│   │   │   └── presentation/
│   │   │       ├── receipt_upload_page.dart
│   │   │       ├── receipt_upload_notifier.dart
│   │   │       ├── receipt_review_page.dart
│   │   │       ├── receipt_review_notifier.dart
│   │   │       ├── receipt_confirm_page.dart
│   │   │       └── widgets/
│   │   │           ├── receipt_status_chip.dart
│   │   │           ├── receipt_item_form.dart
│   │   │           └── price_display.dart
│   │   ├── ranking/
│   │   │   ├── data/
│   │   │   │   ├── ranking_api_client.dart
│   │   │   │   └── ranking_models.dart
│   │   │   ├── domain/
│   │   │   │   └── ranking_repository.dart
│   │   │   └── presentation/
│   │   │       ├── product_search_page.dart
│   │   │       ├── product_search_notifier.dart
│   │   │       └── widgets/
│   │   │           ├── store_price_card.dart
│   │   │           └── ranking_list.dart
│   │   └── profile/
│   │       ├── data/
│   │       │   ├── profile_api_client.dart
│   │       │   └── profile_models.dart
│   │       ├── domain/
│   │       │   └── profile_repository.dart
│   │       └── presentation/
│   │           ├── profile_page.dart
│   │           └── profile_notifier.dart
│   └── shared/
│       └── widgets/
│           ├── app_scaffold.dart    # Common scaffold with nav bar
│           ├── loading_indicator.dart
│           ├── empty_state.dart
│           ├── error_state.dart
│           ├── currency_selector.dart
│           └── points_badge.dart
├── test/
│   ├── unit/
│   │   ├── auth/
│   │   ├── receipt/
│   │   ├── ranking/
│   │   └── profile/
│   ├── widget/
│   │   ├── auth/
│   │   ├── receipt/
│   │   ├── ranking/
│   │   └── profile/
│   └── integration/
│       └── receipt_flow_test.dart
├── pubspec.yaml
├── analysis_options.yaml
└── l10n/
    ├── app_en.arb                # English strings
    └── app_es.arb                # Spanish strings (Venezuela market)
```

**Structure Decision**: Feature-first Flutter architecture mirroring backend's separation: `data/` (API clients, models), `domain/` (entities, repositories), `presentation/` (screens, widgets, state notifiers). Each feature is self-contained. Shared widgets and core infrastructure live at the top level. This satisfies constitution Article VIII (ready to grow) and keeps features independently testable.

## Phase 0: Research

1. Riverpod state management patterns for async flows (upload, polling, search)
2. Dio HTTP client configuration with JWT interceptor and retry strategy
3. Camera integration with photo capture for receipt scanning
4. Multipart/form-data image upload with progress tracking
5. JWT secure storage and session restoration flow
6. GoRouter nested navigation with bottom navigation bar and deep linking
7. Offline-tolerant upload retry with persistent retry queue

Output: `research.md`

## Phase 1: Design and Contracts

1. Define Flutter-side data models mirroring API contracts in `data-model.md`
2. Define auth API contract in `contracts/auth-api-contract.md`
3. Reference existing receipt API contract from `specs/003-receipt-ocr-flow/contracts/`
4. Define ranking/search API contract in `contracts/ranking-api-contract.md`
5. Define profile/points API contract in `contracts/profile-api-contract.md`
6. Define validation flow and setup in `quickstart.md`

Output: design docs ready for `/speckit.tasks`

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
