# Architecture

## Overview

Solid VPN is split into two strictly separated planes:

| Plane | Language | Responsibility |
|-------|----------|----------------|
| Control plane | Go | Business logic, API, auth, billing, configuration |
| Data plane | Rust | WireGuard management, networking, packet forwarding |

These planes communicate over an authenticated internal control channel. The Go API **never** forwards VPN packets. The Rust engine **never** handles business logic.

---

## Component Diagram

```
                         SOLID VPN
                              │
              ┌───────────────┴───────────────┐
              │                               │
         CONTROL PLANE                   DATA PLANE
              │                               │
              ▼                               ▼
       ┌────────────┐                  ┌─────────────┐
       │   Go API   │                  │ Rust Engine │
       │            │                  │             │
       │ Auth       │                  │ WireGuard   │
       │ Users      │                  │ Routing     │
       │ Devices    │◄── Control API ──│ NAT         │
       │ Servers    │                  │ Firewall    │
       │ Sessions   │                  │ DNS         │
       │ Billing    │                  │ Sessions    │
       └─────┬──────┘                  └──────┬──────┘
             │                                │
             │ HTTPS                          │ WireGuard UDP
             ▼                                ▼
          Clients                          Internet
```

---

## Go Control Plane (`apps/api`)

### Responsibilities

- User registration and authentication
- JWT issuance and validation
- Device registration
- VPN server management and health
- Server selection algorithm
- VPN peer lifecycle (create / update / remove)
- Session tracking
- Subscription and payment management
- DNS configuration
- Audit logging
- Administrative operations

### Non-responsibilities

The Go API must NOT:
- Forward VPN packets
- Implement WireGuard cryptography
- Act as a raw network proxy

### Internal structure

```
apps/api/
├── cmd/api/main.go         Entry point, wiring
├── config/                 Environment-based config
├── internal/
│   ├── auth/               Registration, login, tokens
│   ├── users/              Profile management
│   ├── devices/            Device registration
│   ├── vpn/                Server selection, peer lifecycle, config generation
│   ├── servers/            VPN server and region management
│   ├── sessions/           Session lifecycle
│   ├── subscriptions/      Plan management
│   ├── payments/           Payment gateway integration
│   ├── dns/                DNS configuration
│   ├── health/             Liveness and readiness probes
│   └── middleware/         Request logging, recovery, auth
├── migrations/             SQL migration files
└── routes/                 HTTP router wiring
```

### Request flow

```
HTTP Request
     │
     ▼
  Middleware (logger, auth, recovery)
     │
     ▼
  Handler (HTTP concerns only)
     │
     ▼
  Service (business logic)
     │
     ▼
  Repository (database queries)
     │
     ▼
  PostgreSQL
```

---

## Rust Data Plane (`apps/vpn-engine`)

### Responsibilities

- WireGuard interface lifecycle
- Peer configuration (add / remove)
- Network routing
- NAT masquerade
- Firewall rules (iptables / nftables)
- DNS handling
- Session state and expiry
- Health and metrics reporting
- Secure communication with the control plane

### Non-responsibilities

The Rust engine must NOT:
- Implement authentication or user management
- Store user data
- Make billing decisions
- Implement WireGuard cryptographic primitives

### Module structure

```
apps/vpn-engine/src/
├── main.rs             Entry point, startup, graceful shutdown
├── errors.rs           Unified error type
├── config/
│   └── settings.rs     Environment-based configuration
├── tunnel/
│   ├── wireguard.rs    WireGuard interface management
│   ├── interface.rs    Network interface lifecycle
│   └── peer.rs         Peer add / remove
├── network/
│   ├── routing.rs      IP routing
│   ├── firewall.rs     iptables / nftables rules
│   ├── nat.rs          NAT masquerade
│   └── dns.rs          DNS configuration
├── sessions/
│   ├── manager.rs      In-memory session store
│   └── tracker.rs      Session expiry
├── security/
│   ├── auth.rs         Token validation
│   └── keys.rs         Key format validation
└── telemetry/
    └── mod.rs          Prometheus metrics
```

---

## Control Plane ↔ Data Plane Communication

```
Go API
  │
  │  POST /peers          (create peer)
  │  DELETE /peers/:id    (remove peer)
  │  GET /health          (engine health)
  │  GET /status          (engine status)
  │
  ▼
Rust Engine
  │
  ├── validate bearer token
  ├── validate request payload
  ├── apply WireGuard configuration
  └── return result
```

- All requests carry a shared bearer token
- Token is validated on every request — source IP is not trusted alone
- TLS is required in staging and production
- In local development, plain HTTP over Docker network is acceptable

---

## Database

PostgreSQL 16. Schema managed via sequential SQL migrations in `apps/api/migrations/`.

### Core tables

| Table | Description |
|-------|-------------|
| `users` | User accounts |
| `devices` | Registered client devices |
| `vpn_servers` | Available VPN server nodes |
| `vpn_regions` | Geographic regions |
| `vpn_peers` | WireGuard peer records |
| `vpn_sessions` | Active and historical VPN sessions |
| `subscriptions` | User subscription plans |
| `plans` | Available subscription tiers |
| `payments` | Payment records |
| `ip_allocations` | IP address pool management |
| `audit_logs` | Immutable audit event log |

---

## VPN Connection Flow

```
1.  Client authenticates
2.  Client registers device
3.  API validates subscription
4.  API selects optimal VPN server
5.  API creates/updates WireGuard peer record
6.  API sends peer config to Rust engine
7.  Rust engine configures WireGuard interface
8.  API returns WireGuard config to client
9.  Client establishes WireGuard tunnel
10. Session becomes active
11. Metrics and health are tracked continuously
```

### Disconnect flow

```
Client → POST /api/v1/vpn/disconnect
  → API marks session ended
  → API calls Rust engine: remove peer
  → Engine removes peer from WireGuard
  → Engine removes firewall / routing rules
  → Session closed in database
  → Audit log entry written
```

---

## Infrastructure Topology (Target)

```
                    Control Plane
                      Go API
                         │
           ┌─────────────┼─────────────┐
           │             │             │
           ▼             ▼             ▼
       VPN Node       VPN Node       VPN Node
       Nigeria           UK            USA
           │             │             │
         Rust          Rust          Rust
         Engine        Engine        Engine
           │             │             │
       WireGuard     WireGuard     WireGuard
```

For Phase 1, the entire stack runs in Docker Compose on a single machine. Multi-region deployment is planned for Phase 9.

---

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| VPN protocol | WireGuard | Audited, modern, minimal attack surface |
| Control plane language | Go | Strong networking stdlib, simple concurrency, fast builds |
| Data plane language | Rust | Memory safety without GC, ideal for systems/networking code |
| Database | PostgreSQL | Mature, reliable, strong consistency |
| HTTP router (Go) | chi | Lightweight, idiomatic, composable middleware |
| Logging (Go) | zap | Structured JSON, high performance |
| Async runtime (Rust) | tokio | De-facto standard for async Rust |
| Metrics | Prometheus | Industry standard, well-supported |
| Local dev | Docker Compose | Simple, reproducible, no Kubernetes overhead for MVP |
