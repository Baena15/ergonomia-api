#!/bin/bash
# Script de deploy automático a Railway (Linux/Mac)
# Uso: ./deploy.sh

set -e

PROJECT_NAME="ergonomia-api"
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

function print_step() {
    echo -e "${CYAN}\n▶ $1${NC}"
}

function print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

function print_error() {
    echo -e "${RED}❌ $1${NC}"
}

function print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Banner
echo -e "${CYAN}"
echo "🚀 ERGONOMIA API - DEPLOY AUTOMÁTICO A RAILWAY"
echo "============================================="
echo -e "${NC}"

# Verificar Railway CLI
print_step "Verificando Railway CLI..."
if ! command -v railway &> /dev/null; then
    print_info "Instalando Railway CLI..."
    npm install -g @railway/cli
fi
print_success "Railway CLI instalado"

# Verificar directorio
if [ ! -f "./Dockerfile" ]; then
    print_error "No se encontró Dockerfile. Ejecuta este script desde la carpeta ergonomia-api/"
    exit 1
fi

# Git
print_step "Verificando Git..."
if [ ! -d ".git" ]; then
    print_error "No es un repositorio Git. Inicializando..."
    git init
    git add .
    git commit -m "Initial commit"
fi

if [ -n "$(git status --short)" ]; then
    print_info "Haciendo commit de cambios pendientes..."
    git add .
    git commit -m "Deploy: $(date '+%Y-%m-%d %H:%M')"
    print_success "Cambios commiteados"
fi

# Login
print_step "Iniciando sesión en Railway..."
print_info "Se abrirá tu navegador. Autoriza y vuelve aquí."
sleep 2
railway login
print_success "Login exitoso"

# Crear proyecto
print_step "Creando proyecto '$PROJECT_NAME'..."
railway init --name "$PROJECT_NAME" || print_info "Usando proyecto existente"
print_success "Proyecto configurado"

# PostgreSQL
print_step "Configurando PostgreSQL..."
railway add --database postgres || print_info "PostgreSQL ya configurado"
sleep 30
print_success "PostgreSQL configurado"

# Variables de entorno
print_step "Configurando variables de entorno..."
JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
print_info "JWT Secret generado"

railway variables set JWT_SECRET="$JWT_SECRET"
railway variables set ENV="production"
railway variables set PORT="8080"
print_success "Variables configuradas"

# Deploy
print_step "Iniciando deploy..."
print_info "Esto puede tardar 2-5 minutos..."
echo ""
railway up
print_success "Deploy completado!"

# Obtener URL
print_step "Obteniendo URL..."
sleep 5
DOMAIN=$(railway domain 2>/dev/null || echo "")

if [ -n "$DOMAIN" ]; then
    URL="https://$DOMAIN"
    print_success "URL: $URL"
    
    # Health check
    print_step "Verificando API..."
    sleep 10
    
    if command -v curl &> /dev/null; then
        curl -s "$URL/health" && echo "" || print_info "Verifica manualmente: curl $URL/health"
    fi
    
    # Resumen
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  🎉 DEPLOY COMPLETADO!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "  📍 URL: ${CYAN}$URL${NC}"
    echo ""
    echo "  🔗 Endpoints:"
    echo "     • $URL/health"
    echo "     • $URL/api/v1/products"
    echo ""
    echo "  😴 Configura UptimeRobot para mantenerla despierta:"
    echo "     https://uptimerobot.com"
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    
    echo "$URL" > .deploy-url.txt
else
    print_info "Ve a https://railway.app para ver tu URL"
fi

print_success "¡Listo!"
