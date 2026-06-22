# Contract: Health Endpoint

**Endpoint**: `GET /api/v1/health`
**Feature**: 001-backend-skeleton
**Version**: v1

## Request

```
GET /api/v1/health
```

- No authentication required (this is the skeleton; auth added in E2).
- No query parameters.
- No request body.

## Response

### 200 OK — All Dependencies Healthy

```json
{
  "status": "ok",
  "dependencies": {
    "postgres": {
      "name": "postgres",
      "reachable": true
    },
    "redis": {
      "name": "redis",
      "reachable": true
    }
  }
}
```

### 200 OK — Degraded (One or More Dependencies Unreachable)

```json
{
  "status": "degraded",
  "dependencies": {
    "postgres": {
      "name": "postgres",
      "reachable": false,
      "error": "failed to ping postgres: dial tcp 127.0.0.1:5432: connect: connection refused"
    },
    "redis": {
      "name": "redis",
      "reachable": true
    }
  }
}
```

**Rules**:
- HTTP status is always `200` when the handler itself is functioning. Do NOT use `503`
  for degraded dependencies — `503` means the service is unavailable, but the service IS
  available; its dependencies are not.
- `error` is present only when `reachable` is `false`. Omit `error` when `reachable` is
  `true`.
- Each dependency is checked independently. One failing dependency does not prevent
  checking others.

### Response Schema (TypeScript-like for documentation)

```typescript
interface HealthResponse {
  status: "ok" | "degraded";
  dependencies: {
    [name: string]: {
      name: string;
      reachable: boolean;
      error?: string;    // present only when reachable === false
    };
  };
}
```

## Error Responses

None expected from this endpoint. The handler catches internal panics via middleware and
returns `500` with a generic error body, but the health-check logic itself does not throw
— it always produces a valid `HealthResponse`.

| Status | Condition |
|--------|-----------|
| `200`  | Handler functional (dependencies may be healthy or degraded) |
| `500`  | Unexpected internal error (handler crash, middleware panic) |

## Contract Evolution

- **v1**: `status` + `dependencies` object with per-dependency `reachable` boolean.
- **v2 (future)**: May add `dependencies.minio` when MinIO health-check is required, or
  add a `timestamp` field. These are backward-compatible additions that do not break v1
  clients.
- **Breaking change** (e.g., renaming `dependencies` to `checks`): triggers a new path
  version `/api/v2/health` per Art. IV.1 of the constitution.
