use sqlx::{query, query_as, PgPool, FromRow};
use serde::Deserialize;
use std::sync::Arc;
use crate::hub::common::ChatHub;
use crate::client::Client;
use serde_json::json;
use chrono::NaiveDateTime;

#[derive(Debug, FromRow)]
pub struct RoomMessage {
    pub id: i32,
    pub username: String,
    pub content: String,
    pub timestamp: Option<NaiveDateTime>,
    pub room: Option<String>,
}

pub async fn join_room(hub: &ChatHub, room: &str, user_id: i32) {
    let mut rooms = hub.rooms.write().await;
    let entry = rooms.entry(room.to_string()).or_default();

    if !entry.contains(&user_id) {
        entry.push(user_id);
        tracing::debug!(user_id, room, "✅ Ajout à la room en mémoire");
    } else {
        tracing::debug!(user_id, room, "⏩ Déjà membre de la room");
    }

    tracing::info!(room, user_id, "👥 Rejoint la room");
}

pub async fn broadcast_to_room(
    hub: &ChatHub,
    user_id: i32,
    username: &str,
    room: &str,
    msg: &str
) {
    let rec = query!(
        "INSERT INTO messages (from_user, room, content) VALUES ($1, $2, $3) RETURNING id, CURRENT_TIMESTAMP as timestamp",
        user_id,
        room,
        msg
    )
    .fetch_one(&hub.db)
    .await
    .unwrap();

    let clients = hub.clients.read().await;
    let rooms = hub.rooms.read().await;

    let payload = json!({
        "type": "message",
        "data": {
            "id": rec.id,
            "fromUser": user_id,
            "username": username,
            "content": msg,
            "timestamp": rec.timestamp,
            "room": room
        }
    });

    if let Some(user_ids) = rooms.get(room) {
        for id in user_ids {
            if let Some(client) = clients.get(id) {
                client.send_text(&payload.to_string());
            }
        }
    }

    tracing::info!(room, user_id, "📨 Message room enregistré et diffusé");
}


pub async fn fetch_room_history(hub: &ChatHub, room: &str, limit: i64) -> Vec<RoomMessage> {
    query_as!(
        RoomMessage,
        r#"
        SELECT m.id, u.username, m.content, m.timestamp, m.room
        FROM messages m
        JOIN users u ON u.id = m.from_user
        WHERE m.room = $1
        ORDER BY m.timestamp ASC
        LIMIT $2
        "#,
        room,
        limit
    )
    .fetch_all(&hub.db)
    .await
    .unwrap_or_default()
}

pub async fn room_exists(hub: &ChatHub, room: &str) -> bool {
    sqlx::query_scalar!(
        "SELECT EXISTS(SELECT 1 FROM rooms WHERE name = $1)",
        room
    )
    .fetch_one(&hub.db)
    .await
    .unwrap_or(Some(false))
    .unwrap_or(false)
}
