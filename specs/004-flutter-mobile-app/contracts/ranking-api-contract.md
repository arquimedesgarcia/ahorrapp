# Contract: Product Ranking API (`/api/v1/ranking`)

Flutter app MUST consume ONLY these documented endpoints. No undocumented fields or behaviors.

## GET `/api/v1/ranking/products/search`

Search products and return stores ranked by cheapest price.

- **Auth**: Required (`Authorization: Bearer <token>`)

### Query Parameters

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `q` | string | yes | Search query (product name) |

### Success (`200`)

```json
{
  "results": [
    {
      "product_id": "uuid",
      "product_name": "Arroz Blanco",
      "unit": "kg",
      "stores": [
        {
          "store_id": "uuid",
          "store_name": "Central Market",
          "branch": "Downtown",
          "average_price": 1.25,
          "currency": "USD",
          "sample_count": 45
        },
        {
          "store_id": "uuid",
          "store_name": "SuperMaxi",
          "branch": "North",
          "average_price": 1.35,
          "currency": "USD",
          "sample_count": 32
        }
      ]
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `results` | array | Product matches (may be empty) |
| `results[].product_id` | string (UUID) | Canonical product identifier |
| `results[].product_name` | string | Normalized product name |
| `results[].unit` | string? | Unit of measure (kg, lt, unit) |
| `results[].stores` | array | Stores ranked cheapest first |
| `results[].stores[].store_id` | string (UUID) | Store identifier |
| `results[].stores[].store_name` | string | Store name |
| `results[].stores[].branch` | string? | Branch/location name |
| `results[].stores[].average_price` | number | Average price from observations |
| `results[].stores[].currency` | string | "USD" or "Bs." (prices never mixed) |
| `results[].stores[].sample_count` | int | Number of observations |

### Empty Results (`200`)

```json
{
  "results": []
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `400` | `{"error": "query parameter 'q' is required"}` | Missing query |
| `401` | `{"error": "invalid or expired token"}` | Token invalid |
