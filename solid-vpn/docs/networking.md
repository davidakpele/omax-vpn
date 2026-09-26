# Networking

## WireGuard Overview

Solid VPN uses WireGuard as its VPN protocol. WireGuard operates at Layer 3 (network layer) and creates an encrypted tunnel between the client and the VPN server.

Key properties:
- Uses the **Noise Protocol Framework** for cryptographic handshakes
- ChaCha20-Poly1305 for data encryption
- Curve25519 for key exchange
- BLAKE2s for hashing
- No custom cryptography is implemented in this codebase

---

## IP Addressing

### VPN address space

By default, peers are allocated addresses from `10.8.0.0/24`. This is configurable via `VPN_ENGINE_PEER_CIDR`.

| Address | Assignment |
|---------|-----------|
| `10.8.0.1` | VPN server (WireGuard interface) |
| `10.8.0.2 – 10.8.0.254` | Allocated to peers |

IP allocation is managed by the Go API's `ip_allocations` table to prevent conflicts.

### Multi-server

Each VPN server node uses a separate CIDR subnet to avoid routing conflicts. Example:

| Node | Subnet |
|------|--------|
| Nigeria | `10.8.0.0/24` |
| UK | `10.9.0.0/24` |
| USA | `10.10.0.0/24` |

---

## WireGuard Interface

The Rust engine manages a WireGuard interface (default: `wg0`) on the VPN server node.

### Interface configuration

```ini
[Interface]
Address    = 10.8.0.1/24
ListenPort = 51820
PrivateKey = <server private key>
```

### Peer configuration (pushed to client)

```ini
[Interface]
PrivateKey = <client private key>
Address    = 10.8.0.X/32
DNS        = 1.1.1.1

[Peer]
PublicKey           = <server public key>
Endpoint            = <server public IP>:51820
AllowedIPs          = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
```

`AllowedIPs = 0.0.0.0/0` routes all client traffic through the tunnel (full-tunnel mode).

---

## Routing

### Server-side routing

When a peer connects, the Rust engine adds a host route for the peer's assigned IP:

```bash
ip route add 10.8.0.X/32 dev wg0
```

When a peer disconnects:

```bash
ip route del 10.8.0.X/32
```

### IP forwarding

The server must have IP forwarding enabled:

```bash
# Persistent — add to /etc/sysctl.d/99-wireguard.conf
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
```

---

## NAT (Masquerade)

Outbound traffic from VPN peers is masqueraded so it appears to originate from the VPN server's public IP:

```bash
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
```

`eth0` is the server's outbound interface (`VPN_ENGINE_OUTBOUND_INTERFACE`).

---

## Firewall Rules

The Rust engine applies the following rule set. Rules are constructed from validated, typed values — never from raw user input.

### iptables baseline

```bash
# Allow WireGuard inbound
iptables -A INPUT -p udp --dport 51820 -j ACCEPT

# Allow VPN traffic forwarding
iptables -A FORWARD -i wg0 -j ACCEPT
iptables -A FORWARD -o wg0 -j ACCEPT

# Allow established/related traffic
iptables -A INPUT   -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -m state --state ESTABLISHED,RELATED -j ACCEPT
```

### Per-peer rules

On peer add:
```bash
iptables -A FORWARD -s 10.8.0.X/32 -j ACCEPT
```

On peer remove:
```bash
iptables -D FORWARD -s 10.8.0.X/32 -j ACCEPT
```

---

## DNS

To prevent DNS leaks, peers are pushed a DNS server that is routed through the VPN tunnel.

Default DNS: `1.1.1.1` (configurable via `VPN_ENGINE_DNS_SERVER`).

In a full deployment, consider running a private DNS resolver on the VPN server to prevent queries leaking to the upstream provider.

---

## DNS Leak Prevention

A DNS leak occurs when DNS queries bypass the VPN tunnel and go directly to the ISP's resolver, revealing browsing activity.

Mitigations:
- Push a VPN-internal or trusted DNS server to all peers
- Route all traffic through the tunnel (`AllowedIPs = 0.0.0.0/0`)
- Block DNS queries on non-VPN interfaces at the firewall level (planned for Phase 6)
- Implement a kill switch on the client to block all traffic if the VPN drops (planned for Phase 10)

---

## WireGuard Port

| Protocol | Port | Purpose |
|----------|------|---------|
| UDP | 51820 | WireGuard VPN tunnel |
| TCP | 9090 | Rust engine control API (internal only) |
| TCP | 8080 | Go API (behind load balancer / TLS termination) |

Port 9090 must never be exposed to the public internet.

---

## Multi-Region Architecture

Each VPN node runs an independent Rust engine instance. The Go API communicates with each node's engine over the secure control channel.

```
Go API
  │
  ├── → vpn-node-nigeria:9090
  ├── → vpn-node-uk:9090
  └── → vpn-node-usa:9090
```

Server selection is handled by the Go API's `vpn/selector.go` component, which considers:
- Server health status
- Active connection count vs capacity
- User's requested region
- Server availability and maintenance flags
