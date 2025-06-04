--file: backend/db/migrations/listings.sql

CREATE TABLE listings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    state TEXT NOT NULL,
    price INTEGER,
    exchange_for TEXT,
    images TEXT[],
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
