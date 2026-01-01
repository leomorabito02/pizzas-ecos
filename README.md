# 🍕 Pizzas ECOS - Visión General del Proyecto

## 📋 Tabla de Contenidos

1. [Descripción General](#descripción-general)
2. [Arquitectura](#arquitectura)
3. [Stack Tecnológico](#stack-tecnológico)
4. [Estructura del Proyecto](#estructura-del-proyecto)
5. [Características Principales](#características-principales)
6. [Flujos de Datos](#flujos-de-datos)
7. [Deployment](#deployment)
8. [Ambientes](#ambientes)

---

## 📖 Descripción General

**Pizzas ECOS** es un sistema profesional de gestión y registro de ventas para una pequeña/mediana empresa productora de pizzas. La aplicación permite:

- 📝 Registro en tiempo real de ventas
- 📊 Análisis y estadísticas detalladas
- 👥 Gestión de vendedores y clientes
- 🔐 Panel de administración con autenticación
- 📱 Interfaz responsive para móvil y escritorio
- ☁️ Deployment en Google Cloud Run

---

## 🏗️ Arquitectura

### Arquitectura General

```
┌─────────────────────────────────────────────────────────────┐
│                        USUARIO                              │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │          FRONTEND (Netlify)                          │  │
│  │  - Vanilla JavaScript                               │  │
│  │  - HTML5 + CSS3                                     │  │
│  │  - Responsive Design                               │  │
│  │  - Client-side rendering                           │  │
│  └───────────────────────────────────────────────────────┘  │
│                          ↕                                   │
│                    HTTP/HTTPS                              │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │          BACKEND API (GCP Cloud Run)                 │  │
│  │  - Go 1.25                                           │  │
│  │  - RESTful API v1                                    │  │
│  │  - JWT Authentication                               │  │
│  │  - Rate Limiting & DDoS Protection                  │  │
│  └───────────────────────────────────────────────────────┘  │
│                          ↕                                   │
│                    TCP Connection                          │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │           DATABASE (Aiven MySQL)                      │  │
│  │  - MySQL 8.0                                         │  │
│  │  - SSL/TLS Encryption                               │  │
│  │  - Automatic Backups                                │  │
│  │  - Production & QA instances                         │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### Capas de la Aplicación

#### 1. **Frontend (Vanilla JavaScript)**
- Aplicación Single Page (SPA)
- Sin frameworks pesados
- Componentes reutilizables
- Localización en localStorage/sessionStorage

#### 2. **Backend (Go)**
- Servidor HTTP puro (stdlib)
- Arquitectura por capas:
  - **Routes**: Definición de endpoints
  - **Controllers**: Lógica de HTTP
  - **Services**: Lógica de negocio
  - **Database**: Capa de acceso a datos
  - **Middleware**: Auth, CORS, Logging
  - **Models**: DTOs y estructuras

#### 3. **Base de Datos (MySQL)**
- Tablas principales:
  - `usuarios` (autenticación)
  - `vendedores` (gestión de vendedores)
  - `clientes` (información de clientes)
  - `productos` (catálogo)
  - `ventas` (transacciones)
  - `detalle_ventas` (items por venta)

---

## 💻 Stack Tecnológico

### Frontend
| Tecnología | Propósito |
|-----------|----------|
| **HTML5** | Estructura semántica |
| **CSS3** | Estilos y responsive design |
| **Vanilla JavaScript** | Lógica de cliente |
| **Fetch API** | Comunicación con backend |
| **localStorage** | Almacenamiento persistente |
| **sessionStorage** | Almacenamiento de sesión |

### Backend
| Tecnología | Propósito |
|-----------|----------|
| **Go 1.25** | Lenguaje servidor |
| **net/http** | Servidor HTTP |
| **database/sql** | Driver de BD |
| **github.com/go-sql-driver/mysql** | Conector MySQL |
| **github.com/golang-jwt/jwt** | JWT authentication |
| **crypto/bcrypt** | Hashing de contraseñas |

### Infraestructura
| Tecnología | Propósito |
|-----------|----------|
| **Docker** | Containerización |
| **GCP Cloud Run** | Hosting serverless |
| **Aiven MySQL** | Base de datos hosteada |
| **Docker Hub** | Registry privado |
| **Netlify** | Hosting frontend |
| **GitHub Actions** | CI/CD pipeline |

### Herramientas de Desarrollo
| Herramienta | Propósito |
|-----------|----------|
| **Make** | Automatización de tareas |
| **Docker Compose** | Desarrollo local |
| **Git** | Control de versiones |

---

## 📁 Estructura del Proyecto

```
pizzas-ecos/
│
├── 📄 README.md                    # Documentación principal
├── 📄 PROJECT_OVERVIEW.md          # Este archivo
├── 🔧 Makefile                     # Automatización de tareas
├── 🐳 docker-compose.yml           # Configuración local
├── 🐳 Dockerfile                   # Imagen backend
│
├── 📁 backend/                     # API REST (Go)
│   ├── main.go                     # Punto de entrada
│   ├── go.mod / go.sum             # Dependencias
│   ├── 📁 config/                  # Configuración
│   ├── 📁 controllers/             # Controladores HTTP
│   ├── 📁 services/                # Lógica de negocio
│   ├── 📁 database/                # Acceso a datos
│   ├── 📁 models/                  # Estructuras de datos
│   ├── 📁 middleware/              # Middleware HTTP
│   ├── 📁 routes/                  # Definición de rutas
│   ├── 📁 validators/              # Validadores
│   ├── 📁 errors/                  # Manejo de errores
│   ├── 📁 logger/                  # Logging
│   ├── 📁 security/                # DDoS, seguridad
│   ├── 📁 ratelimit/               # Rate limiting
│   └── 📁 httputil/                # Utilidades HTTP
│
├── 📁 frontend/                    # SPA (Vanilla JS)
│   ├── index.html                  # Página de ventas
│   ├── estadisticas.html           # Página de estadísticas
│   ├── admin.html                  # Panel de administración
│   ├── login.html                  # Página de login
│   │
│   ├── 📁 css/                     # Estilos
│   │   ├── styles.css              # Estilos generales
│   │   ├── estadisticas.css        # Estilos estadísticas
│   │   ├── admin.css               # Estilos admin
│   │   └── login.css               # Estilos login
│   │
│   ├── 📁 js/                      # Scripts
│   │   ├── api-service.js          # Cliente API
│   │   ├── controllers.js          # Controladores frontend
│   │   ├── models.js               # Modelos de datos
│   │   ├── form.js                 # Formulario de ventas
│   │   ├── estadisticas.js         # Lógica de estadísticas
│   │   ├── admin.js                # Lógica de admin
│   │   ├── ui-utils.js             # Utilidades UI
│   │   ├── env.js                  # Variables de entorno
│   │   └── backend-config.js       # Configuración backend
│   │
│   └── 📁 images/                  # Imágenes/iconos
│
├── 📁 .github/workflows/           # CI/CD
│   └── deploy-multi-env.yml        # Pipeline deployment
│
├── 📁 scripts/                     # Scripts útiles
│   ├── healthcheck.sh              # Verificación de salud
│   └── pre-commit.sh               # Pre-commit hooks
│
├── 🔐 .env                         # Variables de entorno (no commitear)
├── 🔐 .env.example                 # Plantilla de .env
└── .gitignore                      # Archivos ignorados por Git
```

---

## ✨ Características Principales

### 1. **Registro de Ventas** 📝
- Formulario intuitivo
- Selección de vendedor y cliente
- Múltiples productos por venta
- Cálculo automático de totales
- Métodos de pago: Efectivo / Transferencia
- Tipos de entrega: Delivery / Retiro
- Validación en tiempo real

### 2. **Estadísticas y Reportes** 📊
- Dashboard en tiempo real
- Filtros múltiples:
  - Por vendedor
  - Por tipo de entrega
  - Por estado de pago
  - Por estado de vendedor (con/sin ventas)
- Contador de productos vendidos por tipo
- Desglose de ingresos por método de pago
- Listado de deudores por vendedor
- Estado de cobranza por vendedor

### 3. **Panel de Administración** 🔐
- Gestión de usuarios (admin only)
- Gestión de productos
- Gestión de vendedores
- Visualización de últimas ventas
- Dashboard con estadísticas

### 4. **Autenticación y Seguridad** 🔒
- JWT tokens (sessionStorage)
- Hashing de contraseñas con bcrypt
- CORS configurado
- Rate limiting por IP
- Protección DDoS
- HTTPS en producción

### 5. **Interfaz Responsive** 📱
- Diseño mobile-first
- Breakpoints: Mobile (≤768px), Tablet (769-1024px), Desktop (≥1025px)
- Touch-friendly buttons
- Optimizado para datos móviles

---

## 🔄 Flujos de Datos

### Flujo de Login
```
Usuario → Login HTML → POST /auth/login → Backend
  ↓
Validar credenciales → Hash check → JWT token generado
  ↓
Token → sessionStorage → Redirect a index.html
  ↓
API llamadas incluyen JWT en header
```

### Flujo de Crear Venta
```
Usuario completa formulario
  ↓
Validación local (JavaScript)
  ↓
POST /ventas {vendedor, cliente, items, pago, entrega}
  ↓
Backend valida datos
  ↓
INSERT en BD (transacción)
  ↓
Response con venta creada → Toast "✅ Venta registrada"
  ↓
Actualizar tablas y estadísticas
```

### Flujo de Estadísticas
```
Click en "Estadísticas"
  ↓
GET /estadisticas-sheet
  ↓
Backend calcula:
  - Resumen (totales, dinero, estado)
  - Lista de vendedores con stats
  - Todas las ventas
  ↓
Frontend renderiza tablas y gráficos
  ↓
Filtros actualizan vistas localmente (sin llamadas API)
```

---

## 🚀 Deployment

### Ambientes

#### **LOCAL (Desarrollo)**
```
Frontend: http://localhost:3000
Backend: http://localhost:8080
Database: localhost (docker-compose)
```

#### **QA**
```
Frontend: https://qa-ecos-ventas-pizzas.netlify.app
Backend: https://pizzas-ecos-backend-qa.run.app
Database: Aiven MySQL (instancia QA)
```

#### **PROD**
```
Frontend: https://ecos-ventas-pizzas.netlify.app
Backend: https://pizzas-ecos-backend-prod.run.app
Database: Aiven MySQL (instancia PROD)
```

### Pipeline CI/CD

```
┌─────────────────────────────────────┐
│     Git Push a develop / main        │
└────────────────┬────────────────────┘
                 ↓
        ┌─────────────────┐
        │  GitHub Actions │
        └────────┬────────┘
                 ↓
    ┌────────────────────────────┐
    │ 1. Build & Push Docker Hub │
    └────────────┬───────────────┘
                 ↓
        ┌─────────────────┐
        │   Si develop    │
        └────────┬────────┘
                 ↓
    ┌────────────────────────────┐
    │ 2. Deploy Backend a QA      │
    │ 3. Deploy Frontend a QA     │
    └────────────┬───────────────┘
                 ↓
        ┌─────────────────┐
        │   Si main       │
        └────────┬────────┘
                 ↓
    ┌────────────────────────────┐
    │ 2. Deploy Backend a PROD    │
    │ 3. Deploy Frontend a PROD   │
    │ (Espera aprobación manual)  │
    └────────────────────────────┘
```

### Variables de Entorno

**Backend:**
```
DATABASE_URL          # URL de conexión MySQL
DATABASE_URL_QA       # URL para ambiente QA
JWT_SECRET            # Clave para firmar tokens
CORS_ALLOWED_ORIGINS  # Orígenes permitidos
ENV                   # local|qa|prod
DEBUG                 # true|false
```

**Frontend:**
```
BACKEND_URL           # URL de la API (auto-detectada)
NETLIFY_SITE_ID_QA    # ID del site QA en Netlify
NETLIFY_SITE_ID_PROD  # ID del site PROD en Netlify
```

---

## 📊 Modelo de Datos

### Tabla: usuarios
```sql
id (PK)
username (UNIQUE)
password (bcrypt hash)
email
role (admin|usuario)
created_at
```

### Tabla: vendedores
```sql
id (PK)
nombre (UNIQUE)
email
telefono
comision_porcentaje
activo
created_at
```

### Tabla: clientes
```sql
id (PK)
nombre
direccion
telefono
email
created_at
```

### Tabla: productos
```sql
id (PK)
tipo_pizza
descripcion
precio
activo
created_at
```

### Tabla: ventas
```sql
id (PK)
vendedor_id (FK)
cliente_id (FK)
total
estado (sin pagar|pagada|entregada|cancelada)
payment_method (efectivo|transferencia)
tipo_entrega (delivery|retiro)
created_at
updated_at
```

### Tabla: detalle_ventas
```sql
id (PK)
venta_id (FK)
producto_id (FK)
cantidad
precio_unitario
```

---

## 🔌 Endpoints API

### Autenticación
- `POST /auth/login` - Login
- `POST /auth/logout` - Logout

### Datos Generales
- `GET /data` - Vendedores, clientes, productos
- `GET /estadisticas-sheet` - Estadísticas completas

### Ventas
- `POST /ventas` - Crear venta
- `GET /ventas` - Listar ventas
- `PUT /ventas/:id` - Actualizar venta
- `DELETE /ventas/:id` - Cancelar venta

### Productos
- `GET /productos` - Listar
- `POST /productos` - Crear
- `PUT /productos/:id` - Actualizar
- `DELETE /productos/:id` - Eliminar

### Vendedores
- `GET /vendedores` - Listar
- `POST /vendedores` - Crear
- `PUT /vendedores/:id` - Actualizar
- `DELETE /vendedores/:id` - Eliminar

### Usuarios (Admin)
- `GET /usuarios` - Listar
- `POST /usuarios` - Crear
- `PUT /usuarios/:id` - Actualizar
- `DELETE /usuarios/:id` - Eliminar

---

## 🛠️ Comandos Útiles

```bash
# Desarrollo Local
make docker-build    # Compilar imagen
make docker-up       # Iniciar servicios
make docker-down     # Detener servicios
make docker-logs     # Ver logs

# Backend
make backend-build   # Compilar backend

# Frontend
python server.py     # Servidor local (puerto 5000)

# Testing
go test ./...        # Tests backend
```

---

## 📝 Notas de Desarrollo

### Convenciones
- **Nombres de variables**: camelCase en JS, snake_case en Go
- **Commits**: Mensajes en español, descriptivos
- **Branches**: `develop` para QA, `main` para PROD
- **PRs**: Requieren aprobación antes de merge

### Mejoras Futuras
- [ ] Exportar reportes a PDF/CSV
- [ ] Búsqueda avanzada en tablas
- [ ] Modo oscuro
- [ ] Notificaciones por email
- [ ] App móvil nativa
- [ ] Integración con sistemas de pago

---

## 📞 Contacto y Soporte

- **Organización**: ECOS de Esperanza
- **Desarrollador**: Leonardo Morabito
- **Instagram**: [@ecos.jovenesfybp](https://www.instagram.com/ecos.jovenesfybp/)

---

**Última actualización**: Enero 2026  
**Versión**: 1.0.0  
**Estado**: ✅ Producción
