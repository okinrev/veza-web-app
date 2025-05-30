//file: backend/modules/chat_server/src/messages.rs

use serde::Deserialize;

#[derive(Debug, Deserialize)]
#[serde(tag = "type")]
pub enum WsInbound {
    #[serde(rename = "join")]
    Join {
        room: String,
    },

    #[serde(rename = "message")]
    Message {
        room: String,
        content: String,
    },

    #[serde(rename = "dm")]
    DirectMessage {
        to: i32,
        content: String,
    },

    #[serde(rename = "room_history")]
    RoomHistory {
        room: String,
        limit: Option<i64>,
    },

    #[serde(rename = "dm_history")]
    DmHistory {
        with: i32,
        limit: Option<i64>,
    }

}
