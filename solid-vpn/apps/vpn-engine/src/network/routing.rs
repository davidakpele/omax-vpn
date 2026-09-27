use tracing::debug;

pub async fn add_peer_route(interface: &str, peer_ip: &str) {
    debug!(interface, peer_ip, "add_peer_route");
}

pub async fn remove_peer_route(peer_ip: &str) {
    debug!(peer_ip, "remove_peer_route");
}
