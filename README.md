# 🍕 Pizzas ECOS - Sistema de Ventas

[![Go](https://img.shields.io/badge/Go-1.25-blue)](https://golang.org)
[![Frontend](https://img.shields.io/badge/Frontend-Vanilla%20JS-yellow)](https://developer.mozilla.org/en-US/docs/Web/JavaScript)
[![Database](https://img.shields.io/badge/Database-MySQL%208.0-orange)](https://www.mysql.com)
[![Deployment](https://img.shields.io/badge/Deployment-GCP%20Cloud%20Run-red)](https://cloud.google.com/run)
[![Docker](https://img.shields.io/badge/Docker-46.5MB-2496ED)](DOCKER.md)
[![Status](https://img.shields.io/badge/Status-Stable-brightgreen)](STATUS.md)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

Sistema profesional de gestión y registro de ventas con análisis en tiempo real, integración a Google Sheets y panel de administración.

> ⚡ **[Quick Start (5 min)](QUICK_START.md)** | 📊 **[Project Status](STATUS.md)** | 🐳 **[Docker Guide](DOCKER.md)** | 📚 **[Full Docs](docs/)**

## 🚀 Inicio rápido

> 👉 **[Ir a Quick Start (5 minutos)](QUICK_START.md)** para comenzar inmediatamente

O sigue el resumen abajo:

### Con Docker (recomendado)

```bash
git clone https://github.com/leomorabito02/pizzas-ecos.git
cd pizzas-ecos

# 1. Configurar variables de entorno
cp backend/.env.example backend/.env
# Editar .env con tus credenciales:
# - DATABASE_URL
# - credsJSON (Google Sheets)
# - SpreadsheetID

# 2. Construir imagen Docker
make docker-build

# 3. Iniciar servicios
make docker-up

# 4. Verificar que está corriendo
docker-compose ps

# La API estará en: http://localhost:8080
# Frontend: abre frontend/index.html en el navegador
```

**Comandos útiles:**
```bash
make docker-logs       # Ver logs en tiempo real
make docker-down       # Detener servicios
make docker-clean      # Limpiar todo
make help              # Ver todos los comandos disponibles
```

### Sin Docker

**Terminal 1 - Backend:**
```bash
cd backend
cp .env.example .env
nano .env  # Configurar
go run main.go
# API corriendo en http://localhost:8080
```

**Terminal 2 - Frontend:**
```bash
cd frontend
python3 -m http.server 3000
# o: npx http-server . -p 3000
```

Abre http://localhost:3000 en tu navegador.

---

## 📋 Características

- ✅ **Formulario de ventas** con selección de productos y clientes
- ✅ **Dashboard de estadísticas** en tiempo real
- ✅ **Panel de administración** para gestionar productos, vendedores y usuarios
- ✅ **Autenticación JWT** segura
- ✅ **Integración Google Sheets** para respaldos automáticos
- ✅ **API RESTful** robusta y bien documentada
- ✅ **Responsive design** para móvil y escritorio
- ✅ **Base de datos MySQL** con SSL/TLS
- ✅ **Deployment** en Google Cloud Run con CI/CD automático

---

## 📁 Estructura del Proyecto

```
pizzas-ecos/
├── backend/                    # Go API REST
│   ├── main.go                # Punto de entrada
│   ├── go.mod                 # Dependencias
│   ├── config/                # Configuración
│   ├── controllers/           # Controladores HTTP
│   ├── services/              # Lógica de negocio
│   ├── database/              # Queries y conexión DB
│   ├── models/                # Estructuras de datos
│   ├── middleware/            # Auth, CORS, Rate limiting
│   ├── routes/                # Definición de rutas
│   ├── validators/            # Validación de entrada
│   ├── logger/                # Sistema de logging
│   ├── Dockerfile             # Imagen Docker multi-stage
│   └── .env.example           # Plantilla de configuración
│
├── frontend/                   # HTML/CSS/JavaScript vanilla
│   ├── index.html             # Formulario de ventas
│   ├── admin.html             # Panel de administración
│   ├── estadisticas.html      # Dashboard de estadísticas
│   ├── components.html        # Componentes reutilizables
│   ├── js/
│   │   ├── api-service.js     # Comunicación con backend
│   │   ├── controllers.js     # Controladores de vistas
│   │   ├── models.js          # Modelos de datos
│   │   ├── form.js            # Validación de formularios
│   │   ├── ui-utils.js        # Utilidades de UI
│   │   └── env.js             # Configuración de entorno
│   └── css/
│       ├── styles.css         # Estilos globales
│       ├── admin.css          # Estilos del admin
│       ├── estadisticas.css   # Estilos del dashboard
│       ├── login.css          # Estilos del login
│       └── components.css     # Estilos de componentes
│
├── docker-compose.yml         # Configuración para desarrollo local
├── .github/
│   ├── workflows/
│   │   └── deploy-gcp.yml    # Pipeline CI/CD a Google Cloud Run
│   └── GITHUB_SECRETS_SETUP.md # Guía de configuración de secretos
│
├── DEPLOYMENT_GCP.md          # Guía detallada de despliegue en GCP
├── DOCKER.md                  # Guía completa de Docker y docker-compose
├── DEVELOPMENT.md             # Guía completa de desarrollo local
├── QUICK_REFERENCE.md         # Referencia rápida de comandos
└── README.md                  # Este archivo
```

---

## 🔧 Requisitos previos

### Para desarrollo local
- **Docker & Docker Compose** (recomendado)
  - O: **Go 1.21+**, **MySQL 8.0**, **Node.js 18+**
- **Git**
- **Navegador moderno** (Chrome, Firefox, Safari, Edge)

### Para desplegar en Google Cloud Run
- **Cuenta de Google Cloud**
- **gcloud CLI** instalado
- **GitHub** con repositorio configurado
- **Docker** (para construir imágenes)

---

## 🌐 URLs en desarrollo

| Servicio | URL |
|----------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| API Docs | http://localhost:8080/api/v1 |
| MySQL | localhost:3306 |
| pprof (debug) | http://localhost:8080/debug/pprof |

---

## 📚 Documentación completa

- **🐳 [Guía de Docker](DOCKER.md)** - Docker, docker-compose y deployment
- **🚀 [Guía de Despliegue en GCP](DEPLOYMENT_GCP.md)** - Desplegar a Google Cloud Run
- **🏗️ [Guía de Desarrollo Local](DEVELOPMENT.md)** - Setup completo y desarrollo
- **🧪 [Testing Guide](TESTING.md)** - Tests en backend y frontend
- **⚡ [Quick Reference](QUICK_REFERENCE.md)** - Comandos frecuentes
- **📋 [CI/CD Pipeline](CI_CD.md)** - GitHub Actions y automatización
- **🔄 [CI/CD Multi-Environment Setup](docs/GITHUB_CI_CD_SETUP.md)** - Setup completo de QA + Prod
- **⚡ [Pipeline Quick Start](docs/PIPELINE_QUICK_START.md)** - Guía rápida del pipeline

---

## 🔐 Variables de Entorno

Ver [backend/.env.example](backend/.env.example) para todas las variables disponibles:

```env
# Database
DATABASE_URL=mysql://user:pass@host/dbname?tls=true

# Authentication
JWT_SECRET=tu_secreto_muy_fuerte_aqui

# API
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://tudominio.com
PORT=8080

# Google Sheets (opcional)
SHEETS_SPREADSHEET_ID=tu_id_aqui
GOOGLE_CREDENTIALS_JSON={"type":"service_account",...}

# Debug
DEBUG=false
LOG_LEVEL=info
```

> **⚠️ IMPORTANTE**: No commitear `.env` con datos reales. Usar `.env.example` como plantilla.

---

## 📡 APIs principales

### Authentication
```bash
# Login
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Response
{"token":"eyJhbGc...","user":{"id":1,"username":"admin","role":"admin"}}
```

### Ventas
```bash
# Listar ventas
curl http://localhost:8080/api/v1/ventas \
  -H "Authorization: Bearer $TOKEN"

# Crear venta
curl -X POST http://localhost:8080/api/v1/ventas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vendedor_id": 1,
    "cliente_id": 2,
    "productos": [{"producto_id": 1, "cantidad": 2, "precio": 100}],
    "pagado": true,
    "entregado": true
  }'

# Obtener venta específica
curl http://localhost:8080/api/v1/ventas/123 \
  -H "Authorization: Bearer $TOKEN"
```

### Estadísticas
```bash
# Estadísticas procesadas (recomendado para dashboard)
curl http://localhost:8080/api/v1/estadisticas-sheet

# Response: {"resumen":{...},"vendedores":[...],"ventas":[...]}

# Datos crudos
curl http://localhost:8080/api/v1/estadisticas
```

### Productos
```bash
# Listar productos
curl http://localhost:8080/api/v1/productos

# Crear producto (requiere admin)
curl -X POST http://localhost:8080/api/v1/productos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nombre":"Pizza Margherita","precio":250,"descripcion":""}'
```

### Health Check
```bash
# Verificar que el API está vivo
curl http://localhost:8080/api/v1/data

# Response: {"status":200,"data":[...],"message":""}
```

Consulta [backend/routes/routes.go](backend/routes/routes.go) para todas las rutas disponibles.

---

## 🏗️ Arquitectura de la aplicación

```
┌─────────────────────────────────────────────────────────────────┐
│                     FRONTEND (JavaScript Vanilla)               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  index.html  │  │  admin.html  │  │estadisticas │          │
│  │  Formulario  │  │ Administración│  │ Dashboard   │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘          │
│         └──────────────────┼──────────────────┘               │
│                            │ HTTP API calls                    │
│                            ▼                                   │
├─────────────────────────────────────────────────────────────────┤
│                   BACKEND (Go 1.21)                            │
│  ┌──────────────────────────────────────────────────┐         │
│  │ Controllers → Services → Database Queries        │         │
│  │ Middleware: JWT Auth, CORS, Rate Limit, Logger  │         │
│  └──────────────────────────────────────────────────┘         │
│                            │                                   │
│                            ▼                                   │
├─────────────────────────────────────────────────────────────────┤
│                    DATABASE (MySQL 8.0)                        │
│  • vendedores                                                  │
│  • clientes                                                    │
│  • productos                                                   │
│  • ventas_detalles                                             │
│  • usuarios                                                    │
├─────────────────────────────────────────────────────────────────┤
│            GOOGLE SHEETS API (Respaldos y Analytics)           │
│  • Sincronización automática                                   │
│  • Cálculos y fórmulas en tiempo real                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔐 Autenticación y Seguridad

- **JWT Tokens**: Tokens seguros en `sessionStorage`
- **Hashing de contraseñas**: bcrypt con salt
- **HTTPS/TLS**: Conexión encriptada a base de datos
- **CORS**: Solo orígenes permitidos
- **Rate Limiting**: Protección contra ataques
- **SQL Injection**: Prepared statements
- **Non-root user**: Contenedor Docker sin privilegios root

---

## 🧪 Testing

### Backend
```bash
cd backend

# Todos los tests
go test ./...

# Con coverage
go test -cover ./...

# Generar reporte HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend
- Chrome DevTools (F12)
- Consola de JavaScript
- Network tab para ver peticiones API

---

## 📊 Monitoreo

### Logs en desarrollo
```bash
# Backend logs
docker-compose logs -f backend

# Ver errores específicos
docker-compose logs backend | grep ERROR
```

### Performance
```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap
```

---

## 🚀 Despliegue en producción

### Google Cloud Run (recomendado)
1. **Setup inicial**: Seguir [DEPLOYMENT_GCP.md](DEPLOYMENT_GCP.md)
2. **Variables de entorno**: Configurar en Cloud Run UI o gcloud CLI
3. **CI/CD automático**: GitHub Actions disparado en cada push
4. **Scaling automático**: Cloud Run escala según demanda

### Costos estimados
- Cloud Run: ~$0.000024/request → ~$24/10M requests/mes
- Container Registry: ~$0.10/GB almacenado
- Cloud SQL: Varía según configuración
- **Primeros 2M requests/mes**: GRATIS

---

## 🐛 Troubleshooting

| Problema | Solución |
|----------|----------|
| "Cannot connect to database" | Verificar DATABASE_URL y conectividad a MySQL |
| "CORS error" | Agregar origen a CORS_ALLOWED_ORIGINS en .env |
| "JWT token expired" | Hacer login nuevamente (token guardado en sessionStorage) |
| "Port already in use" | Cambiar puerto en docker-compose.yml o matar proceso |
| "Build failed en GitHub Actions" | Ver logs: `gh run view RUN_ID --log` |

Más en [DEVELOPMENT.md](DEVELOPMENT.md#troubleshooting)

---

## 📞 Soporte y Contribuciones

- **Issues**: [GitHub Issues](https://github.com/leomorabito02/pizzas-ecos/issues)
- **Email**: leonardo.morabito@example.com
- **Documentación**: Ver archivos `.md` en la raíz del proyecto

---

## 📄 Licencia

Este proyecto está licenciado bajo la Licencia MIT. Ver [LICENSE](LICENSE) para más detalles.

---

## 🎯 Roadmap

- [ ] Modo oscuro
- [ ] Exportar reportes a PDF
- [ ] Notificaciones por email
- [ ] App móvil nativa
- [ ] Integraciones con otros servicios
- [ ] Sistema de inventario

---

**Última actualización**: 2024  
**Versión**: 1.0.0  
**Mantenedor**: Leonardo Morabito  
**Stack**: Go + MySQL + Vanilla JS + Docker + GCP Cloud Run

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
