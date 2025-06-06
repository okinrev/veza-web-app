//file: backend/modules/chat_server/src/client.rs

use tokio::sync::mpsc;
use tokio_tungstenite::tungstenite::protocol::Message;

#[derive(Debug)]
pub struct Client {
    pub user_id: i32,
    pub username: String,
    pub sender: mpsc::UnboundedSender<Message>,
}

impl Client {
    /// Envoie un message texte au client
    pub fn send_text(&self, text: &str) {
        let _ = self.sender.send(Message::Text(text.to_string()));
    }

    /// Envoie un message JSON sérialisé
    pub fn send_json<T: serde::Serialize>(&self, value: &T) {
        if let Ok(payload) = serde_json::to_string(value) {
            let _ = self.sender.send(Message::Text(payload));
        }
    }
}