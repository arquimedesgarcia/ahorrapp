# PROMPTS — Spec-Driven Development for AhorraApp

> How to use: in Windsurf or OpenCode, the commands `/speckit.constitution`, `/speckit.specify`, `/speckit.plan`,
> `/speckit.tasks`, `/speckit.implement` are **slash commands** that Spec Kit installed. Type the command and
> paste the prompt text as its argument.
>
> These prompts are NOT "copy, paste, and pray". Each one produces an artifact that YOU review
> and approve before moving on. That is the SDD method.
>
> Before starting, make sure `docs/00_GLOBAL_DESIGN.md` is in the repo: the prompts ask the
> agent to read it as the root context.

---

## PROMPT 1 — Constitution
**Command:** `/speckit.constitution`

```
Read docs/00_GLOBAL_DESIGN.md as the project's source of truth. Adopt and refine the
constitution that already exists in .specify/memory/constitution.md. Do not replace it with a
generic one: respect its nine articles (Clean Architecture in Go with OCR as a replaceable
detail behind the OCRProvider interface and StorageProvider as a replaceable S3-compatible
port, spec-first, tests, versioned REST/JSON contracts, mandatory currency per price
observation, store/merchant extraction, product normalization, simplicity/cost and local-first
development, minimal security with JWT, readiness to grow without over-engineering, and English
as the working language).
Return the consolidated constitution and flag any ambiguity you detect.
```

**Your role:** Read the output. Confirm it keeps the principles. Approve.

---

## PROMPT 2 — Backend skeleton spec (Epic E1)
**Command:** `/speckit.specify`

```
Feature: Go backend skeleton with Clean Architecture.

Context: read docs/00_GLOBAL_DESIGN.md and .specify/memory/constitution.md.

I want to specify the foundation of the backend server. Describe the EXPECTED BEHAVIOR, not
the implementation:

- The service exposes a versioned REST API at /api/v1.
- There is a health endpoint GET /api/v1/health that confirms the service, the PostgreSQL
  database, and Redis are reachable.
- The project structure reflects Clean Architecture with separated layers: HTTP delivery, use
  cases, domain entities, and infrastructure adapters (repositories, Redis client, object
  storage client, OCR provider) behind interfaces defined in the domain.
- The whole application runs with Docker Compose: Go API, PostgreSQL 16 with the PostGIS
  extension, Redis, and a MinIO container (S3-compatible storage) so the full stack runs
  locally with no cloud accounts.
- There is a versioned database migration system.
- Configuration via environment variables (no secrets in the code).

Acceptance criteria:
- `docker compose up` brings up API + Postgres + Redis without errors.
- GET /api/v1/health responds 200 with the status of each dependency.
- The domain does not import HTTP or database packages.
- There is at least one test that verifies the health endpoint.

Do not include authentication or receipt logic yet; this spec is only the foundation.
```

**Your role:** Check that the spec talks about behavior (not code). Approve.

---

## PROMPT 3 — Backend technical plan
**Command:** `/speckit.plan`

```
Generate the technical plan for the backend skeleton spec.

Mandatory stack (from docs/00_GLOBAL_DESIGN.md):
- Language: Go 1.23+, standard HTTP router (net/http with chi or similar lightweight).
- Database: PostgreSQL 16 + PostGIS. Access via pgx. Migrations with golang-migrate.
- Cache/queue: Redis (go-redis client).
- Clean Architecture folder structure: /cmd, /internal/domain (entities, ports),
  /internal/usecase, /internal/adapter (http, postgres, redis, storage, ocr),
  /internal/config, /migrations.
- Containers: multi-stage Dockerfile for Go (minimal binary) and docker-compose.yml.

Respect articles I, IV, and VI of the constitution. Cite which decisions satisfy which
article. List the files to create and the creation order.
```

**Your role:** Verify domain dependencies point inward. Approve. Then run `/speckit.tasks` and
`/speckit.implement`. Test `docker compose up`.

---

## PROMPT 4 — Receipt flow + OCR spec (Epics E3, E4, E5)
**Command:** `/speckit.specify`

```
Feature: Receipt upload, OCR processing, parsing, and editable review.

Context: read docs/00_GLOBAL_DESIGN.md and the constitution. Assume the backend skeleton and
authentication already exist.

Expected behavior:

1. An authenticated user uploads a receipt image (POST /api/v1/receipts).
   - The image is stored in object storage with an unguessable URL.
   - A receipt is created in PENDING state and a processing job is enqueued.
   - The response is immediate (202) with the receipt id; processing is asynchronous.

2. A worker picks up the job and calls the OCR provider through the OCRProvider interface. The
   first implementation is a self-hosted PaddleOCR microservice accessed over HTTP. The
   interface must allow swapping it for a paid API without touching the business logic.

3. From the OCR text, the system parses:
   - The STORE/MERCHANT (name and, if present, branch/address) and associates it with a Store.
   - The purchase date and the total.
   - The line items: for each, raw text, quantity, and unit price, with its currency.
   The receipt moves to NEEDS_REVIEW state.

4. The app fetches an EDITABLE TEXT SUMMARY of the receipt (GET /api/v1/receipts/{id}): store,
   date, total, and list of line items, all editable.

5. The user corrects and confirms (POST /api/v1/receipts/{id}/confirm with the corrected
   content). On confirmation:
   - Each line item is normalized to a canonical Product.
   - PriceObservation records are created per product, with their Store, price, and currency.
   - The receipt moves to CONFIRMED.
   - (Point awarding and average recomputation are specified in other epics; here only the
     corresponding events/calls are emitted.)

Acceptance criteria:
- Uploading an image returns 202 and creates the receipt in PENDING.
- After processing, the receipt ends in NEEDS_REVIEW with store, date, and line items.
- The summary is editable and confirmation persists the user's corrections.
- Changing the OCRProvider implementation requires no changes in the use cases.
- Every PriceObservation has a mandatory currency.

Edge cases to handle: unreadable image (receipt ends in NEEDS_REVIEW with empty line items so
the user can enter them by hand), unrecognized store (allow creating it), duplicate receipt
(same image/user).
```

**Your role:** This is the heart of the product. Review carefully, adjust the edge cases to the
reality of the Venezuelan receipts you have, approve. Then `/speckit.plan`, `/speckit.tasks`, `/speckit.implement`.

---

## PROMPT 4-bis — Receipt flow technical plan
**Command:** `/speckit.plan`

```
Generate the technical plan for the receipt and OCR flow.

Mandatory decisions:
- Job queue over Redis; a Go worker consumes concurrently.
- Domain interface OCRProvider { Extract(ctx, imageRef) (RawOCRResult, error) }.
- PaddleOCRProvider adapter that POSTs to a Python microservice (FastAPI + PaddleOCR) deployed
  separately in Docker. Generate that minimal microservice too.
- Object storage: StorageProvider interface with an S3-compatible adapter. In development it
  targets a MinIO container (S3 API) defined in docker-compose; in production the same adapter
  targets Hetzner Object Storage. Only the endpoint URL and credentials differ (via env vars).
  Add the MinIO service to docker-compose and document the bucket/credentials env vars.
- Receipt parsing as a domain use case, with a decoupled, testable parser using OCR-text
  fixtures.
- Receipt state machine: PENDING → NEEDS_REVIEW → CONFIRMED | REJECTED.

Respect articles I, IV, V of the constitution. Include migrations for receipts, receipt_items,
stores, products, price_observations. Provide OCR-text fixtures of supermarket receipts for the
parser tests.
```

**Your role:** Approve. `/speckit.tasks` → `/speckit.implement`. Test with a real image.

---

## PROMPT 5 — Flutter app spec (Epic E6)
**Command:** `/speckit.specify`

```
Feature: Flutter mobile application for iOS and Android.

Context: read docs/00_GLOBAL_DESIGN.md. The app consumes the existing REST API /api/v1.

MVP screens and behavior:
- Onboarding + register/login.
- Main screen with a prominent "Scan receipt" button that opens the camera.
- Receipt photo capture and upload; shows processing status.
- EDITABLE SUMMARY screen: shows store, date, total, and line items (product, quantity, price,
  currency), all editable. Allows correcting and confirming.
- After confirming, shows the points earned.
- Product search screen that shows in which store it is cheaper (consumes the backend ranking).
- Profile screen with accumulated points.

Non-functional requirements:
- Layered architecture in Flutter (presentation, domain, data) consistent with the backend's
  Clean Architecture philosophy.
- State management with Riverpod (or Bloc); HTTP client with error handling and retries.
- Offline-tolerant: if the network fails on upload, it retries.

Acceptance criteria:
- I can register, scan a receipt, see the editable summary, correct, and confirm.
- I see my points. I search a product and see the cheapest store.
```

**Your role:** Approve. `/speckit.plan` (specify Riverpod, dio/http, go_router) → `/speckit.tasks` →
`/speckit.implement`. Test on your phone.

---

## PROMPT 6 — Price engine and ranking spec (Epic E7 + E8)
**Command:** `/speckit.specify`

```
Feature: Average-price engine and "where to buy cheaper" ranking.

Context: read docs/00_GLOBAL_DESIGN.md and the constitution (article V on currency).

Behavior:
- When a receipt is confirmed and generates PriceObservation records, the system recomputes
  the PriceAggregate per (product × store × currency): average price, minimum, and sample count.
- To avoid distortion, observations have an age; the average weights or filters observations
  older than a configurable threshold.
- Endpoint GET /api/v1/products/{id}/prices: returns, per currency, the list of stores ordered
  from cheapest to most expensive for that product, with average price and freshness.
- Endpoint GET /api/v1/search?q=...: searches products by normalized name and returns the best
  store (cheapest) per currency.
- Optional with PostGIS: order/filter by proximity to the user's location.

Acceptance criteria:
- Confirming a receipt updates the affected averages.
- Searching a product returns the cheapest store per currency.
- Currencies are never mixed within the same average.
```

**Your role:** Approve. `/speckit.plan` → `/speckit.tasks` → `/speckit.implement`.

---

## PROMPT 7 — Gamification / loyalty spec (Epic E9)
**Command:** `/speckit.specify`

```
Feature: Loyalty system for uploading receipts.

Context: read docs/00_GLOBAL_DESIGN.md.

Behavior:
- On confirming a valid receipt, the user earns points (configurable base amount, with possible
  bonuses for completing data or for being the first observation of a product/store).
- Each movement is recorded in LoyaltyTransaction (reason, points, date).
- Endpoint GET /api/v1/me/loyalty: point balance and history.
- Anti-abuse rules: the same receipt does not grant points twice; a configurable daily limit of
  point-granting receipts to mitigate fraud.

Acceptance criteria:
- Confirming a new receipt grants points only once.
- The balance and history are queryable.
- Resubmitting the same receipt grants no additional points.
```

**Your role:** Approve. `/speckit.plan` → `/speckit.tasks` → `/speckit.implement`.

---

## Pattern for the remaining epics (E2 auth, E10 hardening)
Use the same mold: a `/speckit.specify` that describes ONLY behavior and acceptance criteria (taken
from the global design document), followed by `/speckit.plan`, `/speckit.tasks`, `/speckit.implement`, with your review
and approval at each `/speckit.specify` and `/speckit.plan`.

### Reusable /speckit.specify template
```
Feature: <name>.
Context: read docs/00_GLOBAL_DESIGN.md and .specify/memory/constitution.md.
Expected behavior: <list of observable behaviors>.
Acceptance criteria: <verifiable list>.
Edge cases: <list>.
Constraints: respect articles <N> of the constitution. Do not implement <out of scope>.
```

---

## Rules that keep this "SDD" and not "vibe coding"
1. Never jump from an idea straight to `/speckit.implement`. Always: `/speckit.specify` → review → `/speckit.plan` →
   review → `/speckit.tasks` → `/speckit.implement`.
2. Specs talk about WHAT; plans talk about HOW. Do not mix them.
3. If the agent proposes something that violates the constitution, reject it citing the article.
4. Keep `docs/00_GLOBAL_DESIGN.md` updated: it is the project's memory across sessions.
5. Commit after each finished and tested epic.
