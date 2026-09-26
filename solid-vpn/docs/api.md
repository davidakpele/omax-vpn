# API Reference

Base URL: `/api/v1`

Full OpenAPI specification: [`packages/protocol/openapi.yaml`](../packages/protocol/openapi.yaml)

---

## Authentication

All endpoints except `/health`, `/ready`, and auth endpoints require a valid JWT access token:

```
Authorization: Bearer <access_token>
```

---

## Error Format

All error responses use a consistent structure:

```json
{
  "error": {
    "code": "DEVICE_NOT_FOUND",
    "message": "The requested device does not exist."
  }
}
```

Internal stack traces are never exposed in responses. Use the request ID (`X-Request-ID` header) for debugging.

---

## Health

### GET /health

Liveness probe — always returns 200 while the process is alive.

**Response 200**
```json
{
  "status": "ok",
  "timestamp": "2026-09-26T10:00:00Z",
  "service": "vpn-api"
}
```

### GET /ready

Readiness probe — returns 200 only when all dependencies are healthy.

**Response 200**
```json
{
  "status": "ready",
  "checks": {
    "database": "healthy"
  },
  "timestamp": "2026-09-26T10:00:00Z"
}
```

**Response 503** (dependency unavailable)
```json
{
  "status": "not ready",
  "checks": {
    "database": "unhealthy: connection refused"
  },
  "timestamp": "2026-09-26T10:00:00Z"
}
```

---

## Authentication Endpoints

### POST /api/v1/auth/register

Register a new user account.

**Request**
```json
{
  "email": "user@example.com",
  "password": "strongpassword"
}
```

**Response 201**
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "created_at": "2026-09-26T10:00:00Z"
}
```

---

### POST /api/v1/auth/login

Authenticate and receive tokens.

**Request**
```json
{
  "email": "user@example.com",
  "password": "strongpassword"
}
```

**Response 200**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 900
}
```

---

### POST /api/v1/auth/refresh

Exchange a refresh token for a new access token.

**Request**
```json
{
  "refresh_token": "eyJ..."
}
```

**Response 200**
```json
{
  "access_token": "eyJ...",
  "expires_in": 900
}
```

---

### POST /api/v1/auth/logout

Revoke the current refresh token.

**Response 204** (no content)

---

## User Endpoints

### GET /api/v1/users/me

Get the authenticated user's profile.

**Response 200**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "created_at": "2026-09-26T10:00:00Z",
  "updated_at": "2026-09-26T10:00:00Z"
}
```

### PATCH /api/v1/users/me

Update user profile.

### DELETE /api/v1/users/me

Delete user account.

---

## Device Endpoints

### GET /api/v1/devices

List the authenticated user's registered devices.

### POST /api/v1/devices

Register a new device.

**Request**
```json
{
  "name": "My Laptop",
  "public_key": "<wireguard public key>"
}
```

**Response 201**
```json
{
  "id": "uuid",
  "name": "My Laptop",
  "public_key": "<wireguard public key>",
  "created_at": "2026-09-26T10:00:00Z"
}
```

### GET /api/v1/devices/:id

Get a specific device.

### DELETE /api/v1/devices/:id

Delete a device and remove its WireGuard peer from all servers.

---

## VPN Endpoints

### GET /api/v1/vpn/servers

List available VPN servers.

**Response 200**
```json
{
  "servers": [
    {
      "id": "uuid",
      "name": "Nigeria-01",
      "country": "NG",
      "city": "Lagos",
      "status": "healthy",
      "load": 0.42
    }
  ]
}
```

### GET /api/v1/vpn/servers/:id

Get a specific VPN server.

### POST /api/v1/vpn/connect

Initiate a VPN connection.

**Request**
```json
{
  "device_id": "uuid",
  "server_id": "uuid"
}
```

**Response 200** — returns a WireGuard configuration
```json
{
  "session_id": "uuid",
  "wireguard_config": "[Interface]\nPrivateKey=...\n\n[Peer]\n...",
  "server_public_key": "...",
  "assigned_ip": "10.8.0.5",
  "dns": "1.1.1.1"
}
```

### POST /api/v1/vpn/disconnect

Terminate a VPN session.

**Request**
```json
{
  "session_id": "uuid"
}
```

**Response 204**

### GET /api/v1/vpn/config

Get the WireGuard configuration for an existing peer (without creating a new session).

---

## Session Endpoints

### GET /api/v1/vpn/sessions

List VPN sessions for the authenticated user.

### GET /api/v1/vpn/sessions/:id

Get a specific session.

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request — validation error |
| 401 | Unauthorized — missing or invalid token |
| 403 | Forbidden — insufficient permissions |
| 404 | Not Found |
| 409 | Conflict — e.g. device already registered |
| 422 | Unprocessable Entity |
| 503 | Service Unavailable — e.g. no healthy VPN server |
| 500 | Internal Server Error |
