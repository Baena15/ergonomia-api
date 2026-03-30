#!/usr/bin/env pwsh
# Script de deploy automático a Railway
# Uso: .&deploy.ps1

param(
    [string]$ProjectName = "ergonomia-api"
)

$ErrorActionPreference = "Stop"
$green = "`e[32m"
$yellow = "`e[33m"
$red = "`e[31m"
$cyan = "`e[36m"
$reset = "`e[0m"

function Write-Step {
    param($Message)
    Write-Host "`n${cyan}▶ $Message${reset}" -ForegroundColor Cyan
}

function Write-Success {
    param($Message)
    Write-Host "${green}✅ $Message${reset}" -ForegroundColor Green
}

function Write-Error {
    param($Message)
    Write-Host "${red}❌ $Message${reset}" -ForegroundColor Red
}

function Write-Info {
    param($Message)
    Write-Host "${yellow}ℹ️  $Message${reset}" -ForegroundColor Yellow
}

# ============================================
# VERIFICACIONES INICIALES
# ============================================

Write-Host @"
🚀 ERGONOMIA API - DEPLOY AUTOMÁTICO A RAILWAY
=============================================
"@ -ForegroundColor Cyan

# Verificar que railway CLI está instalado
Write-Step "Verificando Railway CLI..."
$railwayVersion = railway --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Info "Instalando Railway CLI..."
    npm install -g @railway/cli
    if ($LASTEXITCODE -ne 0) {
        Write-Error "No se pudo instalar Railway CLI. Instala Node.js primero: https://nodejs.org"
        exit 1
    }
}
Write-Success "Railway CLI instalado"

# Verificar que estamos en el directorio correcto
if (-not (Test-Path "./Dockerfile")) {
    Write-Error "No se encontró Dockerfile. Ejecuta este script desde la carpeta ergonomia-api/"
    exit 1
}

# Verificar Git
Write-Step "Verificando Git..."
$gitStatus = git status --short 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Error "No es un repositorio Git. Inicializando..."
    git init
    git add .
    git commit -m "Initial commit"
}

# Commit cambios pendientes
$gitStatus = git status --short
if ($gitStatus) {
    Write-Info "Hay cambios sin commitear. Haciendo commit..."
    git add .
    git commit -m "Deploy: $(Get-Date -Format 'yyyy-MM-dd HH:mm')"
    Write-Success "Cambios commiteados"
} else {
    Write-Info "No hay cambios pendientes"
}

# ============================================
# LOGIN EN RAILWAY
# ============================================

Write-Step "Iniciando sesión en Railway..."
Write-Info "Se abrirá tu navegador. Haz click en 'Authorize' y vuelve aquí."
Start-Sleep -Seconds 2

railway login
if ($LASTEXITCODE -ne 0) {
    Write-Error "No se pudo iniciar sesión en Railway"
    exit 1
}
Write-Success "Login exitoso"

# ============================================
# CREAR PROYECTO
# ============================================

Write-Step "Creando proyecto '$ProjectName'..."

# Verificar si ya existe un proyecto
try {
    $projectInfo = railway status 2>$null
    if ($projectInfo -match "Project") {
        Write-Info "Ya existe un proyecto vinculado. Usando proyecto existente..."
    }
} catch {
    # Inicializar nuevo proyecto
    Write-Info "Inicializando nuevo proyecto..."
    
    # Crear proyecto vacío
    $null = railway init --name $ProjectName 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "No se pudo crear el proyecto. Intenta manualmente: railway init"
        exit 1
    }
}
Write-Success "Proyecto configurado"

# ============================================
# AÑADIR POSTGRESQL
# ============================================

Write-Step "Configurando PostgreSQL..."

# Verificar si ya existe una base de datos
try {
    $services = railway status 2>$null
    if ($services -match "postgres") {
        Write-Info "PostgreSQL ya configurado"
    } else {
        Write-Info "Añadiendo PostgreSQL (esto puede tardar un minuto)..."
        $null = railway add --database postgres 2>&1
        Start-Sleep -Seconds 30
    }
} catch {
    Write-Info "Añadiendo PostgreSQL..."
    $null = railway add --database postgres 2>&1
    Start-Sleep -Seconds 30
}
Write-Success "PostgreSQL configurado"

# ============================================
# CONFIGURAR VARIABLES DE ENTORNO
# ============================================

Write-Step "Configurando variables de entorno..."

# Generar JWT Secret seguro
$jwtSecret = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 43 | ForEach-Object { [char]$_ })
Write-Info "JWT Secret generado: $($jwtSecret.Substring(0, 10))..."

# Configurar variables
Write-Info "Estableciendo JWT_SECRET..."
railway variables set JWT_SECRET="$jwtSecret" 2>$null

Write-Info "Estableciendo ENV=production..."
railway variables set ENV="production" 2>$null

Write-Info "Estableciendo PORT=8080..."
railway variables set PORT="8080" 2>$null

Write-Success "Variables configuradas"

# ============================================
# DEPLOY
# ============================================

Write-Step "Iniciando deploy..."
Write-Info "Esto puede tardar 2-5 minutos. Espera..."
Write-Host ""

railway up

if ($LASTEXITCODE -ne 0) {
    Write-Error "El deploy falló. Revisa los logs arriba."
    exit 1
}

Write-Success "Deploy completado!"

# ============================================
# OBTENER URL
# ============================================

Write-Step "Obteniendo URL de la aplicación..."
Start-Sleep -Seconds 5

$domain = railway domain 2>$null
if (-not $domain) {
    # Intentar obtener desde el dashboard
    Write-Info "Obteniendo dominio desde Railway..."
    Start-Sleep -Seconds 5
    $domain = railway domain 2>$null
}

if ($domain) {
    $url = $domain.Trim()
    if (-not $url.StartsWith("http")) {
        $url = "https://$url"
    }
    
    Write-Success "URL obtenida: $url"
    
    # ============================================
    # VERIFICAR HEALTH CHECK
    # ============================================
    
    Write-Step "Verificando que la API esté funcionando..."
    Start-Sleep -Seconds 10
    
    try {
        $healthResponse = Invoke-RestMethod -Uri "$url/health" -Method GET -TimeoutSec 30
        Write-Success "Health check OK!"
        Write-Host ""
        Write-Host "Response:" -ForegroundColor Gray
        $healthResponse | ConvertTo-Json -Depth 2 | Write-Host -ForegroundColor Gray
    } catch {
        Write-Info "La app puede estar iniciando. Espera 30 segundos y prueba manualmente:"
        Write-Host "   curl $url/health" -ForegroundColor Cyan
    }
    
    # ============================================
    # RESUMEN FINAL
    # ============================================
    
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
    Write-Host "  🎉 DEPLOY COMPLETADO!" -ForegroundColor Green
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
    Write-Host ""
    Write-Host "  📍 URL de tu API: $url" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  🔗 Endpoints disponibles:" -ForegroundColor White
    Write-Host "     • Health:     $url/health" -ForegroundColor Gray
    Write-Host "     • Products:   $url/api/v1/products" -ForegroundColor Gray
    Write-Host "     • Auth:       $url/api/v1/auth/login" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  👤 Usuarios de prueba:" -ForegroundColor White
    Write-Host "     • test@test.com / password" -ForegroundColor Gray
    Write-Host "     • admin@ergonomia.pro / password" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  🛠️  Comandos útiles:" -ForegroundColor White
    Write-Host "     • Ver logs:   railway logs -f" -ForegroundColor Gray
    Write-Host "     • Abrir DB:   railway connect postgres" -ForegroundColor Gray
    Write-Host "     • Dashboard:  https://railway.app" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  😴 Para mantener la app despierta (GRATIS):" -ForegroundColor Yellow
    Write-Host "     Ve a https://uptimerobot.com y crea un monitor:" -ForegroundColor Gray
    Write-Host "     URL: $url/health" -ForegroundColor Gray
    Write-Host "     Intervalo: Every 5 minutes" -ForegroundColor Gray
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
    
    # Guardar URL en archivo
    $url | Out-File -FilePath "./.deploy-url.txt" -Encoding UTF8
    Write-Info "URL guardada en .deploy-url.txt"
    
} else {
    Write-Info "No se pudo obtener la URL automáticamente."
    Write-Info "Ve a https://railway.app y busca tu proyecto para ver la URL"
}

Write-Host ""
Write-Success "¡Proceso completado!"
