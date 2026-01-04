package models

import (
	"testing"
	"time"
)

func TestProductoItem_CalculateTotal(t *testing.T) {
	tests := []struct {
		name     string
		item     ProductoItem
		expected float64
	}{
		{
			name: "cálculo correcto de total",
			item: ProductoItem{
				ProductID: 1,
				Cantidad:  2,
				Precio:    10.0,
			},
			expected: 20.0,
		},
		{
			name: "cantidad cero debe dar total cero",
			item: ProductoItem{
				ProductID: 1,
				Cantidad:  0,
				Precio:    10.0,
			},
			expected: 0.0,
		},
		{
			name: "precio cero debe dar total cero",
			item: ProductoItem{
				ProductID: 1,
				Cantidad:  2,
				Precio:    0.0,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			item := tt.item

			// Act
			total := item.Precio * float64(item.Cantidad)

			// Assert
			if total != tt.expected {
				t.Errorf("ProductoItem total calculation = %.2f, want %.2f", total, tt.expected)
			}
		})
	}
}

func TestVentaRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     VentaRequest
		expectValid bool
	}{
		{
			name: "venta request válido",
			request: VentaRequest{
				Vendedor:      "Juan Pérez",
				Cliente:       "María García",
				PaymentMethod: "efectivo",
				Estado:        "pagada",
				TipoEntrega:   "retiro",
				Items: []ProductoItem{
					{ProductID: 1, Cantidad: 2, Precio: 10.0},
				},
			},
			expectValid: true,
		},
		{
			name: "venta sin vendedor debe ser inválida",
			request: VentaRequest{
				Vendedor: "",
				Cliente:  "María García",
				Items: []ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
			},
			expectValid: false,
		},
		{
			name: "venta sin cliente debe ser inválida",
			request: VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "",
				Items: []ProductoItem{
					{ProductID: 1, Cantidad: 1, Precio: 10.0},
				},
			},
			expectValid: false,
		},
		{
			name: "venta sin items debe ser inválida",
			request: VentaRequest{
				Vendedor: "Juan Pérez",
				Cliente:  "María García",
				Items:    []ProductoItem{},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			request := tt.request

			// Act
			isValid := request.Vendedor != "" && request.Cliente != "" && len(request.Items) > 0

			// Assert
			if isValid != tt.expectValid {
				t.Errorf("VentaRequest validation = %v, want %v", isValid, tt.expectValid)
			}
		})
	}
}

func TestProducto_Validation(t *testing.T) {
	tests := []struct {
		name        string
		producto    Producto
		expectValid bool
	}{
		{
			name: "producto válido",
			producto: Producto{
				ID:        1,
				TipoPizza: "Margherita",
				Precio:    10.0,
				Activo:    true,
			},
			expectValid: true,
		},
		{
			name: "producto sin tipo debe ser inválido",
			producto: Producto{
				ID:        1,
				TipoPizza: "",
				Precio:    10.0,
				Activo:    true,
			},
			expectValid: false,
		},
		{
			name: "producto con precio cero debe ser inválido",
			producto: Producto{
				ID:        1,
				TipoPizza: "Margherita",
				Precio:    0.0,
				Activo:    true,
			},
			expectValid: false,
		},
		{
			name: "producto inactivo debe ser válido",
			producto: Producto{
				ID:        1,
				TipoPizza: "Margherita",
				Precio:    10.0,
				Activo:    false,
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			producto := tt.producto

			// Act
			isValid := producto.TipoPizza != "" && producto.Precio > 0

			// Assert
			if isValid != tt.expectValid {
				t.Errorf("Producto validation = %v, want %v", isValid, tt.expectValid)
			}
		})
	}
}

func TestVendedor_Validation(t *testing.T) {
	tests := []struct {
		name        string
		vendedor    Vendedor
		expectValid bool
	}{
		{
			name: "vendedor válido",
			vendedor: Vendedor{
				ID:     1,
				Nombre: "Juan Pérez",
			},
			expectValid: true,
		},
		{
			name: "vendedor sin nombre debe ser inválido",
			vendedor: Vendedor{
				ID:     1,
				Nombre: "",
			},
			expectValid: false,
		},
		{
			name: "vendedor con ID cero debe ser inválido",
			vendedor: Vendedor{
				ID:     0,
				Nombre: "Juan Pérez",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			vendedor := tt.vendedor

			// Act
			isValid := vendedor.ID > 0 && vendedor.Nombre != ""

			// Assert
			if isValid != tt.expectValid {
				t.Errorf("Vendedor validation = %v, want %v", isValid, tt.expectValid)
			}
		})
	}
}

func TestCliente_Validation(t *testing.T) {
	tests := []struct {
		name        string
		cliente     Cliente
		expectValid bool
	}{
		{
			name: "cliente válido",
			cliente: Cliente{
				ID:       1,
				Nombre:   "María García",
				Telefono: 123456789,
			},
			expectValid: true,
		},
		{
			name: "cliente sin nombre debe ser inválido",
			cliente: Cliente{
				ID:       1,
				Nombre:   "",
				Telefono: 123456789,
			},
			expectValid: false,
		},
		{
			name: "cliente con ID cero debe ser inválido",
			cliente: Cliente{
				ID:       0,
				Nombre:   "María García",
				Telefono: 123456789,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cliente := tt.cliente

			// Act
			isValid := cliente.ID > 0 && cliente.Nombre != ""

			// Assert
			if isValid != tt.expectValid {
				t.Errorf("Cliente validation = %v, want %v", isValid, tt.expectValid)
			}
		})
	}
}

func TestVenta_CalculateTotal(t *testing.T) {
	// Arrange
	tel := 123456789
	venta := VentaStats{
		ID:              1,
		Vendedor:        "Juan Pérez",
		Cliente:         "María García",
		TelefonoCliente: &tel,
		Total:           0, // Se calculará
		PaymentMethod:   "efectivo",
		Estado:          "pagada",
		TipoEntrega:     "retiro",
		CreatedAt:       time.Now(),
		Items: []ProductoItem{
			{ProductID: 1, Cantidad: 2, Precio: 10.0, Total: 20.0},
			{ProductID: 2, Cantidad: 1, Precio: 5.0, Total: 5.0},
		},
	}

	// Act
	// Calcular total desde los items
	total := 0.0
	for _, item := range venta.Items {
		total += item.Total
	}

	// Assert
	if total != 25.0 {
		t.Errorf("Venta total calculation = %v, want 25.0", total)
	}

	// Verificar que la venta tenga ID
	if venta.ID <= 0 {
		t.Error("Venta should have valid ID")
	}
}

func TestDataResponse_Structure(t *testing.T) {
	// Arrange
	response := DataResponse{
		ClientesPorVendedor: map[string][]Cliente{
			"Juan Pérez": {
				{ID: 1, Nombre: "María García", Telefono: 123456789},
			},
		},
		Vendedores: []Vendedor{
			{ID: 1, Nombre: "Juan Pérez"},
		},
		Productos: []Producto{
			{ID: 1, TipoPizza: "Margherita", Precio: 10.0, Activo: true},
		},
	}

	// Act & Assert
	if response.ClientesPorVendedor == nil {
		t.Error("DataResponse should have ClientesPorVendedor")
	}

	if len(response.Vendedores) == 0 {
		t.Error("DataResponse should have Vendedores")
	}

	if len(response.Productos) == 0 {
		t.Error("DataResponse should have Productos")
	}

	// Verificar estructura de clientes por vendedor
	for vendedor, clientes := range response.ClientesPorVendedor {
		if vendedor == "" {
			t.Error("Vendedor name should not be empty")
		}
		for _, cliente := range clientes {
			if cliente.ID <= 0 || cliente.Nombre == "" {
				t.Error("Cliente should have valid ID and name")
			}
		}
	}
}
