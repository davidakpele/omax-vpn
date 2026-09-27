use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize)]
pub struct AddPeerRequest {
    pub peer_id: String,
    pub public_key: String,
    pub assigned_ip: String,
}

#[derive(Debug, Serialize)]
pub struct AddPeerResponse {
    pub peer_id: String,
    pub assigned_ip: String,
}

#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: &'static str,
    pub service: &'static str,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub code: &'static str,
    pub message: String,
}
