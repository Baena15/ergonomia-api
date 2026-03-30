#!/bin/bash

# Script para poblar la base de datos con datos de ejemplo

echo "🌱 Seeding database..."

# Verificar que PostgreSQL está corriendo
if ! docker-compose ps | grep -q "postgres.*Up"; then
    echo "⚠️  PostgreSQL no está corriendo. Iniciando..."
    docker-compose up -d postgres
    sleep 5
fi

# Ejecutar seed SQL
docker-compose exec -T postgres psql -U ergonomia -d ergonomia << EOF

-- Crear usuario admin
INSERT INTO users (email, password_hash, first_name, last_name, is_admin)
VALUES (
    'admin@ergonomia.pro',
    '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: 'password'
    'Admin',
    'User',
    true
)
ON CONFLICT (email) DO NOTHING;

-- Crear usuario de prueba
INSERT INTO users (email, password_hash, first_name, last_name, is_admin)
VALUES (
    'test@test.com',
    '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: 'password'
    'Test',
    'User',
    false
)
ON CONFLICT (email) DO NOTHING;

-- Mostrar resultados
SELECT 'Usuarios creados:' as info;
SELECT id, email, first_name, is_admin FROM users;

SELECT 'Productos disponibles:' as info;
SELECT id, name, slug, price, category FROM products LIMIT 5;

EOF

echo "✅ Seed completado!"
echo ""
echo "Usuarios de prueba:"
echo "  Admin: admin@ergonomia.pro / password"
echo "  User:  test@test.com / password"
