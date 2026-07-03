# Contract: Per-Product Price Ranking API (`/api/v1/products/{id}/prices`)

Returns the list of stores that carry a given product, grouped per
currency and ordered from cheapest to most expensive by average price.

- **Auth**: Required (`Authorization: Bearer <token>`)

## GET `/api/v1/products/{id}/prices`

### Path Parameters

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id`  | string (UUID) | yes | Canonical product identifier |

### Query Parameters (optional)

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `lat` | number | no | User latitude (for proximity ordering) |
| `long` | number | no | User longitude (for proximity ordering) |
| `radius_km` | number | no | Max distance in km from user (filter) |

When `lat` and `long` are both provided, stores are ordered by proximity
(unless `radius_km` is also provided, in which case stores beyond the
radius are excluded). Stores without geolocation are included and sorted
last by average price ascending. When location is not provided, ranking
is by average price ascending.

### Success (`200`)

```json
{
  "product_id": "uuid",
  "product_name": "Arroz Blanco",
  "currency_rankings": {
    "USD": [
      {
        "store_id": "uuid",
        "store_name": "Central Market",
        "branch": "Downtown",
        "average_price": 1.25,
        "min_price": 1.10,
        "currency": "USD",
        "sample_count": 45,
        "last_observed_at": "2025-06-28T14:30:00Z",
        "distance_km": null
      },
      {
        "store_id": "uuid",
        "store_name": "SuperMaxi",
        "branch": "North",
        "average_price": 1.35,
        "min_price": 1.20,
        "currency": "USD",
        "sample_count": 32,
        "last_observed_at": "2025-06-27T10:00:00Z",
        "distance_km": null
      }
    ],
    "Bs.": [
      {
        "store_id": "uuid",
        "store_name": "Central Market",
        "branch": "Downtown",
        "average_price": 125.00,
        "min_price": 110.00,
        "currency": "Bs.",
        "sample_count": 20,
        "last_observed_at": "2025-06-28T14:30:00Z",
        "distance_km": null
      }
    ]
  }
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `product_id` | string (UUID) | Canonical product identifier |
| `product_name` | string | Normalized product name |
| `currency_rankings` | object | Map keyed by currency code ("USD", "Bs.") |
| `currency_rankings[currency][]` | array | Stores ranked cheapest first |
| `…[].store_id` | string (UUID) | Store identifier |
| `…[].store_name` | string | Store name |
| `…[].branch` | string? | Branch name (omitted if null) |
| `…[].average_price` | number | Average price from fresh observations |
| `…[].min_price` | number | Minimum observed price |
| `…[].currency` | string | Currency code (matches the grouping key) |
| `…[].sample_count` | int | Number of fresh observations included |
| `…[].last_observed_at` | string (ISO 8601) | Most recent observation timestamp |
| `…[].distance_km` | number? | Distance from user (null when no location provided) |

### Empty Ranking (`200`)

A valid product with no observations returns an empty map:

```json
{
  "product_id": "uuid",
  "product_name": "Arroz Blanco",
  "currency_rankings": {}
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `400` | `{"error": "invalid product id"}` | `id` is not a valid UUID |
| `401` | `{"error": "invalid or expired token"}` | Token invalid or missing |
| `404` | `{"error": "product not found"}` | Product does not exist |