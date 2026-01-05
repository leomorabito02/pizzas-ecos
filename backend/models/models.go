package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// ProductoItem representa un item en una venta
type ProductoItem struct {
	DetalleID int     `json:"detalle_id"` // id de la fila en detalle_ventas (para edición)
	Tipo      string  `json:"tipo"`       // "producto"
	ProductID int     `json:"product_id"` // producto_id
	Cantidad  int     `json:"cantidad"`   // cantidad
	Precio    float64 `json:"precio"`     // precio unitario
	Total     float64 `json:"total"`      // total (precio * cantidad)
}

// VentaRequest representa la solicitud para crear una venta
type VentaRequest struct {
	Vendedor        string         `json:"vendedor"`
	Cliente         string         `json:"cliente"`
	Items           []ProductoItem `json:"items"` // array de items con producto_id
	PaymentMethod   string         `json:"payment_method"`
	Estado          string         `json:"estado"`
	TipoEntrega     string         `json:"tipo_entrega"` // retiro o envio
	TelefonoCliente *int           `json:"telefono_cliente,omitempty"`
}

// DataResponse retorna vendedores, clientes y productos
type DataResponse struct {
	ClientesPorVendedor map[string][]Cliente `json:"clientesPorVendedor"`
	Vendedores          []Vendedor           `json:"vendedores"`
	Productos           []Producto           `json:"productos"`
}

// Pizza estructura para pizzas (legado)
type Pizza struct {
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Precios     []float64 `json:"precios"`
}

// VentaStats retorna estadísticas de una venta
type VentaStats struct {
	ID              int            `json:"id"`
	Vendedor        string         `json:"vendedor"`
	Cliente         string         `json:"cliente"`
	TelefonoCliente *int           `json:"telefono_cliente"`
	Total           float64        `json:"total"`
	PaymentMethod   string         `json:"payment_method"`
	Estado          string         `json:"estado"`
	TipoEntrega     string         `json:"tipo_entrega"`
	CreatedAt       time.Time      `json:"created_at"`
	Items           []ProductoItem `json:"items"`
}

// Cliente representa un cliente con teléfono
type Cliente struct {
	ID       int    `json:"id"`
	Nombre   string `json:"nombre"`
	Telefono int    `json:"telefono"`
}

// Auth structs
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Rol      string `json:"rol"`
}

type TokenClaims struct {
	Username string `json:"username"`
	Rol      string `json:"rol"`
	jwt.RegisteredClaims
}

// CreateUsuarioRequest estructura para crear usuario
type CreateUsuarioRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Rol      string `json:"rol"`
}

// UpdateUsuarioRequest estructura para actualizar usuario
type UpdateUsuarioRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Rol      string `json:"rol"`
}

// Producto estructura para productos
type Producto struct {
	ID          int       `json:"id"`
	TipoPizza   string    `json:"tipo_pizza"`
	Descripcion string    `json:"descripcion"`
	Precio      float64   `json:"precio"`
	Activo      bool      `json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
}

// Vendedor estructura para vendedores
type Vendedor struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

// CrearProductoRequest estructura para crear producto
type CrearProductoRequest struct {
	TipoPizza   string  `json:"tipo_pizza"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
}

// ActualizarProductoRequest estructura para actualizar producto
type ActualizarProductoRequest struct {
	TipoPizza   string  `json:"tipo_pizza"`
	Precio      float64 `json:"precio"`
	Descripcion string  `json:"descripcion"`
	Activo      bool    `json:"activo"`
}
