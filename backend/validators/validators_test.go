package validators

import (
	"testing"

	"pizzas-ecos/models"
)

func TestValidateRequest_Add(t *testing.T) {
	// Arrange
	vr := &ValidateRequest{}

	// Act
	vr.Add("campo1", "mensaje1")
	vr.Add("campo2", "mensaje2")

	// Assert
	if len(vr.Errors) != 2 {
		t.Errorf("Add() len = %v, want 2", len(vr.Errors))
	}

	if vr.Errors[0].Field != "campo1" || vr.Errors[0].Message != "mensaje1" {
		t.Errorf("Add() first error = %v, want {campo1, mensaje1}", vr.Errors[0])
	}
}

func TestValidateRequest_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		errors   []ValidationError
		expected bool
	}{
		{
			name:     "sin errores debe ser válido",
			errors:   []ValidationError{},
			expected: true,
		},
		{
			name: "con errores debe ser inválido",
			errors: []ValidationError{
				{Field: "campo1", Message: "mensaje1"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			vr := &ValidateRequest{Errors: tt.errors}

			// Act
			result := vr.IsValid()

			// Assert
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateRequest_GetMessage(t *testing.T) {
	tests := []struct {
		name     string
		errors   []ValidationError
		expected string
	}{
		{
			name:     "sin errores debe retornar string vacío",
			errors:   []ValidationError{},
			expected: "",
		},
		{
			name: "con un error debe formatear correctamente",
			errors: []ValidationError{
				{Field: "campo1", Message: "mensaje1"},
			},
			expected: "campo1: mensaje1",
		},
		{
			name: "con múltiples errores debe formatear correctamente",
			errors: []ValidationError{
				{Field: "campo1", Message: "mensaje1"},
				{Field: "campo2", Message: "mensaje2"},
			},
			expected: "campo1: mensaje1; campo2: mensaje2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			vr := &ValidateRequest{Errors: tt.errors}

			// Act
			result := vr.GetMessage()

			// Assert
			if result != tt.expected {
				t.Errorf("GetMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateVentaRequest(t *testing.T) {
	tests := []struct {
		name           string
		vendedorID     int
		clienteID      int
		monto          float64
		items          int
		expectValid    bool
		expectedErrors int
	}{
		{
			name:           "venta válida debe pasar validación",
			vendedorID:     1,
			clienteID:      1,
			monto:          10.0,
			items:          1,
			expectValid:    true,
			expectedErrors: 0,
		},
		{
			name:           "vendedor inválido debe fallar",
			vendedorID:     0,
			clienteID:      1,
			monto:          10.0,
			items:          1,
			expectValid:    false,
			expectedErrors: 1,
		},
		{
			name:           "cliente inválido debe fallar",
			vendedorID:     1,
			clienteID:      0,
			monto:          10.0,
			items:          1,
			expectValid:    false,
			expectedErrors: 1,
		},
		{
			name:           "monto negativo debe fallar",
			vendedorID:     1,
			clienteID:      1,
			monto:          -10.0,
			items:          1,
			expectValid:    false,
			expectedErrors: 1,
		},
		{
			name:           "sin items debe fallar",
			vendedorID:     1,
			clienteID:      1,
			monto:          10.0,
			items:          0,
			expectValid:    false,
			expectedErrors: 1,
		},
		{
			name:           "múltiples errores deben acumularse",
			vendedorID:     0,
			clienteID:      0,
			monto:          -5.0,
			items:          0,
			expectValid:    false,
			expectedErrors: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := ValidateVentaRequest(tt.vendedorID, tt.clienteID, tt.monto, tt.items)

			// Assert
			if result.IsValid() != tt.expectValid {
				t.Errorf("ValidateVentaRequest() IsValid = %v, want %v", result.IsValid(), tt.expectValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateVentaRequest() errors count = %v, want %v", len(result.Errors), tt.expectedErrors)
			}
		})
	}
}

func TestValidateProductoRequest(t *testing.T) {
	tests := []struct {
		name           string
		tipoPizza      string
		precio         float64
		expectValid    bool
		expectedErrors int
	}{
		{
			name:           "producto válido debe pasar validación",
			tipoPizza:      "Margherita",
			precio:         10.0,
			expectValid:    true,
			expectedErrors: 0,
		},
		{
			name:           "tipo de pizza vacío debe fallar",
			tipoPizza:      "",
			precio:         10.0,
			expectValid:    false,
			expectedErrors: 2,
		},
		{
			name:           "precio cero debe fallar",
			tipoPizza:      "Margherita",
			precio:         0,
			expectValid:    false,
			expectedErrors: 1,
		},
		{
			name:           "precio negativo debe fallar",
			tipoPizza:      "Margherita",
			precio:         -5.0,
			expectValid:    false,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := ValidateProductoRequest(tt.tipoPizza, tt.precio)

			// Assert
			if result.IsValid() != tt.expectValid {
				t.Errorf("ValidateProductoRequest() IsValid = %v, want %v", result.IsValid(), tt.expectValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateProductoRequest() errors count = %v, want %v", len(result.Errors), tt.expectedErrors)
			}
		})
	}
}

func TestValidateVendedorRequest(t *testing.T) {
	tests := []struct {
		name           string
		nombre         string
		expectValid    bool
		expectedErrors int
	}{
		{
			name:           "vendedor válido debe pasar validación",
			nombre:         "Juan Pérez",
			expectValid:    true,
			expectedErrors: 0,
		},
		{
			name:           "nombre vacío debe fallar",
			nombre:         "",
			expectValid:    false,
			expectedErrors: 2,
		},
		{
			name:           "nombre solo espacios debe fallar",
			nombre:         "   ",
			expectValid:    false,
			expectedErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			result := ValidateVendedorRequest(tt.nombre)

			// Assert
			if result.IsValid() != tt.expectValid {
				t.Errorf("ValidateVendedorRequest() IsValid = %v, want %v", result.IsValid(), tt.expectValid)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateVendedorRequest() errors count = %v, want %v", len(result.Errors), tt.expectedErrors)
			}
		})
	}
}

func TestValidateVentaRequestCompleto(t *testing.T) {
	tests := []struct {
		name        string
		request     interface{}
		expectValid bool
	}{
		{
			name: "venta válida debe pasar validación",
			request: &models.VentaRequest{
				Vendedor:        "Juan Pérez",
				Cliente:         "María García",
				TelefonoCliente: 123456789,
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 2, Precio: 10.0},
				},
				PaymentMethod: "efectivo",
				Estado:        "sin_pagar",
				TipoEntrega:   "retiro",
			},
			expectValid: true,
		},
		{
			name: "venta sin teléfono debe ser válida",
			request: &models.VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 2, Precio: 10.0},
				},
				PaymentMethod: "efectivo",
				Estado:        "sin_pagar",
				TipoEntrega:   "retiro",
			},
			expectValid: true,
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
			expectValid: false,
		},
		{
			name: "venta con teléfono inválido debe fallar",
			request: &models.VentaRequest{
				Vendedor:        "Juan Pérez",
				Cliente:         "María García",
				TelefonoCliente: 1234567, // 7 dígitos - inválido
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
			},
			expectValid: false,
		},
		{
			name: "venta con payment method inválido debe fallar",
			request: &models.VentaRequest{
				Vendedor:      "Juan Pérez",
				Cliente:       "María García",
				PaymentMethod: "bitcoin",
				Items: []models.ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := ValidateVentaRequestCompleto(tt.request)

			// Assert
			if result.IsValid() != tt.expectValid {
				t.Errorf("ValidateVentaRequestCompleto() valid = %v, want %v", result.IsValid(), tt.expectValid)
				if !result.IsValid() {
					t.Logf("Errors: %s", result.GetMessage())
				}
			}
		})
	}
}

func TestValidateProductoRequestCompleto(t *testing.T) {
	tests := []struct {
		name        string
		request     interface{}
		expectValid bool
	}{
		{
			name: "producto válido debe pasar validación",
			request: &models.CrearProductoRequest{
				TipoPizza:   "Margherita",
				Descripcion: "Deliciosa pizza",
				Precio:      15.0,
			},
			expectValid: true,
		},
		{
			name: "producto sin tipo debe fallar",
			request: &models.CrearProductoRequest{
				TipoPizza: "",
				Precio:    15.0,
			},
			expectValid: false,
		},
		{
			name: "producto con precio negativo debe fallar",
			request: &models.CrearProductoRequest{
				TipoPizza: "Margherita",
				Precio:    -5.0,
			},
			expectValid: false,
		},
		{
			name: "producto con precio demasiado alto debe fallar",
			request: &models.CrearProductoRequest{
				TipoPizza: "Margherita",
				Precio:    600.0,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := ValidateProductoRequestCompleto(tt.request)

			// Assert
			if result.IsValid() != tt.expectValid {
				t.Errorf("ValidateProductoRequestCompleto() valid = %v, want %v", result.IsValid(), tt.expectValid)
				if !result.IsValid() {
					t.Logf("Errors: %s", result.GetMessage())
				}
			}
		})
	}
}

func TestValidateVendedorRequestCompleto(t *testing.T) {
	tests := []struct {
		name           string
		nombre         string
		expectValid    bool
		expectedErrors int
	}{
		{
			name:           "vendedor válido debe pasar validación",
			nombre:         "Juan Pérez",
			expectValid:    true,
			expectedErrors: 0,
		},
		{
			name:           "vendedor sin nombre debe fallar",
			nombre:         "",
			expectValid:    false,
			expectedErrors: 1,
		},
		{
			name:           "vendedor con nombre muy corto debe fallar",
			nombre:         "A",
			expectValid:    false,
			expectedErrors: 1,
		},
		{
			name:           "vendedor con caracteres inválidos debe fallar",
			nombre:         "Juan@Pérez",
			expectValid:    false,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := ValidateVendedorRequestCompleto(tt.nombre)

			// Assert
			if result.IsValid() != tt.expectValid {
				t.Errorf("ValidateVendedorRequestCompleto() valid = %v, want %v", result.IsValid(), tt.expectValid)
			}
			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidateVendedorRequestCompleto() errors count = %v, want %v", len(result.Errors), tt.expectedErrors)
			}
		})
	}
}
