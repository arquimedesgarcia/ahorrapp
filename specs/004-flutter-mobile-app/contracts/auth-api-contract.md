# Contract: Authentication API (`/api/v1/auth`)

Flutter app MUST consume ONLY these documented endpoints. No undocumented fields or behaviors.

## POST `/api/v1/auth/register`

Register a new user account.

- **Auth**: None
- **Content-Type**: `application/json`

### Request Body

```json
{
  "email": "user@example.com",
  "password": "secret1234",
  "display_name": "Maria Lopez"
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `email` | string | yes | Valid email format |
| `password` | string | yes | Min 8 characters |
| `display_name` | string | yes | 1-100 characters |

### Success (`201`)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "Maria Lopez"
  }
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `400` | `{"error": "invalid email format"}` | Malformed email |
| `400` | `{"error": "password too short"}` | Password < 8 chars |
| `409` | `{"error": "email already registered"}` | Duplicate email |

---

## POST `/api/v1/auth/login`

Authenticate an existing user.

- **Auth**: None
- **Content-Type**: `application/json`

### Request Body

```json
{
  "email": "user@example.com",
  "password": "secret1234"
}
```

### Success (`200`)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "Maria Lopez"
  }
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `401` | `{"error": "invalid credentials"}` | Wrong email or password |

---

## GET `/api/v1/auth/me`

Return the currently authenticated user's profile.

- **Auth**: Required (`Authorization: Bearer <token>`)

### Success (`200`)

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "display_name": "Maria Lopez"
}
```

### Errors

| Status | Body | Condition |
|--------|------|-----------|
| `401` | `{"error": "invalid or expired token"}` | Token invalid/expired |
