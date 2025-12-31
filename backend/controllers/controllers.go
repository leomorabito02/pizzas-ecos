package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"pizzas-ecos/errors"
	"pizzas-ecos/httputil"
	"pizzas-ecos/logger"
	"pizzas-ecos/middleware"
	"pizzas-ecos/models"
	"pizzas-ecos/services"
	"pizzas-ecos/validators"
)

// VentaController maneja requests relacionados con ventas
type VentaController struct {
	ventaService *services.VentaService
}

func NewVentaController() *VentaController {
	return &VentaController{
		ventaService: &services.VentaService{},
	}
}

// CrearVenta crea una nueva venta
func (c *VentaController) CrearVenta(w http.ResponseWriter, r *http.Request) {
	var req models.VentaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("CrearVenta: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	// Validar request - solo validar que no esté vacío
	if req.Vendedor == "" || req.Cliente == "" || len(req.Items) == 0 {
		logger.Warn("CrearVenta: Validación fallida", map[string]interface{}{
			"vendedor": req.Vendedor,
			"cliente":  req.Cliente,
			"items":    len(req.Items),
		})
		errors.WriteError(w, errors.ErrBadRequest, "Vendedor, cliente e items requeridos")
		return
	}

	// Crear venta
	ventaID, err := c.ventaService.CrearVenta(&req)
	if err != nil {
		logger.Error("CrearVenta: Error al crear", "VENTA_CREATE_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al crear venta")
		return
	}

	logger.Info("CrearVenta: Venta creada exitosamente", map[string]interface{}{"venta_id": ventaID})
	errors.WriteSuccess(w, http.StatusCreated, map[string]interface{}{"id": ventaID}, "Venta creada")
}

// ActualizarVenta actualiza una venta existente
func (c *VentaController) ActualizarVenta(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de parámetro de ruta
	idStr := httputil.GetParam(r, "id")
	ventaID, err := strconv.Atoi(idStr)
	if err != nil || ventaID <= 0 {
		logger.Warn("ActualizarVenta: ID inválido", map[string]interface{}{"id": idStr})
		errors.WriteError(w, errors.ErrBadRequest, "ID de venta inválido")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("ActualizarVenta: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	// Extraer campos con validación
	estado, _ := req["estado"].(string)
	paymentMethod, _ := req["payment_method"].(string)
	tipoEntrega, _ := req["tipo_entrega"].(string)

	var productosEliminar []int
	if eliminar, ok := req["productos_eliminar"].([]interface{}); ok {
		for _, id := range eliminar {
			if idVal, ok := id.(float64); ok {
				productosEliminar = append(productosEliminar, int(idVal))
			}
		}
	}

	var productos []map[string]interface{}
	if productosArr, ok := req["productos"].([]interface{}); ok {
		for _, p := range productosArr {
			if pMap, ok := p.(map[string]interface{}); ok {
				productos = append(productos, pMap)
			}
		}
	}

	// Actualizar venta
	err = c.ventaService.ActualizarVenta(ventaID, estado, paymentMethod, tipoEntrega, productosEliminar, productos)
	if err != nil {
		logger.Error("ActualizarVenta: Error al actualizar", "VENTA_UPDATE_ERROR", map[string]interface{}{
			"venta_id": ventaID,
			"error":    err.Error(),
		})
		errors.WriteError(w, errors.ErrServerError, "Error al actualizar venta")
		return
	}

	logger.Info("ActualizarVenta: Venta actualizada", map[string]interface{}{"venta_id": ventaID})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"id": ventaID}, "Venta actualizada")
}

// ObtenerEstadisticas retorna estadísticas de ventas
func (c *VentaController) ObtenerEstadisticas(w http.ResponseWriter, r *http.Request) {
	stats, err := c.ventaService.ObtenerEstadisticas()
	if err != nil {
		logger.Error("ObtenerEstadisticas: Error", "STATS_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al obtener estadísticas")
		return
	}

	errors.WriteSuccess(w, http.StatusOK, stats, "")
}

// ObtenerTodasVentas retorna todas las ventas
func (c *VentaController) ObtenerTodasVentas(w http.ResponseWriter, r *http.Request) {
	ventas, err := c.ventaService.ObtenerTodasVentas()
	if err != nil {
		logger.Error("ObtenerTodasVentas: Error", "VENTAS_LIST_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al obtener ventas")
		return
	}

	errors.WriteSuccess(w, http.StatusOK, ventas, "")
}

// ProductoController maneja requests relacionados con productos
type ProductoController struct {
	productoService *services.ProductoService
}

func NewProductoController() *ProductoController {
	return &ProductoController{
		productoService: &services.ProductoService{},
	}
}

// Listar obtiene lista de productos
func (c *ProductoController) Listar(w http.ResponseWriter, r *http.Request) {
	productos, err := c.productoService.ObtenerProductos()
	if err != nil {
		logger.Error("Listar productos: Error", "PRODUCTOS_LIST_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al obtener productos")
		return
	}

	if len(productos) == 0 {
		productos = []models.Producto{}
	}

	errors.WriteSuccess(w, http.StatusOK, productos, "")
}

// Crear crea un nuevo producto
func (c *ProductoController) Crear(w http.ResponseWriter, r *http.Request) {
	var req models.CrearProductoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Crear producto: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	// Validar
	if req.TipoPizza == "" || req.Precio <= 0 {
		logger.Warn("Crear producto: Validación fallida", map[string]interface{}{
			"tipo_pizza": req.TipoPizza,
			"precio":     req.Precio,
		})
		errors.WriteError(w, errors.ErrBadRequest, "Tipo de pizza y precio requeridos")
		return
	}

	id, err := c.productoService.CrearProducto(&req)
	if err != nil {
		logger.Error("Crear producto: Error", "PRODUCTO_CREATE_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al crear producto")
		return
	}

	logger.Info("Crear producto: Éxito", map[string]interface{}{"producto_id": id})
	errors.WriteSuccess(w, http.StatusCreated, map[string]interface{}{"id": id}, "Producto creado")
}

// Actualizar actualiza un producto
func (c *ProductoController) Actualizar(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de parámetro de ruta
	idStr := httputil.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn("Actualizar producto: ID inválido", map[string]interface{}{"id": idStr})
		errors.WriteError(w, errors.ErrBadRequest, "ID de producto inválido")
		return
	}

	var req models.ActualizarProductoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Actualizar producto: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	// Validar
	if req.TipoPizza == "" || req.Precio <= 0 {
		logger.Warn("Actualizar producto: Validación fallida", map[string]interface{}{
			"tipo_pizza": req.TipoPizza,
			"precio":     req.Precio,
		})
		errors.WriteError(w, errors.ErrBadRequest, "Tipo de pizza y precio requeridos")
		return
	}

	err = c.productoService.ActualizarProducto(id, &req)
	if err != nil {
		logger.Error("Actualizar producto: Error", "PRODUCTO_UPDATE_ERROR", map[string]interface{}{
			"producto_id": id,
			"error":       err.Error(),
		})
		errors.WriteError(w, errors.ErrServerError, "Error al actualizar producto")
		return
	}

	logger.Info("Actualizar producto: Éxito", map[string]interface{}{"producto_id": id})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"id": id}, "Producto actualizado")
}

// Eliminar elimina un producto
func (c *ProductoController) Eliminar(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de parámetro de ruta
	idStr := httputil.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn("Eliminar producto: ID inválido", map[string]interface{}{"id": idStr})
		errors.WriteError(w, errors.ErrBadRequest, "ID de producto inválido")
		return
	}

	err = c.productoService.EliminarProducto(id)
	if err != nil {
		logger.Warn("Eliminar producto: No encontrado", map[string]interface{}{"producto_id": id})
		errors.WriteError(w, errors.ErrNotFound, "Producto no encontrado")
		return
	}

	logger.Info("Eliminar producto: Éxito", map[string]interface{}{"producto_id": id})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"id": id}, "Producto eliminado")
}

// VendedorController maneja requests relacionados con vendedores
type VendedorController struct {
	vendedorService *services.VendedorService
}

func NewVendedorController() *VendedorController {
	return &VendedorController{
		vendedorService: &services.VendedorService{},
	}
}

// Crear crea un nuevo vendedor
func (c *VendedorController) Crear(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Crear vendedor: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	nombre := req["nombre"]

	// Validar
	validation := validators.ValidateVendedorRequest(nombre)
	if !validation.IsValid() {
		logger.Warn("Crear vendedor: Validación fallida", map[string]interface{}{"errors": validation.Errors})
		errors.WriteError(w, errors.ErrBadRequest, validation.GetMessage())
		return
	}

	id, err := c.vendedorService.CrearVendedor(nombre)
	if err != nil {
		logger.Error("Crear vendedor: Error", "VENDEDOR_CREATE_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al crear vendedor")
		return
	}

	logger.Info("Crear vendedor: Éxito", map[string]interface{}{"vendedor_id": id})
	errors.WriteSuccess(w, http.StatusCreated, map[string]interface{}{"id": id}, "Vendedor creado")
}

// Actualizar actualiza un vendedor
func (c *VendedorController) Actualizar(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de parámetro de ruta
	idStr := httputil.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn("Actualizar vendedor: ID inválido", map[string]interface{}{"id": idStr})
		errors.WriteError(w, errors.ErrBadRequest, "ID de vendedor inválido")
		return
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Actualizar vendedor: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	nombre := req["nombre"]

	// Validar
	validation := validators.ValidateVendedorRequest(nombre)
	if !validation.IsValid() {
		logger.Warn("Actualizar vendedor: Validación fallida", map[string]interface{}{"errors": validation.Errors})
		errors.WriteError(w, errors.ErrBadRequest, validation.GetMessage())
		return
	}

	err = c.vendedorService.ActualizarVendedor(id, nombre)
	if err != nil {
		logger.Error("Actualizar vendedor: Error", "VENDEDOR_UPDATE_ERROR", map[string]interface{}{
			"vendedor_id": id,
			"error":       err.Error(),
		})
		errors.WriteError(w, errors.ErrServerError, "Error al actualizar vendedor")
		return
	}

	logger.Info("Actualizar vendedor: Éxito", map[string]interface{}{"vendedor_id": id})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"id": id}, "Vendedor actualizado")
}

// Eliminar elimina un vendedor
func (c *VendedorController) Eliminar(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de parámetro de ruta
	idStr := httputil.GetParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn("Eliminar vendedor: ID inválido", map[string]interface{}{"id": idStr})
		errors.WriteError(w, errors.ErrBadRequest, "ID de vendedor inválido")
		return
	}

	err = c.vendedorService.EliminarVendedor(id)
	if err != nil {
		logger.Warn("Eliminar vendedor: No encontrado", map[string]interface{}{"vendedor_id": id})
		errors.WriteError(w, errors.ErrNotFound, "Vendedor no encontrado")
		return
	}

	logger.Info("Eliminar vendedor: Éxito", map[string]interface{}{"vendedor_id": id})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"id": id}, "Vendedor eliminado")
}

// DataController maneja requests de datos generales
type DataController struct {
	dataService *services.DataService
}

func NewDataController() *DataController {
	return &DataController{
		dataService: &services.DataService{},
	}
}

// ObtenerData retorna vendedores, clientes y productos
func (c *DataController) ObtenerData(w http.ResponseWriter, r *http.Request) {
	data, err := c.dataService.ObtenerDataInicial()
	if err != nil {
		logger.Error("ObtenerData: Error", "DATA_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al obtener datos")
		return
	}

	errors.WriteSuccess(w, http.StatusOK, data, "")
}

// LimpiarBaseDatos limpia todos los datos excepto usuarios
func (c *DataController) LimpiarBaseDatos(w http.ResponseWriter, r *http.Request) {
	err := c.dataService.LimpiarBaseDatos()
	if err != nil {
		logger.Error("LimpiarBaseDatos: Error", "DATABASE_CLEAR_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al limpiar la base de datos")
		return
	}

	logger.Info("LimpiarBaseDatos: Base de datos limpiada exitosamente", map[string]interface{}{})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"status": "cleared"}, "Base de datos limpiada exitosamente")
}

// AuthController maneja requests de autenticación
type AuthController struct {
	authService *services.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService: &services.AuthService{},
	}
}

// Login autentica usuario
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Login: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	// Validar
	validation := validators.ValidateLoginRequest(req.Username, req.Password)
	if !validation.IsValid() {
		logger.Warn("Login: Validación fallida", map[string]interface{}{"errors": validation.Errors})
		errors.WriteError(w, errors.ErrBadRequest, validation.GetMessage())
		return
	}

	// Autenticar
	user, err := c.authService.AutenticarUsuario(req.Username, req.Password)
	if err != nil {
		logger.Warn("Login: Credenciales inválidas", map[string]interface{}{"username": req.Username})
		errors.WriteError(w, errors.ErrUnauthorized, "Usuario o contraseña incorrectos")
		return
	}

	// Generar JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.TokenClaims{
		Username: user.Username,
		Rol:      user.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	tokenString, err := token.SignedString(middleware.JWTSecret)
	if err != nil {
		logger.Error("Login: Error generando token", "TOKEN_GENERATION_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
		errors.WriteError(w, errors.ErrServerError, "Error al generar token")
		return
	}

	// Retornar respuesta con token y usuario
	response := models.LoginResponse{
		Token: tokenString,
		User:  *user,
	}

	logger.Info("Login: Éxito", map[string]interface{}{"username": req.Username})
	errors.WriteSuccess(w, http.StatusOK, response, "Autenticado")
}

// UsuarioController maneja requests relacionados con usuarios
type UsuarioController struct {
	usuarioService *services.UsuarioService
}

func NewUsuarioController() *UsuarioController {
	return &UsuarioController{
		usuarioService: &services.UsuarioService{},
	}
}

// Listar obtiene todos los usuarios
func (c *UsuarioController) Listar(w http.ResponseWriter, r *http.Request) {
	usuarios, err := c.usuarioService.ObtenerTodos()
	if err != nil {
		logger.Error("Listar usuarios: Error al obtener", "USUARIOS_LIST_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al obtener usuarios")
		return
	}

	logger.Info("Listar usuarios: Éxito", map[string]interface{}{"count": len(usuarios)})
	errors.WriteSuccess(w, http.StatusOK, usuarios, "Usuarios obtenidos")
}

// Crear crea un nuevo usuario (siempre como admin)
func (c *UsuarioController) Crear(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Crear usuario: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	// Validar
	if req.Username == "" || req.Password == "" {
		logger.Warn("Crear usuario: Validación fallida", map[string]interface{}{
			"username": req.Username,
			"password": req.Password,
		})
		errors.WriteError(w, errors.ErrBadRequest, "Username y password requeridos")
		return
	}

	// Crear usuario siempre como admin
	usuarioID, err := c.usuarioService.CrearUsuario(req.Username, req.Password, "admin")
	if err != nil {
		logger.Error("Crear usuario: Error al crear", "USUARIO_CREATE_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al crear usuario")
		return
	}

	logger.Info("Crear usuario: Éxito", map[string]interface{}{"usuario_id": usuarioID, "username": req.Username, "rol": "admin"})
	errors.WriteSuccess(w, http.StatusCreated, map[string]interface{}{"id": usuarioID}, "Usuario creado como admin")
}

// Actualizar actualiza un usuario existente
func (c *UsuarioController) Actualizar(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de parámetro de ruta
	idStr := httputil.GetParam(r, "id")
	usuarioID, err := strconv.Atoi(idStr)
	if err != nil || usuarioID <= 0 {
		logger.Warn("Actualizar usuario: ID inválido", map[string]interface{}{"id": idStr})
		errors.WriteError(w, errors.ErrBadRequest, "ID de usuario inválido")
		return
	}

	var req models.UpdateUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Actualizar usuario: JSON inválido", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrBadRequest, "JSON inválido")
		return
	}

	// Validar
	if req.Username == "" || req.Rol == "" {
		logger.Warn("Actualizar usuario: Validación fallida", map[string]interface{}{
			"username": req.Username,
			"rol":      req.Rol,
		})
		errors.WriteError(w, errors.ErrBadRequest, "Username y rol requeridos")
		return
	}

	// Validar rol
	if req.Rol != "admin" && req.Rol != "vendedor" {
		logger.Warn("Actualizar usuario: Rol inválido", map[string]interface{}{"rol": req.Rol})
		errors.WriteError(w, errors.ErrBadRequest, "Rol debe ser 'admin' o 'vendedor'")
		return
	}

	// Actualizar usuario
	err = c.usuarioService.ActualizarUsuario(usuarioID, req.Username, req.Password, req.Rol)
	if err != nil {
		logger.Error("Actualizar usuario: Error al actualizar", "USUARIO_UPDATE_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al actualizar usuario")
		return
	}

	logger.Info("Actualizar usuario: Éxito", map[string]interface{}{"usuario_id": usuarioID})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"id": usuarioID}, "Usuario actualizado")
}

// Eliminar elimina un usuario
func (c *UsuarioController) Eliminar(w http.ResponseWriter, r *http.Request) {
	// Extraer ID de parámetro de ruta
	idStr := httputil.GetParam(r, "id")
	usuarioID, err := strconv.Atoi(idStr)
	if err != nil || usuarioID <= 0 {
		logger.Warn("Eliminar usuario: ID inválido", map[string]interface{}{"id": idStr})
		errors.WriteError(w, errors.ErrBadRequest, "ID de usuario inválido")
		return
	}

	// Eliminar usuario
	err = c.usuarioService.EliminarUsuario(usuarioID)
	if err != nil {
		logger.Error("Eliminar usuario: Error al eliminar", "USUARIO_DELETE_ERROR", map[string]interface{}{"error": err.Error()})
		errors.WriteError(w, errors.ErrServerError, "Error al eliminar usuario")
		return
	}

	logger.Info("Eliminar usuario: Éxito", map[string]interface{}{"usuario_id": usuarioID})
	errors.WriteSuccess(w, http.StatusOK, map[string]interface{}{"id": usuarioID}, "Usuario eliminado")
}
