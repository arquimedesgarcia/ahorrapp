# Contract: `GET /api/v1/me/loyalty`

**Version**: v1 (Constitution Art. IV — versioned contracts; breaking
change requires `/api/v2`). Coexists with the legacy
`GET /api/v1/users/me/points`. The legacy endpoint keeps its existing
shape (`total_points` + `recent_transactions`) for the deployed Flutter
client; the new canonical endpoint `/api/v1/me/loyalty` returns the
richer shape documented below. Both return data sourced from the same
underlying use case / tables.

## Endpoint

| Method | Path                | Auth         | Purpose                                            |
|--------|---------------------|--------------|----------------------------------------------------|
| GET    | `/api/v1/me/loyalty`| JWT required | Returns the authenticated user's point balance and recent movement history. |

## Authentication

- `Authorization: Bearer <JWT>`.
- The user identity (`user_id` claim) is extracted by `JWTMiddleware`
  and used as the only filter; the request body and query string are
  ignored.
- A request without a valid token → `401 Unauthorized` (Art. VII, FR-010).

## Request

No request body. No query parameters in the MVP.

## Response — `200 OK`

```json
{
  "balance": 18,
  "history": [
    {
      "id": "f3b4...",
      "points": 18,
      "reason": "receipt_confirmed;first_observation_product;data_completion",
      "created_at": "2026-06-29T14:32:01Z",
      "receipt_id": "8d1e..."
    },
    {
      "id": "aa01...",
      "points": 0,
      "reason": "daily_limit_reached",
      "created_at": "2026-06-29T20:01:55Z",
      "receipt_id": "c2c2..."
    }
  ]
}
```

| Field        | Type    | Always Present | Notes |
|--------------|---------|----------------|-------|
| `balance`    | integer | yes            | Equivalent to `SUM(history.points)`. Fits the spec FR-011 and SC-003. |
| `history`    | array   | yes            | Empty array `[]` for a user with no movements (FR user story 3, scenario 2). |
| `history[].id` | string | yes          | UUID of the `loyalty_transactions` row. |
| `history[].points` | integer | yes     | May be `0` (daily-limit) or positive. (Negative reserved for future redemption, out of scope.) |
| `history[].reason` | string | yes       | One of the documented reason codes, `;`-joined when multiple. |
| `history[].created_at` | string (RFC 3339) | yes | UTC ISO 8601 timestamp. |
| `history[].receipt_id` | string (UUID) | no  | Present iff the movement is tied to a confirmed receipt; omitted for future non-receipt movements. |

### Ordering and limits

- History is ordered by `created_at DESC`.
- Limited to the latest 100 movements (R-06). Pagination is out of scope
  for this feature (Art. VI.1) and may be added in a later feature
  without changing this contract semantically (cursor query parameters).

## Errors

| Status | Body                                                       | When |
|--------|------------------------------------------------------------|------|
| 401    | `{ "error": "invalid or expired token" }`                 | Missing/invalid JWT (FR-010, SC-007). |
| 405    | `{ "error": "method not allowed" }`                       | Non-GET request. |
| 500    | `{ "error": "internal server error" }`                    | Unexpected DB failure. |

## Cross-user isolation

The endpoint MUST NOT accept a user identifier in the request; the user
is always taken from the JWT (FR-009, SC-007). Two simultaneous users
with different tokens always observe their own data.

## Idempotency / safety

This endpoint is read-only (it does not award points). The data it
returns is the consequence of the award use case invoked by
`POST /api/v1/receipts/{id}/confirm`. Resubmitting the same receipt to
the confirm endpoint grants no additional points (FR-004) and therefore
this endpoint's balance does not change between calls.

## Example: curl

```bash
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/me/loyalty
```

## Relationship with `/api/v1/users/me/points`

The legacy `GET /api/v1/users/me/points` endpoint returns its existing
shape (`{ "total_points": int, "recent_transactions": [ {...} ] }`,
capped at 10 movements). It is left unchanged so the deployed Flutter
client continues to work. The new `GET /api/v1/me/loyalty` is the
canonical endpoint going forward, returning the richer shape documented
above. Both are backed by the same query use case and the same
`loyalty_transactions` / `users.points` data. A future feature may
deprecate `/users/me/points` with a v2 breaking-change notice per
Constitution Art. IV.