/// Firewall rule management (iptables / nftables).
///
/// SECURITY: Rules are never constructed from raw user input.
/// All values are validated and sanitised before any system call.
/// Phase 4 implementation.
use tracing::info;

/// Apply the baseline firewall rules for a VPN interface.
pub async fn apply_base_rules(interface: &str) {
    // TODO(phase-4): apply iptables/nftables base rules
    // - Allow WireGuard UDP inbound
    // - Allow forwarding for VPN CIDR
    // - Default deny for unexpected traffic
    info!(interface, "apply_base_rules (placeholder)");
}

/// Allow traffic for a specific peer IP.
pub async fn allow_peer(peer_ip: &str) {
    // TODO(phase-4): iptables -A FORWARD -s <peer_ip>/32 -j ACCEPT
    info!(peer_ip, "allow_peer (placeholder)");
}

/// Remove firewall rules for a peer IP.
pub async fn revoke_peer(peer_ip: &str) {
    // TODO(phase-4): iptables -D FORWARD -s <peer_ip>/32 -j ACCEPT
    info!(peer_ip, "revoke_peer (placeholder)");
}
