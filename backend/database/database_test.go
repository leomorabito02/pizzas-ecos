package database

import (
	"testing"
)

// TestGetClientesPorVendedor verifica que la función retorne clientes agrupados por vendedor correctamente
func TestGetClientesPorVendedor(t *testing.T) {
	// Skip test if database is not available (for CI/CD or when running unit tests)
	if testing.Short() {
		t.Skip("Skipping database integration test")
	}

	// Este test requiere una base de datos con datos de prueba
	// Por ahora solo verificamos que la función no retorne error y retorne un map válido

	// Arrange
	// (En un test real, aquí configuraríamos una base de datos de prueba)

	// Act
	clientesPorVendedor, err := GetClientesPorVendedor()

	// Assert
	if err != nil {
		t.Fatalf("GetClientesPorVendedor() retornó error: %v", err)
	}

	// Verificar que el resultado no sea nil
	if clientesPorVendedor == nil {
		t.Error("GetClientesPorVendedor() retornó nil, se esperaba un map")
	}

	// Verificar que sea un map
	if clientesPorVendedor != nil {
		// Log para ver qué datos retorna (útil para debugging)
		t.Logf("Clientes por vendedor: %+v", clientesPorVendedor)

		// Verificar que las claves sean strings (nombres de vendedores)
		for vendedor, clientes := range clientesPorVendedor {
			if vendedor == "" {
				t.Error("GetClientesPorVendedor() retornó vendedor con nombre vacío")
			}

			// Verificar que los clientes tengan estructura válida
			for _, cliente := range clientes {
				if cliente.ID <= 0 {
					t.Errorf("GetClientesPorVendedor() cliente con ID inválido: %d", cliente.ID)
				}
				if cliente.Nombre == "" {
					t.Error("GetClientesPorVendedor() retornó cliente con nombre vacío")
				}
			}
		}
	}
}

// TestGetVendedorID verifica la obtención de ID de vendedor por nombre
func TestGetVendedorID(t *testing.T) {
	// Skip test if database is not available (for CI/CD or when running unit tests)
	if testing.Short() {
		t.Skip("Skipping database integration test")
	}
	tests := []struct {
		name           string
		vendedorNombre string
		expectError    bool
		expectedID     int
	}{
		{
			name:           "vendedor existente debe retornar ID válido",
			vendedorNombre: "Juan Pérez",
			expectError:    false,
			expectedID:     1, // Este valor depende de los datos de prueba
		},
		{
			name:           "vendedor inexistente debe retornar error",
			vendedorNombre: "Vendedor Inexistente",
			expectError:    true,
			expectedID:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// (Configuración de base de datos de prueba)

			// Act
			id, err := GetVendedorID(tt.vendedorNombre)

			// Assert
			if tt.expectError && err == nil {
				t.Error("GetVendedorID() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("GetVendedorID() unexpected error: %v", err)
			}

			if !tt.expectError && id != tt.expectedID {
				t.Errorf("GetVendedorID() id = %v, want %v", id, tt.expectedID)
			}
		})
	}
}

// TestGetOrCreateCliente verifica la creación o obtención de cliente
func TestGetOrCreateCliente(t *testing.T) {
	// Skip test if database is not available (for CI/CD or when running unit tests)
	if testing.Short() {
		t.Skip("Skipping database integration test")
	}
	tests := []struct {
		name          string
		clienteNombre string
		expectError   bool
		expectedID    int
	}{
		{
			name:          "cliente existente debe retornar ID",
			clienteNombre: "María García",
			expectError:   false,
			expectedID:    1, // Depende de datos de prueba
		},
		{
			name:          "cliente nuevo debe crearse y retornar ID",
			clienteNombre: "Nuevo Cliente",
			expectError:   false,
			expectedID:    0, // No podemos predecir el ID exacto
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// (Configuración de base de datos de prueba)

			// Act
			id, err := GetOrCreateCliente(tt.clienteNombre)

			// Assert
			if tt.expectError && err == nil {
				t.Error("GetOrCreateCliente() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("GetOrCreateCliente() unexpected error: %v", err)
			}

			if !tt.expectError && id <= 0 {
				t.Errorf("GetOrCreateCliente() returned invalid id: %v", id)
			}
		})
	}
}

// TestGetProductos verifica la obtención de productos
func TestGetProductos(t *testing.T) {
	// Skip test if database is not available (for CI/CD or when running unit tests)
	if testing.Short() {
		t.Skip("Skipping database integration test")
	}
	// Arrange
	// (Configuración de base de datos de prueba)

	// Act
	productos, err := GetProductos()

	// Assert
	if err != nil {
		t.Fatalf("GetProductos() retornó error: %v", err)
	}

	// Verificar que sea un slice
	if productos == nil {
		t.Error("GetProductos() retornó nil")
	}

	// Verificar estructura de productos
	for _, producto := range productos {
		if producto.ID <= 0 {
			t.Errorf("GetProductos() producto con ID inválido: %d", producto.ID)
		}
		if producto.TipoPizza == "" {
			t.Error("GetProductos() retornó producto con tipo vacío")
		}
		if producto.Precio <= 0 {
			t.Errorf("GetProductos() producto con precio inválido: %.2f", producto.Precio)
		}
	}

	t.Logf("GetProductos() retornó %d productos", len(productos))
}
