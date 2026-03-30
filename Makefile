.PHONY: build run test clean docker-up docker-down seed lint fmt deps coverage

# Variables
BINARY_NAME=api
DOCKER_COMPOSE=docker-compose

# Colores para output
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

# Build
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/api

# Run local (requiere PostgreSQL en 5433)
run: build
	@echo "🚀 Starting server..."
	./$(BINARY_NAME)

# Development (con Docker)
dev:
	@echo "$(GREEN)Starting development environment...$(NC)"
	$(DOCKER_COMPOSE) up --build

# Detener Docker
down:
	@echo "🛑 Stopping containers..."
	$(DOCKER_COMPOSE) down

# Seed database
seed:
	@echo "🌱 Seeding database..."
	@./scripts/seed.sh 2>/dev/null || pwsh -ExecutionPolicy Bypass -File ./scripts/seed.ps1

# Tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Tests rápidos (sin integración)
test-short:
	@echo "🧪 Running short tests..."
	go test -v -short ./...

# Tests con coverage
coverage:
	@echo "📊 Running tests with coverage..."
	go test -cover ./...

# Coverage report detallado
coverage-report:
	@echo "📊 Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Report generated: coverage.html"

# Lint
lint:
	@echo "🔍 Running linter..."
	golangci-lint run

# Formatear código
fmt:
	@echo "✨ Formatting code..."
	go fmt ./...

# Descargar dependencias
deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

# Verificar que todo compila
check:
	@echo "🔍 Checking if code compiles..."
	go build ./...

# Limpiar
 clean:
	@echo "🧹 Cleaning..."
	rm -f $(BINARY_NAME) coverage.out coverage.html
	go clean

# Ver logs de Postgres
logs-db:
	$(DOCKER_COMPOSE) logs -f postgres

# Ver logs de API
logs-api:
	$(DOCKER_COMPOSE) logs -f api

# Acceder a PostgreSQL
psql:
	$(DOCKER_COMPOSE) exec postgres psql -U ergonomia -d ergonomia

# Backup de la base de datos
backup:
	@echo "💾 Creating backup..."
	docker-compose exec postgres pg_dump -U ergonomia ergonomia > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Restaurar base de datos
restore:
	@echo "⚠️  Esto eliminará los datos actuales. ¿Continuar? [y/N]"
	@read confirm && [ \$\$confirm = "y" ] && docker-compose exec -T postgres psql -U ergonomia -d ergonomia < $(file) || echo "Cancelado"

# Reset completo (elimina volúmenes)
reset:
	@echo "$(YELLOW)⚠️  WARNING: Esto eliminará TODOS los datos$(NC)"
	@echo "¿Estás seguro? [y/N]"
	@read confirm && [ \$\$confirm = "y" ] && $(DOCKER_COMPOSE) down -v || echo "Cancelado"

# Help
help:
	@echo "$(GREEN)Comandos disponibles:$(NC)"
	@echo "  make build          - Compilar el proyecto"
	@echo "  make run            - Ejecutar localmente (requiere DB)"
	@echo "  make dev            - Iniciar con Docker Compose"
	@echo "  make down           - Detener containers"
	@echo "  make seed           - Poblar DB con datos de prueba"
	@echo "  make test           - Ejecutar tests"
	@echo "  make coverage       - Tests con cobertura"
	@echo "  make fmt            - Formatear código"
	@echo "  make lint           - Ejecutar linter"
	@echo "  make psql           - Acceder a PostgreSQL"
	@echo "  make logs-db        - Ver logs de PostgreSQL"
	@echo "  make logs-api       - Ver logs de la API"
	@echo "  make clean          - Limpiar binarios"
	@echo "  make reset          - ⚠️  Eliminar TODOS los datos"
