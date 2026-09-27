use tracing::info;

#[derive(Debug, Clone)]
pub struct PeerConfig {
    pub public_key: String,
    pub assigned_ip: String,
    pub peer_id: String,
}

pub async fn add_peer(interface: &str, peer: &PeerConfig) {
    info!(
        interface,
        peer_id = %peer.peer_id,
        assigned_ip = %peer.assigned_ip,
        "add_peer"
    );
}

pub async fn remove_peer(interface: &str, public_key: &str) {
    info!(interface, public_key, "remove_peer");
}
