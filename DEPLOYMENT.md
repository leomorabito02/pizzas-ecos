# Guía de Despliegue - Pizzas ECOS

## 🎯 Tu Setup: Render (Backend) + Netlify (Frontend)

```
GitHub Repo → Render.com (Backend Go) + Netlify.com (Frontend)
                ↓                          ↓
          Backend API                   Frontend SPA
         (tu-backend.onrender.com)   (tu-frontend.netlify.app)
                ↓
          Google Sheets
```

---

## 📋 Pasos de Despliegue

### **Paso 1: Preparar el repositorio**

```bash
# Asegúrate que todo esté commiteado
git add .
git commit -m "Preparar para despliegue en Render + Netlify"
git push origin main
```

**Archivos importantes:**
- ✅ `render.yaml` - Configuración para Render
- ✅ `netlify.toml` - Configuración para Netlify
- ✅ `backend/.env` - NO debe estar en GitHub (en .gitignore)
- ✅ `venta-pizzas-ecos.json` - NO debe estar en GitHub (en .gitignore)

---

## 🔧 Despliegue del Backend en Render

### 1. Crear proyecto en Render

```
1. Ir a https://render.com
2. Hacer login / Registrarse
3. Click en "New +"
4. Seleccionar "Web Service"
5. Conectar GitHub repo
```

### 2. Configurar el servicio

**Configuración básica:**
- **Name**: `pizzas-ecos-backend`
- **Region**: `Oregon` (Gratis en algunos casos)
- **Branch**: `main`
- **Runtime**: `Go`
- **Build Command**: `cd backend && go build -o pizzas-ecos`
- **Start Command**: `./pizzas-ecos`
- **Plan**: `Free` (tendrá sleep después de 15 min inactividad)

### 3. Variables de entorno

En "Environment", agregar:

```
PORT=8080
credsJSON=venta-pizzas-ecos.json
SpreadsheetID=1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0
```

### 4. Agregar credenciales de Google

**Opción A: Como archivo (Recomendado)**

1. En "Environment", agregar como "Secret File"
2. Filename: `venta-pizzas-ecos.json`
3. Contenido: Pega el contenido de tu archivo JSON

**Opción B: Como variable (Si el JSON es pequeño)**

1. Convertir JSON a una línea
2. Agregar como variable de entorno

### 5. Deploy

Click en "Create Web Service"

**Resultado:**
- Tu backend estará en: `https://pizzas-ecos-backend.onrender.com`
- API disponible en: `https://pizzas-ecos-backend.onrender.com/api`

---

## 🌐 Despliegue del Frontend en Netlify

### 1. Crear proyecto en Netlify

```
1. Ir a https://netlify.com
2. Hacer login / Registrarse
3. Click en "Add new site"
4. "Import an existing project"
5. Seleccionar GitHub
6. Conectar y autorizar
7. Seleccionar repo "pizzas-ecos"
```

### 2. Configurar build

**Build settings:**
- **Base directory**: (dejar vacío)
- **Build command**: (dejar vacío - no compilar)
- **Publish directory**: `frontend`

### 3. Variables de entorno

En "Site settings" → "Build & deploy" → "Environment":

```
REACT_APP_API_URL=https://pizzas-ecos-backend.onrender.com/api
```

### 4. Deploy

Netlify detectará cambios en GitHub y hará deploy automático.

**Resultado:**
- Tu frontend estará en: `https://tu-nombre.netlify.app`
- Conectará automáticamente al backend en Render

---

## ✅ Verificar que todo funcione

### 1. Backend en Render

```bash
# Desde terminal, o desde navegador:
curl https://pizzas-ecos-backend.onrender.com/api/data

# Deberías recibir un JSON con vendedores y datos
```

### 2. Frontend en Netlify

```
1. Abrir https://tu-nombre.netlify.app
2. Abrir Dev Tools (F12)
3. En Console, deberías ver:
   API Base URL: https://pizzas-ecos-backend.onrender.com/api
4. El formulario debería cargar datos
```

### 3. Test completo

1. Ir a estadísticas
2. Debería mostrar datos del Google Sheets
3. Intentar agregar una venta
4. Verificar que aparezca en Google Sheets

---

## 🔄 Flujo de desarrollo y deploy

### Para hacer cambios:

```bash
# 1. Cambios en el código
# 2. Commit local
git add .
git commit -m "Descripción del cambio"

# 3. Push a GitHub
git push origin main

# 4. Deploy automático
# - Netlify detecta cambio → Redeploy frontend
# - Render detecta cambio → Redeploy backend
```

---

## ⚙️ Configuración según ambiente

El código automáticamente detecta:

### En desarrollo local:
```javascript
getAPIBase() → http://localhost:8080/api
```

### En Netlify (producción):
```javascript
// Si está definida REACT_APP_API_URL:
getAPIBase() → https://pizzas-ecos-backend.onrender.com/api

// O auto-detecta:
getAPIBase() → https://tu-nombre.netlify.app:8080/api (fallará)
```

**Importante**: Netlify debe tener la variable de entorno para saber dónde está el backend.

---

## 🐛 Troubleshooting

### "API not found" / "Cannot connect to backend"

**Solución:**
1. Verificar que Render está corriendo (puede estar en sleep)
2. Visitar `https://pizzas-ecos-backend.onrender.com` en navegador
3. Ver que responda con datos
4. Verificar que `REACT_APP_API_URL` esté correcta en Netlify
5. En navegador F12 → Console → Ver `API Base URL:`

### "Google Sheets error" en Render

**Solución:**
1. Verificar que `venta-pizzas-ecos.json` está subido como Secret File
2. Verificar que `SpreadsheetID` es correcto
3. Verificar que la cuenta de servicio tiene acceso al Sheet
4. Ver logs en Render para más detalles

### "CORS error"

**Solución:**
- Backend ya tiene CORS habilitado
- Si persiste, ir a Render y reiniciar el servicio

### Frontend se ve bien pero no carga datos

**Solución:**
1. F12 → Network → Ver si `/api/data` retorna 200
2. F12 → Console → Ver errores
3. Verificar que `REACT_APP_API_URL` sea correcta
4. Verificar que backend en Render está activo

---

## 💡 Tips

### Render
- Plan Free duerme después de 15 min inactividad
- Primer request tarda ~30 segundos a despertar
- Subir a pago ($7/mes) para mantener siempre activo
- Ver logs en "Logs" tab de Render

### Netlify
- Auto-deploy con cada push a GitHub
- Subdominio gratis incluido
- Comprar dominio en Netlify o conectar el tuyo
- Environment variables en UI, no en código

### Seguridad
- Credenciales JSON no están en GitHub ✓
- Variables de entorno no están en JavaScript ✓
- CORS habilitado para desarrollo ✓
- Considerar autenticación para producción

---

## 📚 Próximos pasos

1. ✅ **Desplegar backend en Render**
2. ✅ **Desplegar frontend en Netlify**
3. ⏭️ **Agregar dominio personalizado**
4. ⏭️ **Configurar SSL (Automático en Netlify + Render)**
5. ⏭️ **Monitoreo y logs**
6. ⏭️ **Considerar autenticación de usuarios**

---

## 📞 URLs importantes

Después del deploy:

| Servicio | URL |
|----------|-----|
| GitHub | https://github.com/leomorabito02/pizzas-ecos |
| Render (Backend) | https://pizzas-ecos-backend.onrender.com |
| Netlify (Frontend) | https://tu-nombre.netlify.app |
| Google Sheets | [Tu Sheet](https://docs.google.com/spreadsheets/d/1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0) |

---

## Archivo de Configuración

Ver archivos:
- `render.yaml` - Configuración para Render
- `netlify.toml` - Configuración para Netlify
