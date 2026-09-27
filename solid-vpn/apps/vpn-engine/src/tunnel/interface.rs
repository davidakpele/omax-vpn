use tracing::debug;

pub async fn bring_up(interface: &str) {
    debug!(interface, "bring_up");
}

pub async fn bring_down(interface: &str) {
    debug!(interface, "bring_down");
}
