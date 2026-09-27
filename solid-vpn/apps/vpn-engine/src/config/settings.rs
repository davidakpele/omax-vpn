use config::{Config, Environment, File};
use serde::Deserialize;

#[derive(Debug, Deserialize, Clone)]
pub struct Settings {
    #[serde(default = "default_control_port")]
    pub control_port: u16,

    pub control_plane_url: String,

    pub control_plane_token: String,

    pub server_id: String,

    #[serde(default = "default_wg_interface")]
    pub wireguard_interface: String,

    #[serde(default = "default_wg_port")]
    pub wireguard_port: u16,

    #[serde(default = "default_peer_cidr")]
    pub peer_cidr: String,

    #[serde(default = "default_dns")]
    pub dns_server: String,

    #[serde(default = "default_log_level")]
    pub log_level: String,

    #[serde(default = "default_environment")]
    pub environment: String,
}

impl Settings {
    pub fn load() -> Result<Self, config::ConfigError> {
        let cfg = Config::builder()
            .add_source(File::with_name("engine").required(false))
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
