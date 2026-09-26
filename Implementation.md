Below is a master prompt you can give to an AI coding agent such as Cursor, Claude Code, or another coding assistant. It is written to enforce the architecture and prevent it from collapsing everything into one Go service.

````markdown
# Solid VPN — Master Development Prompt

## 1. Project Overview

Build a production-oriented VPN platform called **Solid VPN**.

The system must use:

- **Go** for the control plane/API.
- **Rust** for the VPN/data plane.
- **WireGuard** as the VPN protocol.
- **PostgreSQL** for persistent data.
- **React/TypeScript** for the web/admin interface when required.
- **Docker** for local development.
- **Terraform/Ansible** for infrastructure and server provisioning where appropriate.

The project must be designed for security, scalability, observability, maintainability, and clear separation between the control plane and data plane.

Do **not** implement a custom cryptographic protocol.

Do **not** invent custom encryption.

Use established, audited technologies and libraries wherever possible.

---

# 2. Core Architecture

The system must follow this architecture:

```text
                         SOLID VPN
                            │
             ┌──────────────┴──────────────┐
             │                             │
        CONTROL PLANE                 DATA PLANE
             │                             │
             ▼                             ▼
       ┌───────────┐                 ┌────────────┐
       │    Go     │                 │   Rust     │
       │    API    │                 │ VPN Engine │
       └─────┬─────┘                 └─────┬──────┘
             │                             │
             │ Secure Control API          │
             │                             │
             └──────────────┬──────────────┘
                            │
                            ▼
                      WireGuard Tunnel
                            │
                            ▼
                         Internet
````

The Go API controls the VPN infrastructure.

The Rust VPN engine handles VPN/networking operations.

The Go API must NOT become the packet forwarding layer.

The Rust VPN engine must NOT become the business/API layer.

Maintain this separation throughout the entire project.

---

# 3. Exact Project Structure

Use this exact top-level project structure:

```text
solid-vpn/
│
├── README.md
├── LICENSE
├── Makefile
├── docker-compose.yml
├── .env.example
├── .gitignore
│
├── apps/
│   │
│   ├── api/
│   │   ├── cmd/
│   │   │   └── api/
│   │   │       └── main.go
│   │   │
│   │   ├── internal/
│   │   │   ├── auth/
│   │   │   ├── users/
│   │   │   ├── devices/
│   │   │   ├── vpn/
│   │   │   ├── servers/
│   │   │   ├── sessions/
│   │   │   ├── subscriptions/
│   │   │   ├── payments/
│   │   │   ├── dns/
│   │   │   ├── health/
│   │   │   └── middleware/
│   │   │
│   │   ├── config/
│   │   ├── migrations/
│   │   ├── routes/
│   │   └── go.mod
│   │
│   ├── vpn-engine/
│   │   ├── src/
│   │   │   ├── main.rs
│   │   │   │
│   │   │   ├── config/
│   │   │   │   ├── mod.rs
│   │   │   │   └── settings.rs
│   │   │   │
│   │   │   ├── tunnel/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── wireguard.rs
│   │   │   │   ├── interface.rs
│   │   │   │   └── peer.rs
│   │   │   │
│   │   │   ├── network/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── routing.rs
│   │   │   │   ├── firewall.rs
│   │   │   │   ├── nat.rs
│   │   │   │   └── dns.rs
│   │   │   │
│   │   │   ├── sessions/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── manager.rs
│   │   │   │   └── tracker.rs
│   │   │   │
│   │   │   ├── security/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── keys.rs
│   │   │   │   └── auth.rs
│   │   │   │
│   │   │   ├── telemetry/
│   │   │   └── errors.rs
│   │   │
│   │   ├── tests/
│   │   ├── Cargo.toml
│   │   └── Cargo.lock
│   │
│   └── agent/
│       ├── cmd/
│       ├── internal/
│       └── go.mod
│
├── clients/
│   │
│   ├── desktop/
│   │   ├── windows/
│   │   ├── macos/
│   │   └── linux/
│   │
│   ├── mobile/
│   │   ├── android/
│   │   └── ios/
│   │
│   └── shared/
│       └── vpn-config/
│
├── packages/
│   │
│   ├── protocol/
│   │   ├── openapi.yaml
│   │   └── protobuf/
│   │
│   ├── config/
│   └── schemas/
│
├── infrastructure/
│   │
│   ├── docker/
│   │   ├── api.Dockerfile
│   │   └── vpn.Dockerfile
│   │
│   ├── terraform/
│   │   ├── providers.tf
│   │   ├── network.tf
│   │   ├── servers.tf
│   │   └── variables.tf
│   │
│   ├── ansible/
│   │   ├── playbooks/
│   │   └── roles/
│   │
│   ├── kubernetes/
│   │   ├── api/
│   │   ├── vpn/
│   │   └── monitoring/
│   │
│   └── scripts/
│       ├── setup-server.sh
│       ├── setup-wireguard.sh
│       ├── firewall.sh
│       └── deploy.sh
│
├── deployments/
│   │
│   ├── development/
│   ├── staging/
│   └── production/
│
├── monitoring/
│   ├── prometheus/
│   ├── grafana/
│   └── alerts/
│
├── docs/
│   ├── architecture.md
│   ├── api.md
│   ├── security.md
│   ├── networking.md
│   ├── deployment.md
│   └── threat-model.md
│
└── tests/
    ├── integration/
    ├── e2e/
    └── load/
```

Do not replace this architecture with a different project structure unless there is a strong technical reason.

If a structural change is necessary, document the reason first.

---

# 4. Development Philosophy

Follow these principles:

1. Security first.
2. Simplicity before premature optimization.
3. Clear separation of responsibilities.
4. Strong typing.
5. Explicit error handling.
6. Test critical functionality.
7. Avoid unnecessary dependencies.
8. Never implement cryptography yourself.
9. Never store private keys unnecessarily.
10. Never log secrets.
11. Never expose internal infrastructure credentials.
12. Follow least-privilege principles.
13. Make production configuration explicit.
14. Make failures observable.
15. Prefer deterministic behavior.
16. Write maintainable code rather than clever code.

---

# 5. Go Control Plane

The Go API is responsible for:

* Authentication
* User management
* Device management
* VPN server management
* VPN server selection
* VPN configuration generation
* Peer lifecycle management
* Session lifecycle
* Subscription management
* Payment integration
* DNS configuration management
* Server health
* Administrative operations
* Audit logging

The Go API is NOT responsible for:

* Forwarding VPN packets
* Implementing cryptography
* Reimplementing WireGuard
* Acting as a raw packet proxy
* Handling the VPN data path

---

# 6. Go API Architecture

Use a clean architecture approach.

Prefer:

```text
Handler
   ↓
Service
   ↓
Repository
   ↓
Database
```

Example:

```text
internal/vpn/

handler.go
service.go
repository.go
models.go
dto.go
errors.go
```

Handlers should contain HTTP concerns.

Services should contain business logic.

Repositories should contain persistence logic.

Do not place business logic directly inside HTTP handlers.

---

# 7. API Endpoints

Implement versioned endpoints.

Base path:

```text
/api/v1
```

Authentication:

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

Users:

```http
GET    /api/v1/users/me
PATCH  /api/v1/users/me
DELETE /api/v1/users/me
```

Devices:

```http
GET    /api/v1/devices
POST   /api/v1/devices
GET    /api/v1/devices/:id
DELETE /api/v1/devices/:id
```

VPN:

```http
GET  /api/v1/vpn/servers
GET  /api/v1/vpn/servers/:id
POST /api/v1/vpn/connect
POST /api/v1/vpn/disconnect
GET  /api/v1/vpn/config
```

Sessions:

```http
GET /api/v1/vpn/sessions
GET /api/v1/vpn/sessions/:id
```

Health:

```http
GET /health
GET /ready
```

Administrative APIs should use a separate authorization layer.

---

# 8. PostgreSQL Data Model

Use PostgreSQL.

Initial tables:

```text
users
devices
vpn_servers
vpn_regions
vpn_peers
vpn_sessions
subscriptions
plans
payments
ip_allocations
audit_logs
```

Suggested relationships:

```text
users
  │
  ├── devices
  │      │
  │      └── vpn_sessions
  │
  └── subscriptions


vpn_servers
  │
  ├── vpn_peers
  │
  ├── vpn_sessions
  │
  └── ip_allocations
```

Use UUIDs for externally exposed identifiers.

Add:

* created_at
* updated_at

where appropriate.

Use database constraints.

Use foreign keys.

Use indexes for frequently queried fields.

Never rely solely on application-level validation for data integrity.

---

# 9. VPN Server Model

A VPN server should contain information such as:

```text
id
name
hostname
public_ip
region_id
country
city
provider
status
capacity
active_connections
wireguard_port
public_key
created_at
updated_at
```

Do not store private WireGuard keys in plaintext in the normal database.

If private keys must be stored, use an appropriate secret-management system.

---

# 10. VPN Peer Model

A peer should represent a user's VPN identity on a server.

Include concepts such as:

```text
id
user_id
device_id
server_id
public_key
assigned_ip
status
created_at
updated_at
```

Private keys should remain on the client whenever possible.

---

# 11. Rust VPN Engine

The Rust VPN engine is the data plane.

It is responsible for:

* WireGuard interface management
* Peer management
* Network routing
* NAT
* Firewall integration
* DNS handling
* Session tracking
* Health reporting
* Metrics
* Secure communication with the control plane

It should be designed primarily for Linux VPN servers initially.

Support for other server operating systems can be added later.

---

# 12. WireGuard

Use WireGuard.

Do not invent:

```text
Custom encryption
Custom handshake
Custom VPN protocol
Custom key exchange
Custom packet encryption
```

Do not write cryptographic primitives unless absolutely necessary and reviewed by security experts.

Use well-established WireGuard implementations/libraries or the system WireGuard interface.

The Rust engine should manage WireGuard rather than replacing WireGuard's cryptographic design.

---

# 13. Rust Module Responsibilities

## config/

Responsible for:

* Configuration loading
* Environment variables
* Validation
* Runtime settings

Example:

```rust
pub struct Settings {
    pub control_plane_url: String,
    pub server_id: String,
    pub wireguard_interface: String,
}
```

---

## tunnel/

Responsible for:

* WireGuard interface
* Peer configuration
* Tunnel lifecycle
* Interface state

Files:

```text
wireguard.rs
interface.rs
peer.rs
```

---

## network/

Responsible for:

```text
routing
firewall
NAT
DNS
```

Never blindly execute shell commands from user input.

Validate all configuration before applying it.

---

## sessions/

Responsible for:

* Active sessions
* Connection tracking
* Session expiration
* Server connection counts
* Session telemetry

---

## security/

Responsible for:

* Authentication with the control plane
* Key handling
* Credential validation
* Secure communication

Never log:

```text
private keys
tokens
passwords
session secrets
API credentials
```

---

## telemetry/

Expose:

* Health
* Metrics
* Connection count
* Errors
* Resource usage

Use Prometheus-compatible metrics where appropriate.

---

# 14. Control Plane ↔ VPN Engine Communication

The Go control plane and Rust VPN engine must communicate through a secure authenticated channel.

Conceptually:

```text
Go API
   │
   │ authenticated control request
   ▼
Rust VPN Engine
   │
   ├── configure peer
   ├── remove peer
   ├── health
   └── status
```

Never expose administrative engine endpoints publicly without authentication.

Prefer:

* TLS
* short-lived credentials
* service identity
* authorization
* request validation

Do not trust a request simply because it originated from a known IP.

---

# 15. VPN Connection Flow

Implement this logical flow:

```text
1. User authenticates
        ↓
2. User registers device
        ↓
3. API validates subscription/access
        ↓
4. API selects suitable VPN server
        ↓
5. API creates/updates VPN peer
        ↓
6. API communicates with VPN engine
        ↓
7. VPN engine configures WireGuard peer
        ↓
8. Client receives VPN configuration
        ↓
9. Client establishes WireGuard tunnel
        ↓
10. Session becomes active
        ↓
11. Metrics/health/session state are tracked
```

Disconnect:

```text
Client
  ↓
Disconnect
  ↓
API
  ↓
Rust VPN Engine
  ↓
Remove/disable peer
  ↓
Session closed
```

---

# 16. Server Selection

Create a server-selection component.

Consider:

* server status
* region
* country
* active connection count
* configured capacity
* latency when available
* health
* maintenance status

Do not select unhealthy servers.

Example:

```text
vpn/
└── selector.go
```

The selection algorithm must be deterministic and testable.

---

# 17. Security Requirements

Implement:

### Authentication

Use secure password hashing.

Never store plaintext passwords.

Use secure session/token mechanisms.

### Authorization

Implement roles such as:

```text
USER
ADMIN
SYSTEM
```

Do not rely on frontend authorization.

All privileged operations must be authorized server-side.

### Secrets

Never commit:

```text
.env
private keys
JWT secrets
database passwords
API keys
cloud credentials
```

Provide:

```text
.env.example
```

instead.

---

# 18. Logging

Use structured logging.

Example:

```json
{
  "level": "info",
  "service": "vpn-api",
  "event": "vpn_session_created",
  "session_id": "..."
}
```

Never log:

```text
passwords
private keys
access tokens
refresh tokens
VPN credentials
sensitive user data
```

---

# 19. Audit Logging

Create immutable audit events for important operations.

Examples:

```text
USER_CREATED
DEVICE_REGISTERED
VPN_SERVER_CREATED
VPN_PEER_CREATED
VPN_PEER_REMOVED
VPN_SESSION_STARTED
VPN_SESSION_ENDED
SUBSCRIPTION_CREATED
ADMIN_ACTION
```

Audit logs should contain:

```text
id
actor_id
action
resource_type
resource_id
metadata
created_at
```

Do not allow ordinary users to modify audit history.

---

# 20. Infrastructure

Support Linux VPN nodes.

Infrastructure should eventually support:

```text
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
            │             │             │
        WireGuard     WireGuard     WireGuard
```

Terraform should manage infrastructure.

Ansible can configure VPN nodes.

Docker should be used for local development.

Kubernetes should not be required for the initial MVP.

---

# 21. Docker Development

Provide a local development environment.

At minimum:

```text
Go API
PostgreSQL
VPN engine
```

The Docker setup should support hot reload where practical.

Provide:

```bash
make dev
make build
make test
make lint
make format
make migrate
make docker-up
make docker-down
```

---

# 22. Testing

Testing is mandatory.

Go:

```text
unit tests
service tests
repository tests
HTTP handler tests
integration tests
```

Rust:

```text
unit tests
network tests
configuration tests
WireGuard management tests
security tests
```

End-to-end:

```text
register user
    ↓
register device
    ↓
select VPN server
    ↓
create peer
    ↓
configure WireGuard
    ↓
connect
    ↓
verify session
    ↓
disconnect
```

Do not require a real public VPN server for every automated test.

Use mocks/test doubles where appropriate.

---

# 23. Observability

Implement:

### Metrics

Track:

```text
vpn_active_sessions
vpn_sessions_total
vpn_connection_errors
vpn_server_health
vpn_peer_count
vpn_api_requests
vpn_api_latency
```

### Health checks

Go:

```text
/health
/ready
```

Rust:

```text
health endpoint or local health mechanism
```

Monitor:

* CPU
* memory
* network traffic
* connection count
* errors
* server availability

---

# 24. Documentation

Create:

```text
docs/architecture.md
docs/api.md
docs/security.md
docs/networking.md
docs/deployment.md
docs/threat-model.md
```

Document:

* architecture
* data flow
* authentication
* WireGuard integration
* server provisioning
* networking
* firewall requirements
* deployment
* security assumptions
* known limitations

---

# 25. Threat Model

Create a threat model covering at least:

```text
Compromised client
Compromised VPN server
Compromised API
Stolen authentication token
Stolen device
Malicious administrator
Database compromise
Network interception
Credential leakage
Misconfigured firewall
DNS leakage
IP leakage
Replay attempts
Unauthorized peer creation
```

For each threat, document:

```text
Threat
Impact
Likelihood
Mitigation
Residual risk
```

Do not claim the system is "secure" without explaining its security assumptions.

---

# 26. DNS and Leak Prevention

The system should be designed to prevent:

* DNS leaks
* accidental traffic outside the tunnel
* stale routes
* unauthorized access to the VPN server

Implement appropriate:

```text
DNS handling
routing
firewall rules
kill-switch support
```

The client implementation should eventually support a kill switch.

---

# 27. Client Architecture

The client must eventually support:

```text
Windows
macOS
Linux
Android
iOS
```

However, do NOT implement every client initially.

Start with one client platform and prove the end-to-end architecture.

The client should be responsible for:

```text
authentication
device registration
VPN configuration
WireGuard tunnel control
connection state
disconnect
kill switch
server selection
```

Private VPN keys should preferably be generated and stored on the client.

---

# 28. Configuration

Use environment-based configuration.

Example:

```text
DATABASE_URL=
JWT_SECRET=
CONTROL_PLANE_URL=
VPN_ENGINE_TOKEN=
WIREGUARD_INTERFACE=
WIREGUARD_PORT=
LOG_LEVEL=
```

Provide:

```text
.env.example
```

Never provide real production secrets.

---

# 29. API Specification

Maintain:

```text
packages/protocol/openapi.yaml
```

The OpenAPI specification should document:

* authentication
* users
* devices
* VPN servers
* VPN configuration
* sessions
* subscriptions
* errors

Keep implementation and API documentation synchronized.

---

# 30. Error Handling

Use consistent API errors.

Example:

```json
{
  "error": {
    "code": "VPN_SERVER_UNAVAILABLE",
    "message": "No healthy VPN server is currently available."
  }
}
```

Do not expose internal stack traces to clients.

Use internal error IDs for debugging where appropriate.

---

# 31. Database Migrations

Every schema change must use a migration.

Do not manually modify production databases.

Example:

```text
apps/api/migrations/

000001_create_users.sql
000002_create_devices.sql
000003_create_vpn_servers.sql
000004_create_vpn_peers.sql
000005_create_sessions.sql
```

---

# 32. Git Practices

Use clear commits:

```text
feat:
fix:
refactor:
test:
docs:
chore:
security:
```

Example:

```text
feat(vpn): add WireGuard peer management
feat(api): add device registration
security(auth): harden refresh token handling
test(vpn): add peer lifecycle tests
```

Do not commit secrets.

---

# 33. Development Phases

Build the project incrementally.

## Phase 1 — Foundation

Create:

```text
project structure
Go module
Rust Cargo project
Docker Compose
PostgreSQL
environment configuration
Makefile
README
```

Verify everything builds.

---

## Phase 2 — Go API

Implement:

```text
health
database connection
migrations
user registration
login
authentication
device registration
```

---

## Phase 3 — VPN Server Management

Implement:

```text
vpn_servers
vpn_regions
server health
server registration
server selection
```

---

## Phase 4 — Rust VPN Engine

Implement:

```text
configuration
WireGuard interface management
peer management
network routing
NAT
firewall integration
health
metrics
```

Start with Linux.

---

## Phase 5 — Go ↔ Rust Communication

Implement:

```text
authenticated control channel
peer creation
peer removal
server health
status reporting
```

---

## Phase 6 — End-to-End VPN

Implement:

```text
User
 ↓
Device
 ↓
VPN Server Selection
 ↓
Peer Creation
 ↓
WireGuard Configuration
 ↓
Tunnel
 ↓
Internet
```

Verify:

* tunnel connects
* traffic passes
* DNS works
* routing works
* disconnect works
* peer removal works

---

## Phase 7 — Security Hardening

Perform:

```text
threat modeling
dependency audit
secret review
authentication review
authorization review
network security review
logging review
firewall review
```

---

## Phase 8 — Monitoring

Add:

```text
Prometheus
Grafana
structured logs
health checks
alerts
```

---

## Phase 9 — Infrastructure

Add:

```text
Terraform
Ansible
server provisioning
deployment scripts
staging environment
production environment
```

---

## Phase 10 — Client

Start with one desktop client.

Then expand to:

```text
Windows
macOS
Linux
Android
iOS
```

---

# 34. MVP Definition

Do NOT attempt to build every feature immediately.

The first functional MVP must be able to do:

```text
1. Create user
2. Authenticate user
3. Register device
4. Register VPN server
5. Select VPN server
6. Generate WireGuard identity/configuration
7. Create VPN peer
8. Connect client
9. Establish WireGuard tunnel
10. Route traffic
11. Track VPN session
12. Disconnect
13. Remove/disable peer
14. Record audit events
15. Report server health
```

---

# 35. What NOT To Do

Never:

* invent custom encryption
* invent a custom VPN protocol
* store private keys unnecessarily
* hardcode credentials
* commit secrets
* expose admin endpoints publicly
* trust client-provided authorization
* put packet forwarding in Go API
* put business logic in Rust VPN engine
* log VPN private keys
* log authentication tokens
* blindly execute shell commands
* accept arbitrary firewall commands from API requests
* use user input directly in shell commands
* skip input validation
* skip authorization
* claim security without testing
* add Kubernetes before it is needed
* over-engineer the MVP

---

# 36. Code Quality Requirements

Before considering a feature complete:

```text
[ ] Code compiles
[ ] Tests pass
[ ] Error handling exists
[ ] Input validation exists
[ ] Authorization exists where required
[ ] No secrets are committed
[ ] Logs contain no sensitive credentials
[ ] Database migrations exist
[ ] API documentation is updated
[ ] Relevant tests exist
[ ] README/documentation is updated
```

---

# 37. Implementation Rules for the AI Coding Agent

When implementing this project:

1. Inspect the existing project before modifying it.
2. Do not overwrite working code unnecessarily.
3. Follow the existing architecture.
4. Implement one logical feature at a time.
5. After each major change, run relevant tests.
6. Fix compilation errors immediately.
7. Do not create placeholder implementations for security-critical components.
8. Do not silently change architectural decisions.
9. Explain major architectural changes before applying them.
10. Prefer small, reviewable changes.
11. Keep functions focused.
12. Avoid massive files.
13. Avoid global mutable state.
14. Use dependency injection where appropriate.
15. Keep infrastructure configuration separate from application logic.

---

# 38. First Task

Start by creating the complete project foundation.

Do NOT implement the complete VPN yet.

First create:

```text
solid-vpn/
├── apps/api
├── apps/vpn-engine
├── apps/agent
├── clients
├── packages
├── infrastructure
├── deployments
├── monitoring
├── docs
└── tests
```

Then:

### Go

Initialize:

```text
apps/api/go.mod
```

Create the API entry point:

```text
apps/api/cmd/api/main.go
```

Implement a basic:

```text
GET /health
```

endpoint.

### Rust

Initialize:

```text
apps/vpn-engine/Cargo.toml
```

Create:

```text
apps/vpn-engine/src/main.rs
```

Implement a minimal startup process with structured logging and configuration loading.

Do not implement fake VPN functionality.

### Database

Configure PostgreSQL through Docker Compose.

### Docker

Create:

```text
infrastructure/docker/api.Dockerfile
infrastructure/docker/vpn.Dockerfile
docker-compose.yml
```

### Makefile

Provide:

```text
make dev
make build
make test
make lint
make format
make docker-up
make docker-down
```

### Documentation

Create:

```text
README.md
docs/architecture.md
docs/security.md
docs/networking.md
docs/deployment.md
docs/threat-model.md
```

Explain the initial architecture.

---

# 39. First Milestone

The first milestone is complete only when:

```bash
docker compose up
```

starts the development environment successfully.

The Go API must respond to:

```text
GET /health
```

The Rust VPN engine must start successfully.

PostgreSQL must be reachable.

The project must compile successfully.

No secrets should be hardcoded.

No fake VPN tunnel should be presented as functional.

After completing the foundation, report:

1. Files created
2. Architecture implemented
3. Commands used to run it
4. Tests performed
5. Any unresolved issues
6. Recommended next implementation step

Then stop and wait for the next instruction.

---

# 40. Final Architectural Principle

Keep this distinction throughout the entire project:

```text
                 SOLID VPN

              ┌──────────────┐
              │   Clients    │
              └──────┬───────┘
                     │
                     │ HTTPS
                     ▼
              ┌──────────────┐
              │  Go Control  │
              │    Plane     │
              ├──────────────┤
              │ Auth         │
              │ Users        │
              │ Devices      │
              │ Billing      │
              │ Server Mgmt  │
              │ Sessions     │
              └──────┬───────┘
                     │
              Secure Control
                     │
                     ▼
              ┌──────────────┐
              │ Rust Data    │
              │    Plane     │
              ├──────────────┤
              │ WireGuard    │
              │ Routing      │
              │ NAT          │
              │ Firewall     │
              │ DNS          │
              │ Networking   │
              └──────┬───────┘
                     │
                     ▼
                  Internet
```

**Go controls the VPN infrastructure.**

**Rust moves and manages VPN traffic.**

**WireGuard provides the VPN protocol and cryptographic tunnel.**

Keep these responsibilities separated as the system grows.

```
```
