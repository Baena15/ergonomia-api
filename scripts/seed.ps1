# Script para poblar la base de datos con datos de ejemplo (PowerShell)

Write-Host "🌱 Seeding database..." -ForegroundColor Green

# Verificar que PostgreSQL está corriendo
$containerRunning = docker-compose ps | Select-String "postgres.*Up"

if (-not $containerRunning) {
    Write-Host "⚠️  PostgreSQL no está corriendo. Iniciando..." -ForegroundColor Yellow
    docker-compose up -d postgres
    Start-Sleep -Seconds 5
}

# Ejecutar seed SQL
$seedSQL = @"
-- Crear usuario admin
INSERT INTO users (email, password_hash, first_name, last_name, is_admin)
VALUES (
    'admin@ergonomia.pro',
    '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'Admin',
    'User',
    true
)
ON CONFLICT (email) DO NOTHING;

-- Crear usuario de prueba
INSERT INTO users (email, password_hash, first_name, last_name, is_admin)
VALUES (
    'test@test.com',
    '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
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
"@

$seedSQL | docker-compose exec -T postgres psql -U ergonomia -d ergonomia

Write-Host "✅ Seed completado!" -ForegroundColor Green
Write-Host ""
Write-Host "Usuarios de prueba:"
Write-Host "  Admin: admin@ergonomia.pro / password"
Write-Host "  User:  test@test.com / password"
