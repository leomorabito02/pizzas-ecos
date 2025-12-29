# 🚀 Render + Netlify - Guía Rápida

## Resumen

Tu app de Pizzas ECOS deployada en dos servicios:
- **Backend (Go)**: Render → https://pizzas-ecos-backend.onrender.com/api
- **Frontend (HTML/CSS/JS)**: Netlify → https://tu-nombre.netlify.app

## Pasos Rápidos

### 1. Verificar que todo esté en GitHub
```bash
git status      # Debe estar clean
git push origin main
```

### 2. Desplegar Backend (Render)

| Paso | Detalles |
|------|----------|
| 1. | Ir a https://render.com |
| 2. | Hacer login con GitHub |
| 3. | Click "New Web Service" |
| 4. | Conectar repo `pizzas-ecos` |
| 5. | **Build**: `cd backend && go build -o pizzas-ecos` |
| 6. | **Start**: `./pizzas-ecos` |
| 7. | **Plan**: Free (o Starter $7/mo) |
| 8. | **Env vars**: PORT, credsJSON, SpreadsheetID |
| 9. | **Secret File**: venta-pizzas-ecos.json |
| 10. | Esperar 2-3 minutos |

**Resultado**: `https://pizzas-ecos-backend.onrender.com/api/data`

### 3. Desplegar Frontend (Netlify)

| Paso | Detalles |
|------|----------|
| 1. | Ir a https://netlify.com |
| 2. | Hacer login con GitHub |
| 3. | Click "Add new site" → "Import existing" |
| 4. | Seleccionar repo `pizzas-ecos` |
| 5. | **Publish dir**: `frontend` |
| 6. | Dejar Build command vacío |
| 7. | Deploy |
| 8. | Ir a "Build & deploy" → "Environment" |
| 9. | Agregar: `REACT_APP_API_URL=https://pizzas-ecos-backend.onrender.com/api` |
| 10. | Hacer "Trigger deploy" |

**Resultado**: `https://tu-nombre.netlify.app`

## URLs después del Deploy

```
Frontend: https://pizzas-ecos-123.netlify.app
Backend:  https://pizzas-ecos-backend.onrender.com
API:      https://pizzas-ecos-backend.onrender.com/api
```

## Cómo Actualizar

Cualquier cambio en GitHub:
1. `git commit` y `git push`
2. Render → auto-deploy backend
3. Netlify → auto-deploy frontend
4. ¡Listo!

## Troubleshooting Rápido

| Problema | Solución |
|----------|----------|
| "API not found" | Verificar `REACT_APP_API_URL` en Netlify |
| "Google Sheets error" | Verificar Secret File en Render |
| Backend tarda 30s | Normal en Plan Free (wake up) |
| Frontend se ve roto | Refresh (Ctrl+F5) o clear cache |

## Variables Importantes

**En Render:**
```
PORT=8080
credsJSON=venta-pizzas-ecos.json
SpreadsheetID=1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0
```

**En Netlify:**
```
REACT_APP_API_URL=https://pizzas-ecos-backend.onrender.com/api
```

## Archivos de Configuración

```
render.yaml          ← Configuración para Render
netlify.toml         ← Configuración para Netlify
DEPLOYMENT.md        ← Guía detallada
DEPLOY_CHECKLIST.md  ← Checklist paso a paso
```

## ¿Necesitas más ayuda?

- Guía detallada: Ver [DEPLOYMENT.md](./DEPLOYMENT.md)
- Checklist: Ver [DEPLOY_CHECKLIST.md](./DEPLOY_CHECKLIST.md)
- README: Ver [README.md](./README.md)

---

**¡Tu app está lista para producción!** 🍕🚀
