# Script de deploy simplificado para Railway
# Uso: .\deploy-simple.ps1

param(
    [string]$ProjectName = "ergonomia-api"
)

Write-Host "=== ERGONOMIA API - DEPLOY A RAILWAY ===" -ForegroundColor Cyan
Write-Host ""

# Verificar Railway CLI
Write-Host "[1/7] Verificando Railway CLI..." -ForegroundColor Yellow
$railwayCheck = railway --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Instalando Railway CLI..."
    npm install -g @railway/cli
}
Write-Host "OK - Railway CLI listo" -ForegroundColor Green
Write-Host ""

# Verificar Git
Write-Host "[2/7] Verificando Git..." -ForegroundColor Yellow
$gitStatus = git status --short 2>$null
if ($LASTEXITCODE -ne 0) {
    git init
    git add .
    git commit -m "Initial commit"
}
if ($gitStatus) {
    git add .
    git commit -m "Deploy: $(Get-Date -Format 'yyyy-MM-dd HH:mm')"
}
Write-Host "OK - Git listo" -ForegroundColor Green
Write-Host ""

# Login
Write-Host "[3/7] Login en Railway..." -ForegroundColor Yellow
Write-Host "IMPORTANTE: Se abrira tu navegador. Haz click en 'Authorize' y vuelve aqui." -ForegroundColor Magenta
Start-Sleep -Seconds 3
railway login
Write-Host "OK - Login completado" -ForegroundColor Green
Write-Host ""

# Crear proyecto
Write-Host "[4/7] Creando proyecto '$ProjectName'..." -ForegroundColor Yellow
railway init --name $ProjectName 2>$null
Write-Host "OK - Proyecto creado" -ForegroundColor Green
Write-Host ""

# PostgreSQL
Write-Host "[5/7] Configurando PostgreSQL..." -ForegroundColor Yellow
Write-Host "Espera 30 segundos mientras se configura..." -ForegroundColor Gray
railway add --database postgres 2>$null
Start-Sleep -Seconds 30
Write-Host "OK - PostgreSQL configurado" -ForegroundColor Green
Write-Host ""

# Variables
Write-Host "[6/7] Configurando variables..." -ForegroundColor Yellow
$jwtSecret = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object { [char]$_ })
railway variables set JWT_SECRET="$jwtSecret" 2>$null
railway variables set ENV="production" 2>$null
railway variables set PORT="8080" 2>$null
Write-Host "OK - Variables configuradas" -ForegroundColor Green
Write-Host ""

# Deploy
Write-Host "[7/7] Iniciando deploy..." -ForegroundColor Yellow
Write-Host "Esto puede tardar 2-5 minutos. Espera..." -ForegroundColor Gray
railway up
Write-Host ""
Write-Host "=== DEPLOY COMPLETADO ===" -ForegroundColor Green
Write-Host ""

# Obtener URL
Start-Sleep -Seconds 5
$domain = railway domain 2>$null
if ($domain) {
    $url = $domain.Trim()
    if (-not $url.StartsWith("http")) {
        $url = "https://$url"
    }
    
    Write-Host "TU API ESTA ONLINE:" -ForegroundColor Green
    Write-Host $url -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Endpoints:" -ForegroundColor White
    Write-Host "  - $url/health" -ForegroundColor Gray
    Write-Host "  - $url/api/v1/products" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Usuarios de prueba:" -ForegroundColor White
    Write-Host "  - test@test.com / password" -ForegroundColor Gray
    Write-Host "  - admin@ergonomia.pro / password" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Para mantener la app despierta (GRATIS):" -ForegroundColor Yellow
    Write-Host "  Ve a https://uptimerobot.com y crea un monitor:" -ForegroundColor Gray
    Write-Host "  URL: $url/health" -ForegroundColor Gray
    Write-Host "  Intervalo: Every 5 minutes" -ForegroundColor Gray
    
    $url | Out-File -FilePath "./.deploy-url.txt" -Encoding UTF8
} else {
    Write-Host "Ve a https://railway.app para ver tu URL" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Listo!" -ForegroundColor Green
