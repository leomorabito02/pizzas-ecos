package validators

import (
	"fmt"
	"strings"

	"pizzas-ecos/models"
)

// ValidationError contiene errores de validación
type ValidationError struct {
	Field   string
	Message string
}

// ValidateRequest valida un request genérico
type ValidateRequest struct {
	Errors []ValidationError
}

// Add agrega un error de validación
func (vr *ValidateRequest) Add(field, message string) {
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// IsValid retorna true si no hay errores
func (vr *ValidateRequest) IsValid() bool {
	return len(vr.Errors) == 0
}

// GetMessage retorna mensaje de error formateado
func (vr *ValidateRequest) GetMessage() string {
	if len(vr.Errors) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range vr.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// ValidateVentaRequest valida request de creación de venta
func ValidateVentaRequest(vendedorID, clienteID int, monto float64, items int) *ValidateRequest {
	v := &ValidateRequest{}

	if vendedorID <= 0 {
		v.Add("vendedorID", "Vendedor inválido")
	}
	if clienteID <= 0 {
		v.Add("clienteID", "Cliente inválido")
	}
	if monto <= 0 {
		v.Add("monto", "Monto debe ser mayor a 0")
	}
	if items <= 0 {
		v.Add("items", "Debe haber al menos 1 producto")
	}

	return v
}

// ValidateProductoRequest valida request de producto
func ValidateProductoRequest(nombre string, precio float64) *ValidateRequest {
	v := &ValidateRequest{}

	if strings.TrimSpace(nombre) == "" {
		v.Add("nombre", "Nombre requerido")
	}
	if len(strings.TrimSpace(nombre)) < 3 {
		v.Add("nombre", "Nombre debe tener al menos 3 caracteres")
	}
	if precio <= 0 {
		v.Add("precio", "Precio debe ser mayor a 0")
	}

	return v
}

// ValidateVendedorRequest valida request de vendedor
func ValidateVendedorRequest(nombre string) *ValidateRequest {
	v := &ValidateRequest{}

	if strings.TrimSpace(nombre) == "" {
		v.Add("nombre", "Nombre requerido")
	}
	if len(strings.TrimSpace(nombre)) < 2 {
		v.Add("nombre", "Nombre debe tener al menos 2 caracteres")
	}

	return v
}

// ValidateLoginRequest valida request de login
func ValidateLoginRequest(username, password string) *ValidateRequest {
	v := &ValidateRequest{}

	if strings.TrimSpace(username) == "" {
		v.Add("username", "Usuario requerido")
	}
	if strings.TrimSpace(password) == "" {
		v.Add("password", "Contraseña requerida")
	}
	if len(strings.TrimSpace(password)) < 4 {
		v.Add("password", "Contraseña debe tener al menos 4 caracteres")
	}

	return v
}

// ValidateID valida que un ID sea válido
func ValidateID(id interface{}) *ValidateRequest {
	v := &ValidateRequest{}

	switch val := id.(type) {
	case int:
		if val <= 0 {
			v.Add("id", "ID inválido")
		}
	case string:
		if strings.TrimSpace(val) == "" {
			v.Add("id", "ID requerido")
		}
	}

	return v
}

// ValidateVentaRequestCompleto valida una solicitud completa de venta
func ValidateVentaRequestCompleto(req interface{}) *ValidateRequest {
	v := &ValidateRequest{}

	// Type assertion para VentaRequest
	ventaReq, ok := req.(*models.VentaRequest)
	if !ok {
		v.Add("request", "Tipo de request inválido")
		return v
	}

	// Validar vendedor
	if strings.TrimSpace(ventaReq.Vendedor) == "" {
		v.Add("vendedor", "Vendedor es requerido")
	} else if len(strings.TrimSpace(ventaReq.Vendedor)) < 2 {
		v.Add("vendedor", "Nombre de vendedor debe tener al menos 2 caracteres")
	} else if len(ventaReq.Vendedor) > 100 {
		v.Add("vendedor", "Nombre de vendedor demasiado largo (máximo 100 caracteres)")
	}

	// Validar cliente
	if strings.TrimSpace(ventaReq.Cliente) == "" {
		v.Add("cliente", "Cliente es requerido")
	} else if len(strings.TrimSpace(ventaReq.Cliente)) < 2 {
		v.Add("cliente", "Nombre de cliente debe tener al menos 2 caracteres")
	} else if len(ventaReq.Cliente) > 100 {
		v.Add("cliente", "Nombre de cliente demasiado largo (máximo 100 caracteres)")
	}

	// Validar teléfono (opcional - 0 significa vacío)
	if ventaReq.TelefonoCliente != 0 {
		if ventaReq.TelefonoCliente < 10 || ventaReq.TelefonoCliente > 999999999999999 {
			v.Add("telefono_cliente", "Teléfono debe tener entre 2 y 15 dígitos")
		}
	}

	// Validar items
	if len(ventaReq.Items) == 0 {
		v.Add("items", "Al menos un producto es requerido")
	} else if len(ventaReq.Items) > 50 {
		v.Add("items", "Demasiados items (máximo 50)")
	} else {
		// Validar cada item
		for i, item := range ventaReq.Items {
			if item.ProductID <= 0 {
				v.Add(fmt.Sprintf("items[%d].product_id", i), "ID de producto inválido")
			}
			if item.Cantidad <= 0 {
				v.Add(fmt.Sprintf("items[%d].cantidad", i), "Cantidad debe ser mayor a 0")
			} else if item.Cantidad > 100 {
				v.Add(fmt.Sprintf("items[%d].cantidad", i), "Cantidad demasiado grande (máximo 100)")
			}
			if item.Precio < 0 {
				v.Add(fmt.Sprintf("items[%d].precio", i), "Precio no puede ser negativo")
			}
		}
	}

	// Validar payment method
	if strings.TrimSpace(ventaReq.PaymentMethod) == "" {
		v.Add("payment_method", "Método de pago es requerido")
	} else {
		validPayments := []string{"efectivo", "tarjeta", "transferencia", "qr"}
		if !contains(validPayments, strings.ToLower(ventaReq.PaymentMethod)) {
			v.Add("payment_method", "Método de pago inválido (debe ser: efectivo, tarjeta, transferencia, qr)")
		}
	}

	// Validar estado
	if strings.TrimSpace(ventaReq.Estado) == "" {
		ventaReq.Estado = "sin_pagar" // Default
	} else {
		// Normalizar el estado
		estadoNormalizado := strings.ToLower(strings.TrimSpace(ventaReq.Estado))
		validEstados := []string{"sin_pagar", "pagada", "entregada", "cancelada"}
		if !contains(validEstados, estadoNormalizado) {
			v.Add("estado", "Estado inválido (debe ser: sin_pagar, pagada, entregada, cancelada)")
		}
	}

	// Validar tipo de entrega
	if strings.TrimSpace(ventaReq.TipoEntrega) == "" {
		ventaReq.TipoEntrega = "retiro" // Default
	} else {
		validTipos := []string{"retiro", "envio", "delivery"}
		if !contains(validTipos, strings.ToLower(ventaReq.TipoEntrega)) {
			v.Add("tipo_entrega", "Tipo de entrega inválido (debe ser: retiro, envio, delivery)")
		}
	}

	return v
}

// ValidateProductoRequestCompleto valida una solicitud completa de producto
func ValidateProductoRequestCompleto(req interface{}) *ValidateRequest {
	v := &ValidateRequest{}

	// Type assertion para CrearProductoRequest
	prodReq, ok := req.(*models.CrearProductoRequest)
	if !ok {
		v.Add("request", "Tipo de request inválido")
		return v
	}

	// Validar tipo de pizza
	if strings.TrimSpace(prodReq.TipoPizza) == "" {
		v.Add("tipo_pizza", "Tipo de pizza es requerido")
	} else if len(strings.TrimSpace(prodReq.TipoPizza)) < 2 {
		v.Add("tipo_pizza", "Tipo de pizza debe tener al menos 2 caracteres")
	} else if len(prodReq.TipoPizza) > 50 {
		v.Add("tipo_pizza", "Tipo de pizza demasiado largo (máximo 50 caracteres)")
	}

	// Validar descripción
	if len(strings.TrimSpace(prodReq.Descripcion)) > 200 {
		v.Add("descripcion", "Descripción demasiado larga (máximo 200 caracteres)")
	}

	// Validar precio
	if prodReq.Precio <= 0 {
		v.Add("precio", "Precio debe ser mayor a 0")
	} else if prodReq.Precio > 500 {
		v.Add("precio", "Precio demasiado alto (máximo $500)")
	}

	return v
}

// ValidateVendedorRequestCompleto valida una solicitud completa de vendedor
func ValidateVendedorRequestCompleto(nombre string) *ValidateRequest {
	v := &ValidateRequest{}

	if strings.TrimSpace(nombre) == "" {
		v.Add("nombre", "Nombre es requerido")
	} else if len(strings.TrimSpace(nombre)) < 2 {
		v.Add("nombre", "Nombre debe tener al menos 2 caracteres")
	} else if len(nombre) > 100 {
		v.Add("nombre", "Nombre demasiado largo (máximo 100 caracteres)")
	} else if !isValidName(nombre) {
		v.Add("nombre", "Nombre contiene caracteres inválidos")
	}

	return v
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidName(name string) bool {
	// Solo permitir letras, espacios, apóstrofes, guiones y caracteres acentuados comunes
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			r == ' ' || r == '\'' || r == '-' ||
			r == 'á' || r == 'é' || r == 'í' || r == 'ó' || r == 'ú' ||
			r == 'Á' || r == 'É' || r == 'Í' || r == 'Ó' || r == 'Ú' ||
			r == 'ñ' || r == 'Ñ') {
			return false
		}
	}
	return true
}
