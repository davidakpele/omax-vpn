/// Security module — control-plane authentication and key hygiene.
///
/// SECURITY: Private keys are never logged.
/// Tokens are validated on every inbound request; never trusted by IP alone.
/// Phase 4 implementation.
pub mod auth;
pub mod keys;
