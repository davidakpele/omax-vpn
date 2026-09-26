/// WireGuard peer management.
/// Phase 4 implementation.
use tracing::info;

/// Represents a WireGuard peer configuration.
#[derive(Debug, Clone)]
pub struct PeerConfig {
    pub public_key: String,
    pub assigned_ip: String,
    pub peer_id: String,
}

/// Add a peer to the WireGuard interface.
pub async fn add_peer(interface: &str, peer: &PeerConfig) {
    // TODO(phase-4): wg set <interface> peer <public_key> allowed-ips <assigned_ip>/32
    info!(
        interface,
        peer_id = %peer.peer_id,
        assigned_ip = %peer.assigned_ip,
        "add_peer (placeholder)"
    );
}

/// Remove a peer from the WireGuard interface.
pub async fn remove_peer(interface: &str, public_key: &str) {
    // TODO(phase-4): wg set <interface> peer <public_key> remove
    info!(interface, public_key, "remove_peer (placeholder)");
}
