# Deployment

## Environments

| Environment | Purpose |
|-------------|---------|
| Development | Local Docker Compose — all services on one machine |
| Staging | Production-like, separate VPN nodes, real WireGuard tunnels |
| Production | Multi-region, hardened, monitored |

Kubernetes is not required for the MVP. Add it when horizontal scaling demands it.

---

## Prerequisites

### All environments

- Docker 24+, Docker Compose v2
- Go 1.23+
- Rust 1.82+
- PostgreSQL 16+

### VPN nodes (staging / production)

- Linux (Ubuntu 22.04 LTS recommended)
- WireGuard kernel module loaded: `modprobe wireguard`
- `wireguard-tools`, `iproute2`, `iptables` installed
- IP forwarding enabled (see [networking.md](networking.md))
- `NET_ADMIN` capability available to the Rust engine process

---

## Local Development

```bash
# 1. Copy and configure environment
cp .env.example .env
# Edit .env — fill in all secrets

# 2. Start all services
make docker-up

# 3. Run migrations
make migrate

# 4. Verify
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

### Services started by `make docker-up`

| Service | URL |
|---------|-----|
| Go API | http://localhost:8080 |
| Prometheus | http://localhost:9091 |
| Grafana | http://localhost:3000 |
| PostgreSQL | localhost:5432 |

---

## Building Production Images

```bash
# Build both images
make docker-build

# Or build individually
docker build -f infrastructure/docker/api.Dockerfile -t solid-vpn/api:latest .
docker build -f infrastructure/docker/vpn.Dockerfile -t solid-vpn/vpn-engine:latest .
```

The API image uses a `scratch` base — the binary is fully statically linked. The VPN engine image uses `debian:bookworm-slim` to retain access to system networking tools.

---

## Database Migrations

Migrations are located in `apps/api/migrations/` and are numbered sequentially:

```
000001_create_users.sql
000002_create_devices.sql
...
```

Apply migrations:

```bash
# With DATABASE_URL set in .env
make migrate

# Or manually with golang-migrate
migrate -path apps/api/migrations -database "$DATABASE_URL" up

# Roll back one migration
make migrate-down
```

Never modify an already-applied migration. Add a new migration instead.

---

## Environment Variable Reference

See [`.env.example`](../.env.example) for all variables. Critical production values:

| Variable | Required | Notes |
|----------|----------|-------|
| `DATABASE_URL` | ✅ | Full PostgreSQL connection string |
| `JWT_SECRET` | ✅ | 64-byte hex, generated with `openssl rand -hex 64` |
| `VPN_ENGINE_TOKEN` | ✅ | 32-byte hex, generated with `openssl rand -hex 32` |
| `VPN_ENGINE_SERVER_ID` | ✅ | UUID per node, generated with `uuidgen` |
| `POSTGRES_PASSWORD` | ✅ | Strong random password |
| `ENVIRONMENT` | ✅ | `production` in prod |
| `LOG_LEVEL` | ❌ | `info` recommended for production |

---

## VPN Node Provisioning (Phase 9)

VPN nodes are provisioned with Ansible and managed with Terraform.

```
infrastructure/
├── terraform/
│   ├── providers.tf     Cloud provider config
│   ├── network.tf       VPC, subnets, security groups
│   ├── servers.tf       VPN node instances
│   └── variables.tf     Input variables
└── ansible/
    ├── playbooks/       Node setup playbooks
    └── roles/           Reusable roles (wireguard, firewall, vpn-engine)
```

Provisioning steps (Phase 9 planned):

```bash
# Provision infrastructure
cd infrastructure/terraform
terraform init
terraform plan
terraform apply

# Configure VPN nodes
cd infrastructure/ansible
ansible-playbook playbooks/setup-vpn-node.yml -i inventory/production
```

---

## Health Checks

### Go API

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Liveness — process is running |
| `GET /ready` | Readiness — database connected, dependencies up |

```bash
curl http://localhost:8080/health
# {"status":"ok","timestamp":"...","service":"vpn-api"}

curl http://localhost:8080/ready
# {"status":"ready","checks":{"database":"healthy"},"timestamp":"..."}
```

### Rust VPN engine

- Prometheus metrics exposed on `:9090/metrics`
- Health endpoint planned for Phase 4

---

## Monitoring

Prometheus scrapes metrics from:
- Go API at `:8080/metrics`
- Rust engine at `:9090/metrics`

Grafana is pre-configured with a Prometheus datasource and is available at `http://localhost:3000`.

Dashboards and alert rules are in `monitoring/grafana/` and `monitoring/alerts/`.

---

## Graceful Shutdown

Both the Go API and Rust engine handle `SIGTERM` and `SIGINT`:

- **Go API**: stops accepting new connections, waits up to 30 seconds for in-flight requests to complete.
- **Rust engine**: tears down NAT rules and brings down the WireGuard interface cleanly.

Docker Compose sends `SIGTERM` on `docker compose down`. Allow sufficient stop grace period in production orchestration.

---

## Security Checklist (Pre-Production)

```
[ ] All secrets are in a secrets manager, not .env files
[ ] TLS is terminated at the load balancer for the Go API
[ ] Go API ↔ Rust engine communication uses mTLS
[ ] Port 9090 (engine control API) is not publicly accessible
[ ] Database is not publicly accessible
[ ] IP forwarding is enabled on VPN nodes
[ ] WireGuard kernel module is loaded on VPN nodes
[ ] Firewall baseline rules are applied
[ ] Audit logging is enabled and logs are shipped to a SIEM
[ ] Monitoring alerts are configured
[ ] Backups for PostgreSQL are configured and tested
[ ] Dependency audit has been run (govulncheck, cargo audit)
```
