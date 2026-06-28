---
name: flutter-development
description: Flutter/Dart mobile app development for AhorraApp — architecture, API consumption, state management, testing, and constitution compliance
license: MIT
compatibility: opencode
metadata:
  audience: developers
  framework: flutter
  language: dart
  project: ahorrapp
---

## What I do

- Guide Flutter project structure following feature-first architecture
- Ensure Flutter app consumes ONLY published `/api/v1/...` REST contracts (Constitution Article IV)
- Enforce clean separation: data layer (API clients) vs domain layer (models/services) vs presentation layer (widgets/blocs)
- Generate Dart model classes from Go backend entity definitions and API contracts
- Create and maintain API client wrappers that match `specs/*/contracts/` exactly
- Implement state management with Riverpod or Bloc pattern consistently
- Write widget tests, integration tests, and golden tests
- Ensure receipt upload, review, and confirmation flows match backend state machine (`PENDING -> NEEDS_REVIEW -> CONFIRMED | REJECTED`)
- Validate that every price/observation UI element displays mandatory currency (Constitution Article V)
- Support image upload with multipart/form-data matching backend `POST /api/v1/receipts` contract
- Handle async processing states (loading, pending, needs review, confirmed) in UI

## When to use me

Use this skill when:
- Creating or modifying Flutter/Dart code in a `mobile/` or `app/` directory
- Building mobile screens that consume the AhorraApp backend API
- Generating Dart models from backend contracts or entities
- Implementing receipt upload, review, or confirmation flows in Flutter
- Writing Flutter tests (widget, integration, unit)
- Setting up Flutter project structure or state management
- Integrating with `/api/v1/...` endpoints from the Go backend

## Project conventions

### Architecture (Constitution Article I alignment)

```
mobile/
├── lib/
│   ├── main.dart
│   ├── app/                    # App-level setup, routing, theme
│   │   ├── app.dart
│   │   ├── router.dart
│   │   └── theme.dart
│   ├── features/               # Feature-first modules
│   │   ├── receipt/
│   │   │   ├── data/
│   │   │   │   ├── receipt_api_client.dart
│   │   │   │   └── receipt_models.dart
│   │   │   ├── domain/
│   │   │   │   ├── receipt_entity.dart
│   │   │   │   └── receipt_repository.dart
│   │   │   └── presentation/
│   │   │       ├── receipt_upload_page.dart
│   │   │       ├── receipt_review_page.dart
│   │   │       └── receipt_confirm_page.dart
│   │   └── auth/
│   ├── core/                   # Shared utilities, constants
│   │   ├── api/
│   │   │   ├── api_client.dart
│   │   │   └── api_exceptions.dart
│   │   ├── config/
│   │   │   └── environment.dart
│   │   └── utils/
│   └── shared/                 # Reusable widgets
│       └── widgets/
├── test/
│   ├── unit/
│   ├── widget/
│   └── integration/
└── pubspec.yaml
```

### API contract consumption rules

1. Flutter app MUST consume ONLY endpoints documented in `specs/*/contracts/`
2. Every API response model MUST match the contract JSON shapes exactly
3. Undocumented fields MUST NOT be used (Constitution Article IV.3)
4. Base URL comes from environment config (`API_BASE_URL`), never hardcoded
5. Auth token sent via `Authorization: Bearer <token>` header
6. `X-User-ID` header used for testing/local dev only

### Receipt flow state mapping

Backend state machine maps to Flutter UI states:

| Backend status    | Flutter UI state     | User action              |
|-------------------|----------------------|--------------------------|
| `PENDING`         | Loading spinner      | Wait                     |
| `NEEDS_REVIEW`    | Editable form        | Edit + confirm/reject    |
| `CONFIRMED`       | Success view         | View only                |
| `REJECTED`        | Rejection view       | Re-upload or dismiss     |

### Currency display rules (Constitution Article V)

- Every price display MUST include currency code (`USD` or `Bs.`)
- Currency selector MUST be present in every item edit form
- No price observation can be submitted without currency
- Averages are per-currency; never mix in UI displays

### State management

Use Riverpod as default state management:
- `Notifier`/`AsyncNotifier` for feature state
- `Provider` for dependencies (API clients, repositories)
- `ConsumerWidget` / `ConsumerStatefulWidget` for UI

### Testing requirements (Constitution Article III)

- Unit tests for repositories and API client models
- Widget tests for key screens (upload, review, confirm)
- Integration tests for critical user journeys
- Golden tests for consistent UI rendering

### Code style

- Follow `dart format` defaults
- Use `analysis_options.yaml` with `flutter_lints` package
- Prefer `const` constructors
- Use null safety properly (`?` only when genuinely nullable)
- File names: `snake_case.dart`
- Class names: `PascalCase`
