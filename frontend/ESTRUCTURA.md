# 📁 Estructura del Frontend - ECOS

## Organización de Carpetas

```
frontend/
├── index.html                 # Página principal (Formulario de ventas)
├── serve.py                   # Server Python para desarrollo
├── server.py                  # Server alternativo
│
├── js/                        # JavaScript - Lógica de aplicación (MVC)
│   ├── api-service.js        # Capa de datos - Comunicación con API
│   ├── models.js             # DTOs y estado global de la aplicación
│   ├── ui-utils.js           # Utilidades compartidas para UI
│   ├── controllers.js        # Controladores (lógica de negocio)
│   ├── form.js               # Vista - Manejo del formulario principal
│   ├── config.js             # Configuración de la aplicación
│   ├── env.js                # Variables de entorno
│   └── estadisticas.js       # Lógica de estadísticas
│
├── css/                       # Estilos
│   ├── styles.css            # Estilos principales
│   ├── admin.css             # Estilos del panel admin
│   ├── login.css             # Estilos del login
│   └── estadisticas.css      # Estilos de estadísticas
│
├── views/                     # Páginas HTML
│   ├── login.html            # Página de autenticación
│   ├── admin.html            # Panel de administración
│   └── estadisticas.html     # Página de estadísticas
│
└── images/                    # Imágenes de la aplicación
    └── ecoslogo.jpeg
```

## Descripción de Archivos

### Raíz
- **index.html**: Página principal con formulario de creación de ventas

### Carpeta `js/` (MVC Architecture)

#### Capa de Datos
- **api-service.js**: Clase `APIService` que encapsula todas las llamadas HTTP a `/api/v1/*`
  - Manejo automático de tokens JWT
  - Métodos: login(), criarVenta(), obtenerVentas(), criarProducto(), etc.
  - Inyección de headers de autenticación

#### Modelos
- **models.js**: Clases de datos (DTOs) para la aplicación
  - `Producto`, `Vendedor`, `Cliente`, `ProductoItem`, `Venta`
  - `AppState`: Estado global de la aplicación
  - Propiedades calculadas (ej: Venta.calcularTotal())

#### Controladores (Lógica de Negocio)
- **controllers.js**: Orquestadores entre Vista y Datos
  - `VentaController`: Crear y listar ventas, obtener estadísticas
  - `ProductoController`: CRUD de productos
  - `VendedorController`: CRUD de vendedores
  - `AuthController`: Autenticación y sesión

#### Utilidades
- **ui-utils.js**: Funciones compartidas
  - Spinners y mensajes (showSpinner, showMessage)
  - Formateo (formatCurrency, formatDate)
  - Validación (validateRequired, validatePositive)

#### Vistas (Presentación)
- **form.js**: Manejo del DOM del formulario principal
  - setupEventListeners(): Registra event listeners
  - agregarProductoAlPedido(): Agrega productos a la venta
  - renderizarPedido(): Dibuja la lista de productos
  - onSubmitVenta(): Envía la venta usando VentaController

#### Configuración
- **config.js**: Constantes y variables de configuración globales
- **env.js**: Inyección de variables de entorno
- **estadisticas.js**: Lógica de la página de estadísticas

### Carpeta `css/`
- **styles.css**: Estilos base (página principal)
- **login.css**: Estilos para login
- **admin.css**: Estilos para panel admin
- **estadisticas.css**: Estilos para estadísticas

### Carpeta `views/`
- **login.html**: Página de autenticación
- **admin.html**: Panel de administración (productos, vendedores)
- **estadisticas.html**: Página de estadísticas con gráficos

## Flujo de Datos (MVC Pattern)

```
Usuario interactúa con DOM (views) 
    ↓
Event listener en form.js 
    ↓
Llama método del Controller (ej: ventaController.criarVenta())
    ↓
Controller valida datos con UIUtils
    ↓
Controller llama APIService (ej: api.criarVenta())
    ↓
APIService hace POST a /api/v1/* con token JWT
    ↓
Backend procesa y retorna datos
    ↓
Controller actualiza AppState (models.js)
    ↓
form.js re-renderiza el DOM
    ↓
UIUtils muestra spinner/mensaje al usuario
```

## Estructura de Rutas API

Todas las rutas utilizan `/api/v1/*` como prefijo:

### Autenticación
- `POST /api/v1/auth/login` - Login con usuario/contraseña

### Ventas
- `POST /api/v1/ventas` - Crear nueva venta
- `GET /api/v1/ventas` - Listar todas las ventas
- `GET /api/v1/ventas/{id}` - Obtener venta específica

### Productos
- `POST /api/v1/productos` - Crear producto
- `GET /api/v1/productos` - Listar productos
- `PUT /api/v1/productos/{id}` - Actualizar producto
- `DELETE /api/v1/productos/{id}` - Eliminar producto

### Vendedores
- `POST /api/v1/vendedores` - Crear vendedor
- `GET /api/v1/vendedores` - Listar vendedores
- `PUT /api/v1/vendedores/{id}` - Actualizar vendedor
- `DELETE /api/v1/vendedores/{id}` - Eliminar vendedor

## Migraciones de Paths Completadas

✅ Actualizado `index.html`:
- CSS: `styles.css` → `css/styles.css`
- Scripts: Todos los scripts prefijados con `js/`

✅ Actualizado `views/login.html`:
- CSS: `login.css` → `../css/login.css`
- Scripts: `config.js` → `../js/config.js`

✅ Actualizado `views/admin.html`:
- CSS: `admin.css` → `../css/admin.css`
- Scripts: `config.js` → `../js/config.js`

✅ Actualizado `views/estadisticas.html`:
- CSS: `styles.css` → `../css/styles.css`, `estadisticas.css` → `../css/estadisticas.css`
- Scripts: `config.js` → `../js/config.js`, `estadisticas.js` → `../js/estadisticas.js`

## Próximos Pasos

1. ✅ Reorganización de carpetas completada
2. ⏳ Refactorizar login.html para usar AuthController
3. ⏳ Refactorizar admin.html para usar ProductoController y VendedorController
4. ⏳ Refactorizar estadisticas.html con APIService
5. ⏳ Tests unitarios para controllers
6. ⏳ Documentación OpenAPI/Swagger

## Notas

- Todos los paths en `views/*.html` usan `../` para acceder a `js/` y `css/`
- La ruta `index.html` usa paths directos `js/` y `css/` (en la raíz)
- CORS configurado para: `localhost:5000` (dev) y `https://ecos-ventas-pizzas.netlify.app` (prod)
- JWT tokens almacenados en localStorage automaticamente por APIService
