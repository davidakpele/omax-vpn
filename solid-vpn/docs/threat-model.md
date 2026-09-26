# Threat Model

## Scope

This threat model covers the Solid VPN platform as designed for Phase 1 (foundation) through Phase 6 (end-to-end VPN). It will be updated as the system evolves.

## Assumptions

- The Go API runs on a trusted, access-controlled server.
- The Rust VPN engine runs on dedicated Linux VPN nodes with restricted inbound network access.
- Clients are untrusted — all authorization is enforced server-side.
- WireGuard's cryptographic design is correct and audited — we do not evaluate it here.
- The database is not publicly accessible.

---

## Threats

### T-01 — Compromised Client

| Field | Detail |
|-------|--------|
| **Threat** | An attacker gains access to a user's device or steals their session token. |
| **Impact** | Unauthorized VPN access using the victim's identity and subscription. |
| **Likelihood** | Medium |
| **Mitigations** | Short-lived JWT access tokens (15 min). Refresh tokens are revocable server-side. Users can delete devices via API, which removes the associated WireGuard peer. Per-device keys mean compromise of one device does not compromise others. |
| **Residual risk** | If both access and refresh tokens are stolen before expiry, the attacker can maintain access until manual revocation. |

---

### T-02 — Compromised VPN Server Node

| Field | Detail |
|-------|--------|
| **Threat** | An attacker gains root access to a VPN node running the Rust engine. |
| **Impact** | Attacker can intercept traffic of connected peers, extract the server's WireGuard private key, and impersonate the server. |
| **Likelihood** | Low |
| **Mitigations** | VPN nodes are hardened Linux instances with minimal attack surface. The engine control API (9090) is not publicly accessible. WireGuard private keys are stored in environment variables, not on disk. Server private key rotation is planned. |
| **Residual risk** | Traffic of currently-connected peers could be decrypted. Historical traffic (before the key was stolen) is protected by WireGuard's forward secrecy. |

---

### T-03 — Compromised Go API

| Field | Detail |
|-------|--------|
| **Threat** | An attacker gains code execution in the Go API process. |
| **Impact** | Full access to the database, ability to create rogue peers, ability to issue tokens, access to `VPN_ENGINE_TOKEN`. |
| **Likelihood** | Low |
| **Mitigations** | Input validation on all API inputs. Parameterized queries — no SQL injection. Dependency auditing. Minimal permissions for the API's database user. `VPN_ENGINE_TOKEN` in environment only. |
| **Residual risk** | If the API is compromised, the attacker controls the full control plane. Audit logs may be the only forensic record. |

---

### T-04 — Stolen Authentication Token

| Field | Detail |
|-------|--------|
| **Threat** | An access or refresh JWT is intercepted or exfiltrated from logs or storage. |
| **Impact** | Impersonation of the token's subject until expiry or revocation. |
| **Likelihood** | Medium |
| **Mitigations** | Access tokens expire in 15 minutes. Tokens are never logged. HTTPS is required between clients and the API. Refresh tokens are stored server-side and can be revoked. |
| **Residual risk** | A stolen access token is valid until expiry. A stolen refresh token can be used until detected and revoked. |

---

### T-05 — Stolen Device

| Field | Detail |
|-------|--------|
| **Threat** | A user's physical device is stolen. |
| **Impact** | Attacker may be able to extract the WireGuard private key and establish a VPN session. |
| **Likelihood** | Low |
| **Mitigations** | Users can delete the device via the API, triggering immediate peer removal from the WireGuard interface. Per-device keys limit blast radius. Device-level encryption (OS-level) is recommended. |
| **Residual risk** | If the private key is extracted before the device is reported, the attacker could establish a tunnel with it until the peer is removed. |

---

### T-06 — Malicious Administrator

| Field | Detail |
|-------|--------|
| **Threat** | A platform administrator abuses privileged access. |
| **Impact** | Access to all user data, ability to create/remove peers, access to billing data. |
| **Likelihood** | Low |
| **Mitigations** | All admin actions are written to the immutable audit log. Admin roles are separate from user roles. Principle of least privilege for admin accounts. |
| **Residual risk** | An admin can read user data. This is an inherent trust assumption. Audit trails enable detection after the fact. |

---

### T-07 — Database Compromise

| Field | Detail |
|-------|--------|
| **Threat** | The PostgreSQL database is accessed directly by an attacker. |
| **Impact** | Exposure of all user data, hashed passwords, device records, session history, billing data. |
| **Likelihood** | Low |
| **Mitigations** | Database is not publicly accessible. Passwords are stored as bcrypt hashes. WireGuard private keys are not stored in the database. Database user has minimal permissions (no superuser). |
| **Residual risk** | Bcrypt hashes are exposed but not immediately usable. A brute-force attack against weak passwords is feasible over time. |

---

### T-08 — Network Interception (Between Client and API)

| Field | Detail |
|-------|--------|
| **Threat** | An attacker intercepts traffic between the client and the Go API. |
| **Impact** | Credential theft, token theft, configuration exposure. |
| **Likelihood** | Medium |
| **Mitigations** | All client-to-API communication uses HTTPS with TLS. Certificate pinning is recommended for production clients. |
| **Residual risk** | If TLS is misconfigured or a trusted CA is compromised, interception is possible. |

---

### T-09 — Credential Leakage via Logs

| Field | Detail |
|-------|--------|
| **Threat** | Secrets are inadvertently written to log output. |
| **Impact** | Credential exposure to anyone with log access. |
| **Likelihood** | Low |
| **Mitigations** | Strict coding convention: passwords, tokens, and private keys are never passed to log functions. Code review and linting enforce this. |
| **Residual risk** | Future code changes could introduce accidental logging. Automated secret scanning (e.g., truffleHog, gitleaks) is recommended in CI. |

---

### T-10 — Misconfigured Firewall

| Field | Detail |
|-------|--------|
| **Threat** | Firewall rules are missing or misconfigured, allowing unauthorized traffic through the VPN node. |
| **Impact** | VPN traffic could bypass intended access controls. Engine control API could be exposed publicly. |
| **Likelihood** | Medium |
| **Mitigations** | The Rust engine applies firewall rules programmatically with validated inputs. Baseline rules are applied at startup. Engine control API binds to a private interface only. |
| **Residual risk** | A bug in the engine's firewall management could leave gaps. Infrastructure-level firewall rules (security groups, host firewall) should provide a second layer. |

---

### T-11 — DNS Leakage

| Field | Detail |
|-------|--------|
| **Threat** | DNS queries from VPN clients bypass the tunnel and go to the ISP's resolver. |
| **Impact** | Browsing activity is visible to the ISP despite the VPN being active. |
| **Likelihood** | Medium |
| **Mitigations** | All peers are pushed a DNS server within the VPN (`VPN_ENGINE_DNS_SERVER`). Full tunnel mode (`AllowedIPs = 0.0.0.0/0`) routes all traffic including DNS through the tunnel. |
| **Residual risk** | Client-side DNS leak prevention depends on the client OS respecting the pushed DNS. A client-side kill switch is planned for Phase 10. |

---

### T-12 — IP Leakage (WebRTC / IPv6)

| Field | Detail |
|-------|--------|
| **Threat** | The client's real IP is revealed via WebRTC or IPv6 while the VPN is active. |
| **Impact** | User identity and location exposed despite VPN usage. |
| **Likelihood** | Medium |
| **Mitigations** | IPv6 is included in AllowedIPs (`AllowedIPs = 0.0.0.0/0, ::/0`). Clients should disable WebRTC where possible. Kill switch will block non-tunnel traffic. |
| **Residual risk** | WebRTC leak prevention depends on browser/application behavior. Not fully controllable by the VPN platform. |

---

### T-13 — Replay Attacks

| Field | Detail |
|-------|--------|
| **Threat** | An attacker replays a captured API request or WireGuard handshake. |
| **Impact** | Unauthorized actions or tunnel establishment. |
| **Likelihood** | Low |
| **Mitigations** | WireGuard uses timestamp-based anti-replay protection natively. JWTs have short expiry windows. HTTPS uses TLS session nonces. |
| **Residual risk** | Replay of a valid JWT within its 15-minute window is possible if the token is stolen. |

---

### T-14 — Unauthorized Peer Creation

| Field | Detail |
|-------|--------|
| **Threat** | An attacker creates a WireGuard peer without authorization, gaining VPN access. |
| **Impact** | Unauthorized VPN usage; potential for lateral movement on the VPN network. |
| **Likelihood** | Low |
| **Mitigations** | Peer creation requires a valid authenticated session with an active subscription. The Rust engine only accepts peer commands authenticated with `VPN_ENGINE_TOKEN`. Direct access to the engine's control API requires network-level access to the private interface. |
| **Residual risk** | If `VPN_ENGINE_TOKEN` is compromised, an attacker can create arbitrary peers. Token rotation and mTLS are planned for Phase 7. |

---

## Residual Risk Summary

The following risks are accepted for Phase 1 and are tracked for future mitigation:

| Risk | Phase for mitigation |
|------|---------------------|
| Pre-shared token for Go↔Rust auth | Phase 5 (mTLS) |
| No rate limiting on API | Phase 2 |
| No kill switch on clients | Phase 10 |
| No automated WireGuard key rotation on servers | Phase 7 |
| No secret manager integration | Phase 7 |
| No WebRTC/IPv6 leak protection in client | Phase 10 |

This document does not claim the system is secure. It documents the current security posture, known limitations, and planned improvements.
