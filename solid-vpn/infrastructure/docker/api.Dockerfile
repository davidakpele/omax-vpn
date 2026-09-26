# ──────────────────────────────────────────────────────────────────────────────
# Stage 1: Build
# ──────────────────────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy dependency manifests first for layer caching
COPY apps/api/go.mod apps/api/go.sum ./

# Download dependencies (cached unless go.mod/go.sum change)
RUN go mod download

# Copy the rest of the API source
COPY apps/api/ .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/vpn-api \
    ./cmd/api

# ──────────────────────────────────────────────────────────────────────────────
# Stage 2: Runtime (minimal scratch image)
# ──────────────────────────────────────────────────────────────────────────────
FROM scratch

# Copy timezone data and CA certificates for HTTPS outbound calls
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /bin/vpn-api /vpn-api

# Run as non-root uid 65534 (nobody)
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/vpn-api"]
