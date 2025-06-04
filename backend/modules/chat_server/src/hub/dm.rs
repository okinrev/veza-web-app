//file: backend/modules/chat_server/src/hub/dm.rs

use sqlx::{query, query_as, PgPool, FromRow};
use serde::Deserialize;
use crate::hub::common::ChatHub;
use crate::client::Client;
use serde_json::json;
use chrono::NaiveDateTime;

#[derive(Debug, FromRow)]
pub struct DmMessage {
    pub id: i32,
    pub from_user: Option<i32>,
    pub username: String,
    pub content: String,
    pub timestamp: Option<NaiveDateTime>,
}

pub async fn send_dm(hub: &ChatHub, from_user: i32, to_user: i32, username: &str, content: &str) {
    let rec = query!(
        "INSERT INTO messages (from_user, to_user, content) VALUES ($1, $2, $3) RETURNING id, CURRENT_TIMESTAMP as timestamp",
        from_user,
        to_user,
        content
    )
    .fetch_one(&hub.db)
    .await
    .unwrap();

    let clients = hub.clients.read().await;
    if let Some(client) = clients.get(&to_user) {
        let payload = json!({
            "type": "dm",
            "data": {
                "id": rec.id,
                "fromUser": from_user,
                "username": username,
                "content": content,
                "timestamp": rec.timestamp
            }
        });
        client.send_text(&payload.to_string());
    }

    tracing::info!(from_user, to_user, "📨 DM envoyé et enregistré");
}

pub async fn fetch_dm_history(hub: &ChatHub, user_id: i32, with: i32, limit: i64) -> Vec<DmMessage> {
    query_as!(
        DmMessage,
        r#"
        SELECT m.id, u.username, m.from_user, m.content, m.timestamp
        FROM messages m
        JOIN users u ON u.id = m.from_user
        WHERE ((m.from_user = $1 AND m.to_user = $2)
            OR (m.from_user = $2 AND m.to_user = $1))
        ORDER BY m.timestamp ASC
        LIMIT $3
        "#,
        user_id,
        with,
        limit
    )
    .fetch_all(&hub.db)
    .await
    .unwrap_or_default()
}

pub async fn user_exists(hub: &ChatHub, user_id: i32) -> bool {
    sqlx::query_scalar!(
        "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)",
        user_id
    )
    .fetch_one(&hub.db)
    .await
    .unwrap_or(Some(false))
    .unwrap_or(false)
}
