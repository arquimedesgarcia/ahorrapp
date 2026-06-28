# Tasks: Flutter Mobile Application MVP

**Input**: Design documents from `/specs/004-flutter-mobile-app/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Not included in this task list. Tests follow project conventions (flutter-development skill). Add test tasks in a follow-up if needed.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this belongs to (US1, US2, US3, US4, US5, US6)
- Include exact file paths in descriptions

## Path Conventions

All paths relative to repository root. Flutter app lives under `mobile/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize Flutter project, configure dependencies, and set up build tooling

- [x] T001 Create Flutter project with `flutter create mobile` from repo root
- [x] T002 [P] Configure `mobile/pubspec.yaml` with all dependencies (flutter_riverpod, riverpod_annotation, dio, go_router, camera, image_picker, flutter_secure_storage, json_serializable, freezed_annotation, build_runner, riverpod_generator, freezed, json_annotation, flutter_lints, flutter_localizations, intl)
- [x] T003 [P] Configure `mobile/analysis_options.yaml` with flutter_lints and strict mode rules
- [x] T004 [P] Create English ARB localization file in `mobile/l10n/app_en.arb` with initial app metadata strings
- [x] T005 [P] Create Spanish (Venezuela) ARB localization file in `mobile/l10n/app_es.arb` with initial app metadata strings
- [x] T006 Run `flutter pub get` and verify project compiles with `flutter analyze` in `mobile/`

**Checkpoint**: Flutter project skeleton compiles cleanly. All dependencies resolved.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

### Theme & Design System

- [x] T007 [P] Define Material Design 3 color tokens (light + dark) in `mobile/lib/app/theme/colors.dart`
- [x] T008 [P] Define typography scale tokens in `mobile/lib/app/theme/typography.dart`
- [x] T009 [P] Define spacing, elevation, and border radius tokens in `mobile/lib/app/theme/spacing.dart`
- [x] T010 Build light and dark ThemeData with MD3 ColorScheme in `mobile/lib/app/theme/app_theme.dart`

### Core API Layer

- [x] T011 [P] Define typed API exception classes (NetworkException, AuthException, ServerException, ValidationException) in `mobile/lib/core/api/api_exceptions.dart`
- [x] T012 [P] Implement JWT secure storage wrapper (read/write/delete/clear) in `mobile/lib/core/storage/secure_storage.dart`
- [x] T013 Implement AuthInterceptor (attach Bearer token, handle 401 redirect) in `mobile/lib/core/api/auth_interceptor.dart`
- [x] T014 Implement RetryInterceptor (3 retries with 1s/2s/4s backoff for 5xx and network errors) in `mobile/lib/core/api/retry_interceptor.dart`
- [x] T015 Build Dio ApiClient singleton with interceptors, base URL from environment, and timeouts in `mobile/lib/core/api/api_client.dart`

### Routing

- [x] T016 Implement GoRouter with ShellRoute (bottom nav) and auth redirect guard in `mobile/lib/app/router.dart`

### App Entry Point

- [x] T017 Wire MaterialApp with theme, router, localizations, and ProviderScope in `mobile/lib/app/app.dart`
- [x] T018 Configure main() with WidgetsFlutterBinding, system UI overlays, and runApp in `mobile/lib/main.dart`

### Utility Helpers

- [x] T019 [P] Implement currency display helpers (format with symbol, currency badge colors) in `mobile/lib/core/utils/currency_utils.dart`

### Shared Widgets

- [x] T020 [P] Build reusable LoadingIndicator (centered spinner + optional message) in `mobile/lib/shared/widgets/loading_indicator.dart`
- [x] T021 [P] Build reusable EmptyState (illustration + title + subtitle + optional action button) in `mobile/lib/shared/widgets/empty_state.dart`
- [x] T022 [P] Build reusable ErrorState (error message + retry button) in `mobile/lib/shared/widgets/error_state.dart`
- [x] T023 [P] Build CurrencySelector segmented button (USD | Bs.) in `mobile/lib/shared/widgets/currency_selector.dart`
- [x] T024 [P] Build PointsBadge (points icon + count) in `mobile/lib/shared/widgets/points_badge.dart`
- [x] T025 Build AppScaffold with bottom navigation bar (Home, Receipts, Search, Profile) in `mobile/lib/shared/widgets/app_scaffold.dart`

**Checkpoint**: Foundation ready — app launches with working theme, routing, DI, and shared widgets. All 6 user stories can now start implementation.

---

## Phase 3: User Story 1 - Onboarding and Account Creation (Priority: P1) 🎯 MVP

**Goal**: New user installs app, completes onboarding, registers, logs in, and session persists across restarts.

**Independent Test**: Clean install → onboarding screens → register with email/password → logout → login → close and reopen app (session restored).

### Implementation for User Story 1

- [x] T026 [P] [US1] Generate Dart data models (LoginRequest, RegisterRequest, AuthResponse, UserProfile) from auth contract in `mobile/lib/features/auth/data/auth_models.dart`
- [x] T027 [P] [US1] Implement AuthApiClient (register, login, getMe) consuming contracts in `mobile/lib/features/auth/data/auth_api_client.dart`
- [x] T028 [US1] Create AuthEntity and AuthRepository (login, register, logout, restoreSession, isAuthenticated) in `mobile/lib/features/auth/domain/auth_repository.dart`
- [x] T029 [US1] Implement AuthNotifier (Riverpod AsyncNotifier) with states: unauthenticated, loading, authenticated, error in `mobile/lib/features/auth/presentation/auth_notifier.dart`
- [x] T030 [US1] Build OnboardingPage with PageView, value proposition slides, and skip/next controls in `mobile/lib/features/auth/presentation/onboarding_page.dart`
- [x] T031 [US1] Build RegisterPage with email, password, displayName fields and validation in `mobile/lib/features/auth/presentation/register_page.dart`
- [x] T032 [US1] Build LoginPage with email, password fields, error display, and link to register in `mobile/lib/features/auth/presentation/login_page.dart`
- [x] T033 [US1] Update router.dart with auth redirect guard (unauthenticated → /login, authenticated with no onboarding done → /onboarding)

**Checkpoint**: US1 independently functional — users can onboard, register, login, and session persists.

---

## Phase 4: User Story 2 - Capture and Upload Receipt (Priority: P1)

**Goal**: Authenticated user taps scan button, captures receipt photo, uploads with progress and retry, sees processing status.

**Independent Test**: Tap scan → camera opens → take photo → upload progress shown → receipt ID received → status "PENDING" displayed.

### Implementation for User Story 2

- [x] T034 [P] [US2] Generate ReceiptUploadResponse model from receipt API contract in `mobile/lib/features/receipt/data/receipt_models.dart`
- [x] T035 [US2] Implement ReceiptApiClient (upload multipart, getById, confirm) in `mobile/lib/features/receipt/data/receipt_api_client.dart`
- [x] T036 [US2] Create ReceiptEntity and ReceiptRepository (upload, getDetail, confirm) in `mobile/lib/features/receipt/domain/receipt_repository.dart`
- [x] T037 [US2] Implement UploadNotifier (Riverpod AsyncNotifier) with states: idle, capturing, uploading(progress), retrying(attempt), success, duplicate, error in `mobile/lib/features/receipt/presentation/receipt_upload_notifier.dart`
- [x] T038 [US2] Build camera capture screen with receipt framing overlay guide in `mobile/lib/features/receipt/presentation/receipt_camera_page.dart`
- [x] T039 [US2] Build ReceiptUploadPage with scan button, upload progress indicator, retry status display, and duplicate notification in `mobile/lib/features/receipt/presentation/receipt_upload_page.dart`
- [x] T040 [US2] Build ReceiptStatusChip widget (PENDING amber, NEEDS_REVIEW orange, CONFIRMED green, REJECTED red) in `mobile/lib/features/receipt/presentation/widgets/receipt_status_chip.dart`
- [x] T041 [US2] Integrate camera permission handling: request permission, show denied message with settings link

**Checkpoint**: US2 independently functional — users can capture and upload receipts with progress tracking.

---

## Phase 5: User Story 3 - Review and Edit Receipt Summary (Priority: P1)

**Goal**: After processing, user sees editable summary (store, date, total, line items) and can modify any field.

**Independent Test**: Upload receipt → wait for NEEDS_REVIEW → open summary → edit store name, change line item price → verify changes reflected.

### Implementation for User Story 3

- [x] T042 [P] [US3] Add ReceiptDetail, StoreInfo, ReceiptItem models to `mobile/lib/features/receipt/data/receipt_models.dart`
- [x] T043 [US3] Implement polling mechanism for receipt status (poll GET /api/v1/receipts/{id} every 5s until NEEDS_REVIEW) in `mobile/lib/features/receipt/domain/receipt_repository.dart`
- [x] T044 [US3] Implement ReviewNotifier (Riverpod AsyncNotifier) with states: loading, ready(detail), saving in `mobile/lib/features/receipt/presentation/receipt_review_notifier.dart`
- [x] T045 [US3] Build PriceDisplay widget (currency badge + formatted amount + currency color) in `mobile/lib/features/receipt/presentation/widgets/price_display.dart`
- [x] T046 [US3] Build ReceiptItemForm widget (editable raw_text, quantity, unit_price, currency) in `mobile/lib/features/receipt/presentation/widgets/receipt_item_form.dart`
- [x] T047 [US3] Build ReceiptReviewPage with editable store (name, branch, address), date picker, total field, and list of ReceiptItemForm in `mobile/lib/features/receipt/presentation/receipt_review_page.dart`
- [x] T048 [US3] Add "Add item" functionality to dynamically append new line items to review form
- [x] T049 [US3] Handle partial OCR results: display available fields, leave empty fields editable with placeholder hints

**Checkpoint**: US3 independently functional — users can review and edit receipt summary fields.

---

## Phase 6: User Story 4 - Confirm Receipt and Earn Points (Priority: P1)

**Goal**: User confirms corrected receipt, sees points earned, receipt transitions to CONFIRMED.

**Independent Test**: Edit receipt → tap confirm → missing currency blocked → fix currency → confirm → points earned screen shown → receipt status changes to CONFIRMED.

### Implementation for User Story 4

- [x] T050 [P] [US4] Add ConfirmReceiptRequest and ConfirmReceiptResponse models to `mobile/lib/features/receipt/data/receipt_models.dart`
- [x] T051 [US4] Implement client-side currency validation (all items must have non-null currency) in `mobile/lib/features/receipt/domain/receipt_repository.dart`
- [x] T052 [US4] Implement ConfirmNotifier (Riverpod AsyncNotifier) with states: idle, confirming, confirmed(points), error in `mobile/lib/features/receipt/presentation/receipt_confirm_notifier.dart`
- [x] T053 [US4] Build ReceiptConfirmPage with points earned celebration display and receipt detail summary in `mobile/lib/features/receipt/presentation/receipt_confirm_page.dart`
- [x] T054 [US4] Wire confirm button on review page to validate currency → submit confirmation → navigate to confirm page on success, show inline errors on failure
- [x] T055 [US4] Handle confirmation retry: preserve user edits on network failure, show retry button
- [x] T056 [US4] Implement receipt list page showing all user receipts with status chips in `mobile/lib/features/receipt/presentation/receipt_list_page.dart`

**Checkpoint**: US4 independently functional — users can confirm receipts and see earned points.

---

## Phase 7: User Story 5 - Search Products and Find Cheapest Store (Priority: P2)

**Goal**: User searches product by name and sees stores ranked from cheapest to most expensive.

**Independent Test**: Search "Arroz" → results show stores ordered by price ascending → cheapest first → each entry shows price + currency.

### Implementation for User Story 5

- [x] T057 [P] [US5] Generate ProductSearchResponse, ProductSearchResult, StorePriceEntry models from ranking API contract in `mobile/lib/features/ranking/data/ranking_models.dart`
- [x] T058 [US5] Implement RankingApiClient (search) in `mobile/lib/features/ranking/data/ranking_api_client.dart`
- [x] T059 [US5] Create RankingRepository in `mobile/lib/features/ranking/domain/ranking_repository.dart`
- [x] T060 [US5] Implement SearchNotifier (Riverpod AsyncNotifier) with states: idle, loading, results, empty, error in `mobile/lib/features/ranking/presentation/product_search_notifier.dart`
- [x] T061 [US5] Build StorePriceCard widget (store name, branch, price with currency badge, sample count) in `mobile/lib/features/ranking/presentation/widgets/store_price_card.dart`
- [x] T062 [US5] Build RankingList widget (sorted store cards with cheapest-first ordering) in `mobile/lib/features/ranking/presentation/widgets/ranking_list.dart`
- [x] T063 [US5] Build ProductSearchPage with search bar, results list, empty state, and error state in `mobile/lib/features/ranking/presentation/product_search_page.dart`

**Checkpoint**: US5 independently functional — users can search products and see store price rankings.

---

## Phase 8: User Story 6 - View Profile and Accumulated Points (Priority: P2)

**Goal**: User views profile with display name, email, and total accumulated points including recent transaction history.

**Independent Test**: Open profile → verify display name/email/total points match account → confirm a receipt → profile points updated.

### Implementation for User Story 6

- [x] T064 [P] [US6] Generate PointsResponse, PointsTransaction models from profile API contract in `mobile/lib/features/profile/data/profile_models.dart`
- [x] T065 [US6] Implement ProfileApiClient (getPoints) in `mobile/lib/features/profile/data/profile_api_client.dart`
- [x] T066 [US6] Create ProfileRepository in `mobile/lib/features/profile/domain/profile_repository.dart`
- [x] T067 [US6] Implement ProfileNotifier (Riverpod AsyncNotifier) with states: loading, ready(profile, points), error in `mobile/lib/features/profile/presentation/profile_notifier.dart`
- [x] T068 [US6] Build ProfilePage with display name, email, points total (PointsBadge), recent transactions list, and logout action in `mobile/lib/features/profile/presentation/profile_page.dart`

**Checkpoint**: US6 independently functional — users can view profile and accumulated points.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and final quality

- [x] T069 [P] Complete Spanish (es) localization: translate all user-facing strings in `mobile/l10n/app_es.arb`
- [x] T070 [P] Verify dark mode rendering on all screens (no hardcoded colors, proper contrast)
- [x] T071 [P] Add Semantics labels to all interactive elements for screen reader accessibility
- [x] T072 Add Hero animation between receipt list thumbnail and receipt review screen
- [x] T073 Verify all touch targets meet minimum 48x48 dp accessibility requirement
- [x] T074 Run `flutter analyze` and fix all warnings/errors
- [x] T075 Run `dart format --set-exit-if-changed lib/` and verify no format violations
- [x] T076 Run `dart run build_runner build --delete-conflicting-outputs` and verify generated code compiles
- [x] T077 Validate all 10 quickstart scenarios from `quickstart.md` pass end-to-end
- [x] T078 Implement app icon and splash screen for iOS and Android

**Checkpoint**: App polished, accessible, localized, and ready for demo/deployment.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Phase 2 — provides auth for all other stories
- **User Story 2 (Phase 4)**: Depends on Phase 2 + US1 (needs auth)
- **User Story 3 (Phase 5)**: Depends on US2 (needs uploaded receipt)
- **User Story 4 (Phase 6)**: Depends on US3 (needs reviewed receipt)
- **User Story 5 (Phase 7)**: Depends on Phase 2 + US1 (needs auth, independent of receipt flow)
- **User Story 6 (Phase 8)**: Depends on US1 (auth) + US4 (points from confirmations)
- **Polish (Phase 9)**: Depends on all desired user stories complete

### User Story Dependency Graph

```
Phase 1 (Setup)
    │
    ▼
Phase 2 (Foundational)
    │
    ├──► US1 (Auth) ──────────────────────┬──► US5 (Search) ──┐
    │         │                           │                    │
    │         ▼                           │                    │
    │    US2 (Upload)                     │                    │
    │         │                           │                    │
    │         ▼                           │                    │
    │    US3 (Review)                     │                    │
    │         │                           │                    │
    │         ▼                           │                    │
    │    US4 (Confirm ─── points) ────────┼──► US6 (Profile) ──┤
    │                                     │                    │
    └─────────────────────────────────────┴────────────────────┘
                                                              │
                                                              ▼
                                                     Phase 9 (Polish)
```

### Within Each User Story

- Models/API clients first (can be parallel)
- Repository depends on API client
- Notifier depends on repository
- Pages/widgets depend on notifier
- Integration last

### Parallel Opportunities

- **Phase 1**: T002, T003, T004, T005 can run in parallel
- **Phase 2**: T007-T009, T011-T012, T019-T024 can all run in parallel within the phase
- **Phase 3 (US1)**: T026, T027 can run in parallel; T030, T031, T032 can run in parallel
- **Phase 4 (US2)**: T034 can run in parallel with T040
- **Phase 5 (US3)**: T042, T045, T046 can run in parallel
- **Phase 6 (US4)**: T050 can run in parallel with T053
- **Phase 7 (US5)**: T057 can run in parallel with T061
- **Phase 8 (US6)**: T064 can run in parallel with T068
- **Cross-story**: US5 and US6 can be built in parallel once US1 is complete
- **Phase 9**: T069, T070, T071 can run in parallel

---

## Parallel Example: User Story 1 (Auth)

```bash
# Launch models and API client together:
Task: "Generate Dart data models in mobile/lib/features/auth/data/auth_models.dart"
Task: "Implement AuthApiClient in mobile/lib/features/auth/data/auth_api_client.dart"

# After models + API client done, launch pages together:
Task: "Build OnboardingPage in mobile/lib/features/auth/presentation/onboarding_page.dart"
Task: "Build RegisterPage in mobile/lib/features/auth/presentation/register_page.dart"
Task: "Build LoginPage in mobile/lib/features/auth/presentation/login_page.dart"
```

---

## Implementation Strategy

### MVP First (US1 + US2 + US3 + US4)

1. Complete Phase 1: Setup → Flutter project compiles
2. Complete Phase 2: Foundational → Theme, API client, routing, shared widgets ready
3. Complete Phase 3: US1 → Users can register and login
4. Complete Phase 4: US2 → Users can scan and upload receipts
5. Complete Phase 5: US3 → Users can review and edit receipt summaries
6. Complete Phase 6: US4 → Users can confirm receipts and earn points
7. **STOP and VALIDATE**: Test full receipt flow end-to-end (onboard → scan → review → confirm → points)
8. Demo the core value loop

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (Auth) → Test independently → App with auth
3. Add US2 (Upload) → Test independently → App with receipt capture
4. Add US3 (Review) → Test independently → App with editable review
5. Add US4 (Confirm) → Test independently → **Core MVP complete** (receipt → points loop)
6. Add US5 (Search) → Test independently → App with savings discovery
7. Add US6 (Profile) → Test independently → App with gamification visibility
8. Polish → Production-ready app

### Suggested MVP Scope

**Core MVP**: Phases 1-6 (US1 through US4) — delivers the receipt-to-points value loop.
**Full MVP**: Phases 1-9 — complete product with search, ranking, and profile.

---

## Notes

- All Dart code, identifiers, comments in English (Constitution Article IX)
- All user-facing strings in Spanish (Venezuela market) with English fallback
- Every price display MUST include currency badge (Constitution Article V)
- API client must consume only published contracts (Constitution Article IV.3)
- JWT stored via flutter_secure_storage, never SharedPreferences (Constitution Article VII)
- Run `dart run build_runner build --delete-conflicting-outputs` after creating/editing models with json_serializable or freezed annotations
- Commit after each task or logical group of parallel tasks
- Stop at any checkpoint to validate story independently before proceeding
