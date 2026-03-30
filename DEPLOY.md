# Guía de Despliegue - Ergonomia API

## 🚀 Opciones de Deploy

### Opción 1: Railway (Recomendado para empezar)

**Ventajas:**
- Deploy automático desde GitHub
- PostgreSQL incluido
- SSL automático
- Precio: Gratis (con límites) o ~$5/mes

**Pasos:**

1. Crear cuenta en [Railway](https://railway.app)

2. Crear nuevo proyecto desde GitHub

3. Añadir PostgreSQL (New > Database > PostgreSQL)

4. Configurar variables de entorno:
   ```
   DATABASE_URL=${{Postgres.DATABASE_URL}}
   JWT_SECRET=tu-secreto-super-seguro-32-caracteres
   ENV=production
   PORT=8080
   ```

5. El deploy es automático en cada push a `main`

---

### Opción 2: Render

**Pasos:**

1. Crear cuenta en [Render](https://render.com)

2. New > Web Service
   - Conectar repo de GitHub
   - Runtime: Go
   - Build Command: `go build -o api ./cmd/api`
   - Start Command: `./api`

3. New > PostgreSQL
   - Copiar Internal Database URL

4. Configurar Environment Variables:
   ```
   DATABASE_URL=postgres://... (de tu DB en Render)
   JWT_SECRET=tu-secreto-super-seguro
   ENV=production
   ```

---

### Opción 3: VPS (Hetzner, DigitalOcean, etc.)

**Para máximo control y aprendizaje.**

#### Paso 1: Preparar servidor

```bash
# Conectar al servidor
ssh root@tu-ip

# Actualizar sistema
apt update && apt upgrade -y

# Instalar Docker
curl -fsSL https://get.docker.com | sh
usermod -aG docker $USER

# Instalar Docker Compose
apt install docker-compose-plugin
```

#### Paso 2: Configurar app

```bash
mkdir -p /opt/ergonomia-api
cd /opt/ergonomia-api

# Crear docker-compose.yml (ver archivo en repo)
# Crear .env
```

#### Paso 3: SSL con Caddy (recomendado)

```bash
# Instalar Caddy
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt install caddy
```

`Caddyfile`:
```
api.tudominio.com {
    reverse_proxy localhost:8080
}
```

```bash
systemctl reload caddy
```

#### Paso 4: Deploy

```bash
cd /opt/ergonomia-api
docker-compose pull
docker-compose up -d
```

#### Paso 5: Backup automático

```bash
# Cron job para backups diarios
0 2 * * * /opt/ergonomia-api/scripts/backup.sh
```

---

## 📋 Checklist Pre-Deploy

- [ ] Cambiar `JWT_SECRET` (mínimo 32 caracteres aleatorios)
- [ ] Configurar `ENV=production`
- [ ] Verificar que `CORS` permite el dominio frontend
- [ ] Ejecutar migraciones
- [ ] Crear usuario admin
- [ ] Configurar backups automáticos
- [ ] Configurar monitoreo (UptimeRobot, etc.)

---

## 🔒 Seguridad

### JWT Secret
```bash
# Generar secreto seguro
openssl rand -base64 32
```

### Firewall (VPS)
```bash
ufw allow 22
ufw allow 80
ufw allow 443
ufw enable
```

### Fail2Ban (VPS)
```bash
apt install fail2ban
systemctl enable fail2ban
```

---

## 📊 Monitoreo

### Health Check
```bash
curl https://api.tudominio.com/health
```

### Logs
```bash
# Docker
docker-compose logs -f api

# Systemd (si usas servicio nativo)
journalctl -u ergonomia-api -f
```

---

## 🔄 CI/CD con GitHub Actions

`.github/workflows/deploy.yml`:
```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: go test ./...

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Railway
        uses: railway/cli@v1
        with:
          railway_token: ${{ secrets.RAILWAY_TOKEN }}
```

---

## 💰 Estimación de Costos

| Servicio | Precio/mes | Notas |
|----------|------------|-------|
| Railway | $0-5 | Free tier disponible |
| Render | $0-7 | Free tier disponible |
| Hetzner CX11 | €4.51 | VPS básico |
| DigitalOcean | $6 | Droplet básico |
| Domain | $10/año | Namecheap/Cloudflare |

---

## 🆘 Troubleshooting

### "connection refused" a PostgreSQL
```bash
# Verificar que PostgreSQL está corriendo
docker-compose ps

# Ver logs
docker-compose logs postgres
```

### "JWT validation failed"
- Verificar que `JWT_SECRET` es el mismo en todos los servicios
- Verificar que el token no expiró

### CORS errors
- Añadir dominio frontend a `ALLOWED_ORIGINS`
- Verificar que el frontend envía `Content-Type: application/json`

---

## 📚 Recursos

- [Railway Docs](https://docs.railway.app/)
- [Render Docs](https://render.com/docs)
- [Caddy Docs](https://caddyserver.com/docs/)
- [Docker Compose](https://docs.docker.com/compose/)
