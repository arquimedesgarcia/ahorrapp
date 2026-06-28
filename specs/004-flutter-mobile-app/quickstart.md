# Quickstart: Flutter Mobile Application MVP

Validation guide for the Flutter mobile app consuming the AhorraApp backend.

## Prerequisites

- Flutter SDK 3.24+ installed and in PATH
- Android Studio (for Android emulator) or Xcode (for iOS simulator)
- Backend stack running locally via `docker compose up` (from repo root)
- A test user account created (or use registration flow in app)

## Setup

```bash
# From repo root
cd mobile/

# Install dependencies
flutter pub get

# Generate code (json_serializable, freezed, riverpod)
dart run build_runner build --delete-conflicting-outputs

# Set environment (for local backend)
# Android emulator: API_BASE_URL=http://10.0.2.2:8080
# iOS simulator:    API_BASE_URL=http://localhost:8080
# Physical device:  API_BASE_URL=http://<your-local-ip>:8080
```

## Run the App

```bash
# Android
flutter run --flavor development

# iOS
flutter run --flavor development
```

## Validation Scenarios

### V1 — Onboarding + Registration

1. Launch the app on a clean install (clear app data first).
2. **Expected**: Onboarding screens appear with app value proposition.
3. Tap through onboarding to the registration screen.
4. Enter email, password (8+ chars), and display name. Submit.
5. **Expected**: Account created, redirected to main screen. Session persists on app restart.

### V2 — Receipt Upload

1. Log in with a test account.
2. From the main screen, tap the prominent scan button.
3. **Expected**: Camera opens with receipt framing overlay.
4. Take a photo of a test receipt (or select from gallery).
5. Tap "Upload".
6. **Expected**: Progress bar shows upload progress. On completion, receipt ID is displayed and status shows "PENDING".

### V3 — Upload Retry (Offline Tolerance)

1. Start receipt upload as in V2.
2. During upload, disable network (airplane mode).
3. **Expected**: Upload fails, retry indicator shows "Retrying (1/3)...".
4. Re-enable network before final retry.
5. **Expected**: Upload succeeds on retry.

### V4 — Editable Summary + Confirmation

1. Wait for a receipt to reach `NEEDS_REVIEW` state (polling every 5s in app).
2. Open the receipt from the list.
3. **Expected**: Editable summary shows store, date, total, and line items.
4. Edit the store name, change a line item price, select currency.
5. Leave one item's currency empty. Tap "Confirm".
6. **Expected**: Confirmation blocked, currency field highlighted as required.
7. Set the missing currency. Tap "Confirm" again.
8. **Expected**: Points earned screen appears (e.g., "+10 points"). Receipt status changes to "CONFIRMED".

### V5 — Product Search

1. Ensure at least one confirmed receipt exists (from V4).
2. Navigate to search tab.
3. Type a product name (e.g., "Arroz"). Submit search.
4. **Expected**: Stores listed from cheapest to most expensive. Each entry shows store name, price, and currency badge.
5. Tap a store card. Verify full store details are shown.

### V6 — Empty Search Results

1. Search for a product name that doesn't exist (e.g., "XXXXXXXXX").
2. **Expected**: Empty state illustration with message "No results found for 'XXXXXXXXX'".

### V7 — Profile + Points

1. Navigate to the profile tab.
2. **Expected**: Display name, email, and total points shown.
3. Confirm another receipt (repeat V4).
4. Navigate back to profile.
5. **Expected**: Points total increased by confirmation amount.

### V8 — Session Expiry

1. Log in. Wait for JWT to expire (or manually corrupt the stored token via debugger).
2. Attempt to scan a receipt or search a product.
3. **Expected**: Redirected to login screen with a message "Session expired. Please log in again."

### V9 — Camera Permission Denied

1. On a device with camera permissions denied for the app.
2. Tap the scan button.
3. **Expected**: Message explaining camera permission is required, with a button to open app settings.

### V10 — Dark Mode

1. Enable dark mode in device settings.
2. Relaunch the app.
3. **Expected**: All screens render correctly with dark theme, proper contrast ratios, no hardcoded light colors.

## Test Commands

```bash
# Unit tests
flutter test test/unit/

# Widget tests
flutter test test/widget/

# Integration tests (requires running backend)
flutter test test/integration/

# All tests with coverage
flutter test --coverage

# Golden tests (update if intentional changes)
flutter test --update-goldens test/widget/**/*_golden_test.dart
```

## Environment Variables

| Variable | Default (dev) | Description |
|----------|---------------|-------------|
| `API_BASE_URL` | `http://localhost:8080` | Backend API base URL |
| `ENABLE_LOGGING` | `true` | Enable HTTP request/response logging |

## Project Structure Validation

```bash
# Verify no undocumented API fields consumed
dart run build_runner build --delete-conflicting-outputs
dart analyze lib/
dart format --set-exit-if-changed lib/
```

All models must match contracts in `contracts/*.md` exactly. The `dart analyze` step catches unused imports and type errors. Format check ensures consistent code style.
