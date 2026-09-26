/// Authentication helpers for validating control-plane requests.
///
/// The engine verifies a bearer token on every incoming control request.
/// It does NOT trust requests based on source IP alone.
/// Phase 4 implementation.
use tracing::warn;

/// Validate a bearer token against the configured control-plane token.
/// Returns true if valid, false otherwise.
///
/// SECURITY: Uses constant-time comparison to prevent timing attacks.
pub fn validate_token(provided: &str, expected: &str) -> bool {
    use std::hint::black_box;
    // Constant-time comparison: iterate all bytes even if they differ early.
    if provided.len() != expected.len() {
        warn!("token length mismatch — rejected");
        return false;
    }
    let result = provided
        .bytes()
        .zip(expected.bytes())
        .fold(0u8, |acc, (a, b)| acc | (a ^ b));
    black_box(result) == 0
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn valid_token_accepted() {
        assert!(validate_token("correct-token", "correct-token"));
    }

    #[test]
    fn invalid_token_rejected() {
        assert!(!validate_token("wrong-token--", "correct-token"));
    }

    #[test]
    fn different_length_rejected() {
        assert!(!validate_token("short", "a-much-longer-token"));
    }
}
