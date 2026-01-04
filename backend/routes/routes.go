package routes

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"pizzas-ecos/controllers"
	"pizzas-ecos/httputil"
)

// RouteGroup agrupa rutas con un prefijo y middleware común
type RouteGroup struct {
	prefix      string
	middlewares []func(http.Handler) http.Handler
	routes      []Route
}

// Route define una ruta individual
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Name        string
	pathPattern *regexp.Regexp
}

// CustomRouter es un router que maneja rutas con parámetros
type CustomRouter struct {
	routes []*Route
}

// Router maneja el registro de todas las rutas
type Router struct {
	groups []*RouteGroup
}

// NewRouter crea un nuevo router
func NewRouter() *Router {
	return &Router{
		groups: []*RouteGroup{},
	}
}

// Group crea un grupo de rutas con prefijo
func (r *Router) Group(prefix string, middlewares ...func(http.Handler) http.Handler) *RouteGroup {
	group := &RouteGroup{
		prefix:      prefix,
		middlewares: middlewares,
		routes:      []Route{},
	}
	r.groups = append(r.groups, group)
	return group
}

// GET registra una ruta GET
func (rg *RouteGroup) GET(path string, handler http.HandlerFunc, name string) {
	rg.routes = append(rg.routes, Route{
		Method:  http.MethodGet,
		Path:    rg.prefix + path,
		Handler: handler,
		Name:    name,
	})
}

// POST registra una ruta POST
func (rg *RouteGroup) POST(path string, handler http.HandlerFunc, name string) {
	rg.routes = append(rg.routes, Route{
		Method:  http.MethodPost,
		Path:    rg.prefix + path,
		Handler: handler,
		Name:    name,
	})
}

// PUT registra una ruta PUT
func (rg *RouteGroup) PUT(path string, handler http.HandlerFunc, name string) {
	rg.routes = append(rg.routes, Route{
		Method:  http.MethodPut,
		Path:    rg.prefix + path,
		Handler: handler,
		Name:    name,
	})
}

// DELETE registra una ruta DELETE
func (rg *RouteGroup) DELETE(path string, handler http.HandlerFunc, name string) {
	rg.routes = append(rg.routes, Route{
		Method:  http.MethodDelete,
		Path:    rg.prefix + path,
		Handler: handler,
		Name:    name,
	})
}

// convertPathToRegex convierte una ruta con parámetros (:id) a regex
func convertPathToRegex(path string) *regexp.Regexp {
	// Escapar caracteres especiales de regex
	escaped := regexp.QuoteMeta(path)
	// Reemplazar :param con patrón regex
	pattern := strings.ReplaceAll(escaped, "\\:", ":")
	pattern = regexp.MustCompile(`:\w+`).ReplaceAllString(pattern, `[^/]+`)
	pattern = "^" + pattern + "$"
	return regexp.MustCompile(pattern)
}

// matchRoute intenta matchear una ruta
func (r *Route) match(method, path string) bool {
	if r.Method != method {
		return false
	}
	if r.pathPattern == nil {
		r.pathPattern = convertPathToRegex(r.Path)
	}
	return r.pathPattern.MatchString(path)
}

// Register registra todas las rutas en el mux
func (r *Router) Register(mux *http.ServeMux) {
	// Recolectar todas las rutas
	var allRoutes []*Route
	for _, group := range r.groups {
		for i := range group.routes {
			route := &group.routes[i]
			route.pathPattern = convertPathToRegex(route.Path)
			allRoutes = append(allRoutes, route)
		}
	}

	// Registrar un único handler que rutea dinámicamente
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		method := req.Method

		// Buscar ruta que matchee
		for _, route := range allRoutes {
			if route.match(method, path) {
				// Extraer parámetros de la ruta
				params := extractParams(route.Path, path)

				// Crear nuevo context con parámetros
				ctx := context.WithValue(req.Context(), httputil.PathParamsKey, params)
				req = req.WithContext(ctx)

				// Aplicar middlewares
				handler := http.HandlerFunc(route.Handler)
				for _, group := range r.groups {
					for _, r := range group.routes {
						if &r == route && len(group.middlewares) > 0 {
							for _, mw := range group.middlewares {
								handler = mw(handler).(http.HandlerFunc)
							}
							break
						}
					}
				}
				handler.ServeHTTP(w, req)
				return
			}
		}

		// No se encontró ruta
		http.NotFound(w, req)
	})
}

// extractParams extrae parámetros nombrados de una ruta
// Ej: ruta="/api/v1/productos/:id", path="/api/v1/productos/123" retorna {"id": "123"}
func extractParams(routePath, requestPath string) httputil.PathParams {
	params := make(httputil.PathParams)

	// Convertir patrón a regex con grupos nombrados
	parts := strings.Split(routePath, "/")
	pathParts := strings.Split(requestPath, "/")

	if len(parts) != len(pathParts) {
		return params
	}

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// GetRoutes retorna todas las rutas registradas (para documentación)
func (r *Router) GetRoutes() []Route {
	var allRoutes []Route
	for _, group := range r.groups {
		allRoutes = append(allRoutes, group.routes...)
	}
	return allRoutes
}

// SetupRoutes configura todas las rutas de la API
func SetupRoutes() *Router {
	router := NewRouter()

	// Inicializar controladores
	ventaCtrl := controllers.NewVentaController()
	productoCtrl := controllers.NewProductoController()
	vendedorCtrl := controllers.NewVendedorController()
	dataCtrl := controllers.NewDataController()
	authCtrl := controllers.NewAuthController()
	usuarioCtrl := controllers.NewUsuarioController()

	// ============================================
	// GRUPO: Autenticación (Sin middleware)
	// ============================================
	authGroup := router.Group("/api/v1/auth")
	authGroup.POST("/login", authCtrl.Login, "Autenticar usuario")

	// ============================================
	// GRUPO: Datos iniciales (Sin middleware de auth)
	// ============================================
	dataGroup := router.Group("/api/v1/data")
	dataGroup.GET("", dataCtrl.ObtenerData, "Obtener vendedores, clientes y productos")

	// ============================================
	// GRUPO: Ventas
	// ============================================
	ventaGroup := router.Group("/api/v1/ventas")
	ventaGroup.POST("", ventaCtrl.CrearVenta, "Crear nueva venta")
	ventaGroup.PUT("/:id", ventaCtrl.ActualizarVenta, "Actualizar venta")
	ventaGroup.GET("/estadisticas", ventaCtrl.ObtenerEstadisticas, "Obtener estadísticas")
	ventaGroup.GET("/todas", ventaCtrl.ObtenerTodasVentas, "Obtener todas las ventas")

	// ============================================
	// GRUPO: Productos (SIN MIDDLEWARE - Auth aplicado globalmente)
	// ============================================
	productoGroup := router.Group("/api/v1/productos")
	productoGroup.GET("", productoCtrl.Listar, "Listar productos")
	productoGroup.POST("", productoCtrl.Crear, "Crear producto")
	productoGroup.PUT("/:id", productoCtrl.Actualizar, "Actualizar producto")
	productoGroup.DELETE("/:id", productoCtrl.Eliminar, "Eliminar producto")

	// ============================================
	// GRUPO: Vendedores (SIN MIDDLEWARE - Auth aplicado globalmente)
	// ============================================
	vendedorGroup := router.Group("/api/v1/vendedores")
	vendedorGroup.GET("", vendedorCtrl.Listar, "Listar vendedores")
	vendedorGroup.POST("", vendedorCtrl.Crear, "Crear vendedor")
	vendedorGroup.PUT("/:id", vendedorCtrl.Actualizar, "Actualizar vendedor")
	vendedorGroup.DELETE("/:id", vendedorCtrl.Eliminar, "Eliminar vendedor")

	// ============================================
	// GRUPO: Usuarios (SIN MIDDLEWARE - Auth aplicado globalmente)
	// ============================================
	usuarioGroup := router.Group("/api/v1/usuarios")
	usuarioGroup.GET("", usuarioCtrl.Listar, "Listar usuarios")
	usuarioGroup.POST("", usuarioCtrl.Crear, "Crear usuario")
	usuarioGroup.PUT("/:id", usuarioCtrl.Actualizar, "Actualizar usuario")
	usuarioGroup.DELETE("/:id", usuarioCtrl.Eliminar, "Eliminar usuario")

	// ============================================
	// GRUPO: Health Check
	// ============================================
	healthGroup := router.Group("/api/v1/health")
	healthGroup.GET("", healthCtrl(), "Health check")

	// ============================================
	// RUTAS LEGACY para Backward Compatibility (ahora en /api/v1)
	// ============================================
	apiGroup := router.Group("/api/v1")
	apiGroup.POST("/login", authCtrl.Login, "Login")
	apiGroup.GET("/data", dataCtrl.ObtenerData, "Datos iniciales")
	apiGroup.POST("/submit", ventaCtrl.CrearVenta, "Crear venta")
	apiGroup.GET("/estadisticas", ventaCtrl.ObtenerTodasVentas, "Estadísticas")
	apiGroup.GET("/estadisticas-sheet", ventaCtrl.ObtenerEstadisticas, "Estadísticas Sheet")
	apiGroup.POST("/actualizar-venta/:id", ventaCtrl.ActualizarVenta, "Actualizar venta")

	apiGroup.GET("/productos", productoCtrl.Listar, "Listar productos")
	apiGroup.POST("/crear-producto", productoCtrl.Crear, "Crear producto")
	apiGroup.PUT("/actualizar-producto/:id", productoCtrl.Actualizar, "Actualizar producto")
	apiGroup.DELETE("/eliminar-producto/:id", productoCtrl.Eliminar, "Eliminar producto")

	apiGroup.POST("/crear-vendedor", vendedorCtrl.Crear, "Crear vendedor")
	apiGroup.PUT("/actualizar-vendedor/:id", vendedorCtrl.Actualizar, "Actualizar vendedor")
	apiGroup.DELETE("/eliminar-vendedor/:id", vendedorCtrl.Eliminar, "Eliminar vendedor")

	apiGroup.GET("/usuarios", usuarioCtrl.Listar, "Listar usuarios")
	apiGroup.POST("/crear-usuario", usuarioCtrl.Crear, "Crear usuario")
	apiGroup.PUT("/actualizar-usuario/:id", usuarioCtrl.Actualizar, "Actualizar usuario")
	apiGroup.DELETE("/eliminar-usuario/:id", usuarioCtrl.Eliminar, "Eliminar usuario")

	apiGroup.POST("/limpiar-base-datos", dataCtrl.LimpiarBaseDatos, "Limpiar base de datos")

	return router
}

// healthCtrl es un handler de health check
func healthCtrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// PrintRoutes imprime todas las rutas registradas
func PrintRoutes(router *Router) {
	routes := router.GetRoutes()
	for _, route := range routes {
		println(route.Method + " " + route.Path + " (" + route.Name + ")")
	}
}
