package services

import (
	"context"
	"fmt"
	"strings"

	"pizzas-ecos/database"
	"pizzas-ecos/logger"
	"pizzas-ecos/models"
)

// VentaServiceInterface define los métodos del servicio de ventas
type VentaServiceInterface interface {
	CrearVenta(req *models.VentaRequest) (int, error)
	ActualizarVenta(ventaID int, estado, paymentMethod, tipoEntrega string, productosEliminar []int, productos []map[string]interface{}) error
	ObtenerEstadisticas() (map[string]interface{}, error)
	ObtenerTodasVentas() ([]models.VentaStats, error)
}

// ProductoServiceInterface define los métodos del servicio de productos
type ProductoServiceInterface interface {
	ObtenerProductos() ([]models.Producto, error)
	CrearProducto(req *models.CrearProductoRequest) (int64, error)
	ActualizarProducto(id int, req *models.ActualizarProductoRequest) error
	EliminarProducto(id int) error
}

// VendedorServiceInterface define los métodos del servicio de vendedores
type VendedorServiceInterface interface {
	ObtenerVendedores() ([]models.Vendedor, error)
	CrearVendedor(nombre string) (int64, error)
	ActualizarVendedor(id int, nombre string) error
	EliminarVendedor(id int) error
}

// AuthServiceInterface define los métodos del servicio de autenticación
type AuthServiceInterface interface {
	AutenticarUsuario(username, password string) (*models.User, error)
}

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
		// Validar que cliente no sea solo espacios (trim)
		cliente := strings.TrimSpace(req.Cliente)
		if cliente == "" {
			logger.Warn("CrearVenta: Cliente vacío después de trim", map[string]interface{}{
				"cliente_original": req.Cliente,
			})
			return 0, fmt.Errorf("cliente no puede estar vacío")
		}
		// Intentar obtener cliente existente para posiblemente actualizar su teléfono
		id, tel, exists, err := database.GetClienteByNombre(cliente)
		if err == nil && exists {
			logger.Info("CrearVenta: Cliente existente encontrado", map[string]interface{}{
				"cliente":          cliente,
				"cliente_id":       id,
				"telefono_actual":  tel,
				"telefono_enviado": req.TelefonoCliente,
			})
			// Si se envió teléfono y es distinto, actualizarlo
			if req.TelefonoCliente != nil && *req.TelefonoCliente != 0 && *req.TelefonoCliente != tel {
				telPtr := *req.TelefonoCliente
				if err := database.UpdateClienteTelefono(id, &telPtr); err != nil {
					logger.Warn("CrearVenta: Error actualizando teléfono de cliente existente", map[string]interface{}{
						"cliente_id":      id,
						"cliente":         cliente,
						"telefono_nuevo":  req.TelefonoCliente,
						"telefono_actual": tel,
						"error":           err.Error(),
					})
					// Continuar con la creación de la venta aunque falle la actualización del teléfono
				} else {
					logger.Info("CrearVenta: Teléfono actualizado para cliente existente", map[string]interface{}{
						"cliente_id":        id,
						"cliente":           cliente,
						"telefono_anterior": tel,
						"telefono_nuevo":    req.TelefonoCliente,
					})
				}
			} else {
				razon := "telefono_igual_al_actual"
				if req.TelefonoCliente == nil || *req.TelefonoCliente == 0 {
					razon = "telefono_enviado_es_0_o_nil"
				}
				logger.Info("CrearVenta: Teléfono no actualizado", map[string]interface{}{
					"cliente":          cliente,
					"telefono_actual":  tel,
					"telefono_enviado": req.TelefonoCliente,
					"razon":            razon,
				})
			}
			clienteID = &id
		} else {
			// Crear cliente nuevo con teléfono opcional
			var telPtr *int
			if req.TelefonoCliente != nil && *req.TelefonoCliente != 0 {
				t := *req.TelefonoCliente
				telPtr = &t
			}
			newID, err := database.CreateClienteWithTelefono(req.Cliente, telPtr)
			if err == nil {
				clienteID = &newID
			}
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
	if len(strings.TrimSpace(req.Vendedor)) < 2 {
		return fmt.Errorf("nombre de vendedor debe tener al menos 2 caracteres")
	}
	if len(req.Vendedor) > 100 {
		return fmt.Errorf("nombre de vendedor demasiado largo")
	}

	// Validar que el vendedor existe
	vendedorID, err := database.GetVendedorID(req.Vendedor)
	if err != nil {
		return fmt.Errorf("error al validar vendedor: %w", err)
	}
	if vendedorID <= 0 {
		return fmt.Errorf("vendedor '%s' no encontrado", req.Vendedor)
	}

	if req.Cliente == "" {
		return fmt.Errorf("cliente es requerido")
	}
	if len(strings.TrimSpace(req.Cliente)) < 2 {
		return fmt.Errorf("nombre de cliente debe tener al menos 2 caracteres")
	}
	if len(req.Cliente) > 100 {
		return fmt.Errorf("nombre de cliente demasiado largo")
	}

	if req.TelefonoCliente != nil && *req.TelefonoCliente != 0 && (*req.TelefonoCliente < 10 || *req.TelefonoCliente > 999999999999999) {
		return fmt.Errorf("teléfono debe tener entre 2 y 15 dígitos")
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("al menos un item es requerido")
	}
	if len(req.Items) > 50 {
		return fmt.Errorf("demasiados items (máximo 50)")
	}

	// Validar que cada item tenga datos válidos y producto existe
	for i, item := range req.Items {
		if item.ProductID <= 0 {
			return fmt.Errorf("item %d: product_id inválido", i)
		}
		if item.Cantidad <= 0 {
			return fmt.Errorf("item %d: cantidad debe ser mayor a 0", i)
		}
		if item.Cantidad > 100 {
			return fmt.Errorf("item %d: cantidad demasiado grande (máximo 100)", i)
		}
		if item.Precio < 0 {
			return fmt.Errorf("item %d: precio no puede ser negativo", i)
		}

		// Validar que el producto existe
		exists, err := database.ExistsProducto(context.Background(), item.ProductID)
		if err != nil {
			return fmt.Errorf("item %d: error al validar producto: %w", i, err)
		}
		if !exists {
			return fmt.Errorf("item %d: producto con ID %d no existe", i, item.ProductID)
		}
	}

	// Validar payment method
	if req.PaymentMethod == "" {
		return fmt.Errorf("método de pago es requerido")
	}
	validPayments := []string{"efectivo", "tarjeta", "transferencia", "qr"}
	found := false
	for _, p := range validPayments {
		if strings.ToLower(req.PaymentMethod) == p {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("método de pago inválido (debe ser: efectivo, tarjeta, transferencia, qr)")
	}

	// Validar estado
	if req.Estado != "" {
		validEstados := []string{"pendiente", "pagada", "cancelada", "en_proceso"}
		found := false
		for _, e := range validEstados {
			if strings.ToLower(req.Estado) == e {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("estado inválido (debe ser: pendiente, pagada, cancelada, en_proceso)")
		}
	}

	// Validar tipo de entrega
	if req.TipoEntrega != "" {
		validTipos := []string{"retiro", "envio", "delivery"}
		found := false
		for _, t := range validTipos {
			if strings.ToLower(req.TipoEntrega) == t {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tipo de entrega inválido (debe ser: retiro, envio, delivery)")
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
	if strings.TrimSpace(req.TipoPizza) == "" {
		return fmt.Errorf("tipo_pizza es requerido")
	}
	if len(strings.TrimSpace(req.TipoPizza)) < 2 {
		return fmt.Errorf("tipo_pizza debe tener al menos 2 caracteres")
	}
	if len(req.TipoPizza) > 50 {
		return fmt.Errorf("tipo_pizza demasiado largo (máximo 50 caracteres)")
	}

	if len(strings.TrimSpace(req.Descripcion)) > 200 {
		return fmt.Errorf("descripcion demasiado larga (máximo 200 caracteres)")
	}

	if req.Precio <= 0 {
		return fmt.Errorf("precio debe ser mayor a 0")
	}
	if req.Precio > 500 {
		return fmt.Errorf("precio demasiado alto (máximo $500)")
	}

	return nil
}

func (s *ProductoService) validarActualizarProducto(req *models.ActualizarProductoRequest) error {
	if strings.TrimSpace(req.TipoPizza) == "" {
		return fmt.Errorf("tipo_pizza es requerido")
	}
	if len(strings.TrimSpace(req.TipoPizza)) < 2 {
		return fmt.Errorf("tipo_pizza debe tener al menos 2 caracteres")
	}
	if len(req.TipoPizza) > 50 {
		return fmt.Errorf("tipo_pizza demasiado largo (máximo 50 caracteres)")
	}

	if req.Precio <= 0 {
		return fmt.Errorf("precio debe ser mayor a 0")
	}
	if req.Precio > 500 {
		return fmt.Errorf("precio demasiado alto (máximo $500)")
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
