use config::{Config, Environment, File};
use serde::Deserialize;

/// Runtime settings for the VPN engine.
/// All values are loaded from environment variables (prefixed with `VPN_ENGINE_`)
/// and optionally an `engine.toml` config file.
#[derive(Debug, Deserialize, Clone)]
pub struct Settings {
    /// Listening port for the engine's internal control API.
    #[serde(default = "default_control_port")]
    pub control_port: u16,

    /// URL of the Go control-plane API.
    pub control_plane_url: String,

    /// Shared token used to authenticate requests from the control plane.
    pub control_plane_token: String,

    /// Unique identifier of this VPN server instance (UUID string).
    pub server_id: String,

    /// Name of the WireGuard network interface to manage (e.g. "wg0").
    #[serde(default = "default_wg_interface")]
    pub wireguard_interface: String,

    /// UDP port WireGuard listens on.
    #[serde(default = "default_wg_port")]
    pub wireguard_port: u16,

    /// CIDR range from which peer IPs are allocated (e.g. "10.8.0.0/24").
    #[serde(default = "default_peer_cidr")]
    pub peer_cidr: String,

    /// DNS server pushed to VPN peers.
    #[serde(default = "default_dns")]
    pub dns_server: String,

    /// Tracing/log level: trace | debug | info | warn | error.
    #[serde(default = "default_log_level")]
    pub log_level: String,

    /// Environment: development | staging | production.
    #[serde(default = "default_environment")]
    pub environment: String,
}

impl Settings {
    /// Load settings from environment variables and optional file.
    ///
    /// Environment variables use the prefix `VPN_ENGINE_` and double-underscores
    /// as hierarchy separators, e.g. `VPN_ENGINE_WIREGUARD_PORT=51820`.
    pub fn load() -> Result<Self, config::ConfigError> {
        let cfg = Config::builder()
            // Optional file-based config (ignored if missing)
            .add_source(File::with_name("engine").required(false))
            // Environment variables override file values
            .add_source(
                Environment::with_prefix("VPN_ENGINE")
                    .prefix_separator("_")
                    .separator("__"),
            )
            .build()?;

        cfg.try_deserialize()
    }
}

fn default_control_port() -> u16 {
    9090
}

fn default_wg_interface() -> String {
    "wg0".to_string()
}

fn default_wg_port() -> u16 {
    51820
}

fn default_peer_cidr() -> String {
    "10.8.0.0/24".to_string()
}

fn default_dns() -> String {
    "1.1.1.1".to_string()
}

fn default_log_level() -> String {
    "info".to_string()
}

fn default_environment() -> String {
    "development".to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn defaults_are_sane() {
        assert_eq!(default_control_port(), 9090);
        assert_eq!(default_wg_port(), 51820);
        assert_eq!(default_wg_interface(), "wg0");
    }
}
