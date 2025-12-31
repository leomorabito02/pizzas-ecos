package validators

import (
	"fmt"
	"strings"
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
