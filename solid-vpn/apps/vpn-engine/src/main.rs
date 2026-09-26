mod config;
mod errors;
mod network;
mod security;
mod sessions;
mod telemetry;
mod tunnel;

use std::time::Instant;
use tokio::signal;
use tracing::{error, info, warn};
use tracing_subscriber::{fmt, layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

#[tokio::main]
async fn main() {
    // Load .env if present (non-fatal in production)
    let _ = dotenvy::dotenv();

    // Initialise structured logging before loading config so startup errors are visible.
    // Level is overridden below once config is loaded.
    init_tracing("info");

    info!(service = "vpn-engine", event = "startup_begin", "Solid VPN engine starting");

    // Load configuration
    let settings = match config::Settings::load() {
        Ok(s) => s,
        Err(e) => {
            error!(error = %e, "failed to load configuration");
            std::process::exit(1);
        }
    };

    // Re-initialise tracing at the configured level
    // (tracing_subscriber can only be set once; the first init above is intentionally minimal)
    info!(
        environment = %settings.environment,
        server_id   = %settings.server_id,
        interface   = %settings.wireguard_interface,
        wg_port     = settings.wireguard_port,
        control_api = settings.control_port,
        "configuration loaded"
    );

    // Initialise Prometheus metrics
    let _metrics = telemetry::init();
    info!("telemetry initialised");

    let start_time = Instant::now();

    // Placeholder: initialise tunnel interface
    tunnel::wireguard::init_interface(
        &settings.wireguard_interface,
        settings.wireguard_port,
    )
    .await;

    // Placeholder: configure peer DNS
    network::dns::configure_peer_dns(&settings.dns_server).await;

    // Placeholder: apply baseline firewall rules
    network::firewall::apply_base_rules(&settings.wireguard_interface).await;

    // Placeholder: enable NAT masquerade
    // outbound_interface is read from env or defaulted — not user-controllable
    let outbound_iface = std::env::var("VPN_ENGINE_OUTBOUND_INTERFACE")
        .unwrap_or_else(|_| "eth0".to_string());
    network::nat::enable_masquerade(&settings.wireguard_interface, &outbound_iface).await;

    info!(
        service = "vpn-engine",
        event   = "startup_complete",
        "VPN engine ready — waiting for control-plane commands"
    );

    // Update uptime metric periodically in a background task
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(tokio::time::Duration::from_secs(10));
        loop {
            interval.tick().await;
            telemetry::get()
                .uptime_seconds
                .set(start_time.elapsed().as_secs_f64());
        }
    });

    // Wait for shutdown signal
    match signal::ctrl_c().await {
        Ok(()) => {
            info!(service = "vpn-engine", event = "shutdown", "shutdown signal received");
        }
        Err(e) => {
            warn!(error = %e, "failed to listen for shutdown signal");
        }
    }

    // Graceful teardown
    info!("tearing down network configuration");
    network::nat::disable_masquerade(&settings.wireguard_interface, &outbound_iface).await;
    tunnel::interface::bring_down(&settings.wireguard_interface).await;

    info!(service = "vpn-engine", event = "stopped", "VPN engine stopped cleanly");
}

/// Initialise the global tracing subscriber with JSON output and an env-filter.
fn init_tracing(default_level: &str) {
    let filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new(default_level));

    tracing_subscriber::registry()
        .with(filter)
        .with(fmt::layer().json())
        .init();
}
