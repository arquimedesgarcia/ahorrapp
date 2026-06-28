# Research: Flutter Mobile Application MVP

## Decision 1: Riverpod for State Management

- **Decision**: Use Riverpod with code generation (`riverpod_generator`, `riverpod_annotation`) as the primary state management solution.
- **Why**:
  - Compile-time safety with generated providers.
  - `AsyncNotifier` pattern maps cleanly to async flows (upload, polling, search).
  - Provider dependency injection eliminates manual DI wiring.
  - Project conventions (Flutter skill) mandate Riverpod.
- **Alternatives considered**:
  - Bloc: more boilerplate, two-class pattern (Event + Bloc) adds overhead without benefit for this app's complexity. Rejected.
  - Provider: lacks async-first design and code generation. Rejected.

### Provider Architecture

```
                    ┌────────────────┐
                    │  ApiClient     │ (Singleton Provider)
                    │  (Dio)         │
                    └───────┬────────┘
                            │ injects via Ref
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │AuthApiClient │ │ReceiptApiCl..│ │RankingApiCl..│
    └──────┬───────┘ └──────┬───────┘ └──────┬───────┘
           │                │                │
           ▼                ▼                ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │AuthRepository│ │ReceiptRepo.. │ │RankingRepo.. │
    └──────┬───────┘ └──────┬───────┘ └──────┬───────┘
           │                │                │
           ▼                ▼                ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │AuthNotifier  │ │UploadNotifier│ │SearchNotifier│
    │(StateNotifier│ │(AsyncNotifier│ │(AsyncNotifier│
    │ Provider)    │ │ Provider)    │ │ Provider)    │
    └──────────────┘ └──────────────┘ └──────────────┘
```

## Decision 2: Dio HTTP Client with Retry Interceptor

- **Decision**: Use Dio with custom interceptors for JWT auth, retry on transient failures, and error normalization.
- **Why**:
  - Dio supports multipart upload with progress callbacks (required for receipt scan).
  - Interceptor chain enables clean separation of auth, retry, and logging concerns.
  - Built-in `RetryInterceptor` pattern with exponential backoff.
  - Cancel tokens for request lifecycle management.
- **Configuration**:
  - Base URL from environment (`API_BASE_URL`, default `http://10.0.2.2:8080` for Android emulator, `http://localhost:8080` for iOS simulator).
  - Default timeout: 30s connect, 60s receive.
  - Retry policy: 3 retries with 1s/2s/4s backoff for 5xx and network errors.
- **Alternatives considered**:
  - `http` package: lacks interceptor pattern and multipart progress. Rejected.
  - Retrofit: adds code generation complexity without significant benefit at MVP scale. Rejected.

### Interceptor Chain

```
Request → AuthInterceptor → RetryInterceptor → LogInterceptor → Network
                                  ↕ (on error)
                            Retry with backoff
```

### Interceptor Details

- **AuthInterceptor**: Reads JWT from secure storage, attaches `Authorization: Bearer <token>` header. On 401 response, clears token and triggers navigation to login.
- **RetryInterceptor**: Retries on `DioException` with type `connectionError`, `connectionTimeout`, `receiveTimeout`, or `badResponse` with status >= 500. Max 3 retries with exponential backoff (1s, 2s, 4s). Increments retry count in progress notifier.
- **LogInterceptor**: Development-only request/response logging.

## Decision 3: Camera Integration for Receipt Capture

- **Decision**: Use `camera` plugin for in-app camera capture with `image_picker` as gallery fallback.
- **Why**:
  - `camera` plugin provides full camera control within the app (required for guided receipt framing).
  - `image_picker` provides system camera/gallery sheet as fallback for devices with camera plugin incompatibility.
  - Both support Android and iOS.
- **Implementation pattern**:
  - Primary: In-app camera with overlay guide frame for receipt positioning.
  - Fallback: `ImagePicker().takePhoto()` when camera initialization fails.
  - Captured image stored as temporary file, uploaded via multipart, then deleted.
- **Alternatives considered**:
  - `image_picker` only: simpler but no guided capture experience. Rejected.
  - CameraX (Android) / AVFoundation (iOS) directly: requires platform channels, excessive complexity. Rejected.

## Decision 4: Multipart/Form-Data Upload with Progress

- **Decision**: Use Dio's `FormData` and `onSendProgress` callback for receipt image uploads.
- **Why**:
  - Maps to backend `POST /api/v1/receipts` contract (`multipart/form-data`).
  - `onSendProgress` provides real upload progress (bytes sent / total).
  - Cancel token allows user to abort in-progress upload.
- **Flow**:
  1. User captures/takes photo → temp file path.
  2. `FormData.fromMap({'image': await MultipartFile.fromFile(path)})`.
  3. POST with `onSendProgress` updating `UploadNotifier` state.
  4. On success: receipt ID returned, polling for processing status begins.
  5. On network error: retry interceptor handles; if exhausted, notify user with retry button.

## Decision 5: JWT Secure Storage and Session Restoration

- **Decision**: Use `flutter_secure_storage` for JWT token persistence with in-memory cache.
- **Why**:
  - Platform-native secure storage (Keychain on iOS, EncryptedSharedPreferences on Android).
  - Survives app restarts for session restoration (FR-004).
  - Simple key-value API.
- **Session flow**:
  - Login/Register success → store JWT in secure storage + in-memory provider.
  - App launch → read JWT from secure storage → validate (decode expiry) → restore session or redirect to login.
  - 401 response → clear token, redirect to login.
  - Logout → clear secure storage + in-memory cache.
- **Alternatives considered**:
  - SharedPreferences: not encrypted, rejected for security (Constitution Article VII.1).
  - Hive with encryption: heavier than needed for single token. Rejected.

## Decision 6: GoRouter with Bottom Navigation

- **Decision**: Use GoRouter with `ShellRoute` and bottom navigation bar for top-level navigation.
- **Why**:
  - Declarative routing with type-safe path parameters.
  - `ShellRoute` enables persistent bottom navigation bar across tabs.
  - Redirect guards for auth state.
  - Deep linking support for future notification integration.
- **Route structure**:
  ```
  /onboarding          → OnboardingPage
  /login               → LoginPage
  /register            → RegisterPage
  / (shell)            → ScaffoldWithNavBar
    /home              → HomePage (main screen + scan button)
    /receipts          → ReceiptsListPage
      /receipts/:id    → ReceiptReviewPage
      /receipts/:id/confirm → ReceiptConfirmPage
    /search            → ProductSearchPage
    /profile           → ProfilePage
  ```
- **Auth guard**: Redirect to `/login` when no valid JWT token exists.
- **Alternatives considered**:
  - Navigator 2.0 directly: too verbose. Rejected.
  - AutoRoute: code generation adds dependency; GoRouter is simpler for MVP scale. Rejected.

## Decision 7: Offline-Tolerant Upload Retry

- **Decision**: In-flight upload retry handled by Dio interceptor. No persistent offline queue for MVP.
- **Why**:
  - Dio retry interceptor handles transient network failures automatically (up to 3 retries).
  - User-corrected edits are preserved in memory during confirmation retry.
  - Full offline queue (saving uploads while completely offline) is out of scope per spec assumptions.
- **User experience**:
  - Upload in progress → progress bar with "Uploading..." label.
  - Retry in progress → progress bar with "Retrying (1/3)..." label.
  - All retries exhausted → error state with "Retry" button.
  - Confirmation retry → corrections preserved, retry button shown.

## Decision 8: Material Design 3 Theming

- **Decision**: Full MD3 theming with light/dark mode support using `ColorScheme.fromSeed`.
- **Why**:
  - Project design system (mobile-ui-design skill) defines MD3 token system.
  - `ColorScheme.fromSeed` generates harmonious palette from brand primary.
  - Dark mode requires no extra tokens — generated from same seed.
- **Currency colors**: USD green (0xFF2E7D32), Bs. blue (0xFF1565C0) — used in price displays and currency badges.

## Decision 9: Localization

- **Decision**: Flutter `flutter_localizations` with ARB files for English (default) and Spanish (Venezuela market).
- **Why**:
  - Venezuelan market requires Spanish UI.
  - English is the code convention (Constitution Article IX) but UI must be in Spanish for end users.
  - ARB format is Flutter's standard localization approach.
- **Languages**: `en` (fallback), `es` (primary for Venezuela).
