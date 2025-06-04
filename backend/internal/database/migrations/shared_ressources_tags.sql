-- file: backend/db/migrations/shared_ressource_tags.sql

CREATE TABLE IF NOT EXISTS shared_ressource_tags (
  shared_ressource_id INTEGER REFERENCES shared_ressources(id) ON DELETE CASCADE,
  tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (shared_ressource_id, tag_id)
);