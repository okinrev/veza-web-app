-- Migration pour étendre la table products
-- file: migrations/extend_products.sql

-- Ajouter les nouvelles colonnes à la table products
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES categories(id);
ALTER TABLE products ADD COLUMN IF NOT EXISTS brand TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS model TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS price DECIMAL(10,2);
ALTER TABLE products ADD COLUMN IF NOT EXISTS warranty_months INTEGER DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS warranty_conditions TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS manufacturer_website TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS specifications TEXT DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active' CHECK (status IN ('active', 'discontinued', 'draft'));
ALTER TABLE products ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE products ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Créer la table categories
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Créer la table product_documents
CREATE TABLE IF NOT EXISTS product_documents (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    file_type TEXT NOT NULL CHECK (file_type IN ('manual', 'datasheet', 'warranty', 'image', 'other')),
    file_path TEXT NOT NULL,
    file_size BIGINT DEFAULT 0,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insérer quelques catégories par défaut
INSERT INTO categories (name, description) VALUES 
    ('Microphones', 'Microphones et accessoires'),
    ('Interfaces Audio', 'Cartes son et interfaces audio'),
    ('Casques', 'Casques et écouteurs'),
    ('Enceintes', 'Enceintes de monitoring et haut-parleurs'),
    ('Contrôleurs', 'Contrôleurs MIDI et surfaces de contrôle'),
    ('Tables de mixage', 'Consoles de mixage et tables'),
    ('Traitement', 'Égaliseurs, compresseurs et effets'),
    ('Câbles', 'Câbles et connectique'),
    ('Accessoires', 'Pieds, supports et autres accessoires')
ON CONFLICT (name) DO NOTHING;

-- Mettre à jour les produits existants avec des catégories appropriées
UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Microphones'),
    warranty_months = 24,
    status = 'active'
WHERE name ILIKE '%microphone%';

UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Interfaces Audio'),
    warranty_months = 24,
    status = 'active'
WHERE name ILIKE '%interface%';

UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Casques'),
    warranty_months = 12,
    status = 'active'
WHERE name ILIKE '%casque%';

UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Enceintes'),
    warranty_months = 24,
    status = 'active'
WHERE name ILIKE '%enceinte%';

UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Contrôleurs'),
    warranty_months = 12,
    status = 'active'
WHERE name ILIKE '%contrôleur%' OR name ILIKE '%midi%';

UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Tables de mixage'),
    warranty_months = 24,
    status = 'active'
WHERE name ILIKE '%table%' OR name ILIKE '%mixage%';

UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Traitement'),
    warranty_months = 24,
    status = 'active'
WHERE name ILIKE '%compresseur%' OR name ILIKE '%égaliseur%' OR name ILIKE '%préampli%';

UPDATE products SET 
    category_id = (SELECT id FROM categories WHERE name = 'Câbles'),
    warranty_months = 6,
    status = 'active'
WHERE name ILIKE '%câble%';

-- Créer des index pour améliorer les performances
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products(updated_at);
CREATE INDEX IF NOT EXISTS idx_product_documents_product_id ON product_documents(product_id);
CREATE INDEX IF NOT EXISTS idx_product_documents_file_type ON product_documents(file_type);

-- Trigger pour mettre à jour automatiquement updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Appliquer le trigger aux tables
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
CREATE TRIGGER update_categories_updated_at 
    BEFORE UPDATE ON categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();