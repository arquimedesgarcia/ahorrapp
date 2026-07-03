# Feature Specification: Average-Price Engine and "Where to Buy Cheaper" Ranking

**Feature Branch**: `005-price-ranking-engine`

**Created**: 2025-06-29

**Status**: Draft

**Input**: User description: "Feature: Average-price engine and 'where to buy cheaper' ranking."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Recompute Averages After Receipt Confirmation (Priority: P1)

A shopper confirms a receipt, which generates price observations for each
line item. The system immediately recomputes the aggregate price
(average, minimum, sample count) for each affected (product, store,
currency) triple. This happens synchronously as part of the confirmation
flow so that the very next query sees fresh data.

**Why this priority**: Without up-to-date aggregates the ranking is stale
and misleading. This is the foundation upon which every other story in
this feature depends.

**Independent Test**: Upload and confirm two receipts for the same product
at the same store in the same currency with different unit prices. Query
the aggregate for that product at that store. Verify that the average and
minimum reflect both observations.

**Acceptance Scenarios**:

1. **Given** a confirmed receipt that generated 3 price observations for
   product P at store S in currency USD with prices 1.00, 1.20, 1.40,
   **When** the confirmation completes,
   **Then** the aggregate for (P, S, USD) shows average 1.20, minimum 1.00,
   and sample count 3.
2. **Given** an existing aggregate with 5 observations,
   **When** a new receipt is confirmed adding 1 more observation for the
   same (product, store, currency),
   **Then** the aggregate is updated to reflect 6 observations with the
   new average and minimum.
3. **Given** two observations with different currencies (USD and Bs.) for
   the same product and store,
   **When** the receipt is confirmed,
   **Then** two separate aggregates are maintained — one per currency —
   and neither average mixes currencies.

---

### User Story 2 - View Per-Store Ranking for a Specific Product (Priority: P1)

A shopper wants to know where a specific product is cheapest. They look up
the product by its identifier and the system returns, per currency, the
list of stores that carry it, ordered from cheapest to most expensive by
average price. Each ranking entry includes the average price, the minimum
observed price, the number of observations, and a freshness indicator
(how recent the most recent observation is).

**Why this priority**: This is the core user-facing output of the price
engine — the "where to buy cheaper" recommendation. Without it the
aggregated data has no user-visible value.

**Independent Test**: Seed price observations for product P at stores S1
($1.00 avg) and S2 ($1.20 avg) in currency USD. Query the per-product
ranking endpoint. Verify that S1 appears before S2 and that the average
prices match.

**Acceptance Scenarios**:

1. **Given** aggregates for product P at stores S1 (avg $1.00 USD),
   S2 (avg $1.50 USD), S3 (avg $1.20 USD),
   **When** the user queries the per-product ranking for P,
   **Then** stores are ordered S1, S3, S2 from cheapest to most expensive
   in the USD grouping.
2. **Given** aggregates for product P in both USD and Bs. at various
   stores,
   **When** the user queries the per-product ranking for P,
   **Then** the response is grouped per currency and no currency's ranking
   mixes prices from another currency.
3. **Given** a product with no observations at all,
   **When** the user queries its per-product ranking,
   **Then** the response returns an empty list per currency with no error.

---

### User Story 3 - Search Products by Name and Get Cheapest Store Per Currency (Priority: P1)

A shopper types the name of a product and the system returns matching
products with, for each match, the cheapest store per currency. This is
the entry-level search experience: the user does not know a product
identifier; they just search by what they remember the product was called
on their receipt.

**Why this priority**: Search is the primary discovery path for mobile
users. Without it, users cannot find the product they care about without
already knowing its identifier, which makes the ranking unusable in
practice.

**Independent Test**: Seed aggregates for products "Arroz Blanco" and
"Arroz Integral" across stores. Search for "arroz". Verify both products
appear, each with the cheapest store per currency.

**Acceptance Scenarios**:

1. **Given** products "Arroz Blanco" and "Arroz Integral" with aggregates
   in USD,
   **When** the user searches for "arroz",
   **Then** both products are returned, each with its cheapest store
   (store name and average price) per currency.
2. **Given** a search query that matches no product names,
   **When** the user searches,
   **Then** the response returns an empty list with no error.
3. **Given** a product with observations in both USD and Bs.,
   **When** the user searches and that product matches,
   **Then** the cheapest store is returned independently per currency.

---

### User Story 4 - Observation Freshness and Age Weighting (Priority: P2)

To avoid distorted averages caused by very old prices in a
high-inflation context, the system applies an age-based weighting or
filtering policy to observations. Observations older than a configurable
threshold are either excluded or contribute with reduced weight. The
aggregated average thus reflects recent market conditions, not historical
data.

**Why this priority**: In high-inflation economies, old prices are
actively misleading. Freshness control is a quality guarantee, but the
core ranking mechanism (Stories 1-3) works without it and is the priority
delivery path.

**Independent Test**: Seed two observations for the same (product, store,
currency): one from 60 days ago at $1.00 and one from yesterday at $2.00.
Apply a 90-day threshold. Confirm the aggregate average is computed with
both observations. Then change the threshold to 30 days and recompute:
confirm that only the $2.00 observation is included.

**Acceptance Scenarios**:

1. **Given** observations dated 120 days ago and 1 day ago with a
   configurable age threshold of 90 days,
   **When** aggregates are recomputed,
   **Then** only the observation from 1 day ago is included in the average
   and the 120-day-old observation is excluded.
2. **Given** observations dated 120 days ago and 1 day ago with a
   configurable age threshold of 180 days,
   **When** aggregates are recomputed,
   **Then** both observations are included in the average.
3. **Given** a store whose only observation for product P was 200 days
   ago and the threshold is 90 days,
   **When** the per-product ranking for P is queried,
   **Then** that store does not appear in the ranking for P because it has
   no fresh observations.

---

### User Story 5 - Optional Proximity-Based Ordering with PostGIS (Priority: P3)

When the user provides their location, store rankings are optionally
ordered or filtered by proximity to the user's location, using the
geolocation stored on each `Store`. This enhances the user experience in
two directions: (a) a user can find nearby stores first, and (b) a user
can filter out stores beyond a certain radius.

**Why this priority**: Proximity is a nice-to-have refinement; the core
value loop ( Stories 1-3) operates without location data. PostGIS
infrastructure is already in place but the user-location input mechanism
is beyond the MVP core of this feature and is deferred to an optional
enhancement.

**Independent Test**: Seed two stores with coordinates, one near the user
and one far. Query the per-product ranking with the user's location and a
radius filter. Verify the ranking excludes the distant store (or appears
last when proximity ordering is used without a radius).

**Acceptance Scenarios**:

1. **Given** stores S1 (lat 10.5, long -66.9) and S2 (lat 11.0, long
   -68.0), and aggregates for product P at both,
   **When** the user queries the per-product ranking with their location
   (lat 10.5, long -66.9) and proximity ordering enabled,
   **Then** S1 appears before S2 in the ranking.
2. **Given** the same stores and a user-specified radius of 50 km,
   **When** the user queries the per-product ranking with the radius
   filter,
   **Then** S2 is excluded from the ranking for product P because it is
   beyond 50 km from the user.

---

### Edge Cases

- What happens when a receipt confirmation produces price observations for
  a product that has never been seen before? The aggregate is created from
  scratch (no pre-existing row to update).
- What happens when all observations for a (product, store, currency)
  triple are older than the age threshold? The aggregate is either removed
  or reported as empty/stale so the ranking does not display it.
- What happens when two stores have the exact same average price for a
  product? Ties are broken deterministically (e.g., by store name) so the
  ranking order is stable across queries.
- What happens when a receipt is rejected after confirmation? The
  observations from that receipt are either invalidated or the
  recomputation accounts for a "reverted" observation. The MVP scope treats
  confirmed receipts as final; reverting is out of scope.
- What happens when a search query is very short (1 character) or very
  long? The search returns only products whose normalized name contains or
  starts with the normalized query; a minimum query length may be enforced
  to prevent excessively expensive scans.
- What happens when the user searches with currency information? The
  search always returns the cheapest store per currency; the response
  structure does not depend on whether the user specified a currency
  preference (the MVP returns all currencies the product has observations
  in).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When a receipt is confirmed and generates PriceObservation
  records, the system MUST recompute the PriceAggregate per
  (product, store, currency) triple, updating average price, minimum price,
  and sample count.
- **FR-002**: Aggregates MUST be computed per currency; currencies MUST
  NEVER be mixed within a single average or within a single ranking
  grouping.
- **FR-003**: The age of each observation MUST be tracked. The aggregate
  computation MUST either exclude observations older than a configurable
  threshold or weight them by age so that stale observations do not
  distort the average in a high-inflation economy.
- **FR-004**: The age threshold MUST be configurable at deployment time
  (via environment variable or configuration) without code changes.
- **FR-005**: The system MUST expose endpoint `GET /api/v1/products/{id}/prices`
  that, for a given product identifier, returns the list of stores that
  carry it, grouped per currency and ordered from cheapest to most
  expensive by average price.
- **FR-006**: Each store entry in the per-product ranking response MUST
  include store identifier, store name, branch (when available), average
  price, minimum observed price, currency, sample count, and a freshness
  indicator (age of the most recent observation or last-updated timestamp).
- **FR-007**: The system MUST expose endpoint `GET /api/v1/search?q=...`
  that searches products by normalized name and, for each matching
  product, returns the cheapest store per currency.
- **FR-008**: The search response MUST return every product whose
  normalized name matches the normalized query; the search is
  case-insensitive and accent-insensitive.
- **FR-009**: When a store has no fresh observations for a product within
  the configured age threshold, the ranking MUST omit that store from the
  ranking for that product (stale stores do not appear).
- **FR-010**: The ranking order MUST be deterministic; ties in average
  price MUST be broken by a stable secondary order (e.g., store name) so
  that equal-priced stores appear in a consistent order.
- **FR-011**: A product with no observations at all MUST return an empty
  per-currency ranking (HTTP 200 with empty results), not an error.
- **FR-012**: Both endpoints MUST require authentication (valid bearer
  token) and MUST reject unauthenticated requests with HTTP 401.
- **FR-013**: Empty or missing search query parameter `q` MUST return
  HTTP 400 with a descriptive error message.
- **FR-014**: An invalid or non-existent product identifier passed to
  `GET /api/v1/products/{id}/prices` MUST return HTTP 404, but a valid
  identifier with no observations MUST return HTTP 200 with an empty
  ranking.
- **FR-015**: Optionally, when the user provides their location, the
  per-product ranking endpoint MUST accept location and optional radius
  parameters, and when provided, MUST either order stores by proximity to
  the user or filter out stores beyond the specified radius, using the
  geolocation stored on each Store.
- **FR-016**: The proximity filtering / ordering is OPTIONAL at the MVP
  stage; when the user does not provide a location, the ranking MUST be
  ordered by average price ascending (cheapest first) without proximity
  considerations.

### Key Entities *(include if feature involves data)*

- **PriceObservation**: A single observed price for a (product, store,
  currency) triple, tied to the receipt that produced it; carries an
  observed-at timestamp that defines its age. Already exists in the
  domain.
- **PriceAggregate**: The computed aggregate for a (product, store,
  currency) triple. Contains average price, minimum price, sample count,
  and a freshness indicator (timestamp of the most recent observation
  included). This entity does not yet exist in the data model and is
  introduced by this feature.
- **Product**: The canonical product record (already exists), used for
  normalized name search and to scope price queries.
- **Store**: The merchant record (already exists), optionally carrying
  geolocation; used for ranking output and proximity filtering.
- **Currency**: A grouping key (USD or Bs.) that scoping every average
  and ranking group; currencies are never mixed (Article V of the
  constitution).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After a receipt confirmation completes, querying the
  per-product ranking for any product on that receipt returns the
  updated average price for the store where the purchase was made,
  including the new observation, typically within a few seconds at MVP
  scale.
- **SC-002**: A user searching for a product by name sees at least one
  matching store with its cheapest-price store within seconds on a
  local development environment.
- **SC-003**: A ranking result for a product never mixes currencies;
  for a product with observations in both USD and Bs., the response shows
  two distinct currency groupings with their own store orderings.
- **SC-004**: A ranking result excludes stores whose only observations
  for that product are older than the configured age threshold, so that
  users never see stale prices in a high-inflation economy.
- **SC-005**: The ranking order is stable: two requests issued in
  quick succession for the same product and currency return the same
  store order (ties broken deterministically).
- **SC-006**: The average-price engine recomputes aggregates for all
  (product, store, currency) triples affected by a single receipt
  confirmation within a time that remains imperceptible to the user at
  MVP scale (a receipt with 20 line items).
- **SC-007**: The product search returns relevant matches for queries
  with 3+ characters; searches with fewer characters may be rejected with
  a descriptive message to prevent expensive scans.
- **SC-008**: When proximity filtering is enabled (optional), the result
  set excludes stores beyond the user-specified radius; otherwise the
  ranking still returns all stores ordered by average price ascending.

## Assumptions

- The receipt confirmation flow already persists `PriceObservation`
  records; this feature builds the aggregation layer on top of that
  existing data.
- A confirmed receipt is final; reverting a confirmation (and
  invalidating its observations) is out of MVP scope. If this becomes
  needed later, the recomputation logic must be extended.
- The age threshold is a single global configuration value at MVP scale;
  per-category or per-currency freshness policies are deferred.
- Product name normalization already exists (basic lowercasing and
  whitespace collapse); accent-insensitive search is achieved through
  database collation or query-time normalization rather than a separate
  normalization pipeline.
- A single tenant / single currency marketplace per region is assumed;
  multi-region isolation is out of MVP scope.
- Proximity filtering requires the Store to have populated lat/long;
  stores without geolocation are included in the ranking without
  proximity ordering when a location is provided, and ordered by average
  price ascending (placed at the end if proximity ordering is active).
- The existing REST `/api/v1/` prefix and the published ranking API
  contract are honored; this feature does not introduce breaking changes
  to already-published endpoints. The stub `RankingHandler` that
  currently returns empty results is replaced with a real implementation
  that satisfies the existing contract.
- An averaged result is expected to be available immediately after
  confirmation (synchronous recomputation); eventual consistency or
  a delayed batch recompute is out of MVP scope and would require a
  separate spec.