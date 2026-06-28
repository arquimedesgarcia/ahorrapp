# Contract: Receipt API (`/api/v1`)

## POST `/api/v1/receipts`

- **Auth**: required
- **Content-Type**: `multipart/form-data`
- **Field**: `image`

### Success (`202`)

```json
{
  "receipt_id": "uuid",
  "status": "PENDING",
  "duplicate": false
}
```

### Duplicate Idempotent Success (`202`)

```json
{
  "receipt_id": "uuid",
  "status": "PENDING",
  "duplicate": true
}
```

## GET `/api/v1/receipts/{id}`

- **Auth**: required (owner)

### Success (`200`)

```json
{
  "receipt_id": "uuid",
  "status": "NEEDS_REVIEW",
  "store": {
    "name": "Central Market",
    "branch": "Downtown",
    "address": "Av. X #123"
  },
  "purchase_date": "2026-06-24",
  "total": 42.5,
  "items": [
    {
      "raw_text": "ARROZ 1KG",
      "quantity": 1,
      "unit_price": 2.4,
      "currency": "USD"
    }
  ]
}
```

## POST `/api/v1/receipts/{id}/confirm`

- **Auth**: required (owner)
- **Body**:

```json
{
  "store": {
    "name": "Central Market",
    "branch": "Downtown",
    "address": "Av. X #123"
  },
  "purchase_date": "2026-06-24",
  "total": 42.5,
  "items": [
    {
      "raw_text": "ARROZ 1KG",
      "quantity": 1,
      "unit_price": 2.4,
      "currency": "USD"
    }
  ]
}
```

### Success (`204`)

- Empty body.

### Validation failure (`400`)

```json
{
  "error": "confirm receipt: item currency is required"
}
```
