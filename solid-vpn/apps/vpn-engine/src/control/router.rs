use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Request, State},
    http::{HeaderMap, StatusCode},
    middleware::{self, Next},
    response::Response,
    routing::{delete, get, post},
    Router,
};
use tokio::sync::RwLock;

use crate::control::handlers::{add_peer_handler, health, metrics_handler, remove_peer_handler};

#[derive(Clone)]
pub struct AppState {
    pub wireguard_interface: String,
    pub token: String,
    pub peer_ips: Arc<RwLock<HashMap<String, String>>>,
}

pub async fn serve(state: AppState, port: u16) {
    let addr = SocketAddr::from(([0, 0, 0, 0], port));

    let app = Router::new()
        .route("/health", get(health))
        .route("/metrics", get(metrics_handler))
        .route("/peers", post(add_peer_handler))
        .route("/peers/{public_key}", delete(remove_peer_handler))
        .layer(middleware::from_fn_with_state(state.clone(), auth_middleware))
        .with_state(state);

    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("failed to bind control API port");

    tracing::info!(addr = %addr, "control API listening");

    axum::serve(listener, app)
        .await
        .expect("control API server error");
}

async fn auth_middleware(
    State(state): State<AppState>,
    headers: HeaderMap,
    req: Request,
    next: Next,
) -> Result<Response, StatusCode> {
    let provided = headers
        .get("authorization")
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "))
        .unwrap_or("");

    if !crate::security::auth::validate_token(provided, &state.token) {
        return Err(StatusCode::UNAUTHORIZED);
    }

    Ok(next.run(req).await)
}
