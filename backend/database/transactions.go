package database

import (
	"context"
	"database/sql"
	"fmt"

	"pizzas-ecos/logger"
)

// Transaction wrapper para manejar transacciones de forma segura
type Transaction struct {
	tx *sql.Tx
}

// BeginTx inicia una nueva transacción
func BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		logger.Error("Error al iniciar transacción", "TX_BEGIN_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return &Transaction{tx: tx}, nil
}

// Commit commits la transacción
func (t *Transaction) Commit() error {
	if err := t.tx.Commit(); err != nil {
		logger.Error("Error al hacer commit", "TX_COMMIT_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	return nil
}

// Rollback revierte la transacción
func (t *Transaction) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		logger.Error("Error al hacer rollback", "TX_ROLLBACK_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	return nil
}

// Exec ejecuta una query en la transacción
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// Query ejecuta una query de lectura en la transacción
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryRow ejecuta una query que retorna una fila
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// ValidateData contiene validaciones de datos
type ValidateData struct {
	Errors []string
}

// Add agrega un error de validación
func (v *ValidateData) Add(msg string) {
	v.Errors = append(v.Errors, msg)
}

// IsValid retorna si no hay errores
func (v *ValidateData) IsValid() bool {
	return len(v.Errors) == 0
}

// GetMessage retorna mensaje de validación
func (v *ValidateData) GetMessage() string {
	if len(v.Errors) == 0 {
		return ""
	}
	return fmt.Sprintf("Validación fallida: %v", v.Errors)
}

// ValidateVenta valida datos de venta antes de insertar
func ValidateVenta(vendedorID, clienteID int, total float64, items int) *ValidateData {
	v := &ValidateData{}

	if vendedorID <= 0 {
		v.Add("Vendedor inválido")
	}
	if clienteID <= 0 {
		v.Add("Cliente inválido")
	}
	if total <= 0 {
		v.Add("Total debe ser mayor a 0")
	}
	if items <= 0 {
		v.Add("Debe haber al menos 1 item")
	}

	return v
}

// ValidateProducto valida datos de producto
func ValidateProducto(nombre string, precio float64) *ValidateData {
	v := &ValidateData{}

	if nombre == "" {
		v.Add("Nombre requerido")
	}
	if len(nombre) < 3 {
		v.Add("Nombre mínimo 3 caracteres")
	}
	if precio <= 0 {
		v.Add("Precio debe ser mayor a 0")
	}

	return v
}

// ValidateVendedor valida datos de vendedor
func ValidateVendedor(nombre string) *ValidateData {
	v := &ValidateData{}

	if nombre == "" {
		v.Add("Nombre requerido")
	}
	if len(nombre) < 2 {
		v.Add("Nombre mínimo 2 caracteres")
	}

	return v
}

// ExistsVendedor verifica si un vendedor existe
func ExistsVendedor(ctx context.Context, id int) (bool, error) {
	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM vendedores WHERE id = ?", id).Scan(&count)
	return count > 0, err
}

// ExistsCliente verifica si un cliente existe
func ExistsCliente(ctx context.Context, id int) (bool, error) {
	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM clientes WHERE id = ?", id).Scan(&count)
	return count > 0, err
}

// ExistsProducto verifica si un producto existe
func ExistsProducto(ctx context.Context, id int) (bool, error) {
	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM productos WHERE id = ?", id).Scan(&count)
	return count > 0, err
}

// ExistsVenta verifica si una venta existe
func ExistsVenta(ctx context.Context, id int) (bool, error) {
	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM ventas WHERE id = ?", id).Scan(&count)
	return count > 0, err
}
