# Contract: User Profile API (`/api/v1/users`)

Flutter app MUST consume ONLY these documented endpoints. No undocumented fields or behaviors.

## GET `/api/v1/users/me/points`

Return the authenticated user's accumulated loyalty points and recent transaction history.

- **Auth**: Required (`Authorization: Bearer <token>`)

### Success (`200`)

```json
{
  "total_points": 350,
  "recent_transactions": [
    {
      "id": "uuid",
      "points": 10,
      "reason": "Receipt confirmed",
      "created_at": "2026-06-25T10:30:00Z"
    },
    {
      "id": "uuid",
      "points": 10,
      "reason": "Receipt confirmed",
      "created_at": "2026-06-24T14:15:00Z"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `total_points` | int | Lifetime accumulated points |
| `recent_transactions` | array | Recent points transactions (optional, may be empty/omitted) |
| `recent_transactions[].id` | string (UUID) | Transaction ID |
| `recent_transactions[].points` | int | Points earned (positive) or deducted (negative) |
| `recent_transactions[].reason` | string | Human-readable reason |
| `recent_transactions[].created_at` | string | ISO 8601 datetime |

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `401` | `{"error": "invalid or expired token"}` | Token invalid |
