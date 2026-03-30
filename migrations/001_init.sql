-- Migración inicial: Usuarios y Productos
-- Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabla de usuarios
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- Tabla de productos
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    subcategory VARCHAR(50),
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'EUR',
    rating DECIMAL(2, 1) CHECK (rating >= 0 AND rating <= 5),
    reviews INTEGER DEFAULT 0,
    image_url VARCHAR(500),
    pros TEXT[], -- Array de strings en PostgreSQL
    cons TEXT[],
    ideal_for TEXT[],
    affiliate_links JSONB, -- Almacenamos links como JSON
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_active ON products(is_active);

-- Tabla de favoritos
CREATE TABLE IF NOT EXISTS favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);

CREATE INDEX idx_favorites_user ON favorites(user_id);

-- Tabla de comparaciones
CREATE TABLE IF NOT EXISTS comparisons (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    slug VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    product_ids INTEGER[], -- Array de IDs de productos
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_comparations_user ON comparisons(user_id);
CREATE INDEX idx_comparations_slug ON comparisons(slug);

-- Función para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_comparisons_updated_at BEFORE UPDATE ON comparisons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insertar datos de ejemplo (productos de tu web actual)
INSERT INTO products (name, slug, description, category, subcategory, price, currency, rating, reviews, pros, cons, ideal_for, affiliate_links) VALUES
('Reposapiés Ergonómico Ajustable HUANUO', 'reposapies-huanuo-ajustable', 'Reposapiés ajustable en altura e inclinación con textura masaje. Mejora la circulación y reduce la presión en las piernas.', 'general', 'reposapies', 29.99, 'EUR', 4.5, 2847, 
 ARRAY['Ajustable en altura (10-17cm)', 'Superficie texturizada masaje', 'Ángulo de inclinación ajustable', 'Base antideslizante'],
 ARRAY['Plástico algo rígido al principio', 'No tiene función de calor'],
 ARRAY['oficina', 'trabajo-casa', 'piernas-cansadas'],
 '{"amazon": "https://www.amazon.es/s?k=HUANUO+reposapies+ajustable&tag=baena15-21"}'),

('SIHOO M57 Silla Ergonómica de Oficina', 'silla-sihoo-m57', 'Silla ergonómica con soporte lumbar dinámico, reposabrazos 3D ajustables y malla transpirable. Ideal para largas jornadas de trabajo.', 'espalda', 'sillas', 189.99, 'EUR', 4.7, 15234,
 ARRAY['Soporte lumbar ajustable dinámicamente', 'Reposabrazos 3D', 'Malla transpirable de alta calidad', 'Reposacabezas ajustable', 'Excelente relación calidad-precio'],
 ARRAY['Montaje puede llevar 30-45 minutos', 'Para usuarios hasta 100kg'],
 ARRAY['programadores', 'oficina', 'dolor-espalda', 'largas-jornadas'],
 '{"amazon": "https://www.amazon.es/s?k=SIHOO+M57+silla+ergonomica&tag=baena15-21"}'),

('Cojín Lumbar Ergonómico Dreamer Car', 'cojin-lumbar-dreamer', 'Cojín lumbar de espuma viscoelástica con funda lavable. Diseño ortopédico que se adapta a la curva natural de la espalda.', 'espalda', 'cojines-lumbares', 24.95, 'EUR', 4.4, 8932,
 ARRAY['Espuma viscoelástica de alta densidad', 'Diseño ergonómico probado', 'Funda transpirable y lavable', 'Correas ajustables', 'Portátil'],
 ARRAY['Puede ser demasiado firme al principio', 'Tamaño estándar, no XL'],
 ARRAY['dolor-espalda', 'oficina', 'coche', 'silla-rigida'],
 '{"amazon": "https://www.amazon.es/s?k=cojin+lumbar+ergonomico&tag=baena15-21"}'),

('Logitech ERGO K860 Teclado Ergonómico Split', 'teclado-logitech-ergo-k860', 'Teclado inalámbrico ergonómico con diseño split curvo y reposamanos integrado. Reduce la tensión en muñecas y antebrazos.', 'programadores', 'teclados', 119.99, 'EUR', 4.6, 3421,
 ARRAY['Diseño split curvo natural', 'Reposamanos acolchado integrado', 'Teclas de perfil bajo silenciosas', 'Conectividad multi-dispositivo', 'Batería de 2 años'],
 ARRAY['Precio elevado', 'Curva de aprendizaje de 1-2 semanas', 'No es mecánico'],
 ARRAY['programadores', 'tunnel-carpiano', 'escritura-larga'],
 '{"amazon": "https://www.amazon.es/s?k=Logitech+ERGO+K860&tag=baena15-21"}');
