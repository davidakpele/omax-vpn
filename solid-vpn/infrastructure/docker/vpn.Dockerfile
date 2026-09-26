# ──────────────────────────────────────────────────────────────────────────────
# Stage 1: Build
# ──────────────────────────────────────────────────────────────────────────────
FROM rust:1.82-slim-bookworm AS builder

# Install system build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy dependency manifests first for layer caching.
# Create a dummy main.rs so `cargo build` can resolve and cache all deps
# before copying the real source.
COPY apps/vpn-engine/Cargo.toml apps/vpn-engine/Cargo.lock* ./
RUN mkdir -p src && echo "fn main() {}" > src/main.rs
RUN cargo build --release || true

# Now copy the real source and rebuild
COPY apps/vpn-engine/src ./src
# Touch main.rs so Cargo notices the change
RUN touch src/main.rs && cargo build --release

# ──────────────────────────────────────────────────────────────────────────────
# Stage 2: Runtime
# Uses Debian slim to retain libc and system networking tools needed by
# WireGuard management (ip, iptables) on a VPN node.
# ──────────────────────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# Install runtime networking tools required by the VPN engine
RUN apt-get update && apt-get install -y --no-install-recommends \
    wireguard-tools \
    iproute2 \
    iptables \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary
COPY --from=builder /build/target/release/vpn-engine /usr/local/bin/vpn-engine

# The VPN engine must run as root to manage network interfaces and iptables.
# In production, use Linux capabilities (NET_ADMIN, SYS_MODULE) instead of
# running as full root where possible.
EXPOSE 9090/tcp
EXPOSE 51820/udp

ENTRYPOINT ["/usr/local/bin/vpn-engine"]
