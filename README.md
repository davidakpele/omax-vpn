# Solid VPN

A VPN platform with a strict separation between the control plane (Go) and the data plane (Rust), using WireGuard as the underlying VPN protocol.

---

## Architecture

```
              ┌──────────────────┐
              │     Clients      │
              │  (any WireGuard  │
              │   compatible)    │
              └────────┬─────────┘
                       │ HTTPS
                       ▼
              ┌──────────────────┐
              │   Go Control     │
              │     Plane        │
              │                  │
              │  Auth & JWT      │
              │  Users           │
              │  Devices         │
              │  Server Select   │
              │  Peer Lifecycle  │
              │  Sessions        │
              │  Audit Logging   │
              └────────┬─────────┘
                       │ Bearer Token (HTTP)
                       │ mTLS (planned)
                       ▼
              ┌──────────────────┐
              │   Rust Data      │
              │     Plane        │
              │                  │
              │  axum HTTP API   │
              │  WireGuard mgmt  │
              │  Routing         │
              │  NAT / iptables  │
              │  Firewall        │
              │  DNS             │
              │  Prometheus      │
              └────────┬─────────┘
                       │
                       ▼
                   Internet
```

**Go** owns all business logic — authentication, user management, device registration, server selection, peer lifecycle, session tracking, and audit logging.

**Rust** owns all network operations — the axum HTTP control API, WireGuard peer management, routing, NAT, firewall, and DNS. Currently the network calls are stubs; the control API is fully wired and operational.

**WireGuard** provides the cryptographic VPN tunnel. No custom cryptography is implemented anywhere in the codebase.

---

## What Is Built

### Go Control Plane (`apps/api`)

**Authentication**

- `POST /api/v1/auth/register` — register with email + password (bcrypt, min 12 chars)
- `POST /api/v1/auth/login` — returns JWT access token (15 min) + refresh token (7 days)
- `POST /api/v1/auth/refresh` — rotating refresh tokens (old token revoked on use)
- `POST /api/v1/auth/logout` — revoke refresh token server-side

**Users**

- `GET /api/v1/users/me` — authenticated profile
- `PATCH /api/v1/users/me` — update email
- `DELETE /api/v1/users/me` — soft delete

**Devices**

- `POST /api/v1/devices` — register a WireGuard public key
- `GET /api/v1/devices` — list user's devices
- `GET /api/v1/devices/:id` — get device
- `DELETE /api/v1/devices/:id` — revoke device

**VPN Servers**

- `GET /api/v1/vpn/servers` — list healthy servers (filterable by country, region)
- `GET /api/v1/vpn/servers/:id` — get a specific server
- `GET /api/v1/vpn/regions` — list regions

**VPN Lifecycle**

- `POST /api/v1/vpn/connect` — allocate IP, create peer DB record, call Rust engine to add WireGuard peer, return `.conf` file string to client
- `POST /api/v1/vpn/disconnect` — end session, call Rust engine to remove peer, release IP
- `GET /api/v1/vpn/config` — retrieve config for an existing peer without creating a new session
- `GET /api/v1/vpn/sessions` — paginated session history
- `GET /api/v1/vpn/sessions/:id` — get a specific session

**Health**

- `GET /health` — liveness (always 200 while running)
- `GET /ready` — readiness (checks PostgreSQL + Rust engine)

**Audit Logging**

Every significant operation writes to the immutable `audit_logs` table:

| Event                 | Trigger                 |
| --------------------- | ----------------------- |
| `USER_CREATED`        | Successful registration |
| `USER_LOGIN`          | Successful login        |
| `USER_LOGIN_FAILED`   | Failed password check   |
| `USER_DELETED`        | Account deletion        |
| `DEVICE_REGISTERED`   | Device creation         |
| `DEVICE_DELETED`      | Device revocation       |
| `VPN_SESSION_STARTED` | Successful connect      |
| `VPN_SESSION_ENDED`   | Disconnect              |

**Server Selection**

The `Select()` algorithm picks the lowest-load healthy server. If a preferred `server_id` is supplied it is used when selectable, otherwise falls back to best-available. Filters by country when candidates exist for the requested country.

### Rust Data Plane (`apps/vpn-engine`)

**axum HTTP control API** on `0.0.0.0:9090`

| Endpoint                    | Auth         | Description                             |
| --------------------------- | ------------ | --------------------------------------- |
| `GET /health`               | None         | Liveness check                          |
| `GET /metrics`              | None         | Prometheus text exposition              |
| `POST /peers`               | Bearer token | Add WireGuard peer + firewall + routing |
| `DELETE /peers/:public_key` | Bearer token | Remove peer + revoke firewall + routing |

All `/peers` requests are validated with constant-time bearer token comparison. `/health` and `/metrics` are intentionally public for Docker healthchecks and Prometheus scraping.

**Structured logging** — JSON output via `tracing-subscriber`.

**Prometheus metrics** — `vpn_active_sessions`, `vpn_sessions_total`, `vpn_connection_errors_total`, `vpn_peer_count`, `vpn_engine_uptime_seconds`.

**In-memory peer state** — `Arc<RwLock<HashMap<PublicKey, AssignedIP>>>` tracks active peers for firewall/routing cleanup on removal.

Network functions (`tunnel/`, `network/`) are wired into the handler chain but are currently stubs that log operations. They are the correct shape for real WireGuard system calls — see [Next Features](#next-features) below.

### Database (`PostgreSQL 16`)

11 tables managed by sequential SQL migrations:

| Table            | Purpose                                           |
| ---------------- | ------------------------------------------------- |
| `users`          | User accounts with bcrypt password hashes         |
| `refresh_tokens` | Server-side refresh token store (revocable)       |
| `devices`        | Registered WireGuard client public keys           |
| `vpn_regions`    | Geographic region definitions                     |
| `vpn_servers`    | VPN server nodes with capacity and status         |
| `vpn_peers`      | WireGuard peer records (public key + assigned IP) |
| `ip_allocations` | IP address pool per server (pessimistic locking)  |
| `vpn_sessions`   | Active and historical session records             |
| `subscriptions`  | Subscription plan records (schema ready)          |
| `plans`          | Available subscription tiers (schema ready)       |
| `payments`       | Payment records (schema ready)                    |
| `audit_logs`     | Append-only immutable event log                   |

### Infrastructure

- **Docker Compose** — `postgres`, `vpn-engine`, `api`, `prometheus`, `grafana` with health checks and ordered startup (`postgres` healthy → `vpn-engine` healthy → `api`)
- **Prometheus** — scrapes `/metrics` from both API and engine
- **Grafana** — auto-provisioned with Prometheus datasource

---

## Quick Start

### Prerequisites

| Tool           | Minimum version |
| -------------- | --------------- |
| Go             | 1.23            |
| Rust           | 1.88            |
| Docker         | 24              |
| Docker Compose | v2              |

### 1. Configure

```bash
cp .env.example .env
```

Edit `.env` and fill in:

```bash
POSTGRES_PASSWORD=$(openssl rand -hex 16)
JWT_SECRET=$(openssl rand -hex 64)
VPN_ENGINE_TOKEN=$(openssl rand -hex 32)
VPN_ENGINE_SERVER_ID=$(uuidgen)
```

### 2. Start

```bash
docker compose up -d
```

Wait for all services to be healthy (~30 seconds on first run due to Rust compilation):

```bash
docker compose ps
```

### 3. Run migrations

```bash
for f in apps/api/migrations/*.sql; do
  docker exec -i solid-vpn-postgres-1 psql -U postgres -d solidvpn < "$f"
done
```

On Windows (PowerShell):

```powershell
Get-ChildItem apps\api\migrations\*.sql | ForEach-Object {
    Get-Content $_.FullName | docker exec -i solid-vpn-postgres-1 psql -U postgres -d solidvpn
}
```

### 4. Verify

```bash
curl http://localhost:8080/health
# {"status":"ok","service":"vpn-api"}

curl http://localhost:8080/ready
# {"status":"ready","checks":{"database":"healthy","vpn_engine":"healthy"}}
```

---

## API Walkthrough

### Register and authenticate

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"strongpassword123"}'

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"strongpassword123"}' \
  | jq -r '.access_token')
```

### Register a device

```bash
curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Laptop","public_key":"mNb8O2FNkBkJo5tXhA3UGN4sbeE6DKBP3gRKe3DXnWk="}'
```

For a real tunnel, generate your own keypair first:

```bash
wg genkey | tee private.key | wg pubkey
```

### List servers

```bash
curl -s http://localhost:8080/api/v1/vpn/servers \
  -H "Authorization: Bearer $TOKEN" | jq '.servers[] | {id, name, country, status}'
```

Seeded servers:

| ID                                     | Name      | Country | City     |
| -------------------------------------- | --------- | ------- | -------- |
| `10000000-0000-0000-0000-000000000001` | NG-LAG-01 | NG      | Lagos    |
| `10000000-0000-0000-0000-000000000002` | GB-LON-01 | GB      | London   |
| `10000000-0000-0000-0000-000000000003` | US-NYC-01 | US      | New York |

### Connect to a server

```bash
curl -s -X POST http://localhost:8080/api/v1/vpn/connect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "<your device id>",
    "server_id": "10000000-0000-0000-0000-000000000002"
  }'
```

Response:

```json
{
  "session_id": "4d20ab3e-...",
  "wireguard_config": "[Interface]\nAddress = 10.8.0.5/32\nDNS = 1.1.1.1\n\n[Peer]\n...",
  "server_public_key": "...",
  "assigned_ip": "10.8.0.5",
  "dns": "1.1.1.1"
}
```

### Disconnect

```bash
curl -s -X POST http://localhost:8080/api/v1/vpn/disconnect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"session_id": "<session_id from connect>"}'
```

### Inspect audit log

```bash
docker exec solid-vpn-postgres-1 psql -U postgres -d solidvpn \
  -c "SELECT action, resource_type, metadata, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 10;"
```

---

## Development Commands

```bash
make build          # compile Go API + Rust engine locally
make test           # run all tests
make lint           # go vet + cargo clippy
make format         # gofmt + rustfmt
make docker-up      # start dev environment
make docker-down    # stop dev environment
make docker-logs    # follow all service logs
make migrate        # apply database migrations
make clean          # remove build artifacts
make help           # list all targets
```

---

## Environment Variables

See [`.env.example`](.env.example) for all variables with descriptions.

| Variable                    | Required | Default       | Description                              |
| --------------------------- | -------- | ------------- | ---------------------------------------- |
| `POSTGRES_PASSWORD`         | Yes      | —             | Database password                        |
| `JWT_SECRET`                | Yes      | —             | 64-byte hex, signs JWT tokens            |
| `VPN_ENGINE_TOKEN`          | Yes      | —             | 32-byte hex, authenticates Go→Rust calls |
| `VPN_ENGINE_SERVER_ID`      | Yes      | —             | UUID identifying this VPN node           |
| `VPN_DNS`                   | No       | `1.1.1.1`     | DNS pushed to VPN peers                  |
| `VPN_ENGINE_PEER_CIDR`      | No       | `10.8.0.0/24` | IP pool for peer allocation              |
| `VPN_ENGINE_WIREGUARD_PORT` | No       | `51820`       | WireGuard listen port                    |
| `LOG_LEVEL`                 | No       | `info`        | trace / debug / info / warn / error      |

---

## Project Layout

```
solid-vpn/
├── apps/
│   ├── api/                    Go control-plane
│   │   ├── cmd/api/main.go     Entry point, dependency wiring
│   │   ├── config/             Environment-based config
│   │   ├── internal/
│   │   │   ├── audit/          Audit log service + repository
│   │   │   ├── auth/           Register, login, JWT, refresh tokens
│   │   │   ├── devices/        Device registration and management
│   │   │   ├── engine/         Authenticated HTTP client for Rust engine
│   │   │   ├── health/         Liveness + readiness handlers
│   │   │   ├── middleware/      JWT auth middleware, request logger
│   │   │   ├── servers/        VPN server management + selection
│   │   │   ├── users/          User profile management
│   │   │   └── vpn/            Peer lifecycle, sessions, WireGuard config
│   │   ├── migrations/         Sequential SQL migrations (000001–000008)
│   │   └── routes/             HTTP router wiring
│   │
│   ├── vpn-engine/             Rust data-plane
│   │   └── src/
│   │       ├── control/        axum HTTP server (health, metrics, peers)
│   │       ├── config/         Settings from VPN_ENGINE_* env vars
│   │       ├── tunnel/         WireGuard interface + peer stubs
│   │       ├── network/        Routing, firewall, NAT, DNS stubs
│   │       ├── sessions/       In-memory session manager
│   │       ├── security/       Token validation (constant-time), key format check
│   │       ├── telemetry/      Prometheus metrics
│   │       └── errors.rs       Unified EngineError type
│   │
│   └── agent/                  Node agent (Phase 7 placeholder)
│
├── infrastructure/
│   ├── docker/                 api.Dockerfile, vpn.Dockerfile
│   └── scripts/                seed.sql
├── monitoring/
│   ├── prometheus/             prometheus.yml
│   └── grafana/                Datasource provisioning
├── packages/protocol/          openapi.yaml
└── docs/                       Architecture, security, networking, deployment, threat model
```

---

## Next Features

### Immediate — Real WireGuard system calls (Rust)

The stubs in `tunnel/` and `network/` need to be replaced with actual system operations. This requires a Linux node (WireGuard is a Linux kernel module).

**`tunnel/wireguard.rs`** — create and configure the WireGuard interface:

```rust
use wireguard_control::{DeviceUpdate, InterfaceName, Key};

pub async fn init_interface(interface: &str, port: u16, private_key: &Key) {
    let name = InterfaceName::from_str(interface).unwrap();
    DeviceUpdate::new()
        .set_private_key(private_key.clone())
        .set_listen_port(port)
        .apply(&name, Backend::Kernel)
        .unwrap();
}
```

**`tunnel/peer.rs`** — add/remove peers:

```rust
DeviceUpdate::new()
    .add_peer(PeerConfigBuilder::new(&public_key)
        .replace_allowed_ips()
        .allow_ip(assigned_ip.parse().unwrap()))
    .apply(&name, Backend::Kernel)
    .unwrap();
```

**`network/firewall.rs`** — iptables rules via `iptables` crate or `std::process::Command` with validated inputs.

**`network/nat.rs`** — masquerade rule for outbound traffic.

**`network/routing.rs`** — `ip route add` via netlink or command.

Crates to add: `wireguard-control`, `iptables`, `rtnetlink`.

---

### Phase 7 — Security Hardening

- **mTLS between Go and Rust** — replace pre-shared bearer token with mutual TLS certificates
- **WireGuard server private key management** — store in HashiCorp Vault or AWS Secrets Manager, never in plain env vars
- **Rate limiting** — per-IP on auth endpoints, per-user on connect
- **JWT rotation** — automated signing key rotation
- **Dependency audit** — `govulncheck` in CI, `cargo audit` in CI
- **Secret scanning** — `gitleaks` pre-commit hook

---

### Phase 8 — Monitoring and Alerting

- Grafana dashboards for active sessions, connection errors, server load per region
- Prometheus alerts: server capacity > 80%, engine unreachable, high auth failure rate
- Log aggregation — ship structured JSON logs to a centralised store (Loki, CloudWatch, etc.)
- Distributed tracing — OpenTelemetry spans across Go→Rust control calls

---

### Phase 9 — Infrastructure

- **Terraform** — provision VPS nodes (DigitalOcean, Hetzner, or AWS) per region
- **Ansible** — configure Linux nodes: WireGuard kernel module, sysctl IP forwarding, iptables baseline, deploy vpn-engine binary
- **Multi-region routing** — register real server rows in the database with actual public IPs and generated WireGuard keypairs
- **Health-based server rotation** — mark servers offline when engine heartbeat is missed

---

### Phase 10 — Client Applications

The `wireguard_config` string returned by `POST /vpn/connect` is a standard WireGuard `.conf` file. It works with any WireGuard client today. Native clients add:

- **Kill switch** — block all non-VPN traffic when tunnel drops
- **Auto-connect** — reconnect on network change
- **Server picker UI** — show latency per region, let user choose
- **Split tunneling** — route only selected traffic through VPN

Platform targets: Linux (CLI first), macOS, Windows, Android, iOS.

---

### Phase 11 — Billing and Subscriptions

The `subscriptions`, `plans`, and `payments` tables are already migrated. Wire in:

- Stripe or Flutterwave for payment processing
- Subscription gating on `POST /vpn/connect` — check active subscription before allocating a peer
- Admin endpoints for plan management
- Webhook handlers for payment events

---

### Phase 12 — Admin API

- Server registration — `POST /admin/servers` to add a new VPN node
- Server status management — set maintenance / offline
- User management — suspend, restore, view audit history
- Subscription overrides
- Separate admin role authorization from user role

---

## Security Model

- Passwords hashed with bcrypt (cost 10)
- JWTs signed with HS256, 15-minute access token lifetime, 7-day rotating refresh tokens stored server-side
- Go→Rust calls use a pre-shared bearer token validated with constant-time comparison
- `/health` and `/metrics` on the engine are unauthenticated — bind engine control port to a private interface in production
- All sensitive values in environment variables, never committed — `.env` is gitignored
- `audit_logs` table is append-only — no delete or update permissions for the API database user in production
- WireGuard handles all VPN cryptography (Noise Protocol Framework, ChaCha20-Poly1305, Curve25519)
- No custom cryptographic primitives anywhere in the codebase

See [docs/security.md](docs/security.md) and [docs/threat-model.md](docs/threat-model.md) for the full security analysis.

---

## Documentation

| Document                                                         | Contents                                                      |
| ---------------------------------------------------------------- | ------------------------------------------------------------- |
| [docs/architecture.md](docs/architecture.md)                     | Component design, data flow, technology decisions             |
| [docs/api.md](docs/api.md)                                       | REST API reference with request/response examples             |
| [docs/security.md](docs/security.md)                             | Security model, secret management, transport security         |
| [docs/networking.md](docs/networking.md)                         | WireGuard setup, routing, NAT, DNS leak prevention            |
| [docs/deployment.md](docs/deployment.md)                         | Docker, migrations, VPN node provisioning checklist           |
| [docs/threat-model.md](docs/threat-model.md)                     | 14 threats with impact, likelihood, mitigation, residual risk |
| [packages/protocol/openapi.yaml](packages/protocol/openapi.yaml) | Full OpenAPI 3.1 specification                                |

---

## License

Copyright (c) 2026 Solid VPN. All rights reserved.
All rights reserved.

PROPRIETARY SOFTWARE LICENSE

This software and all associated source code, documentation, designs, interfaces, algorithms, and other materials are proprietary and confidential property of Willstone Strategic Industries Limited.

Permission is not granted to copy, modify, distribute, publish, sublicense, sell, lease, reverse engineer, decompile, disassemble, or otherwise use, reproduce, or exploit this software or any portion of it without prior written authorization from Willstone Strategic Industries Limited.

Access to or possession of this source code does not grant any ownership, license, or other intellectual property rights.

Unauthorized use, reproduction, distribution, modification, or disclosure of this software is strictly prohibited.

For licensing, commercial use, partnership, or other authorized access, contact Willstone Strategic Industries Limited.

Copyright © 2026 Willstone Strategic Industries Limited. All rights reserved
