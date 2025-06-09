--file: backend/db/migrations/products.sql

CREATE TABLE IF NOT EXISTS products (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  purchase_date DATE NOT NULL,
  warranty_expires DATE NOT NULL
);