package services

import (
	"context"
	"fmt"
	"testing"

	"pizzas-ecos/models"
)

// Mock para database functions
type mockDatabase struct {
	getVendedorIDFunc          func(nombre string) (int, error)
	existsVendedorFunc         func(ctx context.Context, id int) (bool, error)
	existsProductoFunc         func(ctx context.Context, id int) (bool, error)
	getOrCreateClienteFunc     func(nombre string) (int, error)
	crearVentaFunc             func(ctx context.Context, venta *models.VentaRequest) (int, error)
	crearDetalleVentaFunc      func(ctx context.Context, detalles []models.ProductoItem) error
	getProductosFunc           func() ([]models.Producto, error)
	crearProductoFunc          func(producto *models.Producto) (*models.Producto, error)
	actualizarProductoFunc     func(id int, producto *models.Producto) (*models.Producto, error)
	eliminarProductoFunc       func(id int) error
	getVendedoresFunc          func() ([]models.Vendedor, error)
	crearVendedorFunc          func(vendedor *models.Vendedor) (*models.Vendedor, error)
	actualizarVendedorFunc     func(id int, vendedor *models.Vendedor) (*models.Vendedor, error)
	eliminarVendedorFunc       func(id int) error
	validarCredencialesFunc    func(usuario, password string) (*models.User, error)
	getClientesPorVendedorFunc func() (map[string][]models.Cliente, error)
}

func (m *mockDatabase) GetVendedorID(nombre string) (int, error) {
	if m.getVendedorIDFunc != nil {
		return m.getVendedorIDFunc(nombre)
	}
	return 1, nil
}

func (m *mockDatabase) ExistsVendedor(ctx context.Context, id int) (bool, error) {
	if m.existsVendedorFunc != nil {
		return m.existsVendedorFunc(ctx, id)
	}
	return true, nil
}

func (m *mockDatabase) ExistsProducto(ctx context.Context, id int) (bool, error) {
	if m.existsProductoFunc != nil {
		return m.existsProductoFunc(ctx, id)
	}
	return true, nil
}

func (m *mockDatabase) GetOrCreateCliente(nombre string) (int, error) {
	if m.getOrCreateClienteFunc != nil {
		return m.getOrCreateClienteFunc(nombre)
	}
	return 1, nil
}

func (m *mockDatabase) CrearVenta(ctx context.Context, venta *models.VentaRequest) (int, error) {
	if m.crearVentaFunc != nil {
		return m.crearVentaFunc(ctx, venta)
	}
	return 1, nil
}

func (m *mockDatabase) CrearDetalleVenta(ctx context.Context, detalles []models.ProductoItem) error {
	if m.crearDetalleVentaFunc != nil {
		return m.crearDetalleVentaFunc(ctx, detalles)
	}
	return nil
}

func (m *mockDatabase) GetProductos() ([]models.Producto, error) {
	if m.getProductosFunc != nil {
		return m.getProductosFunc()
	}
	return []models.Producto{{ID: 1, TipoPizza: "Margherita", Precio: 10.0}}, nil
}

func (m *mockDatabase) CrearProducto(producto *models.Producto) (*models.Producto, error) {
	if m.crearProductoFunc != nil {
		return m.crearProductoFunc(producto)
	}
	return &models.Producto{ID: 1, TipoPizza: producto.TipoPizza, Precio: producto.Precio}, nil
}

func (m *mockDatabase) ActualizarProducto(id int, producto *models.Producto) (*models.Producto, error) {
	if m.actualizarProductoFunc != nil {
		return m.actualizarProductoFunc(id, producto)
	}
	return &models.Producto{ID: id, TipoPizza: producto.TipoPizza, Precio: producto.Precio}, nil
}

func (m *mockDatabase) EliminarProducto(id int) error {
	if m.eliminarProductoFunc != nil {
		return m.eliminarProductoFunc(id)
	}
	return nil
}

func (m *mockDatabase) GetVendedores() ([]models.Vendedor, error) {
	if m.getVendedoresFunc != nil {
		return m.getVendedoresFunc()
	}
	return []models.Vendedor{{ID: 1, Nombre: "Juan Pérez"}}, nil
}

func (m *mockDatabase) CrearVendedor(vendedor *models.Vendedor) (*models.Vendedor, error) {
	if m.crearVendedorFunc != nil {
		return m.crearVendedorFunc(vendedor)
	}
	return &models.Vendedor{ID: 1, Nombre: vendedor.Nombre}, nil
}

func (m *mockDatabase) ActualizarVendedor(id int, vendedor *models.Vendedor) (*models.Vendedor, error) {
	if m.actualizarVendedorFunc != nil {
		return m.actualizarVendedorFunc(id, vendedor)
	}
	return &models.Vendedor{ID: id, Nombre: vendedor.Nombre}, nil
}

func (m *mockDatabase) EliminarVendedor(id int) error {
	if m.eliminarVendedorFunc != nil {
		return m.eliminarVendedorFunc(id)
	}
	return nil
}

func (m *mockDatabase) ValidarCredenciales(usuario, password string) (*models.User, error) {
	if m.validarCredencialesFunc != nil {
		return m.validarCredencialesFunc(usuario, password)
	}
	return &models.User{Username: "Admin", Rol: "admin"}, nil
}

func (m *mockDatabase) GetClientesPorVendedor() (map[string][]models.Cliente, error) {
	if m.getClientesPorVendedorFunc != nil {
		return m.getClientesPorVendedorFunc()
	}
	return map[string][]models.Cliente{
		"Juan Pérez": {{ID: 1, Nombre: "María García"}},
	}, nil
}

// DatabaseInterface define las operaciones de base de datos
type DatabaseInterface interface {
	GetVendedorID(nombre string) (int, error)
	ExistsVendedor(ctx context.Context, id int) (bool, error)
	ExistsProducto(ctx context.Context, id int) (bool, error)
	GetOrCreateCliente(nombre string) (int, error)
	CrearVenta(ctx context.Context, venta *models.VentaRequest) (int, error)
	CrearDetalleVenta(ctx context.Context, detalles []models.ProductoItem) error
	GetProductos() ([]models.Producto, error)
	CrearProducto(producto *models.Producto) (*models.Producto, error)
	ActualizarProducto(id int, producto *models.Producto) (*models.Producto, error)
	EliminarProducto(id int) error
	GetVendedores() ([]models.Vendedor, error)
	CrearVendedor(vendedor *models.Vendedor) (*models.Vendedor, error)
	ActualizarVendedor(id int, vendedor *models.Vendedor) (*models.Vendedor, error)
	EliminarVendedor(id int) error
	ValidarCredenciales(usuario, password string) (*models.User, error)
	GetClientesPorVendedor() (map[string][]models.Cliente, error)
}

// VentaService con inyección de dependencias
type TestableVentaService struct {
	db DatabaseInterface
}

func NewTestableVentaService(db DatabaseInterface) *TestableVentaService {
	return &TestableVentaService{db: db}
}

func (s *TestableVentaService) CrearVenta(req *models.VentaRequest) (int, error) {
	ctx := context.Background()

	// Validar datos requeridos
	if err := s.validarVentaRequest(req); err != nil {
		return 0, err
	}

	// Obtener ID del vendedor
	vendedorID, err := s.db.GetVendedorID(req.Vendedor)
	if err != nil {
		return 0, fmt.Errorf("vendedor no encontrado: %w", err)
	}

	// Verificar que el vendedor existe
	exists, err := s.db.ExistsVendedor(ctx, vendedorID)
	if err != nil || !exists {
		return 0, fmt.Errorf("vendedor no válido")
	}

	// Crear venta
	ventaID, err := s.db.CrearVenta(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("error al crear venta: %w", err)
	}

	// Crear detalles de venta
	detalles := s.crearDetallesVenta(ventaID, req.Items)
	if err := s.db.CrearDetalleVenta(ctx, detalles); err != nil {
		return 0, fmt.Errorf("error al crear detalles de venta: %w", err)
	}

	return ventaID, nil
}

func (s *TestableVentaService) validarVentaRequest(req *models.VentaRequest) error {
	if req.Vendedor == "" {
		return fmt.Errorf("vendedor es requerido")
	}
	if req.Cliente == "" {
		return fmt.Errorf("cliente es requerido")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("al menos un item es requerido")
	}
	for _, item := range req.Items {
		if item.ProductID <= 0 || item.Cantidad <= 0 || item.Precio <= 0 {
			return fmt.Errorf("item inválido: producto_id=%d, cantidad=%d, precio=%.2f",
				item.ProductID, item.Cantidad, item.Precio)
		}
		// Validar que el producto existe
		exists, err := s.db.ExistsProducto(context.Background(), item.ProductID)
		if err != nil {
			return fmt.Errorf("error al validar producto: %w", err)
		}
		if !exists {
			return fmt.Errorf("producto con ID %d no existe", item.ProductID)
		}
	}
	// Teléfono es opcional, solo validar formato si está presente
	if req.TelefonoCliente != nil && (*req.TelefonoCliente < 10000000 || *req.TelefonoCliente > 999999999) {
		return fmt.Errorf("teléfono debe tener entre 8 y 9 dígitos")
	}
	return nil
}

func (s *TestableVentaService) calcularTotal(items []models.ProductoItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Precio * float64(item.Cantidad)
	}
	return total
}

func (s *TestableVentaService) crearDetallesVenta(ventaID int, items []models.ProductoItem) []models.ProductoItem {
	detalles := make([]models.ProductoItem, len(items))
	for i, item := range items {
		detalles[i] = models.ProductoItem{
			ProductID: item.ProductID,
			Cantidad:  item.Cantidad,
			Precio:    item.Precio,
			Total:     item.Precio * float64(item.Cantidad),
		}
	}
	return detalles
}

func TestVentaService_CrearVenta(t *testing.T) {
	tests := []struct {
		name        string
		request     *models.VentaRequest
		mockSetup   func(*mockDatabase)
		expectError bool
		expectedID  int
	}{
		{
			name: "venta sin teléfono debe ser válida",
			request: &models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 2, Precio: 10.0},
				},
				PaymentMethod: "efectivo",
				Estado:        "pagada",
				TipoEntrega:   "retiro",
			},
			mockSetup: func(m *mockDatabase) {
				m.getVendedorIDFunc = func(nombre string) (int, error) { return 1, nil }
				m.existsVendedorFunc = func(ctx context.Context, id int) (bool, error) { return true, nil }
				m.existsProductoFunc = func(ctx context.Context, id int) (bool, error) { return true, nil }
				m.getOrCreateClienteFunc = func(nombre string) (int, error) { return 1, nil }
				m.crearVentaFunc = func(ctx context.Context, venta *models.VentaRequest) (int, error) { return 1, nil }
				m.crearDetalleVentaFunc = func(ctx context.Context, detalles []models.ProductoItem) error { return nil }
			},
			expectError: false,
			expectedID:  1,
		},
		{
			name: "venta sin vendedor debe fallar",
			request: &models.VentaRequest{
				Vendedor: "",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
			},
			mockSetup:   func(m *mockDatabase) {},
			expectError: true,
			expectedID:  0,
		},
		{
			name: "venta sin cliente debe fallar",
			request: &models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
			},
			mockSetup:   func(m *mockDatabase) {},
			expectError: true,
			expectedID:  0,
		},
		{
			name: "venta sin items debe fallar",
			request: &models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items:    []models.ProductoItem{},
			},
			mockSetup:   func(m *mockDatabase) {},
			expectError: true,
			expectedID:  0,
		},
		{
			name: "vendedor no encontrado debe fallar",
			request: &models.VentaRequest{
				Vendedor: "Vendedor Inexistente",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
			},
			mockSetup: func(m *mockDatabase) {
				m.getVendedorIDFunc = func(nombre string) (int, error) { return 0, fmt.Errorf("vendedor no encontrado") }
			},
			expectError: true,
			expectedID:  0,
		},
		{
			name: "producto inexistente debe fallar",
			request: &models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 999, Cantidad: 1, Precio: 10.0},
				},
			},
			mockSetup: func(m *mockDatabase) {
				m.getVendedorIDFunc = func(nombre string) (int, error) { return 1, nil }
				m.existsVendedorFunc = func(ctx context.Context, id int) (bool, error) { return true, nil }
				m.existsProductoFunc = func(ctx context.Context, id int) (bool, error) { return false, nil }
			},
			expectError: true,
			expectedID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDB := &mockDatabase{}
			tt.mockSetup(mockDB)

			service := NewTestableVentaService(mockDB)

			// Act
			id, err := service.CrearVenta(tt.request)

			// Assert
			if tt.expectError && err == nil {
				t.Error("CrearVenta() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("CrearVenta() unexpected error: %v", err)
			}

			if id != tt.expectedID {
				t.Errorf("CrearVenta() id = %v, want %v", id, tt.expectedID)
			}
		})
	}
}

type TestableProductoService struct {
	db DatabaseInterface
}

func NewTestableProductoService(db DatabaseInterface) *TestableProductoService {
	return &TestableProductoService{db: db}
}

func (s *TestableProductoService) ObtenerProductos() ([]models.Producto, error) {
	return s.db.GetProductos()
}

func (s *TestableProductoService) CrearProducto(req *models.CrearProductoRequest) (int64, error) {
	// Validar
	if req.TipoPizza == "" {
		return 0, fmt.Errorf("tipo de pizza es requerido")
	}
	if req.Precio <= 0 {
		return 0, fmt.Errorf("precio debe ser mayor a 0")
	}

	producto := &models.Producto{
		TipoPizza:   req.TipoPizza,
		Descripcion: req.Descripcion,
		Precio:      req.Precio,
		Activo:      true,
	}

	created, err := s.db.CrearProducto(producto)
	if err != nil {
		return 0, err
	}

	return int64(created.ID), nil
}

func TestProductoService_ObtenerProductos(t *testing.T) {
	// Arrange
	mockDB := &mockDatabase{
		getProductosFunc: func() ([]models.Producto, error) {
			return []models.Producto{
				{ID: 1, TipoPizza: "Margherita", Precio: 10.0},
				{ID: 2, TipoPizza: "Pepperoni", Precio: 12.0},
			}, nil
		},
	}

	service := NewTestableProductoService(mockDB)

	// Act
	productos, err := service.ObtenerProductos()

	// Assert
	if err != nil {
		t.Errorf("ObtenerProductos() error = %v, want nil", err)
	}

	if len(productos) != 2 {
		t.Errorf("ObtenerProductos() len = %v, want 2", len(productos))
	}

	if productos[0].TipoPizza != "Margherita" {
		t.Errorf("ObtenerProductos() first product = %v, want Margherita", productos[0].TipoPizza)
	}
}

func TestProductoService_CrearProducto(t *testing.T) {
	tests := []struct {
		name        string
		request     *models.CrearProductoRequest
		mockSetup   func(*mockDatabase)
		expectError bool
	}{
		{
			name: "producto válido debe crearse exitosamente",
			request: &models.CrearProductoRequest{
				TipoPizza:   "Nueva Pizza",
				Descripcion: "Deliciosa pizza nueva",
				Precio:      15.0,
			},
			mockSetup: func(m *mockDatabase) {
				m.crearProductoFunc = func(producto *models.Producto) (*models.Producto, error) {
					return &models.Producto{ID: 1, TipoPizza: producto.TipoPizza, Precio: producto.Precio}, nil
				}
			},
			expectError: false,
		},
		{
			name: "producto sin tipo debe fallar",
			request: &models.CrearProductoRequest{
				TipoPizza: "",
				Precio:    10.0,
			},
			mockSetup:   func(m *mockDatabase) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDB := &mockDatabase{}
			tt.mockSetup(mockDB)

			service := NewTestableProductoService(mockDB)

			// Act
			id, err := service.CrearProducto(tt.request)

			// Assert
			if tt.expectError && err == nil {
				t.Error("CrearProducto() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("CrearProducto() unexpected error: %v", err)
			}

			if !tt.expectError && id == 0 {
				t.Error("CrearProducto() expected id but got 0")
			}
		})
	}
}
