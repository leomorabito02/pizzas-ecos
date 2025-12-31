package services

import (
	"context"
	"fmt"

	"pizzas-ecos/database"
	"pizzas-ecos/logger"
	"pizzas-ecos/models"
)

// VentaService contiene lógica de negocio para ventas
type VentaService struct{}

// CrearVenta crea una nueva venta con validación de negocio y transacción
func (s *VentaService) CrearVenta(req *models.VentaRequest) (int, error) {
	ctx := context.Background()

	// Validar datos requeridos
	if err := s.validarVentaRequest(req); err != nil {
		logger.Warn("CrearVenta: Validación fallida", map[string]interface{}{"error": err.Error()})
		return 0, err
	}

	// Obtener ID del vendedor
	vendedorID, err := database.GetVendedorID(req.Vendedor)
	if err != nil {
		logger.Error("CrearVenta: Vendedor no encontrado", "VENDOR_NOT_FOUND", map[string]interface{}{
			"vendedor": req.Vendedor,
			"error":    err.Error(),
		})
		return 0, fmt.Errorf("vendedor no encontrado: %w", err)
	}

	// Verificar que el vendedor existe
	exists, err := database.ExistsVendedor(ctx, vendedorID)
	if err != nil || !exists {
		return 0, fmt.Errorf("vendedor no válido")
	}

	// Obtener o crear cliente
	var clienteID *int
	if req.Cliente != "" {
		id, err := database.GetOrCreateCliente(req.Cliente)
		if err == nil {
			clienteID = &id
		}
	}

	// Iniciar transacción
	tx, err := database.BeginTx(ctx)
	if err != nil {
		logger.Error("CrearVenta: Error iniciando transacción", "TX_BEGIN_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
		return 0, fmt.Errorf("error en transacción: %w", err)
	}

	// Calcular total
	total := s.calcularTotal(req.Items)

	// Insertar venta
	ventaID, err := database.InsertVenta(clienteID, vendedorID, total, req.PaymentMethod, req.Estado, req.TipoEntrega)
	if err != nil {
		tx.Rollback()
		logger.Error("CrearVenta: Error insertando venta", "VENTA_INSERT_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
		return 0, fmt.Errorf("error guardando venta: %w", err)
	}

	// Insertar detalles
	for _, item := range req.Items {
		if err := database.InsertDetalle(ventaID, item); err != nil {
			tx.Rollback()
			logger.Error("CrearVenta: Error insertando detalle", "DETAIL_INSERT_ERROR", map[string]interface{}{
				"venta_id": ventaID,
				"error":    err.Error(),
			})
			return 0, fmt.Errorf("error insertando detalle: %w", err)
		}
	}

	// Commit de la transacción
	if err := tx.Commit(); err != nil {
		logger.Error("CrearVenta: Error en commit", "TX_COMMIT_ERROR", map[string]interface{}{
			"venta_id": ventaID,
			"error":    err.Error(),
		})
		return 0, fmt.Errorf("error completando venta: %w", err)
	}

	logger.Info("CrearVenta: Venta creada exitosamente", map[string]interface{}{
		"venta_id": ventaID,
		"total":    total,
	})

	return ventaID, nil
}

// ActualizarVenta actualiza una venta existente
func (s *VentaService) ActualizarVenta(ventaID int, estado, paymentMethod, tipoEntrega string, productosEliminar []int, productos []map[string]interface{}) error {
	// Validar estado válido
	estadosValidos := map[string]bool{
		"sin pagar": true,
		"pagada":    true,
		"entregada": true,
		"cancelada": true,
	}
	if !estadosValidos[estado] {
		return fmt.Errorf("estado inválido: %s", estado)
	}

	// Validar método de pago
	metodosPagos := map[string]bool{
		"efectivo":      true,
		"transferencia": true,
	}
	if !metodosPagos[paymentMethod] {
		return fmt.Errorf("método de pago inválido: %s", paymentMethod)
	}

	// Actualizar en BD
	return database.UpdateVenta(ventaID, estado, paymentMethod, tipoEntrega, productosEliminar, productos)
}

// ObtenerEstadisticas retorna estadísticas completas
func (s *VentaService) ObtenerEstadisticas() (map[string]interface{}, error) {
	resumen, err := database.GetResumen()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo resumen: %w", err)
	}

	vendedores, err := database.GetVendedoresConStats()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo vendedores: %w", err)
	}

	ventas, err := database.GetAllVentas(false)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ventas: %w", err)
	}

	return map[string]interface{}{
		"resumen":    resumen,
		"vendedores": vendedores,
		"ventas":     ventas,
	}, nil
}

// ObtenerTodasVentas retorna todas las ventas incluyendo canceladas
func (s *VentaService) ObtenerTodasVentas() ([]models.VentaStats, error) {
	ventas, err := database.GetAllVentas(true)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ventas: %w", err)
	}
	return ventas, nil
}

// Validaciones privadas
func (s *VentaService) validarVentaRequest(req *models.VentaRequest) error {
	if req.Vendedor == "" {
		return fmt.Errorf("vendedor es requerido")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("al menos un item es requerido")
	}

	// Validar que cada item tenga datos válidos
	for i, item := range req.Items {
		if item.ProductID <= 0 {
			return fmt.Errorf("item %d: product_id inválido", i)
		}
		if item.Cantidad <= 0 {
			return fmt.Errorf("item %d: cantidad debe ser mayor a 0", i)
		}
		if item.Precio < 0 {
			return fmt.Errorf("item %d: precio no puede ser negativo", i)
		}
	}

	return nil
}

func (s *VentaService) calcularTotal(items []models.ProductoItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Total
	}
	return total
}

// ProductoService contiene lógica de negocio para productos
type ProductoService struct{}

// CrearProducto crea un nuevo producto
func (s *ProductoService) CrearProducto(req *models.CrearProductoRequest) (int64, error) {
	// Validar
	if err := s.validarCrearProducto(req); err != nil {
		return 0, err
	}

	id, err := database.CreateProducto(req.TipoPizza, req.Descripcion, req.Precio)
	if err != nil {
		return 0, fmt.Errorf("error creando producto: %w", err)
	}

	return id, nil
}

// ActualizarProducto actualiza un producto
func (s *ProductoService) ActualizarProducto(id int, req *models.ActualizarProductoRequest) error {
	if err := s.validarActualizarProducto(req); err != nil {
		return err
	}

	return database.UpdateProducto(id, req.TipoPizza, req.Descripcion, req.Precio, req.Activo)
}

// EliminarProducto elimina un producto (soft delete)
func (s *ProductoService) EliminarProducto(id int) error {
	return database.DeleteProducto(id)
}

// ObtenerProductos retorna lista de productos activos
func (s *ProductoService) ObtenerProductos() ([]models.Producto, error) {
	productos, err := database.GetProductos()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo productos: %w", err)
	}
	return productos, nil
}

// Validaciones privadas
func (s *ProductoService) validarCrearProducto(req *models.CrearProductoRequest) error {
	if req.TipoPizza == "" {
		return fmt.Errorf("tipo_pizza es requerido")
	}
	if req.Precio <= 0 {
		return fmt.Errorf("precio debe ser mayor a 0")
	}
	return nil
}

func (s *ProductoService) validarActualizarProducto(req *models.ActualizarProductoRequest) error {
	if req.TipoPizza == "" {
		return fmt.Errorf("tipo_pizza es requerido")
	}
	if req.Precio <= 0 {
		return fmt.Errorf("precio debe ser mayor a 0")
	}
	return nil
}

// VendedorService contiene lógica de negocio para vendedores
type VendedorService struct{}

// CrearVendedor crea un nuevo vendedor
func (s *VendedorService) CrearVendedor(nombre string) (int64, error) {
	if nombre == "" {
		return 0, fmt.Errorf("nombre del vendedor es requerido")
	}

	id, err := database.CreateVendedor(nombre)
	if err != nil {
		return 0, fmt.Errorf("error creando vendedor: %w", err)
	}

	return id, nil
}

// ActualizarVendedor actualiza un vendedor
func (s *VendedorService) ActualizarVendedor(id int, nombre string) error {
	if nombre == "" {
		return fmt.Errorf("nombre del vendedor es requerido")
	}

	return database.UpdateVendedor(id, nombre)
}

// EliminarVendedor elimina un vendedor
func (s *VendedorService) EliminarVendedor(id int) error {
	return database.DeleteVendedor(id)
}

// ObtenerVendedores retorna lista de vendedores
func (s *VendedorService) ObtenerVendedores() ([]models.Vendedor, error) {
	vendedores, err := database.GetVendedores()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo vendedores: %w", err)
	}
	return vendedores, nil
}

// DataService contiene lógica para obtener datos generales
type DataService struct{}

// ObtenerDataInicial retorna vendedores, clientes y productos
func (s *DataService) ObtenerDataInicial() (*models.DataResponse, error) {
	vendedores, err := database.GetVendedores()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo vendedores: %w", err)
	}

	clientesPorVendedor, err := database.GetClientesPorVendedor()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo clientes: %w", err)
	}

	productos, err := database.GetProductos()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo productos: %w", err)
	}

	return &models.DataResponse{
		Vendedores:          vendedores,
		ClientesPorVendedor: clientesPorVendedor,
		Productos:           productos,
	}, nil
}

// LimpiarBaseDatos elimina todos los datos excepto usuarios
func (s *DataService) LimpiarBaseDatos() error {
	// Eliminar en el orden correcto para evitar restricciones de foreign keys
	// 1. Eliminar detalles de ventas
	if err := database.ClearDetalleVentas(); err != nil {
		logger.Error("LimpiarBaseDatos: Error eliminando detalles", "CLEAR_DETAIL_ERROR", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("error eliminando detalles: %w", err)
	}

	// 2. Eliminar ventas
	if err := database.ClearVentas(); err != nil {
		logger.Error("LimpiarBaseDatos: Error eliminando ventas", "CLEAR_VENTAS_ERROR", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("error eliminando ventas: %w", err)
	}

	// 3. Eliminar clientes
	if err := database.ClearClientes(); err != nil {
		logger.Error("LimpiarBaseDatos: Error eliminando clientes", "CLEAR_CLIENTES_ERROR", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("error eliminando clientes: %w", err)
	}

	// 4. Eliminar vendedores
	if err := database.ClearVendedores(); err != nil {
		logger.Error("LimpiarBaseDatos: Error eliminando vendedores", "CLEAR_VENDEDORES_ERROR", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("error eliminando vendedores: %w", err)
	}

	// 5. Eliminar productos
	if err := database.ClearProductos(); err != nil {
		logger.Error("LimpiarBaseDatos: Error eliminando productos", "CLEAR_PRODUCTOS_ERROR", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("error eliminando productos: %w", err)
	}

	logger.Info("LimpiarBaseDatos: Base de datos limpiada exitosamente", map[string]interface{}{})
	return nil
}

// AuthService contiene lógica de autenticación
type AuthService struct{}

// AutenticarUsuario autentica un usuario y retorna token
func (s *AuthService) AutenticarUsuario(username, passwordHash string) (*models.User, error) {
	user, err := database.GetUserByCredentials(username, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}
	return user, nil
}

type UsuarioService struct{}

// ObtenerTodos obtiene todos los usuarios sin mostrar contraseñas
func (s *UsuarioService) ObtenerTodos() ([]models.User, error) {
	usuarios, err := database.GetAllUsers()
	if err != nil {
		logger.Error("ObtenerTodos: Error al obtener usuarios", "USUARIOS_GET_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("error obteniendo usuarios: %w", err)
	}
	return usuarios, nil
}

// CrearUsuario crea un nuevo usuario con contraseña hasheada
func (s *UsuarioService) CrearUsuario(username, password, rol string) (int, error) {
	// Validar que el usuario no exista
	exists, err := database.UserExists(username)
	if err != nil {
		logger.Error("CrearUsuario: Error verificando existencia", "USER_CHECK_ERROR", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return 0, fmt.Errorf("error verificando usuario: %w", err)
	}

	if exists {
		logger.Warn("CrearUsuario: Usuario ya existe", map[string]interface{}{"username": username})
		return 0, fmt.Errorf("usuario ya existe")
	}

	// Crear usuario
	usuarioID, err := database.CreateUser(username, password, rol)
	if err != nil {
		logger.Error("CrearUsuario: Error al crear", "USER_CREATE_ERROR", map[string]interface{}{
			"username": username,
			"error":    err.Error(),
		})
		return 0, fmt.Errorf("error creando usuario: %w", err)
	}

	logger.Info("CrearUsuario: Usuario creado", map[string]interface{}{
		"usuario_id": usuarioID,
		"username":   username,
		"rol":        rol,
	})
	return usuarioID, nil
}

// ActualizarUsuario actualiza un usuario existente
func (s *UsuarioService) ActualizarUsuario(usuarioID int, username, password, rol string) error {
	err := database.UpdateUser(usuarioID, username, password, rol)
	if err != nil {
		logger.Error("ActualizarUsuario: Error al actualizar", "USER_UPDATE_ERROR", map[string]interface{}{
			"usuario_id": usuarioID,
			"error":      err.Error(),
		})
		return fmt.Errorf("error actualizando usuario: %w", err)
	}

	logger.Info("ActualizarUsuario: Usuario actualizado", map[string]interface{}{
		"usuario_id": usuarioID,
		"username":   username,
		"rol":        rol,
	})
	return nil
}

// EliminarUsuario elimina un usuario
func (s *UsuarioService) EliminarUsuario(usuarioID int) error {
	err := database.DeleteUser(usuarioID)
	if err != nil {
		logger.Error("EliminarUsuario: Error al eliminar", "USER_DELETE_ERROR", map[string]interface{}{
			"usuario_id": usuarioID,
			"error":      err.Error(),
		})
		return fmt.Errorf("error eliminando usuario: %w", err)
	}

	logger.Info("EliminarUsuario: Usuario eliminado", map[string]interface{}{
		"usuario_id": usuarioID,
	})
	return nil
}
