package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pizzas-ecos/models"
)

// TestVentaService es una versión de test que no llama a database
type TestVentaService struct {
	crearVentaFunc func(req *models.VentaRequest) (int, error)
}

func (s *TestVentaService) CrearVenta(req *models.VentaRequest) (int, error) {
	if s.crearVentaFunc != nil {
		return s.crearVentaFunc(req)
	}
	return 1, nil
}

func (s *TestVentaService) ActualizarVenta(ventaID int, estado, paymentMethod, tipoEntrega string, productosEliminar []int, productos []map[string]interface{}) error {
	return nil
}

func (s *TestVentaService) ObtenerEstadisticas() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *TestVentaService) ObtenerTodasVentas() ([]models.VentaStats, error) {
	return []models.VentaStats{}, nil
}

// TestProductoService es una versión de test que no llama a database
type TestProductoService struct {
	obtenerProductosFunc   func() ([]models.Producto, error)
	crearProductoFunc      func(req *models.CrearProductoRequest) (int64, error)
	actualizarProductoFunc func(id int, req *models.ActualizarProductoRequest) error
	eliminarProductoFunc   func(id int) error
}

func (s *TestProductoService) ObtenerProductos() ([]models.Producto, error) {
	if s.obtenerProductosFunc != nil {
		return s.obtenerProductosFunc()
	}
	return []models.Producto{{ID: 1, TipoPizza: "Margherita", Precio: 10.0}}, nil
}

func (s *TestProductoService) CrearProducto(req *models.CrearProductoRequest) (int64, error) {
	if s.crearProductoFunc != nil {
		return s.crearProductoFunc(req)
	}
	return 1, nil
}

func (s *TestProductoService) ActualizarProducto(id int, req *models.ActualizarProductoRequest) error {
	if s.actualizarProductoFunc != nil {
		return s.actualizarProductoFunc(id, req)
	}
	return nil
}

func (s *TestProductoService) EliminarProducto(id int) error {
	if s.eliminarProductoFunc != nil {
		return s.eliminarProductoFunc(id)
	}
	return nil
}

// TestVendedorService es una versión de test que no llama a database
type TestVendedorService struct {
	obtenerVendedoresFunc  func() ([]models.Vendedor, error)
	crearVendedorFunc      func(nombre string) (int64, error)
	actualizarVendedorFunc func(id int, nombre string) error
	eliminarVendedorFunc   func(id int) error
}

func (s *TestVendedorService) ObtenerVendedores() ([]models.Vendedor, error) {
	if s.obtenerVendedoresFunc != nil {
		return s.obtenerVendedoresFunc()
	}
	return []models.Vendedor{{ID: 1, Nombre: "Juan Pérez"}}, nil
}

func (s *TestVendedorService) CrearVendedor(nombre string) (int64, error) {
	if s.crearVendedorFunc != nil {
		return s.crearVendedorFunc(nombre)
	}
	return 1, nil
}

func (s *TestVendedorService) ActualizarVendedor(id int, nombre string) error {
	if s.actualizarVendedorFunc != nil {
		return s.actualizarVendedorFunc(id, nombre)
	}
	return nil
}

func (s *TestVendedorService) EliminarVendedor(id int) error {
	if s.eliminarVendedorFunc != nil {
		return s.eliminarVendedorFunc(id)
	}
	return nil
}

// TestAuthService es una versión de test que no llama a database
type TestAuthService struct {
	autenticarUsuarioFunc func(username, password string) (*models.User, error)
}

func (s *TestAuthService) AutenticarUsuario(username, password string) (*models.User, error) {
	if s.autenticarUsuarioFunc != nil {
		return s.autenticarUsuarioFunc(username, password)
	}
	return &models.User{ID: 1, Username: "Admin", Rol: "admin"}, nil
}

// Test helpers
func createTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody bytes.Buffer
	if body != nil {
		json.NewEncoder(&reqBody).Encode(body)
	}
	req, _ := http.NewRequest(method, url, &reqBody)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestVentaController_CrearVenta(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.VentaRequest
		mockSetup      func(*TestVentaService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "venta válida debe crearse exitosamente",
			requestBody: models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 2, Precio: 10.0},
				},
				PaymentMethod:   "efectivo",
				Estado:          "pagada",
				TipoEntrega:     "retiro",
				TelefonoCliente: &[]int{12345678}[0],
			},
			mockSetup: func(m *TestVentaService) {
				m.crearVentaFunc = func(req *models.VentaRequest) (int, error) {
					return 1, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "venta sin teléfono debe ser válida",
			requestBody: models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 2, Precio: 10.0},
				},
				PaymentMethod: "efectivo",
				Estado:        "pagada",
				TipoEntrega:   "retiro",
			},
			mockSetup: func(m *TestVentaService) {
				m.crearVentaFunc = func(req *models.VentaRequest) (int, error) {
					return 1, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "venta sin vendedor debe fallar",
			requestBody: models.VentaRequest{
				Vendedor: "",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
				PaymentMethod: "efectivo",
				Estado:        "pagada",
				TipoEntrega:   "retiro",
			},
			mockSetup:      func(m *TestVentaService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "venta sin cliente debe fallar",
			requestBody: models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
				PaymentMethod: "efectivo",
				Estado:        "pagada",
				TipoEntrega:   "retiro",
			},
			mockSetup:      func(m *TestVentaService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "venta sin items debe fallar",
			requestBody: models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items:    []models.ProductoItem{},
			},
			mockSetup:      func(m *TestVentaService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := &TestVentaService{}
			tt.mockSetup(mockService)

			controller := &VentaController{
				ventaService: mockService,
			}

			req := createTestRequest("POST", "/api/v1/ventas", tt.requestBody)
			w := httptest.NewRecorder()

			// Act
			controller.CrearVenta(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("CrearVenta() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestProductoController_Listar(t *testing.T) {
	// Arrange
	mockService := &TestProductoService{
		obtenerProductosFunc: func() ([]models.Producto, error) {
			return []models.Producto{
				{ID: 1, TipoPizza: "Margherita", Precio: 10.0},
				{ID: 2, TipoPizza: "Pepperoni", Precio: 12.0},
			}, nil
		},
	}

	controller := &ProductoController{
		productoService: mockService,
	}

	req := createTestRequest("GET", "/api/v1/productos", nil)
	w := httptest.NewRecorder()

	// Act
	controller.Listar(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Listar() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response struct {
		Status  int               `json:"status"`
		Data    []models.Producto `json:"data"`
		Message string            `json:"message"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Errorf("Listar() returned %d products, want 2", len(response.Data))
	}
}

func TestProductoController_CrearProducto(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.CrearProductoRequest
		mockSetup      func(*TestProductoService)
		expectedStatus int
	}{
		{
			name: "producto válido debe crearse exitosamente",
			requestBody: models.CrearProductoRequest{
				TipoPizza:   "Nueva Pizza",
				Descripcion: "Deliciosa pizza nueva",
				Precio:      15.0,
			},
			mockSetup: func(m *TestProductoService) {
				m.crearProductoFunc = func(req *models.CrearProductoRequest) (int64, error) {
					return 1, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "producto sin tipo debe fallar",
			requestBody: models.CrearProductoRequest{
				TipoPizza: "",
				Precio:    10.0,
			},
			mockSetup:      func(m *TestProductoService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := &TestProductoService{}
			tt.mockSetup(mockService)

			controller := &ProductoController{
				productoService: mockService,
			}

			req := createTestRequest("POST", "/api/v1/productos", tt.requestBody)
			w := httptest.NewRecorder()

			// Act
			controller.Crear(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("CrearProducto() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestVendedorController_Listar(t *testing.T) {
	// Arrange
	mockService := &TestVendedorService{
		obtenerVendedoresFunc: func() ([]models.Vendedor, error) {
			return []models.Vendedor{
				{ID: 1, Nombre: "Juan Pérez"},
				{ID: 2, Nombre: "María García"},
			}, nil
		},
	}

	controller := &VendedorController{
		vendedorService: mockService,
	}

	req := createTestRequest("GET", "/api/v1/vendedores", nil)
	w := httptest.NewRecorder()

	// Act
	controller.Listar(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Listar() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response struct {
		Status  int               `json:"status"`
		Data    []models.Vendedor `json:"data"`
		Message string            `json:"message"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Errorf("Listar() returned %d vendedores, want 2", len(response.Data))
	}
}

func TestAuthController_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.LoginRequest
		mockSetup      func(*TestAuthService)
		expectedStatus int
	}{
		{
			name: "login válido debe retornar token",
			requestBody: models.LoginRequest{
				Username: "admin",
				Password: "password123",
			},
			mockSetup: func(m *TestAuthService) {
				m.autenticarUsuarioFunc = func(username, password string) (*models.User, error) {
					return &models.User{
						Username: "Admin",
						Rol:      "admin",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "login con credenciales inválidas debe fallar",
			requestBody: models.LoginRequest{
				Username: "admin",
				Password: "wrongpassword",
			},
			mockSetup: func(m *TestAuthService) {
				m.autenticarUsuarioFunc = func(username, password string) (*models.User, error) {
					return nil, fmt.Errorf("credenciales inválidas")
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := &TestAuthService{}
			tt.mockSetup(mockService)

			controller := &AuthController{
				authService: mockService,
			}

			req := createTestRequest("POST", "/api/v1/auth/login", tt.requestBody)
			w := httptest.NewRecorder()

			// Act
			controller.Login(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
