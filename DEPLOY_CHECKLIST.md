# Checklist: Despliegue Render + Netlify

## Pre-Deploy (Local)

### Backend
- [ ] Código compila sin errores: `cd backend && go build`
- [ ] Variables de entorno en `.env`:
  - [ ] `credsJSON` apunta al archivo correcto
  - [ ] `SpreadsheetID` es correcto
- [ ] Credenciales JSON existe localmente: `venta-pizzas-ecos.json`
- [ ] Archivo `.env` está en `.gitignore` ✓
- [ ] Archivo JSON está en `.gitignore` ✓
- [ ] CORS está habilitado en `main.go` ✓
- [ ] API endpoints responden: `curl http://localhost:8080/api/data`

### Frontend
- [ ] Código está limpio (sin console.log de debug)
- [ ] `form.js` y `estadisticas.js` importan correctamente
- [ ] No hay rutas hardcodeadas a localhost
- [ ] Todos los archivos estáticos están en `frontend/`
- [ ] `index.html` y `estadisticas.html` existen

### Git
- [ ] Todo commiteado: `git status` (clean)
- [ ] Rama main está actualizada: `git log -1`
- [ ] Push a GitHub: `git push origin main`

---

## Deploy Backend en Render

### Cuenta Render
- [ ] Crear cuenta en https://render.com
- [ ] Conectar GitHub (autorizar)

### Nuevo Web Service
- [ ] Click "New Web Service"
- [ ] Seleccionar repo `pizzas-ecos`
- [ ] Seleccionar rama `main`

### Configuración
- [ ] Name: `pizzas-ecos-backend`
- [ ] Runtime: `Go`
- [ ] Region: `Oregon` (o la más cercana)
- [ ] Build Command: `cd backend && go build -o pizzas-ecos`
- [ ] Start Command: `./pizzas-ecos`
- [ ] Plan: `Free` (opcional: `Starter $7/mo` para evitar sleep)

### Environment Variables
En "Environment", agregar:
- [ ] `PORT=8080`
- [ ] `credsJSON=venta-pizzas-ecos.json`
- [ ] `SpreadsheetID=1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0`

### Secret Files
En "Environment", agregar como "Secret File":
- [ ] Filename: `venta-pizzas-ecos.json`
- [ ] File contents: (Pega el contenido del JSON)

### Deploy
- [ ] Click "Create Web Service"
- [ ] Esperar a que compile (2-3 minutos)
- [ ] Ver que status sea "Live"
- [ ] Anotar URL: `https://pizzas-ecos-backend.onrender.com`

### Verificación
- [ ] Visitar `https://pizzas-ecos-backend.onrender.com/api/data`
- [ ] Deberías ver un JSON con datos
- [ ] Si ves error, ir a Logs en Render para debugging

---

## Deploy Frontend en Netlify

### Cuenta Netlify
- [ ] Crear cuenta en https://netlify.com
- [ ] Conectar GitHub (autorizar)

### Nuevo Site
- [ ] Click "Add new site"
- [ ] "Import an existing project"
- [ ] Seleccionar GitHub
- [ ] Seleccionar repo `pizzas-ecos`

### Build Settings
- [ ] Base directory: (dejar vacío)
- [ ] Build command: (dejar vacío)
- [ ] Publish directory: `frontend`

### Deploy
- [ ] Click "Deploy site"
- [ ] Esperar deploy (30-60 segundos)
- [ ] Ver que status sea "Published"
- [ ] Anotar URL: `https://tu-nombre.netlify.app`

### Site Settings → Build & Deploy → Environment
Agregar variables:
- [ ] `REACT_APP_API_URL=https://pizzas-ecos-backend.onrender.com/api`

### Redeploy después de variable
- [ ] En Netlify, ir a "Deploys"
- [ ] Click "Trigger deploy" → "Deploy site"
- [ ] Esperar que termine

---

## Post-Deploy

### Verificación Backend
```bash
# Terminal
curl https://pizzas-ecos-backend.onrender.com/api/data
# Debería retornar JSON con vendedores
```

### Verificación Frontend
- [ ] Abrir `https://tu-nombre.netlify.app`
- [ ] Abrir DevTools (F12)
- [ ] En Console debería ver: `API Base URL: https://pizzas-ecos-backend.onrender.com/api`
- [ ] El formulario debería cargar vendedores
- [ ] Estadísticas debería mostrar datos

### Test Completo
- [ ] Ir a "Agregar Combos"
- [ ] Llenar formulario
- [ ] Agregar venta
- [ ] Verificar que aparezca en Google Sheets
- [ ] Ir a "Ver Estadísticas"
- [ ] Debería mostrar la venta nueva

---

## Troubleshooting

### Si Backend no funciona:

1. **En Render:**
   - [ ] Ir a Logs tab
   - [ ] Ver error específico
   - [ ] Verificar Secret File está bien
   - [ ] Verificar Build Command compiló sin error

2. **Si error sobre credenciales:**
   - [ ] Verificar que `venta-pizzas-ecos.json` está como Secret File
   - [ ] Verificar que `credsJSON=venta-pizzas-ecos.json` en variables
   - [ ] Hacer redeploy en Render

3. **Si error sobre SpreadsheetID:**
   - [ ] Copiar ID correcto de URL del Sheet
   - [ ] Actualizar en Render environment
   - [ ] Hacer redeploy

### Si Frontend no conecta a Backend:

1. **Verificar variable:**
   - [ ] En Netlify, ir a "Site settings" → "Build & deploy" → "Environment"
   - [ ] Confirmar `REACT_APP_API_URL` está correcta
   - [ ] Hacer "Trigger deploy"

2. **Si persiste:**
   - [ ] F12 Console → Ver `API Base URL:`
   - [ ] F12 Network → Ver si `/api/data` retorna 200
   - [ ] Verificar URL del backend en variable es correcta

### Si ves "Cannot read properties"

- Significa que backend no está respondiendo
- Probablemente esté en sleep (Plan Free)
- Clickear "Wake up" en Render, o cambiar a Paid plan

---

## Post-Launch

### Dominio Personalizado
- [ ] En Netlify: "Domain settings" → "Custom domain"
- [ ] Apuntar DNS al dominio
- [ ] SSL automático (Let's Encrypt)

### Monitoreo
- [ ] Render: Ver Logs regularmente
- [ ] Netlify: Ver Analytics
- [ ] Google Sheets: Verificar que se están guardando ventas

### Mantenimiento
- [ ] Hacer cambios y push a GitHub
- [ ] Deployments automáticos
- [ ] Monitorear logs por errores

---

## URLs Importantes (Después del Deploy)

| Servicio | URL |
|----------|-----|
| GitHub | https://github.com/leomorabito02/pizzas-ecos |
| Render Dashboard | https://dashboard.render.com |
| Render Backend | https://pizzas-ecos-backend.onrender.com |
| Netlify Dashboard | https://app.netlify.com |
| Netlify Frontend | https://tu-nombre.netlify.app |
| Google Sheets | [Tu Sheet](https://docs.google.com/spreadsheets/d/1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0) |

---

## Notas Finales

- ✅ Backend en Render (Plan Free tiene sleep)
- ✅ Frontend en Netlify (sin sleep)
- ✅ Ambos conectados vía variable `REACT_APP_API_URL`
- ✅ Auto-deploy con cada push a GitHub
- ✅ CORS habilitado en backend
- ✅ Credenciales seguras (Secret Files en Render)

**¡Listo para producción!** 🚀
