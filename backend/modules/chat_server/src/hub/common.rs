//file: backend/modules/chat_server/src/hub/common.rs

use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use sqlx::PgPool;

use crate::client::Client;

#[derive(Debug)]
pub struct ChatHub {
    pub db: PgPool,
    pub clients: RwLock<HashMap<i32, Client>>, // user_id → Client
    pub rooms: RwLock<HashMap<String, Vec<i32>>>, // room name → list of user_ids
}

impl ChatHub {
    pub fn new(db: PgPool) -> Arc<Self> {
        Arc::new(Self {
            db,
            clients: Default::default(),
            rooms: Default::default(),
        })
    }

    pub async fn register(&self, user_id: i32, client: Client) {
        tracing::info!(user_id, "👤 Enregistrement du client");
        self.clients.write().await.insert(user_id, client);
    }

    pub async fn unregister(&self, user_id: i32) {
        tracing::info!(user_id, "🚪 Déconnexion du client");
        self.clients.write().await.remove(&user_id);
    }
}
