/// NAT (masquerade) configuration for outbound VPN traffic.
/// Phase 4 implementation.
use tracing::info;

/// Enable NAT masquerading so VPN traffic exits via the server's public interface.
pub async fn enable_masquerade(vpn_interface: &str, outbound_interface: &str) {
    // TODO(phase-4): iptables -t nat -A POSTROUTING -o <outbound> -j MASQUERADE
    info!(vpn_interface, outbound_interface, "enable_masquerade (placeholder)");
}

/// Disable NAT masquerading.
pub async fn disable_masquerade(vpn_interface: &str, outbound_interface: &str) {
    // TODO(phase-4): iptables -t nat -D POSTROUTING -o <outbound> -j MASQUERADE
    info!(vpn_interface, outbound_interface, "disable_masquerade (placeholder)");
}
