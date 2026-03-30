# 🚀 Checklist de Deploy - Ergonomia API

## ✅ Paso 1: Prueba Local

### 1.1 Iniciar servicios
```bash
cd ergonomia-api
docker-compose up --build
```

**Verificación:** Debes ver:
```
✅ Connected to database
🚀 Server starting on http://0.0.0.0:8080
```

### 1.2 Health Check
```bash
curl http://localhost:8080/health
```

**Esperado:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2026-03-30T..."
}
```

### 1.3 Poblar datos
```bash
# PowerShell
.\scripts\seed.ps1

# Bash
make seed
```

### 1.4 Test de API
```bash
# PowerShell
.\scripts\test-api.ps1
```

**Esperado:** Todos los checks en verde (✅)

---

## ✅ Paso 2: Preparar para Producción

### 2.1 Generar JWT Secret
```bash
openssl rand -base64 32
```

**Guardar este valor** lo necesitarás en Railway.

### 2.2 Verificar código
```bash
make fmt        # Formatear
make test       # Tests pasando
go build ./...  # Compila sin errores
```

### 2.3 Commit a GitHub
```bash
git add .
git commit -m "feat: backend v1.0 ready for deploy"
git push origin main
```

---

## ✅ Paso 3: Deploy a Railway

### 3.1 Crear cuenta e instalar CLI
```bash
# Instalar Railway CLI
npm install -g @railway/cli

# Login
railway login
```

### 3.2 Crear proyecto
```bash
# En el directorio ergonomia-api
railway init

# Seleccionar:
# - "Empty Project"
# - Nombre: "ergonomia-api"
```

### 3.3 Añadir PostgreSQL
```bash
railway add --database postgres

# O desde el dashboard web:
# New > Database > Add PostgreSQL
```

### 3.4 Configurar Variables de Entorno

```bash
railway variables

# Añadir estas:
JWT_SECRET=<el-valor-generado-en-paso-2.1>
ENV=production
```

**Variables automáticas** (Railway las genera):
- `DATABASE_URL` ← Se crea automáticamente al añadir PostgreSQL
- `PORT` ← Railway asigna automáticamente

### 3.5 Deploy

**Opción A: Desde CLI**
```bash
railway up
```

**Opción B: GitHub Actions (recomendado)**

1. Obtener token:
```bash
railway token
```

2. En GitHub repo → Settings → Secrets and variables → Actions:
   - `RAILWAY_TOKEN`: El token del paso anterior
   - `RAILWAY_SERVICE_ID`: ID del servicio (ver en `railway status`)

3. Push a main activa deploy automático.

### 3.6 Verificar deploy
```bash
# Obtener URL
railway domain
# Ejemplo: https://ergonomia-api-production.up.railway.app

# Test
curl https://<tu-url>/health
```

---

## ✅ Paso 4: Configurar Dominio Personalizado (Opcional)

### 4.1 En Railway Dashboard
1. Ir a tu servicio
2. Settings → Domains
3. "Generate Domain" o "Custom Domain"

### 4.2 Si usas dominio propio
1. Añadir dominio en Railway
2. Configurar DNS (CNAME a Railway)
3. Esperar propagación (5-30 min)

---

## ✅ Paso 5: Verificación Final

### 5.1 Test endpoints principales
```bash
# Health
curl https://<tu-dominio>/health

# Listar productos
curl https://<tu-dominio>/api/v1/products

# Login
 curl -X POST https://<tu-dominio>/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password"}'
```

### 5.2 Verificar PostgreSQL
```bash
railway connect postgres

# En el prompt:
\dt                    # Ver tablas
SELECT * FROM users;   # Ver usuarios
\q                     # Salir
```

---

## 🔧 Troubleshooting

### "Build failed"
```bash
# Ver logs
railway logs

# Rebuild
railway up --build
```

### "Database connection refused"
- Verificar que PostgreSQL está "Deployed" en el dashboard
- Verificar que `DATABASE_URL` existe en Variables

### "JWT validation failed"
- Verificar que `JWT_SECRET` tiene 32+ caracteres
- Regenerar tokens después de cambiar secret

### "CORS errors" desde frontend
- Actualizar `ALLOWED_ORIGINS` con tu dominio frontend
- Redeploy: `railway up`

---

## 📊 Monitoreo

### Logs en tiempo real
```bash
railway logs -f
```

### Métricas
En Railway Dashboard → Metrics

### Uptime monitoring (recomendado)
1. Crear cuenta en [UptimeRobot](https://uptimerobot.com) (gratis)
2. Añadir monitor: `https://<tu-dominio>/health`
3. Intervalo: 5 minutos

---

## 💰 Costos Estimados (Railway)

| Recurso | Gratis | Starter ($5/mes) |
|---------|--------|------------------|
| API | 500h/mes | Ilimitado |
| PostgreSQL | 500h/mes | Ilimitado |
| Ancho de banda | 100GB/mes | 100GB/mes |
| **Total** | **$0** | **~$5/mes** |

**Nota:** El plan gratuito "duerme" la app después de inactividad (~15 min).
Para producción real, recomendado Starter.

---

## 🎉 Post-Deploy

### Crear usuario admin en producción
```bash
railway connect postgres

INSERT INTO users (email, password_hash, first_name, last_name, is_admin)
VALUES (
    'admin@tudominio.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'Admin',
    'User',
    true
);

\q
```

**Contraseña:** `password` (cambiar inmediatamente vía API)

### Documentar URL
- API: `https://api.tudominio.com`
- Health: `https://api.tudominio.com/health`
- GitHub: Enlace al repo

---

## ✅ Checklist Final

- [ ] Health check pasa local
- [ ] Tests pasan (`make test`)
- [ ] Commit push a GitHub
- [ ] Railway proyecto creado
- [ ] PostgreSQL añadido
- [ ] JWT_SECRET configurado
- [ ] Deploy exitoso
- [ ] Health check pasa en producción
- [ ] Dominio configurado (opcional)
- [ ] UptimeRobot configurado
- [ ] Usuario admin creado
- [ ] Documentación actualizada

**¡Listo para producción! 🚀**
