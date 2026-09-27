use axum::{
    extract::{Path, State},
    http::StatusCode,
    Json,
};

use crate::control::router::AppState;
use crate::control::types::{AddPeerRequest, AddPeerResponse, ErrorResponse, HealthResponse};
use crate::network::firewall::{allow_peer, revoke_peer};
use crate::network::routing::{add_peer_route, remove_peer_route};
use crate::security::keys::is_valid_public_key;
use crate::telemetry;
use crate::tunnel::peer::{add_peer, remove_peer, PeerConfig};

pub async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "ok",
        service: "vpn-engine",
    })
}

pub async fn metrics_handler() -> (StatusCode, String) {
    (StatusCode::OK, telemetry::render())
}

pub async fn add_peer_handler(
    State(state): State<AppState>,
    Json(req): Json<AddPeerRequest>,
) -> Result<(StatusCode, Json<AddPeerResponse>), (StatusCode, Json<ErrorResponse>)> {
    if !is_valid_public_key(&req.public_key) {
        return Err((
            StatusCode::UNPROCESSABLE_ENTITY,
            Json(ErrorResponse {
                code: "INVALID_PUBLIC_KEY",
                message: format!("'{}' is not a valid WireGuard public key", req.public_key),
            }),
        ));
    }

    if req.assigned_ip.is_empty() {
        return Err((
            StatusCode::UNPROCESSABLE_ENTITY,
            Json(ErrorResponse {
                code: "INVALID_IP",
                message: "assigned_ip must not be empty".into(),
            }),
        ));
    }

    let peer = PeerConfig {
        public_key: req.public_key.clone(),
        assigned_ip: req.assigned_ip.clone(),
        peer_id: req.peer_id.clone(),
    };

    add_peer(&state.wireguard_interface, &peer).await;
    allow_peer(&req.assigned_ip).await;
    add_peer_route(&state.wireguard_interface, &req.assigned_ip).await;

    {
        let mut map = state.peer_ips.write().await;
        map.insert(req.public_key.clone(), req.assigned_ip.clone());
    }

    telemetry::get().peer_count.inc();

    tracing::info!(peer_id = %req.peer_id, assigned_ip = %req.assigned_ip, "peer added");

    Ok((
        StatusCode::CREATED,
        Json(AddPeerResponse {
            peer_id: req.peer_id,
            assigned_ip: req.assigned_ip,
        }),
    ))
}

pub async fn remove_peer_handler(
    State(state): State<AppState>,
    Path(public_key): Path<String>,
) -> Result<StatusCode, (StatusCode, Json<ErrorResponse>)> {
    if !is_valid_public_key(&public_key) {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                code: "INVALID_PUBLIC_KEY",
                message: format!("'{}' is not a valid WireGuard public key", public_key),
            }),
        ));
    }

    let assigned_ip = {
        let mut map = state.peer_ips.write().await;
        map.remove(&public_key)
    };

    remove_peer(&state.wireguard_interface, &public_key).await;

    if let Some(ip) = &assigned_ip {
        revoke_peer(ip).await;
        remove_peer_route(ip).await;
    }

    telemetry::get().peer_count.dec();

    tracing::info!(public_key = %public_key, "peer removed");

    Ok(StatusCode::NO_CONTENT)
}
