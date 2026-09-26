/// Session manager — owns the authoritative in-memory session state.
/// Phase 4 implementation.
use std::collections::HashMap;
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct Session {
    pub id: Uuid,
    pub peer_id: String,
    pub user_id: String,
    pub assigned_ip: String,
    pub started_at: DateTime<Utc>,
}

/// In-memory session store. Will be replaced with persistent storage in Phase 4.
#[derive(Default)]
pub struct SessionManager {
    sessions: HashMap<Uuid, Session>,
}

impl SessionManager {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn add(&mut self, session: Session) {
        self.sessions.insert(session.id, session);
    }

    pub fn remove(&mut self, id: &Uuid) -> Option<Session> {
        self.sessions.remove(id)
    }

    pub fn get(&self, id: &Uuid) -> Option<&Session> {
        self.sessions.get(id)
    }

    pub fn count(&self) -> usize {
        self.sessions.len()
    }
}
