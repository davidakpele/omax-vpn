/// IP routing management for VPN traffic.
/// Phase 4 implementation.
use tracing::debug;

/// Add a route for a peer's assigned IP through the VPN interface.
pub async fn add_peer_route(interface: &str, peer_ip: &str) {
    // TODO(phase-4): ip route add <peer_ip>/32 dev <interface>
    debug!(interface, peer_ip, "add_peer_route (placeholder)");
}

/// Remove a peer's route.
pub async fn remove_peer_route(peer_ip: &str) {
    // TODO(phase-4): ip route del <peer_ip>/32
    debug!(peer_ip, "remove_peer_route (placeholder)");
}
