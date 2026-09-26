/// Network interface state management.
/// Phase 4 implementation.
use tracing::debug;

/// Brings up a network interface by name.
pub async fn bring_up(interface: &str) {
    // TODO(phase-4): netlink / ip link set up
    debug!(interface, "bring_up (placeholder)");
}

/// Tears down a network interface by name.
pub async fn bring_down(interface: &str) {
    // TODO(phase-4): netlink / ip link set down
    debug!(interface, "bring_down (placeholder)");
}
