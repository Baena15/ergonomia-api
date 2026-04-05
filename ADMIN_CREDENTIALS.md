# 👤 Credenciales de Administrador

> ⚠️ **IMPORTANTE:** Guarda este archivo en un lugar seguro. No lo compartas.

---

## Cuenta Administrador

| Campo | Valor |
|-------|-------|
| **Email** | `ismael.alfaromarin@icloud.com` |
| **Password** | `AdminPass123!` |
| **Nombre** | Ismael Alfaro |
| **Rol** | ✅ Administrador |
| **ID** | 2 |

---

## Endpoints de Admin Disponibles

Con esta cuenta puedes acceder a:

```bash
# Estadísticas del sistema
GET /api/v1/admin/stats

# Crear producto
POST /api/v1/admin/products

# Actualizar producto
PUT /api/v1/admin/products/{id}

# Eliminar producto
DELETE /api/v1/admin/products/{id}

# Listar todos los usuarios (si existe)
GET /api/v1/admin/users
```

---

## Ejemplo de Uso

```bash
# Login
POST /api/v1/auth/login
Body: {"email":"ismael.alfaromarin@icloud.com","password":"AdminPass123!"}

# Usar el token en headers:
Authorization: Bearer {token}
```

---

## Seguridad

- 🔒 Cambia tu contraseña regularmente
- 🚫 No compartas estas credenciales
- 📝 El acceso admin permite modificar/eliminar productos y ver datos de usuarios
