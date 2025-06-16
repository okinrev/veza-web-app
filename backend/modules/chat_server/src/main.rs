//file: backend/modules/chat_server/src/main.rs

mod auth;
mod client;
mod messages;
mod hub {
    pub mod common;
    pub mod room;
    pub mod dm;
}

use std::env;
use std::net::SocketAddr;
use std::sync::Arc;
use serde_json::json;
use serde::Serialize;

use dotenvy::dotenv;
use futures_util::{SinkExt, StreamExt};
use http::{Response, StatusCode};
use sqlx::postgres::PgPoolOptions;
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::mpsc;
use tokio_tungstenite::{
    accept_hdr_async,
    tungstenite::{handshake::server::Request, protocol::Message},
};

use crate::auth::{validate_token, Claims};
use crate::client::Client;
use crate::hub::common::ChatHub;
use crate::hub::room::*;
use crate::hub::dm::*;
use crate::messages::WsInbound;

#[derive(Serialize)]
struct OutgoingMessage<'a, T> {
    r#type: &'a str,
    data: T,
}

fn make_json_message<T: Serialize>(typ: &str, payload: T) -> Message {
    let msg = OutgoingMessage { r#type: typ, data: payload };
    Message::Text(serde_json::to_string(&msg).unwrap())
}


#[tokio::main]
async fn main() {
    dotenv().ok();

    tracing_subscriber::fmt()
        .with_env_filter("chat_server=debug")
        .with_target(true)
        .init();

    tracing::info!("🟢 Démarrage du serveur WebSocket...");

    let addr = env::var("WS_BIND_ADDR").unwrap_or_else(|_| "127.0.0.1:9001".to_string());
    let db_url = env::var("DATABASE_URL").expect("DATABASE_URL manquant");

    let pool = PgPoolOptions::new()
        .max_connections(5)
        .connect(&db_url)
        .await
        .expect("❌ Connexion DB échouée");

    let hub = ChatHub::new(pool);

    tracing::info!("🔌 Serveur WebSocket lancé sur ws://{}", addr);

    let listener = TcpListener::bind(&addr).await.expect("❌ Bind échoué");

    while let Ok((stream, addr)) = listener.accept().await {
        let hub = hub.clone();
        tokio::spawn(async move {
            if let Err(e) = handle_connection(stream, addr, hub).await {
                tracing::error!("❌ Erreur de connexion WS : {}", e);
            }
        });
    }
}

async fn handle_connection(
    stream: TcpStream,
    addr: SocketAddr,
    hub: Arc<ChatHub>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    tracing::info!(%addr, "🔌 Connexion TCP entrante");

    let mut extracted_claims: Option<Claims> = None;

    let callback = |req: &Request, response: Response<()>| {
        if let Some(header) = req.headers().get("Authorization") {
            if let Ok(auth) = header.to_str() {
                let token = auth.strip_prefix("Bearer ").unwrap_or(auth);
                match validate_token(token) {
                    Ok(token_data) => {
                        tracing::info!(user_id = token_data.claims.user_id, "🔐 Authentification réussie");
                        extracted_claims = Some(token_data.claims.clone());
                        Ok(response)
                    }
                    Err(_) => {
                        tracing::warn!("🔐 JWT invalide");
                        Err(Response::builder()
                            .status(StatusCode::UNAUTHORIZED)
                            .body(Some("JWT invalide".to_string()))
                            .unwrap())
                    }
                }
            } else {
                tracing::warn!("🔐 En-tête Authorization invalide");
                Err(Response::builder()
                    .status(StatusCode::BAD_REQUEST)
                    .body(Some("Authorization mal formé".to_string()))
                    .unwrap())
            }
        } else {
            // 2. Check query param ?token=...
            if let Some(query) = req.uri().query() {
                if let Some(token) = query.strip_prefix("token=") {
                    match validate_token(token) {
                        Ok(token_data) => {
                            extracted_claims = Some(token_data.claims.clone());
                            tracing::info!(user_id = token_data.claims.user_id, "🔐 Auth via query OK");
                            return Ok(response);
                        }
                        Err(_) => {
                            tracing::warn!("❌ JWT (query) invalide");
                        }
                    }
                }
            }
    
            tracing::warn!("🔐 Auth manquante ou invalide (query + header)");
            Err(Response::builder()
                .status(StatusCode::UNAUTHORIZED)
                .body(Some("Authorization manquante".to_string()))
                .unwrap())
        }
    };

    let ws_stream = accept_hdr_async(stream, callback).await?;

    let claims = match extracted_claims {
        Some(c) => c,
        None => return Err("JWT absent après handshake".into()),
    };

    tracing::info!(user_id = claims.user_id, "✅ Connexion WS autorisée");

    let user_id = claims.user_id;
    let username = claims.username;
    let (ws_write, mut ws_read) = ws_stream.split();

    let (tx, mut rx) = mpsc::unbounded_channel::<Message>();
    let client = Client {
        user_id,
        username: username.clone(),
        sender: tx.clone(),
    };
    hub.register(user_id, client).await;

    let write_task = tokio::spawn(async move {
        let mut ws_write = ws_write;
        while let Some(msg) = rx.recv().await {
            if ws_write.send(msg).await.is_err() {
                break;
            }
        }
    });

    while let Some(result) = ws_read.next().await {
        match result {
            Ok(msg) if msg.is_text() => {
                let msg_text = msg.to_text().unwrap();
                tracing::info!(user_id = user_id, message = %msg_text, "📨 Message reçu du client");
                
                if let Ok(parsed) = serde_json::from_str::<WsInbound>(msg_text) {
                    tracing::info!(user_id = user_id, message_type = ?parsed, "🔍 Message parsé avec succès");
                    match parsed {
                        WsInbound::Join { room } => {
                            tracing::info!(user_id = user_id, room = %room, "🚪 Tentative de rejoindre salon");
                            if room_exists(&hub, &room).await {
                                join_room(&hub, &room, user_id).await;
                                tracing::info!(user_id = user_id, room = %room, "✅ Salon rejoint avec succès");
                                let _ = tx.send(make_json_message("join_ack", json!({
                                    "room": room,
                                    "status": "ok"
                                })));                                
                            } else {
                                tracing::warn!(user_id = user_id, room = %room, "❌ Salon inexistant");
                                let _ = tx.send(make_json_message("error", json!({"message": "Room inexistante."})));
                            }
                        }
                        WsInbound::Message { room, content } => {
                            tracing::info!(user_id = user_id, room = %room, content = %content, "💬 Message salon reçu"); 
                            if room_exists(&hub, &room).await {
                                broadcast_to_room(&hub, user_id, &username, &room, &content).await;
                                tracing::info!(user_id = user_id, room = %room, "✅ Message salon diffusé");
                                let _ = tx.send(make_json_message("message_sent", json!({
                                    "room": room,
                                    "status": "ok"
                                })));                                
                            } else {
                                tracing::warn!(user_id = user_id, room = %room, "❌ Salon inexistant pour message");
                                let _ = tx.send(make_json_message("error", json!({"message": "Impossible d'envoyer un message à une room inexistante."})));
                            }
                        }
                        WsInbound::DirectMessage { to, content } => {
                            tracing::info!(user_id = user_id, to = to, content = %content, "📨 Message privé reçu");
                            if user_exists(&hub, to).await {
                                send_dm(&hub, user_id, to, &username, &content).await;
                                tracing::info!(user_id = user_id, to = to, "✅ Message privé envoyé");
                                let _ = tx.send(make_json_message("dm_sent", json!({
                                    "to": to,
                                    "status": "ok"
                                })));                                                                
                            } else {
                                tracing::warn!(user_id = user_id, to = to, "❌ Utilisateur destinataire introuvable");
                                let _ = tx.send(make_json_message("error", json!({"message": "Utilisateur destinataire introuvable."})));
                            }
                        }
                        WsInbound::RoomHistory { room, limit } => {
                            let actual_limit = limit.unwrap_or(50);
                            tracing::info!(user_id = user_id, room = %room, limit = actual_limit, "📜 Demande historique salon");
                            if room_exists(&hub, &room).await {
                                let msgs = fetch_room_history(&hub, &room, actual_limit).await;
                                tracing::info!(user_id = user_id, room = %room, count = msgs.len(), "📜 Historique salon récupéré");
                                for message in msgs {
                                    let payload = json!({
                                        "id": message.id,
                                        "username": message.username,
                                        "content": message.content,
                                        "timestamp": message.timestamp
                                    });
                                    let _ = tx.send(Message::Text(payload.to_string()));
                                }                                
                            } else {
                                tracing::warn!(user_id = user_id, room = %room, "❌ Salon inconnu pour historique");
                                let _ = tx.send(make_json_message("error", json!({"message": "Room inconnue."})));
                            }
                        }
                        WsInbound::DmHistory { with, limit } => {
                            let actual_limit = limit.unwrap_or(50);
                            tracing::info!(user_id = user_id, with = with, limit = actual_limit, "📜 Demande historique DM");
                            if user_exists(&hub, with).await {
                                let msgs = fetch_dm_history(&hub, user_id, with, actual_limit).await;
                                tracing::info!(user_id = user_id, with = with, count = msgs.len(), "📜 Historique DM récupéré");
                                let payload: Vec<_> = msgs.into_iter().map(|message| {
                                    json!({
                                        "username": message.username,
                                        "fromUser": message.from_user,
                                        "content": message.content,
                                        "timestamp": message.timestamp
                                    })
                                }).collect();
                                
                                let _ = tx.send(make_json_message("dm_history", payload));                                                                
                            } else {
                                tracing::warn!(user_id = user_id, with = with, "❌ Utilisateur introuvable pour historique DM");
                                let _ = tx.send(make_json_message("error", json!({"message": "Utilisateur introuvable pour DM."})));
                            }
                        }
                    }
                }
            }
            Ok(_) => continue,
            Err(e) => {
                tracing::error!("❌ Erreur de message WS : {}", e);
                break;
            }
        }
    }

    write_task.abort();
    hub.unregister(user_id).await;
    tracing::info!(user_id, "🚪 Déconnexion de l'utilisateur");

    Ok(())
}
