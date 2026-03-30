# Ergonomia API 🚀

Backend enterprise para web de afiliados de productos ergonómicos.

**Stack:** Go + PostgreSQL + JWT + Docker

[![CI](https://github.com/Gentleman-Programming/ergonomia-api/actions/workflows/ci.yml/badge.svg)](https://github.com/Gentleman-Programming/ergonomia-api/actions)

## ⚡ Quick Start (5 minutos)

### 1. Clonar y entrar al directorio

```bash
git clone <repo>
cd ergonomia-api
```

### 2. Iniciar con Docker Compose

```bash
docker-compose up --build
```

Esto levanta:
- **PostgreSQL 16** en puerto `5433`
- **API Go** en puerto `8080`

### 3. Verificar que funciona

```bash
# Health check
curl http://localhost:8080/health

# Listar productos
curl http://localhost:8080/api/v1/products
```

### 4. Poblar datos de prueba

```bash
# En Windows PowerShell
.\scripts\seed.ps1

# En Linux/Mac
make seed
```

Usuarios de prueba:
- **Admin:** `admin@ergonomia.pro` / `password`
- **Usuario:** `test@test.com` / `password`

### 5. Test completo de la API

```bash
# En Windows PowerShell
.\scripts\test-api.ps1

# O manualmente:
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password"}'
```

---

## 📁 Estructura del Proyecto

```
ergonomia-api/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── config/                  # Configuración (.env)
│   ├── handlers/                # HTTP handlers (REST + HTMX)
│   ├── middleware/              # Auth, Rate Limit, CORS
│   ├── models/                  # Entidades (User, Product, etc.)
│   ├── store/                   # Repository pattern (PostgreSQL)
│   └── pkg/validator/           # Validaciones
├── pkg/auth/                    # JWT service
├── migrations/                  # SQL schema
├── scripts/                     # Seed y tests
├── .github/workflows/           # CI/CD
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── openapi.yaml                 # API Documentation
```

---

## 🔌 API Endpoints

### Auth
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
```

### Products (público)
```
GET  /api/v1/products          # Listar (con filtros)
GET  /api/v1/products/{slug}   # Detalle
```

### User (requiere JWT)
```
GET    /api/v1/me
GET    /api/v1/favorites
POST   /api/v1/favorites
DELETE /api/v1/favorites/{productID}
```

### Comparisons (requiere JWT)
```
GET  /api/v1/comparisons
POST /api/v1/comparisons
GET  /api/v1/comparisons/{slug}
```

### Admin (requiere JWT + Admin)
```
GET    /api/v1/admin/stats
POST   /api/v1/admin/products
PUT    /api/v1/admin/products/{id}
DELETE /api/v1/admin/products/{id}
```

### Health
```
GET /health
```

📖 **Documentación completa:** Ver `openapi.yaml`

---

## 🧪 Testing

```bash
# Tests unitarios
go test ./...

# Tests con coverage
make coverage

# Test de integración (requiere Docker)
make test

# Test rápido de API (PowerShell)
.\scripts\test-api.ps1
```

---

## 🚀 Deploy a Railway (Recomendado)

Railway es la forma más rápida y económica de deployar este backend.

### Paso 1: Crear cuenta

1. Ve a [railway.app](https://railway.app)
2. Regístrate con GitHub
3. Instala la CLI (opcional pero recomendado):
   ```bash
   npm install -g @railway/cli
   ```

### Paso 2: Crear proyecto

```bash
# Login
railway login

# Inicializar proyecto
railway init

# Crear PostgreSQL
railway add --database postgres
```

### Paso 3: Configurar variables

```bash
railway variables

# Añadir:
JWT_SECRET=tu-secreto-super-seguro-minimo-32-caracteres
ENV=production
```

Para generar un secreto seguro:
```bash
openssl rand -base64 32
```

### Paso 4: Deploy

```bash
# Deploy manual
railway up

# O configura GitHub Actions (ya incluido en .github/workflows/)
# Solo necesitas añadir RAILWAY_TOKEN en GitHub Secrets
```

### Paso 5: Verificar deploy

```bash
# Obtener URL
railway domain

# Health check
curl https://tu-app.railway.app/health
```

---

## 🐳 Docker Commands

```bash
# Iniciar todo
docker-compose up --build

# En segundo plano
docker-compose up -d

# Ver logs
docker-compose logs -f api
docker-compose logs -f postgres

# Detener
docker-compose down

# Eliminar todo (incluyendo datos)
docker-compose down -v
```

---

## 📋 Makefile Commands

```bash
make build          # Compilar
make run            # Ejecutar local
make dev            # Docker compose up
make down           # Docker compose down
make test           # Ejecutar tests
make coverage       # Coverage report
make fmt            # Formatear código
make seed           # Poblar DB
make psql           # Acceder a PostgreSQL
```

---

## 🔐 Variables de Entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | postgres://... |
| `JWT_SECRET` | Secret para firmar tokens | required |
| `JWT_EXPIRATION_HOURS` | Expiración access token | 24 |
| `PORT` | Puerto del servidor | 8080 |
| `ENV` | Environment (dev/prod) | development |
| `ALLOWED_ORIGINS` | CORS origins | http://localhost:8080 |

---

## 📝 Scripts Disponibles

### PowerShell
```powershell
.\scripts\seed.ps1      # Crear usuarios de prueba
.\scripts\test-api.ps1  # Test completo de API
```

### Bash
```bash
./scripts/seed.sh
```

---

## 🎯 Roadmap

- [x] Auth JWT completo
- [x] CRUD Productos
- [x] Favoritos y Comparaciones
- [x] Admin dashboard
- [x] Rate limiting
- [x] Tests unitarios
- [x] CI/CD GitHub Actions
- [x] Docker + Compose
- [x] OpenAPI documentation
- [ ] Frontend HTMX
- [ ] Analytics tracking
- [ ] Email service

---

## 📄 Licencia

MIT - Gentleman Programming

---

## 🆘 Troubleshooting

### "connection refused" a PostgreSQL
```bash
# Verificar que está corriendo
docker-compose ps

# Ver logs
docker-compose logs postgres
```

### "JWT validation failed"
- Verificar que `JWT_SECRET` tiene al menos 32 caracteres
- Verificar que no expiró el token

### Errores de CORS
- Añadir tu dominio frontend a `ALLOWED_ORIGINS`

### Puerto ocupado
```bash
# Cambiar puerto en docker-compose.yml o .env
PORT=8081
```
