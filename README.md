# Pizzas ECOS - Sistema de Ventas

Sistema de registro de ventas con integración a Google Sheets.

## Estructura del Proyecto

```
pizzas-ecos/
├── backend/
│   ├── main.go           # Servidor Go (API + servidor estático)
│   ├── go.mod            # Dependencias Go
│   └── go.sum
├── frontend/
│   ├── index.html        # Página principal
│   ├── form.js           # Lógica del formulario
│   └── styles.css        # Estilos CSS
├── venta-pizzas-ecos.json # Credenciales de Google (no commitear)
├── .env                  # Variables de entorno (no commitear)
└── README.md
```

## Requisitos

- **Go 1.25+** (para el backend)
- **Google Sheets API** habilitada
- **Credentials JSON** de Google Cloud

## Instalación

### 1. Backend

```bash
cd backend
go mod tidy
```

### 2. Variables de Entorno

Crea un archivo `.env` en la raíz con:

```
GOOGLE_CREDENTIALS_JSON=<contenido del JSON de credenciales>
PORT=8080
```

O usa el archivo `venta-pizzas-ecos.json` en la raíz.

## Ejecución

### Desarrollo Local - Terminal 1 (Backend)

```bash
cd backend
go run main.go
```

El servidor backend estará disponible en `http://localhost:8080/api`

### Desarrollo Local - Terminal 2 (Frontend)

```bash
cd frontend
python server.py
```

El frontend estará disponible en `http://localhost:5000`

**Abre tu navegador en `http://localhost:5000`**

### Estructura de Rutas Backend

- `POST /api/submit` - Guarda una venta en Google Sheets
- `GET /api/data` - Obtiene vendedores y clientes históricos

### Estructura de Rutas Frontend

- `/` - Página principal (index.html)
- `/styles.css` - Estilos
- `/form.js` - Lógica JavaScript

## Configuración de Google Sheets

El sistema lee:
- **Vendedores**: Hoja `datos`, columna C, a partir de C9 (hasta celda vacía)
- **Clientes**: Extraídos del historial de ventas
- **Guardar ventas**: Hoja `Ventas`

## Desarrollo Frontend

El frontend está completamente separado del backend:

```bash
cd frontend
python server.py
```

Luego abre `http://localhost:5000` en tu navegador.

**Nota**: Asegúrate que el backend está corriendo en otra terminal, porque el frontend comunica con `http://localhost:8080/api`

## Despliegue en Render

1. Crea un repositorio en GitHub
2. Conecta en Render
3. Configura variables de entorno:
   - `GOOGLE_CREDENTIALS_JSON` (contenido del JSON)
   - `PORT` (por defecto 8080)
4. Build command: `cd backend && go mod tidy`
5. Start command: `cd backend && go run main.go`

## Notas

- Los credenciales de Google no deben ser commiteados (incluidos en `.gitignore`)
- El frontend está completamente separado del backend
- Se pueden servir desde orígenes diferentes (CORS habilitado si es necesario)
