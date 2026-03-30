# Deploy Manual - Si el script falla

## Opcion 1: Script simplificado (recomendado)

```powershell
.\deploy-simple.ps1
```

## Opcion 2: Comandos uno por uno

### Paso 1: Instalar Railway CLI
```powershell
npm install -g @railway/cli
```

### Paso 2: Login (abre navegador)
```powershell
railway login
```
**Haz click en "Authorize" en el navegador y vuelve**

### Paso 3: Crear proyecto
```powershell
railway init --name ergonomia-api
```
**Selecciona "Empty Project" con las flechas y Enter**

### Paso 4: Agregar PostgreSQL
```powershell
railway add --database postgres
```
**Espera 30 segundos...**

### Paso 5: Configurar variables
```powershell
# Generar JWT secret
$secret = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object { [char]$_ })

# Guardar variables
railway variables set JWT_SECRET="$secret"
railway variables set ENV="production"
railway variables set PORT="8080"
```

### Paso 6: Deploy
```powershell
railway up
```

### Paso 7: Obtener URL
```powershell
railway domain
```

---

## Verificar que funciona

```powershell
# Reemplaza con tu URL real
$url = "https://TU-URL-REAL.railway.app"

# Health check
Invoke-RestMethod -Uri "$url/health" -Method GET
```

---

## Problemas comunes

### "No se reconoce railway"
```powershell
# Agregar al PATH
$env:Path += ";$env:APPDATA\npm"
```

### "No se puede cargar el script"
```powershell
# Permitir scripts temporalmente
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
```

### "Error al crear proyecto"
```powershell
# Ir al dashboard web
Start-Process "https://railway.app"
# Crear proyecto manualmente, luego:
railway link
```
