use tracing::info;

pub async fn apply_base_rules(interface: &str) {
    info!(interface, "apply_base_rules");
}

pub async fn allow_peer(peer_ip: &str) {
    info!(peer_ip, "allow_peer");
}

pub async fn revoke_peer(peer_ip: &str) {
    info!(peer_ip, "revoke_peer");
}
