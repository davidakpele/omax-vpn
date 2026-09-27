use tracing::info;

pub async fn init_interface(interface: &str, port: u16) {
    info!(interface, port, "wireguard interface init");
}
