//file: backend/modules/chat_server/src/auth.rs

use jsonwebtoken::{decode, DecodingKey, Validation, Algorithm, TokenData, errors::Error};
use serde::{Deserialize, Serialize};
use std::env;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub user_id: i32,
    pub username: String,
    pub exp: usize,
    pub iat: usize,
}

pub fn validate_token(token: &str) -> Result<TokenData<Claims>, Error> {
    let secret = env::var("JWT_SECRET").expect("JWT_SECRET manquant");
    let validation = Validation::new(Algorithm::HS256);
    decode::<Claims>(token, &DecodingKey::from_secret(secret.as_bytes()), &validation)
}
