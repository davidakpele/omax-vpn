use thiserror::Error;

#[derive(Debug, Error)]
pub enum EngineError {
    #[error("configuration error: {0}")]
    Config(#[from] config::ConfigError),

    #[error("tunnel error: {0}")]
    Tunnel(String),

    #[error("network error: {0}")]
    Network(String),

    #[error("security error: {0}")]
    Security(String),

    #[error("session error: {0}")]
    Session(String),

    #[error("control-plane communication error: {0}")]
    ControlPlane(#[from] reqwest::Error),

    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    #[error("internal error: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, EngineError>;
