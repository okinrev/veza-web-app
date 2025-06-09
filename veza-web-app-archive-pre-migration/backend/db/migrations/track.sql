--file: backend/db/migrations/tracks.sql

CREATE TABLE tracks (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    filename TEXT NOT NULL,
    artist TEXT,
    duration_seconds INT,
    tags TEXT[],
    is_public BOOLEAN DEFAULT TRUE,
    uploader_id INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT now()
);
