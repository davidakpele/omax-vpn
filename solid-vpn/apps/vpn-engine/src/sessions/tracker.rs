/// Session tracker — periodic session health and expiry checks.
/// Phase 4 implementation.
use tracing::debug;

/// Check all active sessions and expire stale ones.
pub async fn run_expiry_check() {
    // TODO(phase-4): iterate sessions, compare last_handshake with WireGuard kernel stats
    debug!("session expiry check (placeholder)");
}
