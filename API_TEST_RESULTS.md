# 🧪 Resultados de Pruebas - API Ergonomia

## URL Base
https://ergonomia-api-production.up.railway.app

---

## ✅ Tests Pasados

### 1. Health Check
```bash
GET /health
```
**Resultado:** ✅ OK

### 2. Registro de Usuario
```bash
POST /api/v1/auth/register
Body: {"email":"test@ejemplo.com","password":"password123","first_name":"Test","last_name":"Usuario"}
```
**Resultado:** ✅ Token JWT devuelto correctamente

### 3. Login
```bash
POST /api/v1/auth/login
Body: {"email":"test@ejemplo.com","password":"password123"}
```
**Resultado:** ✅ Token JWT devuelto correctamente

### 4. Endpoint Protegido (/me)
```bash
GET /api/v1/me
Headers: Authorization: Bearer {token}
```
**Resultado:** ✅ Devuelve datos del usuario autenticado

### 5. Favoritos (Protegido)
```bash
GET /api/v1/favorites
Headers: Authorization: Bearer {token}
```
**Resultado:** ✅ Lista vacía (usuario nuevo)

### 6. Listar Productos (Público)
```bash
GET /api/v1/products
```
**Resultado:** ✅ 4 productos disponibles

---

## 🔑 Credenciales de Prueba

| Campo | Valor |
|-------|-------|
| Email | `test@ejemplo.com` |
| Password | `password123` |
| Rol | Usuario normal (no admin) |

---

## 📋 Próximos Pasos Sugeridos

1. **Agregar producto a favoritos:**
   ```bash
   POST /api/v1/favorites
   Body: {"product_id": 1}
   ```

2. **Crear usuario admin:** (para acceso al panel de administración)

3. **Conectar frontend:** Actualizar URLs en el frontend para usar esta API
