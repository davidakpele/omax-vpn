# Security

## Principles

1. No custom cryptography — WireGuard handles all VPN cryptography.
2. Secrets are never committed to version control.
3. Private WireGuard keys remain on the client wherever possible.
4. Least privilege throughout — every component has only the permissions it needs.
5. Defense in depth — multiple independent security controls.
6. Fail closed — unauthenticated or invalid requests are rejected.
7. Audit everything — all privileged operations are logged immutably.

---

## Authentication

### Users

- Passwords are hashed with **bcrypt** (cost factor ≥ 12). Plaintext passwords are never stored or logged.
- **JWT** access tokens (15-minute expiry by default) are issued on login.
- **Refresh tokens** (7-day expiry by default) are stored server-side and can be revoked.
- Token signing uses HS256 with a 64-byte random secret (`JWT_SECRET`).
- Tokens are validated on every protected request — the API does not trust cookies or query parameters.

### Go ↔ Rust engine

- All requests from the Go API to the Rust engine carry a `Authorization: Bearer <token>` header.
- The token is a pre-shared 32-byte random hex value (`VPN_ENGINE_TOKEN`).
- The Rust engine validates the token on every request using constant-time comparison (no timing oracle).
- Source IP is not trusted as an authentication factor.

---

## Authorization

Three roles are defined:

| Role | Description |
|------|-------------|
| `USER` | Authenticated end user |
| `ADMIN` | Platform administrator |
| `SYSTEM` | Internal service-to-service identity |

- All privileged operations are authorized server-side.
- The frontend (client) is never trusted for authorization decisions.
- Administrative endpoints are protected by role checks and separated from user-facing endpoints.

---

## Secret Management

| Secret | Storage |
|--------|---------|
| `POSTGRES_PASSWORD` | Environment variable, never committed |
| `JWT_SECRET` | Environment variable, never committed |
| `VPN_ENGINE_TOKEN` | Environment variable, never committed |
| WireGuard server private key | Environment variable or secret manager — never in plain DB |
| WireGuard client private key | Generated and stored on the client — never sent to the server |

### Rules

- `.env` is gitignored. Only `.env.example` (with placeholder values) is committed.
- No secrets in source code, comments, or log output.
- In production, use a secrets manager (HashiCorp Vault, AWS Secrets Manager, etc.) instead of environment variables.
- Rotate secrets immediately if exposure is suspected.

---

## Transport Security

| Link | Development | Production |
|------|------------|------------|
| Client → Go API | HTTPS (termination at load balancer) | TLS required |
| Go API → Rust engine | HTTP over Docker internal network | mTLS required |
| Client → WireGuard | WireGuard UDP (always encrypted) | WireGuard UDP |

WireGuard provides authenticated encryption for all VPN traffic using its built-in Noise Protocol Framework implementation.

---

## Input Validation

- All API request bodies are validated before processing.
- UUIDs are validated as proper UUID format.
- WireGuard public keys are validated as 44-character base64 strings before use.
- Shell commands in the Rust engine are never constructed from raw user input.
- Firewall and routing rules use validated, typed parameters only — not interpolated strings from API requests.

---

## Logging

Structured JSON logging is used throughout. The following values are **never** logged:

- Passwords (plaintext or hashed)
- JWT secrets
- Access or refresh tokens
- WireGuard private keys
- Database connection strings containing passwords
- Payment card data

Log fields that identify a user use `user_id` (UUID), not email or name.

---

## Audit Log

An immutable `audit_logs` table records all significant operations:

```
USER_CREATED
USER_DELETED
DEVICE_REGISTERED
DEVICE_DELETED
VPN_SERVER_CREATED
VPN_PEER_CREATED
VPN_PEER_REMOVED
VPN_SESSION_STARTED
VPN_SESSION_ENDED
SUBSCRIPTION_CREATED
SUBSCRIPTION_CANCELLED
ADMIN_ACTION
LOGIN_SUCCESS
LOGIN_FAILURE
TOKEN_REFRESH
```

Audit log rows are append-only. Application users cannot modify or delete audit history. Database-level row security enforces this.

---

## Network Security

- The Rust engine applies firewall rules before any peer is allowed to route traffic.
- Default-deny firewall policy — only explicitly permitted traffic flows.
- NAT masquerade is applied only to the VPN interface's traffic.
- WireGuard listens only on the configured port (default 51820/udp).
- The engine's control API (9090/tcp) must not be exposed publicly — bind to a private interface only.

---

## Dependency Security

- Dependencies are pinned to exact versions.
- `go mod tidy` and `cargo update` are run deliberately, not automatically.
- The dependency list is reviewed before each release.
- Security advisories are monitored via `govulncheck` (Go) and `cargo audit` (Rust).

---

## Known Limitations (Phase 1)

- The kill switch is not yet implemented on clients — traffic may escape outside the tunnel if the VPN connection drops unexpectedly.
- The Go ↔ Rust control channel uses a pre-shared token rather than mTLS — mTLS is planned for Phase 5.
- WireGuard private key management on server nodes is not yet automated — planned for Phase 7.
- Rate limiting on the API is not yet implemented — planned for Phase 2.

See [docs/threat-model.md](threat-model.md) for the complete threat analysis.
