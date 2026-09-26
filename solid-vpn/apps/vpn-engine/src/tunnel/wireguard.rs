/// WireGuard interface management.
///
/// This module will wrap the system WireGuard interface (via `wg` / `wg-quick`
/// or the `wireguard-control` crate) to add/remove peers and manage interface
/// state. It deliberately does NOT implement WireGuard cryptography.
///
/// Phase 4 implementation.
use tracing::info;

/// Placeholder: initialise WireGuard interface.
pub async fn init_interface(interface: &str, port: u16) {
    // TODO(phase-4): call wireguard-control or netlink to create wgX interface
    info!(interface, port, "wireguard interface init (placeholder)");
}
