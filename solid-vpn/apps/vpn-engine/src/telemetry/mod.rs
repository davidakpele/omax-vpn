/// Telemetry module — Prometheus metrics and health reporting.
use prometheus::{
    register_gauge, register_int_counter, register_int_gauge, Encoder, Gauge, IntCounter,
    IntGauge, TextEncoder,
};
use std::sync::OnceLock;
use tracing::warn;

/// Global metric handles — initialised once at startup.
pub struct Metrics {
    pub active_sessions: IntGauge,
    pub sessions_total: IntCounter,
    pub connection_errors_total: IntCounter,
    pub peer_count: IntGauge,
    pub uptime_seconds: Gauge,
}

static METRICS: OnceLock<Metrics> = OnceLock::new();

/// Initialise global Prometheus metrics.
/// Must be called once at startup before any metrics are recorded.
pub fn init() -> &'static Metrics {
    METRICS.get_or_init(|| Metrics {
        active_sessions: register_int_gauge!(
            "vpn_active_sessions",
            "Number of currently active VPN sessions"
        )
        .expect("failed to register vpn_active_sessions"),

        sessions_total: register_int_counter!(
            "vpn_sessions_total",
            "Total VPN sessions created since startup"
        )
        .expect("failed to register vpn_sessions_total"),

        connection_errors_total: register_int_counter!(
            "vpn_connection_errors_total",
            "Total VPN connection errors since startup"
        )
        .expect("failed to register vpn_connection_errors_total"),

        peer_count: register_int_gauge!(
            "vpn_peer_count",
            "Number of WireGuard peers currently configured"
        )
        .expect("failed to register vpn_peer_count"),

        uptime_seconds: register_gauge!(
            "vpn_engine_uptime_seconds",
            "Seconds since the VPN engine started"
        )
        .expect("failed to register vpn_engine_uptime_seconds"),
    })
}

/// Retrieve the global metrics handle.
pub fn get() -> &'static Metrics {
    METRICS.get().expect("metrics not initialised — call telemetry::init() at startup")
}

/// Render all registered Prometheus metrics as a text exposition string.
pub fn render() -> String {
    let encoder = TextEncoder::new();
    let metric_families = prometheus::gather();
    let mut buf = Vec::new();
    if let Err(e) = encoder.encode(&metric_families, &mut buf) {
        warn!(error = %e, "failed to encode metrics");
        return String::new();
    }
    String::from_utf8(buf).unwrap_or_default()
}
