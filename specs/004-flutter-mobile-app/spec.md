# Feature Specification: Flutter Mobile Application MVP

**Feature Branch**: `004-flutter-mobile-app`

**Created**: 2026-06-25

**Status**: Draft

**Input**: User description: "Flutter mobile application for iOS and Android. The app consumes the existing REST API /api/v1. MVP screens: onboarding + register/login, main screen with scan receipt button, receipt photo capture and upload with processing status, editable summary screen (store, date, total, line items with product, quantity, price, currency), confirm and see points earned, product search showing cheapest store, profile with accumulated points."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Onboarding and Account Creation (Priority: P1)

A new user installs the app, goes through onboarding, registers an account, and logs in to access the main experience.

**Why this priority**: Without authentication, no receipt data can be associated with a user, blocking the entire value loop.

**Independent Test**: Install the app on a clean device, complete onboarding, register with email and password, log out, and log back in. Verify the session persists across app restarts.

**Acceptance Scenarios**:

1. **Given** a first-time user opening the app, **When** they view the onboarding screens, **Then** they see the app's value proposition and can proceed to registration.
2. **Given** a user on the registration screen, **When** they submit valid email and password, **Then** an account is created and they are logged in.
3. **Given** a registered user on the login screen, **When** they submit correct credentials, **Then** they access the main screen.
4. **Given** incorrect credentials on login, **When** the user attempts to log in, **Then** a clear error message is shown and they can retry.
5. **Given** an authenticated user, **When** they close and reopen the app, **Then** their session is restored without re-authentication.

---

### User Story 2 - Capture and Upload Receipt (Priority: P1)

An authenticated user taps the scan button, captures a receipt photo, and uploads it. The app shows upload progress and processing status.

**Why this priority**: Receipt capture is the primary data input mechanism; without it the app has no value.

**Independent Test**: Tap the scan button, take a photo, upload it, and verify the app shows upload progress followed by a processing status indicator.

**Acceptance Scenarios**:

1. **Given** an authenticated user on the main screen, **When** they tap the scan receipt button, **Then** the device camera opens for photo capture.
2. **Given** a captured receipt photo, **When** the user confirms the upload, **Then** the app shows upload progress and the receipt is sent to the backend.
3. **Given** a successful upload, **When** the receipt is being processed, **Then** the app displays a processing status indicator.
4. **Given** a network failure during upload, **When** the upload fails, **Then** the app retries automatically and notifies the user if retries are exhausted.
5. **Given** the user captures the same receipt image twice, **When** duplicate detection triggers, **Then** the app shows the existing receipt instead of creating a duplicate.

---

### User Story 3 - Review and Edit Receipt Summary (Priority: P1)

After processing completes, the user sees an editable summary with store, date, total, and line items (product, quantity, price, currency). The user can correct any field before confirming.

**Why this priority**: User correction is the trust boundary between raw OCR and reliable price data.

**Independent Test**: Upload a receipt, wait for processing, verify the summary screen shows all fields editable, modify a line item price, and confirm the changes are reflected.

**Acceptance Scenarios**:

1. **Given** a receipt that finished processing, **When** the user opens the receipt summary, **Then** all fields (store, date, total, line items) are displayed and individually editable.
2. **Given** an editable summary, **When** the user modifies the store name, date, total, or any line item field (product name, quantity, price, currency), **Then** the changes are reflected immediately in the UI.
3. **Given** a receipt with partial or low-quality extraction, **When** the summary is displayed, **Then** available fields are shown and empty fields remain editable for manual completion.
4. **Given** the user edits a currency value on a line item, **When** they attempt to confirm without setting a currency, **Then** the app prevents confirmation and highlights the missing field.

---

### User Story 4 - Confirm Receipt and Earn Points (Priority: P1)

After reviewing and correcting the receipt, the user confirms it. The app shows the points earned for this receipt.

**Why this priority**: Confirmation triggers the data pipeline (price observations, averages, rankings) and gamification rewards the user for contributing data.

**Independent Test**: Correct and confirm a receipt, verify a points-earned confirmation is displayed, and check that the points total updates in the profile.

**Acceptance Scenarios**:

1. **Given** a corrected receipt summary, **When** the user taps confirm, **Then** the receipt is submitted and transitions to confirmed state.
2. **Given** a confirmed receipt, **When** confirmation succeeds, **Then** a points-earned screen is displayed showing the points awarded.
3. **Given** a confirmation attempt with missing required fields (e.g., missing currency on a line item), **When** the user taps confirm, **Then** the app rejects the submission and indicates which fields need attention.
4. **Given** a network failure during confirmation, **When** confirmation fails, **Then** the app retries and does not lose the user's corrections.

---

### User Story 5 - Search Products and Find Cheapest Store (Priority: P2)

A user searches for a product by name and sees a ranking of stores ordered by lowest price, helping them decide where to shop.

**Why this priority**: This is the core value proposition — turning receipt data into actionable savings. It depends on confirmed receipt data flowing through the price engine.

**Independent Test**: Search for a known product name, verify the results show stores ranked by price, and confirm the cheapest store appears first.

**Acceptance Scenarios**:

1. **Given** an authenticated user on the product search screen, **When** they type a product name and submit the search, **Then** results show stores with prices for that product, ordered from cheapest to most expensive.
2. **Given** a product with no price data, **When** the user searches for it, **Then** the app shows an empty state indicating no results are available yet.
3. **Given** a search with multiple matching products, **When** results are displayed, **Then** the user can distinguish between product variants (e.g., different sizes or brands).
4. **Given** a search result, **When** the user views a store entry, **Then** the displayed price includes its currency.

---

### User Story 6 - View Profile and Accumulated Points (Priority: P2)

A user views their profile screen to see their display name, total accumulated points, and recent activity.

**Why this priority**: Profile and points visibility reinforces the gamification loop and encourages continued receipt uploads.

**Independent Test**: Open the profile screen and verify the displayed points total matches the sum of points earned from confirmed receipts.

**Acceptance Scenarios**:

1. **Given** an authenticated user, **When** they navigate to the profile screen, **Then** they see their display name, email, and total accumulated points.
2. **Given** a user confirms a receipt and earns points, **When** they return to the profile screen, **Then** the points total reflects the newly earned points.
3. **Given** a user with no confirmed receipts, **When** they view their profile, **Then** they see a zero or empty points state.

---

### Edge Cases

- What does the app show when the camera permission is denied?
- How does the app behave when the backend is unreachable on app launch?
- What happens when the user receives a phone call or switches apps during receipt upload?
- How does the app handle receipt images that exceed the maximum upload size?
- What does the app display when a receipt remains in processing state beyond the expected timeout?
- How does the app behave when the session token expires mid-use?
- What happens when the user searches for a product with special characters or very long names?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: App MUST require authentication (login/registration) before granting access to receipt scanning, product search, and profile features.
- **FR-002**: App MUST present onboarding screens explaining the app's value proposition to first-time users.
- **FR-003**: App MUST support email and password registration and login.
- **FR-004**: App MUST persist the authenticated session across app restarts.
- **FR-005**: App MUST provide a main screen with a prominent scan receipt action.
- **FR-006**: App MUST open the device camera for receipt photo capture when the scan action is triggered.
- **FR-007**: App MUST upload captured receipt images to the backend API and display upload progress.
- **FR-008**: App MUST display receipt processing status (pending, processing, needs review, confirmed) to the user.
- **FR-009**: App MUST present an editable summary screen after processing completes, showing store, date, total, and line items (product, quantity, unit price, currency) as individually editable fields.
- **FR-010**: App MUST allow the user to modify any field on the editable summary and see changes reflected immediately.
- **FR-011**: App MUST validate that all line items have a currency value before allowing confirmation.
- **FR-012**: App MUST submit corrected receipt data to the backend on confirmation and display a points-earned confirmation.
- **FR-013**: App MUST retry failed network operations (upload, confirmation) with user-visible feedback.
- **FR-014**: App MUST provide a product search screen that displays stores ranked by price from cheapest to most expensive.
- **FR-015**: App MUST display the currency alongside every price shown in search results and receipt summaries.
- **FR-016**: App MUST provide a profile screen showing the user's display name, email, and total accumulated points.
- **FR-017**: App MUST handle expired authentication tokens by redirecting to the login screen.
- **FR-018**: App MUST display appropriate empty states when no data is available (no receipts, no search results, no points).
- **FR-019**: App MUST request camera permission and show an appropriate message if permission is denied.
- **FR-020**: App MUST consume only the published API contract under `/api/v1` without assuming undocumented fields or behaviors.

### Key Entities *(include if feature involves data)*

- **User Session**: Authentication token and user identity persisted locally to maintain login state across app restarts.
- **Receipt Summary**: Editable view of parsed receipt data including store, purchase date, total amount, and line items with product name, quantity, unit price, and currency.
- **Product Search Result**: A product match with a list of stores and their prices, ranked from cheapest to most expensive.
- **User Profile**: Display name, email, and total accumulated loyalty points.
- **Points Award**: Points earned from a confirmed receipt submission, displayed immediately after confirmation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can complete onboarding and registration in under 2 minutes.
- **SC-002**: Receipt photo capture to upload confirmation completes in under 10 seconds on a stable network connection.
- **SC-003**: Upload retry succeeds automatically for at least 90% of transient network failures without user intervention.
- **SC-004**: Users can edit and confirm a receipt summary in under 3 minutes.
- **SC-005**: Product search returns results in under 3 seconds under normal conditions.
- **SC-006**: 100% of price displays include a currency label.
- **SC-007**: Session restoration on app restart is successful for at least 99% of valid stored sessions.
- **SC-008**: At least 80% of users who complete registration go on to scan and confirm their first receipt.

## Assumptions

- The backend REST API under `/api/v1` is already implemented and available, including endpoints for auth, receipt upload, receipt detail retrieval, receipt confirmation, product search, and user profile.
- Authentication uses JWT tokens as defined in the backend constitution (Article VII).
- The API contract for each endpoint is published (OpenAPI or equivalent) and stable.
- Camera access is available on target devices (iOS and Android smartphones with rear cameras).
- Points are awarded per confirmed receipt with a flat rate defined by the backend.
- Product search ranking is computed server-side; the app only displays the results.
- Push notifications are out of scope for the MVP.
- Offline receipt creation (queueing uploads while offline) is out of scope; the app retries only for in-flight operations.
- The onboarding flow consists of informational screens followed by registration; social login is out of scope.
