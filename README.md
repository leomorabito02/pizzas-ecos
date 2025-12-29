# 🍕 Pizzas ECOS - Sistema de Ventas

Sistema de registro de ventas con integración a Google Sheets.

## Estructura del Proyecto

```
pizzas-ecos/
├── backend/              # API en Go
│   ├── main.go
│   ├── go.mod
│   └── .env             # credsJSON, SpreadsheetID (no commitear)
│
├── frontend/            # Aplicación web
│   ├── index.html       # Formulario de ventas
│   ├── estadisticas.html # Dashboard de estadísticas
│   ├── form.js
│   ├── estadisticas.js
│   ├── styles.css
│   ├── config.js        # Configuración de API
│   └── .env.example
│
├── DEPLOYMENT.md        # Guía completa de despliegue
└── README.md
```

## Requisitos

- **Go 1.16+** (para el backend)
- **Google Sheets API** habilitada
- **Credentials JSON** de Google Cloud
- **Navegador moderno** para el frontend

## Instalación Local

### 1. Backend

```bash
cd backend

# Crear archivo .env con credenciales
# credsJSON=tu-archivo.json
# SpreadsheetID=tu-sheet-id

# Ejecutar
go run main.go
```

Backend estará en: `http://localhost:8080/api`

### 2. Frontend

```bash
cd frontend

# Con Python 3
python -m http.server 5000

# O con Node.js
npx http-server -p 5000
```

Frontend estará en: `http://localhost:5000`

## Características

✅ **Formulario de Ventas**
- Seleccionar vendedor y cliente
- Agregar múltiples combos (Muzza y Jamón)
- Seleccionar cantidad
- Método de pago
- Tipo de entrega

✅ **Dashboard de Estadísticas**
- Resumen de ventas
- Detalle por vendedor
- Lista completa de transacciones
- Editar venta (estado, pago, combos)

✅ **Integración Google Sheets**
- Sincronización en tiempo real
- Fórmulas automáticas de cálculo
- Historial completo

## Arquitectura

```
Frontend (localhost:5000) ←--API HTTP-→ Backend (localhost:8080) ←→ Google Sheets
```

**CORS habilitado**: El backend permite peticiones desde cualquier origen

## Endpoints API

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/data` | Vendedores, clientes, precios |
| POST | `/api/submit` | Guardar nueva venta |
| GET | `/api/estadisticas` | Todas las ventas (detalle) |
| GET | `/api/estadisticas-sheet` | Resumen y vendedores |
| POST | `/api/actualizar-venta` | Actualizar venta existente |

## Despliegue en Producción

**Ver [DEPLOYMENT.md](./DEPLOYMENT.md)** para instrucciones completas.

**Tu plan: Render (Backend) + Netlify (Frontend)**

### Quick Start Deploy:

#### Backend en Render
```
1. Ir a https://render.com
2. New Web Service
3. Conectar GitHub repo
4. Build: cd backend && go build -o pizzas-ecos
5. Start: ./pizzas-ecos
6. Agregar Secret File: venta-pizzas-ecos.json
7. Deploy!
```

#### Frontend en Netlify
```
1. Ir a https://netlify.com
2. New site from Git
3. Seleccionar repo pizzas-ecos
4. Publish dir: frontend
5. Environment: REACT_APP_API_URL=https://tu-backend.onrender.com/api
6. Deploy!
```

**Ver [DEPLOY_CHECKLIST.md](./DEPLOY_CHECKLIST.md)** para checklist paso a paso

## Variables de Entorno

### Backend (.env)
```
credsJSON=venta-pizzas-ecos.json
SpreadsheetID=1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0
PORT=8080
```

### Frontend (.env.local) - Opcional
```
REACT_APP_API_URL=http://localhost:8080/api
```

Si no está definida, usa automáticamente:
- `http://localhost:8080/api` en desarrollo local
- `http://{mismo-dominio}:8080/api` en producción

## Estructura de Google Sheets

### Sheet "Ventas"
Columnas B-P: ID, Vendedor, Cliente, Muzzas (C1-C3), Jamones (C1-C3), Pago, Estado, Tipo Entrega, Total

### Sheet "estadisticas"
- **C5-C6**: Totales (Muzzas, Jamones)
- **G5-G9**: Dinero (Pendiente, Efectivo, Transferencia, Total, Total+SinCobrar)
- **B24-I**: Vendedores (Nombre, Cantidad, Muzzas, Jamones, Sin Pagar, Pagado, Total)

## Troubleshooting

### Backend error: `Can't find credentials`
```bash
# Verificar archivo .env existe en backend/
cat backend/.env

# Verificar archivo de credenciales existe
ls venta-pizzas-ecos.json
```

### Frontend no se conecta al backend
1. Verificar backend está corriendo: `http://localhost:8080/api/data`
2. Ver console del navegador (F12) para ver URL que intenta
3. Revisar que CORS esté habilitado en backend ✓

### Google Sheets error
1. Verificar SpreadsheetID en `.env`
2. Verificar permisos del servicio account
3. Verificar sheets "Ventas" y "estadisticas" existen

## Stack Tecnológico

- **Backend**: Go, Google Sheets API v4
- **Frontend**: HTML5, CSS3, JavaScript Vanilla
- **Database**: Google Sheets
- **Deployment**: Vercel, Docker, Render

## Desarrollo

```bash
# Terminal 1: Backend
cd backend && go run main.go

# Terminal 2: Frontend
cd frontend && python -m http.server 5000

# Abrir navegador
http://localhost:5000
```

## Notas Importantes

⚠️ **No commitear**:
- `.env` (contiene credenciales)
- `venta-pizzas-ecos.json` (credenciales de Google)

✅ **Incluidos en .gitignore**

## Roadmap

- [ ] Autenticación de usuarios
- [ ] Múltiples espacios de trabajo
- [ ] Reportes PDF
- [ ] Integración de pagos
- [ ] App móvil

## Licencia

Privado - Pizzas ECOS

## Contacto

Leonardo Morabito - [GitHub](https://github.com/leomorabito02)
