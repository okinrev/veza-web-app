--file: backend/db/migrations/messages.sql

CREATE TABLE IF NOT EXISTS messages (
  id SERIAL PRIMARY KEY,
  from_user INTEGER REFERENCES users(id) ON DELETE CASCADE,
  to_user INTEGER REFERENCES users(id), -- NULL si message de salon
  room TEXT, -- NULL si message privé
  content TEXT NOT NULL,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);