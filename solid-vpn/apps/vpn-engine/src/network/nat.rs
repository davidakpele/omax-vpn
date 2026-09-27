use tracing::info;

pub async fn enable_masquerade(vpn_interface: &str, outbound_interface: &str) {
    info!(vpn_interface, outbound_interface, "enable_masquerade");
}

pub async fn disable_masquerade(vpn_interface: &str, outbound_interface: &str) {
    info!(vpn_interface, outbound_interface, "disable_masquerade");
}
