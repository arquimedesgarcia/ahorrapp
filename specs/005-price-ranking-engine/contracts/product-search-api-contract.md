# Contract: Product Search API (`/api/v1/search`)

Searches products by normalized name and, for each matching product,
returns the cheapest store per currency.

- **Auth**: Required (`Authorization: Bearer <token>`)

## GET `/api/v1/search`

### Query Parameters

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `q` | string | yes | Search query (product name); minimum 3 characters |

### Success (`200`)

```json
{
  "results": [
    {
      "product_id": "uuid",
      "product_name": "Arroz Blanco",
      "unit": null,
      "best_prices": {
        "USD": {
          "store_id": "uuid",
          "store_name": "Central Market",
          "branch": "Downtown",
          "average_price": 1.25,
          "currency": "USD",
          "sample_count": 45,
          "last_observed_at": "2025-06-28T14:30:00Z"
        },
        "Bs.": {
          "store_id": "uuid",
          "store_name": "SuperMaxi",
          "branch": null,
          "average_price": 125.00,
          "currency": "Bs.",
          "sample_count": 12,
          "last_observed_at": "2025-06-27T09:00:00Z"
        }
      }
    },
    {
      "product_id": "uuid",
      "product_name": "Arroz Integral",
      "unit": null,
      "best_prices": {
        "USD": {
          "store_id": "uuid",
          "store_name": "Central Market",
          "branch": "Downtown",
          "average_price": 1.80,
          "currency": "USD",
          "sample_count": 10,
          "last_observed_at": "2025-06-26T11:00:00Z"
        }
      }
    }
  ]
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `results` | array | Products matching the query (may be empty) |
| `results[].product_id` | string (UUID) | Canonical product identifier |
| `results[].product_name` | string | Normalized product name |
| `results[].unit` | string? | Unit of measure (null when not known) |
| `results[].best_prices` | object | Map keyed by currency code |
| `best_prices[currency]` | object? | Cheapest store for that currency |
| `….store_id` | string (UUID) | Store identifier |
| `….store_name` | string | Store name |
| `….branch` | string? | Branch name (omitted if null) |
| `….average_price` | number | Average price at the cheapest store |
| `….currency` | string | Currency code (matches the grouping key) |
| `….sample_count` | int | Number of fresh observations |
| `….last_observed_at` | string (ISO 8601) | Most recent observation timestamp |

When a product has no fresh observations for a currency, that currency
key is omitted from `best_prices`.

### Empty Results (`200`)

```json
{
  "results": []
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `400` | `{"error": "query parameter 'q' is required"}` | Missing `q` |
| `400` | `{"error": "query must be at least 3 characters"}` | `q` shorter than 3 chars |
| `401` | `{"error": "invalid or expired token"}` | Token invalid or missing |

## Existing Endpoint Compatibility

The existing `GET /api/v1/ranking/products/search` endpoint (from spec
004) continues to work and returns the same shape as documented in
`specs/004-flutter-mobile-app/contracts/ranking-api-contract.md`. The
new `/api/v1/search` endpoint returns an enriched shape (with
`best_prices` per currency and freshness data); the old endpoint
returns the flat `stores` array per product. Both are served by the same
ranking use case with different response projections. The Flutter app
can migrate to `/api/v1/search` when ready, but is not required to do so
for this feature.