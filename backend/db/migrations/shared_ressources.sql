--file: backend/db/migrations/shared_ressources.sql

CREATE TABLE IF NOT EXISTS shared_resources (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  filename TEXT NOT NULL,
  url TEXT NOT NULL,
  type TEXT NOT NULL, -- "sample", "preset", "project", etc.
  tags TEXT[],
  uploader_id INT NOT NULL REFERENCES users(id),
  is_public BOOLEAN DEFAULT TRUE,
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
