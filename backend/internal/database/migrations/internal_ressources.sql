--file: backend/db/migrations/internal_ressources.sql

CREATE TABLE IF NOT EXISTS internal_documents (
  id SERIAL PRIMARY KEY,
  product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  filename TEXT NOT NULL,
  url TEXT NOT NULL,
  type TEXT NOT NULL, -- "manual", "diagram", "repair_guide", etc.
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
