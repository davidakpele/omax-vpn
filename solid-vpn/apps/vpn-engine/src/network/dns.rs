/// DNS configuration for VPN peers.
///
/// Handles DNS leak prevention: peers should use the configured DNS server
/// exclusively, routed through the VPN tunnel.
/// Phase 4 implementation.
use tracing::info;

/// Configure the DNS server advertised to peers.
pub async fn configure_peer_dns(dns_server: &str) {
    // TODO(phase-4): configure resolved/dnsmasq or push DNS via WireGuard config
    info!(dns_server, "configure_peer_dns (placeholder)");
}
