# Solid VPN

A production-oriented VPN platform built with a strict separation between the control plane (Go) and the data plane (Rust), using WireGuard as the VPN protocol.

---

## Architecture

```
              ┌──────────────┐
              │   Clients    │
              └──────┬───────┘
                     │ HTTPS
                     ▼
              ┌──────────────┐
              │  Go Control  │
              │    Plane     │  Auth · Users · Devices
              │   (API)      │  Billing · Server Mgmt · Sessions
              └──────┬───────┘
                     │ Secure Control Channel
                     ▼
              ┌──────────────┐
              │  Rust Data   │
              │    Plane     │  WireGuard · Routing · NAT
              │  (Engine)    │  Firewall · DNS · Networking
              └──────┬───────┘
                     │
                     ▼
                  Internet
```

**Go** owns all business logic — authentication, user management, device registration, server selection, billing, and session lifecycle.

**Rust** owns all network operations — WireGuard interface management, peer configuration, routing, NAT, firewall, and DNS.

**WireGuard** provides the cryptographic VPN tunnel. No custom cryptography is implemented.

---

## Repository Structure

```
solid-vpn/
├── apps/
│   ├── api/            Go control-plane API
│   ├── vpn-engine/     Rust data-plane engine
│   └── agent/          Go node agent (Phase 4)
├── clients/            Desktop and mobile clients (Phase 10)
├── packages/           Shared protocol definitions and schemas
├── infrastructure/     Docker, Terraform, Ansible, Kubernetes, scripts
├── deployments/        Environment-specific deployment configs
├── monitoring/         Prometheus and Grafana configuration
├── docs/               Architecture and operational documentation
└── tests/              Integration, e2e, and load tests
```

---

## Quick Start

### Prerequisites

| Tool | Version |
|------|---------|
| Go | 1.23+ |
| Rust | 1.82+ |
| Docker | 24+ |
| Docker Compose | v2+ |

### 1. Clone and configure

```bash
git clone https://github.com/your-org/solid-vpn.git
cd solid-vpn
cp .env.example .env
# Edit .env and fill in all required secrets
```

### 2. Start the development stack

```bash
make docker-up
```

This starts:
- **PostgreSQL** on `localhost:5432`
- **Go API** on `http://localhost:8080`
- **Rust VPN engine** on `localhost:9090` (control API) / `51820/udp` (WireGuard)
- **Prometheus** on `http://localhost:9091`
- **Grafana** on `http://localhost:3000` (admin / see `.env`)

### 3. Verify health

```bash
curl http://localhost:8080/health
# {"status":"ok","timestamp":"...","service":"vpn-api"}

curl http://localhost:8080/ready
# {"status":"ready","checks":{"database":"healthy"},"timestamp":"..."}
```

### 4. Run database migrations

```bash
make migrate
```

---

## Development Commands

```bash
make build          # compile Go API + Rust engine
make test           # run all tests
make lint           # run all linters
make format         # auto-format all code
make docker-up      # start dev environment
make docker-down    # stop dev environment
make docker-logs    # follow all service logs
make migrate        # apply database migrations
make clean          # remove build artifacts
make help           # list all available targets
```

---

## Environment Variables

See [`.env.example`](.env.example) for the complete list of required variables with descriptions.

Critical values that **must** be set before starting:

| Variable | Description |
|----------|-------------|
| `POSTGRES_PASSWORD` | Database password |
| `JWT_SECRET` | 64-byte hex secret for JWT signing |
| `VPN_ENGINE_TOKEN` | 32-byte hex shared token for Go↔Rust auth |
| `VPN_ENGINE_SERVER_ID` | UUID identifying this VPN server |

Generate secrets with:

```bash
openssl rand -hex 64   # JWT_SECRET
openssl rand -hex 32   # VPN_ENGINE_TOKEN
uuidgen                # VPN_ENGINE_SERVER_ID
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/architecture.md](docs/architecture.md) | System architecture and component design |
| [docs/api.md](docs/api.md) | REST API reference |
| [docs/security.md](docs/security.md) | Security model and assumptions |
| [docs/networking.md](docs/networking.md) | Networking, routing, and WireGuard setup |
| [docs/deployment.md](docs/deployment.md) | Deployment guide for staging and production |
| [docs/threat-model.md](docs/threat-model.md) | Threat model and mitigations |

---

## Development Phases

| Phase | Status | Description |
|-------|--------|-------------|
| 1 — Foundation | ✅ Complete | Project structure, Go API skeleton, Rust engine skeleton, Docker, PostgreSQL |
| 2 — Go API | 🔜 Next | Auth, users, devices, subscriptions |
| 3 — VPN Server Mgmt | ⬜ Planned | VPN servers, regions, server selection |
| 4 — Rust VPN Engine | ⬜ Planned | WireGuard management, routing, NAT, firewall |
| 5 — Go↔Rust Channel | ⬜ Planned | Authenticated control channel, peer lifecycle |
| 6 — End-to-End VPN | ⬜ Planned | Full tunnel, session tracking, disconnect |
| 7 — Security Hardening | ⬜ Planned | Threat modeling, audit, hardening |
| 8 — Monitoring | ⬜ Planned | Prometheus, Grafana, alerts |
| 9 — Infrastructure | ⬜ Planned | Terraform, Ansible, multi-region |
| 10 — Clients | ⬜ Planned | Desktop and mobile clients |

---

## Security

- No custom cryptography — WireGuard handles all VPN cryptography
- Secrets are never committed — use `.env` (gitignored)
- Private WireGuard keys remain on clients wherever possible
- All control-plane↔engine communication is authenticated
- Least-privilege principles throughout
- See [docs/security.md](docs/security.md) and [docs/threat-model.md](docs/threat-model.md)

---

## License

MIT — see [LICENSE](LICENSE)
