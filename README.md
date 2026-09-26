# Solid VPN

**Solid VPN** is a production-oriented VPN platform designed around a strict separation between the **control plane** and the **VPN data plane**.

The platform is designed to provide secure user authentication, device management, VPN server management, server selection, WireGuard configuration, peer lifecycle management, subscription and payment integration, session tracking, DNS/network management, audit logging, monitoring, and infrastructure automation.

The architecture uses:

- **Go** for the control plane and business/API layer
- **Rust** for the VPN data plane and networking engine
- **WireGuard** for the VPN protocol and encrypted tunnel
- **PostgreSQL** for persistent application data
- **React/TypeScript** for web/admin interfaces where required
- **Docker** for local development
- **Terraform** for infrastructure provisioning
- **Ansible** for server configuration and provisioning
- **Prometheus/Grafana** for observability

> **Project status note:** This README is derived from the supplied Solid VPN master development specification. It documents the software architecture, scope, responsibilities, and intended capabilities described by that specification. It does not by itself certify that every listed component has already been implemented or deployed.

---

## Table of Contents

- [Overview](#overview)
- [What Solid VPN Covers](#what-solid-vpn-covers)
- [Core Architecture](#core-architecture)
- [Technology Stack](#technology-stack)
- [Control Plane](#control-plane)
- [VPN Data Plane](#vpn-data-plane)
- [WireGuard](#wireguard)
- [End-to-End VPN Flow](#end-to-end-vpn-flow)
- [Server Selection](#server-selection)
- [Users, Devices and Peers](#users-devices-and-peers)
- [Subscriptions and Payments](#subscriptions-and-payments)
- [DNS, Routing, NAT and Firewall](#dns-routing-nat-and-firewall)
- [Security Model](#security-model)
- [Authentication and Authorization](#authentication-and-authorization)
- [Secrets and Key Management](#secrets-and-key-management)
- [Audit Logging](#audit-logging)
- [Observability](#observability)
- [Database](#database)
- [API](#api)
- [Project Structure](#project-structure)
- [Infrastructure](#infrastructure)
- [Client Architecture](#client-architecture)
- [Development Environment](#development-environment)
- [Testing](#testing)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [Threat Model](#threat-model)
- [Development Phases](#development-phases)
- [MVP Scope](#mvp-scope)
- [Engineering Rules](#engineering-rules)
- [Known Boundaries](#known-boundaries)
- [Getting Started](#getting-started)
- [Future Expansion](#future-expansion)
- [Conclusion](#conclusion)

---

# Overview

Solid VPN is designed as a scalable VPN service rather than a single VPN application.

Its architecture separates the system into two major responsibilities:

```text
                         SOLID VPN
                              │
               ┌──────────────┴──────────────┐
               │                             │
          CONTROL PLANE                 DATA PLANE
               │                             │
               ▼                             ▼
          Go API Service               Rust VPN Engine
               │                             │
               │ Secure Control API          │
               └──────────────┬──────────────┘
                              │
                              ▼
                         WireGuard
                              │
                              ▼
                           Internet
```

The **Go control plane** manages users, devices, access, servers, sessions, configuration, billing, and administrative operations.

The **Rust data plane** manages the networking operations required to establish and maintain VPN connectivity, including WireGuard interfaces, peers, routing, NAT, firewall integration, DNS handling, and telemetry.

**WireGuard** remains responsible for the VPN protocol and cryptographic tunnel.

This separation is one of the most important architectural decisions in the project.

---

# What Solid VPN Covers

The platform is designed to cover the complete lifecycle of a managed VPN service.

## User Management

The control plane is responsible for:

- User registration
- User authentication
- User profile management
- Session/token management
- Account deletion
- Role management
- Access control

## Device Management

Users can be associated with registered devices.

Device management covers:

- Device registration
- Device identification
- Device lifecycle
- Device-to-user relationships
- Device-to-VPN session relationships
- Device removal

## VPN Server Management

The platform is designed to manage a fleet of VPN servers.

A VPN server can contain information such as:

- Server ID
- Name
- Hostname
- Public IP
- Region
- Country
- City
- Infrastructure provider
- Status
- Capacity
- Active connection count
- WireGuard port
- WireGuard public key

The control plane can use this information to determine which server is appropriate for a connection.

## VPN Server Selection

Server selection considers operational information such as:

- Server health
- Server status
- Region
- Country
- Active connections
- Configured capacity
- Latency when available
- Maintenance state

Unhealthy servers should not be selected.

The selection component is intended to be deterministic and testable.

## VPN Peer Management

A peer represents a user's VPN identity on a particular VPN server.

Peer management covers:

- Peer creation
- Peer configuration
- Assigned VPN IP
- Public key
- Device association
- Server association
- Peer status
- Peer removal or disabling

Private keys should remain on the client whenever possible.

## VPN Sessions

The system tracks VPN sessions so the control plane can understand the current state of user connectivity.

Session capabilities include:

- Session creation
- Active session tracking
- Session expiration
- Connection counts
- Session closure
- Server-level connection metrics
- Session telemetry

## Subscriptions and Payments

The control plane includes a dedicated area for:

- Subscription management
- Plans
- Payment records
- Subscription validation before VPN access
- Billing-related business logic

The master specification defines the architecture for these capabilities but does not prescribe a particular payment provider.

## DNS and Network Management

The VPN engine is responsible for network-level operations including:

- DNS handling
- Routing
- NAT
- Firewall integration
- Tunnel traffic management
- Leak prevention mechanisms

The system is intended to prevent:

- DNS leaks
- Accidental traffic outside the tunnel
- Stale routes
- Unauthorized access to VPN infrastructure

A client-side kill switch is part of the planned client architecture.

## Administration

Administrative capabilities include:

- VPN server management
- Server health
- User/device administration
- Operational management
- Audit review
- Infrastructure-related control operations

Administrative APIs must have separate authorization controls from ordinary user operations.

---

# Core Architecture

Solid VPN is intentionally divided into independent services.

```text
                         Clients
                            │
                            │ HTTPS
                            ▼
                  ┌──────────────────┐
                  │   Go Control     │
                  │      Plane       │
                  ├──────────────────┤
                  │ Authentication   │
                  │ Users            │
                  │ Devices          │
                  │ Billing          │
                  │ VPN Management   │
                  │ Sessions         │
                  │ Server Selection │
                  │ Audit Logs       │
                  └────────┬─────────┘
                           │
                    Secure Control
                           │
                           ▼
                  ┌──────────────────┐
                  │  Rust Data Plane │
                  ├──────────────────┤
                  │ WireGuard        │
                  │ Routing          │
                  │ NAT              │
                  │ Firewall         │
                  │ DNS              │
                  │ Networking       │
                  │ Telemetry        │
                  └────────┬─────────┘
                           │
                           ▼
                        Internet
```

## Architectural Responsibilities

### Go Control Plane

The Go service owns:

- Authentication
- Authorization
- Users
- Devices
- VPN server inventory
- Server selection
- VPN configuration generation
- Peer lifecycle orchestration
- Sessions
- Subscriptions
- Payments
- DNS configuration management
- Server health information
- Administration
- Audit logging

### Rust Data Plane

The Rust service owns:

- WireGuard interface management
- Peer management at the networking layer
- Routing
- NAT
- Firewall integration
- DNS handling
- Session/network telemetry
- VPN server health reporting
- Data-plane networking operations

### WireGuard

WireGuard owns the actual VPN protocol and cryptographic tunnel.

The project does **not** replace WireGuard with a custom protocol.

---

# Technology Stack

| Area | Technology |
|---|---|
| Control plane | Go |
| VPN/data plane | Rust |
| VPN protocol | WireGuard |
| Database | PostgreSQL |
| Web/admin UI | React + TypeScript |
| Local development | Docker / Docker Compose |
| Infrastructure | Terraform |
| Server provisioning | Ansible |
| API specification | OpenAPI |
| Inter-service protocol | Secure authenticated control API |
| Metrics | Prometheus-compatible |
| Dashboards | Grafana |
| Operating system for VPN nodes | Linux initially |

The specification deliberately avoids unnecessary technology choices where they are not required.

---

# Control Plane

The Go API is the business and orchestration layer of Solid VPN.

## Clean Architecture

The API follows a layered approach:

```text
HTTP Handler
     │
     ▼
Service
     │
     ▼
Repository
     │
     ▼
PostgreSQL
```

### Handlers

Handlers deal with HTTP concerns:

- Request parsing
- Validation at the API boundary
- Authentication context
- HTTP responses
- HTTP error mapping

### Services

Services contain business logic such as:

- Selecting a VPN server
- Creating a peer
- Validating subscription access
- Managing device lifecycle
- Managing sessions

### Repositories

Repositories contain persistence logic and database access.

Business logic should not be placed directly in HTTP handlers.

---

# API

The API is versioned under:

```text
/api/v1
```

## Authentication

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

## Current User

```http
GET    /api/v1/users/me
PATCH  /api/v1/users/me
DELETE /api/v1/users/me
```

## Devices

```http
GET    /api/v1/devices
POST   /api/v1/devices
GET    /api/v1/devices/:id
DELETE /api/v1/devices/:id
```

## VPN

```http
GET  /api/v1/vpn/servers
GET  /api/v1/vpn/servers/:id
POST /api/v1/vpn/connect
POST /api/v1/vpn/disconnect
GET  /api/v1/vpn/config
```

## Sessions

```http
GET /api/v1/vpn/sessions
GET /api/v1/vpn/sessions/:id
```

## Health

```http
GET /health
GET /ready
```

Administrative endpoints require separate authorization.

---

# VPN Data Plane

The Rust VPN engine is designed primarily for Linux VPN servers.

Its responsibility is to perform the actual networking operations rather than business/API operations.

## Main Modules

```text
config/
tunnel/
network/
sessions/
security/
telemetry/
```

### Configuration

Responsible for:

- Environment configuration
- Runtime settings
- Configuration validation
- Control-plane connection settings
- WireGuard interface settings

### Tunnel

Responsible for:

- WireGuard interface management
- Peer configuration
- Tunnel lifecycle
- Interface state

### Network

Responsible for:

- Routing
- NAT
- Firewall integration
- DNS

### Sessions

Responsible for:

- Active sessions
- Connection tracking
- Session expiration
- Connection counts
- Session telemetry

### Security

Responsible for:

- Control-plane authentication
- Credential validation
- Key handling
- Secure communication

### Telemetry

Responsible for:

- Health
- Metrics
- Connection counts
- Errors
- Resource usage

---

# WireGuard

Solid VPN uses WireGuard as its VPN protocol.

The project explicitly does **not** implement:

- Custom encryption
- Custom handshakes
- Custom VPN protocols
- Custom key exchange
- Custom packet encryption

The Rust engine manages WireGuard rather than replacing its cryptographic design.

Where possible, the system should use established WireGuard implementations, libraries, or the system WireGuard interface.

---

# Control Plane ↔ VPN Engine Communication

The Go API and Rust engine communicate over a secure, authenticated control channel.

Conceptually:

```text
Go API
  │
  │ Authenticated control request
  ▼
Rust VPN Engine
  │
  ├── Configure peer
  ├── Remove peer
  ├── Report health
  └── Report status
```

The engine's administrative interface must never be publicly exposed without authentication.

The architecture favors:

- TLS
- Short-lived credentials
- Service identity
- Authorization
- Request validation

A request must not be trusted merely because it originated from a known IP address.

---

# End-to-End VPN Flow

A typical connection follows this sequence:

```text
1. User authenticates
        ↓
2. User registers device
        ↓
3. API validates subscription/access
        ↓
4. API selects a suitable VPN server
        ↓
5. API creates or updates VPN peer
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
11. Health and session state are tracked
```

## Disconnect Flow

```text
Client
  ↓
Disconnect
  ↓
Go API
  ↓
Rust VPN Engine
  ↓
Remove/disable peer
  ↓
Session closed
```

---

# Server Selection

Solid VPN is designed to support multiple VPN nodes across regions.

Example infrastructure:

```text
                    Control Plane
                         │
          ┌──────────────┼──────────────┐
          │              │              │
          ▼              ▼              ▼
      VPN Node        VPN Node       VPN Node
       Nigeria           UK             USA
          │              │              │
        Rust           Rust           Rust
          │              │              │
      WireGuard      WireGuard      WireGuard
```

Server selection can consider:

- Health
- Status
- Region
- Country
- Capacity
- Active connections
- Latency
- Maintenance state

The selection logic should be deterministic and independently testable.

---

# Users, Devices and Peers

The data model separates the user identity from the physical/device identity and VPN identity.

```text
User
 │
 ├── Devices
 │      │
 │      └── VPN Sessions
 │
 └── Subscription


VPN Server
 │
 ├── VPN Peers
 │
 ├── VPN Sessions
 │
 └── IP Allocations
```

This allows one user to have multiple devices while maintaining independent VPN identities and sessions.

---

# Subscriptions and Payments

Subscription state is part of the control plane.

Before establishing a VPN connection, the control plane can validate whether the user has the required access.

The data model includes:

```text
subscriptions
plans
payments
```

The supplied specification defines the platform architecture for these areas but does not define a specific external payment gateway.

---

# DNS, Routing, NAT and Firewall

The data plane manages network behavior required for VPN traffic.

## Routing

The engine is responsible for applying validated routes required by the VPN configuration.

## NAT

NAT allows VPN client traffic to reach external networks through the VPN server.

## Firewall

Firewall rules should be explicitly controlled and validated.

The system must never accept arbitrary firewall commands directly from untrusted user input.

## DNS

DNS handling is part of the VPN networking layer and is designed to reduce the risk of DNS leakage.

## Kill Switch

The client architecture is intended to eventually support a kill switch so traffic cannot accidentally bypass the VPN tunnel when the connection is expected to be active.

---

# Security Model

Security is a core design requirement rather than an optional layer.

The system follows principles including:

1. Security first
2. Least privilege
3. Explicit authorization
4. Strong typing
5. Explicit error handling
6. No custom cryptography
7. Minimal secret exposure
8. No secret logging
9. Deterministic behavior
10. Observable failures
11. Secure production configuration
12. Test critical security functionality

---

# Authentication and Authorization

## Authentication

The platform requires secure authentication mechanisms.

Passwords must never be stored in plaintext.

Session and token mechanisms must be designed to prevent unauthorized access.

## Authorization

The architecture includes roles such as:

```text
USER
ADMIN
SYSTEM
```

Authorization must always be enforced server-side.

The frontend must never be treated as the authority for privileged operations.

---

# Secrets and Key Management

Secrets must never be committed to source control.

The following must not be stored in Git:

```text
.env
private keys
JWT secrets
database passwords
API keys
cloud credentials
```

Instead, the repository provides an environment template:

```text
.env.example
```

WireGuard private keys should remain on the client whenever possible.

VPN server private keys should not be stored as plaintext in the normal application database. If they must be stored centrally, an appropriate secret-management system should be used.

---

# Logging

Solid VPN uses structured logging.

Example:

```json
{
  "level": "info",
  "service": "vpn-api",
  "event": "vpn_session_created",
  "session_id": "..."
}
```

Sensitive credentials must never appear in logs.

Do not log:

- Passwords
- Private keys
- Access tokens
- Refresh tokens
- VPN credentials
- API credentials
- Sensitive user information

---

# Audit Logging

Important administrative and lifecycle events should produce immutable audit records.

Examples include:

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

Audit records contain concepts such as:

```text
id
actor_id
action
resource_type
resource_id
metadata
created_at
```

Ordinary users must not be able to modify audit history.

---

# Observability

The platform is designed to be observable in production.

## Metrics

Important metrics include:

```text
vpn_active_sessions
vpn_sessions_total
vpn_connection_errors
vpn_server_health
vpn_peer_count
vpn_api_requests
vpn_api_latency
```

## Health Checks

Go provides:

```text
/health
/ready
```

The Rust engine also exposes a health mechanism appropriate to its deployment.

## Infrastructure Monitoring

The system should monitor:

- CPU
- Memory
- Network traffic
- Connection count
- Errors
- Server availability
- VPN server health

Prometheus-compatible metrics and Grafana dashboards are part of the observability architecture.

---

# Database

Solid VPN uses PostgreSQL for persistent application data.

## Core Tables

The initial model includes:

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

## Database Principles

The database should use:

- UUIDs for externally exposed identifiers
- Foreign keys
- Database constraints
- Appropriate indexes
- Created/updated timestamps where applicable
- Migration-based schema changes

Application-level validation should not be the only mechanism protecting data integrity.

## Migrations

Schema changes must be represented as migrations.

Example:

```text
apps/api/migrations/

000001_create_users.sql
000002_create_devices.sql
000003_create_vpn_servers.sql
000004_create_vpn_peers.sql
000005_create_sessions.sql
```

Production databases should not be modified manually.

---

# API Error Handling

API errors use a consistent structure.

Example:

```json
{
  "error": {
    "code": "VPN_SERVER_UNAVAILABLE",
    "message": "No healthy VPN server is currently available."
  }
}
```

Internal stack traces must not be returned to clients.

Internal error IDs can be used to correlate a client-facing error with server logs.

---

# Project Structure

The project follows a monorepo-style structure:

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
│   ├── api/
│   ├── vpn-engine/
│   └── agent/
│
├── clients/
│   ├── desktop/
│   │   ├── windows/
│   │   ├── macos/
│   │   └── linux/
│   ├── mobile/
│   │   ├── android/
│   │   └── ios/
│   └── shared/
│
├── packages/
│   ├── protocol/
│   ├── config/
│   └── schemas/
│
├── infrastructure/
│   ├── docker/
│   ├── terraform/
│   ├── ansible/
│   ├── kubernetes/
│   └── scripts/
│
├── deployments/
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

---

# Infrastructure

The platform is designed to support multiple Linux VPN nodes.

## Docker

Docker is used for local development and should provide the minimum development environment:

- Go API
- PostgreSQL
- VPN engine

## Terraform

Terraform is intended to manage cloud/infrastructure resources such as:

- Networks
- VPN servers
- Infrastructure resources
- Environment-level configuration

## Ansible

Ansible can configure VPN nodes and perform tasks such as:

- Server preparation
- WireGuard setup
- Firewall configuration
- Runtime configuration

## Kubernetes

Kubernetes is supported as a future deployment option, but it is deliberately not required for the initial MVP.

The architecture avoids introducing Kubernetes before it provides meaningful operational value.

---

# Client Architecture

The long-term client architecture includes:

```text
Windows
macOS
Linux
Android
iOS
```

However, the platform should not attempt to implement every client simultaneously.

The recommended progression is to prove the end-to-end architecture with one client platform first.

## Client Responsibilities

A client is responsible for:

- Authentication
- Device registration
- VPN configuration
- WireGuard tunnel control
- Connection state
- Disconnect
- Server selection
- Kill switch support

Private VPN keys should preferably be generated and stored on the client.

---

# Development Environment

The project is designed to provide a reproducible local development environment.

Expected development commands include:

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

The initial environment should be able to start PostgreSQL, the Go API, and the Rust VPN engine through Docker Compose.

---

# Configuration

Configuration is environment-based.

Example variables include:

```text
DATABASE_URL=
JWT_SECRET=
CONTROL_PLANE_URL=
VPN_ENGINE_TOKEN=
WIREGUARD_INTERFACE=
WIREGUARD_PORT=
LOG_LEVEL=
```

A template should be provided:

```text
.env.example
```

No production secrets belong in the repository.

---

# Testing

Testing is mandatory for critical functionality.

## Go Tests

The API should include:

- Unit tests
- Service tests
- Repository tests
- HTTP handler tests
- Integration tests

## Rust Tests

The VPN engine should include:

- Unit tests
- Configuration tests
- Network tests
- WireGuard management tests
- Security tests

## End-to-End Tests

The target end-to-end flow is:

```text
Register user
    ↓
Register device
    ↓
Select VPN server
    ↓
Create peer
    ↓
Configure WireGuard
    ↓
Connect
    ↓
Verify session
    ↓
Disconnect
```

Automated tests should not require a real public VPN server for every test. Mocks and test doubles should be used where appropriate.

---

# API Specification

The project maintains its API contract in:

```text
packages/protocol/openapi.yaml
```

The specification should document:

- Authentication
- Users
- Devices
- VPN servers
- VPN configuration
- Sessions
- Subscriptions
- Errors

The API implementation and OpenAPI documentation should remain synchronized.

---

# Documentation

The project documentation is organized into focused technical documents:

```text
docs/architecture.md
docs/api.md
docs/security.md
docs/networking.md
docs/deployment.md
docs/threat-model.md
```

These documents should explain:

- System architecture
- Data flow
- Authentication
- WireGuard integration
- Server provisioning
- Networking
- Firewall requirements
- Deployment
- Security assumptions
- Known limitations

---

# Threat Model

Solid VPN should explicitly consider at least the following threats:

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

Each threat should be documented using:

```text
Threat
Impact
Likelihood
Mitigation
Residual risk
```

The project should not make broad claims that the system is "secure" without documenting the assumptions, controls, and remaining risks.

---

# Development Phases

The project is designed to be implemented incrementally.

## Phase 1 — Foundation

Establish:

- Repository structure
- Go module
- Rust project
- Docker Compose
- PostgreSQL
- Environment configuration
- Makefile
- Initial documentation

## Phase 2 — Go API

Implement:

- Health endpoints
- Database connection
- Migrations
- User registration
- Login
- Authentication
- Device registration

## Phase 3 — VPN Server Management

Implement:

- VPN server records
- VPN regions
- Server registration
- Server health
- Server selection

## Phase 4 — Rust VPN Engine

Implement:

- Configuration
- WireGuard interface management
- Peer management
- Routing
- NAT
- Firewall integration
- Health
- Metrics

Linux is the initial target for the VPN engine.

## Phase 5 — Go/Rust Communication

Implement:

- Authenticated control channel
- Peer creation
- Peer removal
- Server health
- Status reporting

## Phase 6 — End-to-End VPN

Connect:

```text
User
 ↓
Device
 ↓
Server Selection
 ↓
Peer Creation
 ↓
WireGuard Configuration
 ↓
VPN Tunnel
 ↓
Internet
```

Verify:

- Tunnel connectivity
- Traffic routing
- DNS
- Routing
- Disconnect
- Peer removal

## Phase 7 — Security Hardening

Perform:

- Threat modeling
- Dependency auditing
- Secret review
- Authentication review
- Authorization review
- Network security review
- Logging review
- Firewall review

## Phase 8 — Monitoring

Add:

- Prometheus
- Grafana
- Structured logs
- Health checks
- Alerts

## Phase 9 — Infrastructure

Add:

- Terraform
- Ansible
- Server provisioning
- Deployment scripts
- Staging environment
- Production environment

## Phase 10 — Client

Begin with one desktop client, then expand to:

```text
Windows
macOS
Linux
Android
iOS
```

---

# MVP Scope

The first functional MVP is intended to support the complete basic VPN lifecycle:

1. Create a user
2. Authenticate the user
3. Register a device
4. Register a VPN server
5. Select a VPN server
6. Generate a WireGuard identity/configuration
7. Create a VPN peer
8. Connect the client
9. Establish a WireGuard tunnel
10. Route traffic
11. Track the VPN session
12. Disconnect
13. Remove or disable the peer
14. Record audit events
15. Report server health

The MVP should prove the architecture before the platform expands into a larger multi-client, multi-region VPN service.

---

# Engineering Rules

The project follows several non-negotiable engineering rules.

## Do

- Keep control-plane and data-plane responsibilities separate
- Use established cryptographic technologies
- Validate inputs
- Enforce authorization server-side
- Use database constraints
- Use migrations
- Test critical behavior
- Use structured logging
- Monitor service health
- Keep infrastructure configuration separate from application logic
- Make security assumptions explicit
- Keep changes small and reviewable

## Do Not

Never:

- Invent custom encryption
- Invent a custom VPN protocol
- Store private keys unnecessarily
- Hardcode credentials
- Commit secrets
- Expose administrative engine endpoints publicly
- Trust client-provided authorization
- Put packet forwarding in the Go API
- Put business logic in the Rust VPN engine
- Log private keys or authentication tokens
- Blindly execute shell commands
- Accept arbitrary firewall commands from API requests
- Skip input validation
- Skip authorization
- Claim security without testing
- Introduce Kubernetes before it is needed
- Over-engineer the MVP

---

# Code Quality Checklist

Before a feature is considered complete:

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

# First Development Milestone

The foundation milestone is considered successful when the local development environment can start successfully with:

```bash
docker compose up
```

The environment should provide:

- A running Go API
- A running Rust VPN engine
- A reachable PostgreSQL database
- Successful project compilation
- No hardcoded secrets

The initial foundation should not present a fake VPN tunnel as functional.

The first milestone should establish the infrastructure and service boundaries needed for the real VPN implementation.

---

# Known Boundaries

The master specification deliberately leaves some implementation decisions open.

These include:

- Specific payment provider
- Specific cloud provider
- Specific client framework
- Exact authentication/token implementation
- Exact WireGuard integration library
- Exact secret-management platform
- Exact infrastructure topology for production
- Final deployment strategy
- Detailed UI design

Those decisions should be made based on the deployment environment, operational requirements, security review, and product needs rather than being invented prematurely.

---

# Future Expansion

Once the core architecture is proven, Solid VPN can expand toward:

- Multiple VPN regions
- More VPN server providers
- Multiple desktop clients
- Android client
- iOS client
- Automated server provisioning
- Advanced server selection
- More detailed analytics
- Subscription tiers
- Payment provider integrations
- Advanced monitoring
- Automated alerting
- Production deployment automation
- Additional client-side privacy controls
- Kill switch support
- Larger-scale VPN infrastructure

The architecture is intentionally designed so these capabilities can be added without collapsing the control plane and data plane into a single service.

---

# Project Philosophy

Solid VPN is built around one central architectural principle:

```text
                 SOLID VPN

              ┌──────────────┐
              │   Clients    │
              └──────┬───────┘
                     │
                     │ HTTPS
                     ▼
              ┌──────────────┐
              │ Go Control   │
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

Keeping these responsibilities separate provides a foundation for maintainability, security, scalability, and future product expansion.

---

# Conclusion

Solid VPN is designed as a complete VPN service platform rather than simply a WireGuard configuration tool.

The project covers the major layers required to operate a managed VPN service:

```text
Users
  ↓
Authentication
  ↓
Devices
  ↓
Subscriptions / Access
  ↓
VPN Server Selection
  ↓
Peer Management
  ↓
Rust VPN Engine
  ↓
WireGuard
  ↓
Routing / NAT / Firewall / DNS
  ↓
Internet
  ↓
Session + Health + Metrics + Audit
```

The architecture intentionally separates business operations from network packet handling.

That separation allows the Go control plane to evolve independently from the Rust networking engine, while WireGuard remains responsible for the underlying VPN protocol and cryptographic tunnel.

The result is a foundation intended to support a secure, observable, maintainable, and scalable VPN platform as the project grows.
