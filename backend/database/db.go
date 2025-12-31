package database

import (
	"database/sql"
	"fmt"
	"log"
	"pizzas-ecos/models"
	"strings"
)

var DB *sql.DB

// GetVendedores retorna lista de vendedores
func GetVendedores() ([]models.Vendedor, error) {
	rows, err := DB.Query("SELECT id, nombre FROM vendedores ORDER BY nombre")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendedores []models.Vendedor
	for rows.Next() {
		var vendedor models.Vendedor
		if err := rows.Scan(&vendedor.ID, &vendedor.Nombre); err != nil {
			return nil, err
		}
		vendedores = append(vendedores, vendedor)
	}

	return vendedores, nil
}

// GetVendedorID obtiene el ID de un vendedor por nombre
func GetVendedorID(nombre string) (int, error) {
	var id int
	err := DB.QueryRow("SELECT id FROM vendedores WHERE nombre = ?", nombre).Scan(&id)
	return id, err
}

// GetClientesPorVendedor obtiene clientes agrupados por vendedor
func GetClientesPorVendedor() (map[string][]string, error) {
	result := make(map[string][]string)

	query := `
		SELECT ve.nombre, c.nombre, c.apellido
		FROM ventas v
		JOIN vendedores ve ON v.vendedor_id = ve.id
		LEFT JOIN clientes c ON v.cliente_id = c.id
		WHERE c.id IS NOT NULL
		ORDER BY ve.nombre, c.nombre, c.apellido
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var vendedorNombre, clienteNombre string
		var clienteApellido sql.NullString
		if err := rows.Scan(&vendedorNombre, &clienteNombre, &clienteApellido); err != nil {
			return nil, err
		}
		// Construir nombre completo del cliente
		fullName := clienteNombre
		if clienteApellido.Valid && clienteApellido.String != "" {
			fullName = clienteNombre + " " + clienteApellido.String
		}
		fullName = strings.TrimSpace(fullName)

		// Evitar duplicados
		encontrado := false
		for _, c := range result[vendedorNombre] {
			if c == fullName {
				encontrado = true
				break
			}
		}
		if !encontrado {
			result[vendedorNombre] = append(result[vendedorNombre], fullName)
		}
	}

	return result, nil
}

// GetOrCreateCliente obtiene o crea un cliente
func GetOrCreateCliente(nombre string) (int, error) {
	// Buscar si existe
	var id int
	err := DB.QueryRow("SELECT id FROM clientes WHERE nombre = ?", nombre).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Crear si no existe
	res, err := DB.Exec("INSERT INTO clientes (nombre) VALUES (?)", nombre)
	if err != nil {
		return 0, err
	}

	idInt, _ := res.LastInsertId()
	return int(idInt), nil
}

// GetProductos retorna lista de productos activos
func GetProductos() ([]models.Producto, error) {
	var productos []models.Producto

	rows, err := DB.Query(`
		SELECT id, tipo_pizza, descripcion, precio, activo, created_at
		FROM productos
		WHERE activo = TRUE
		ORDER BY tipo_pizza
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Producto
		if err := rows.Scan(&p.ID, &p.TipoPizza, &p.Descripcion, &p.Precio, &p.Activo, &p.CreatedAt); err != nil {
			return nil, err
		}
		productos = append(productos, p)
	}

	return productos, nil
}

// InsertVenta inserta una nueva venta
func InsertVenta(clienteID *int, vendedorID int, total float64, payment, estado, tipoEntrega string) (int, error) {
	query := `
		INSERT INTO ventas (cliente_id, vendedor_id, total, payment_method, estado, tipo_entrega)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	res, err := DB.Exec(query, clienteID, vendedorID, total, payment, estado, tipoEntrega)
	if err != nil {
		return 0, err
	}

	id, _ := res.LastInsertId()
	return int(id), nil
}

// InsertDetalle inserta un detalle de venta
func InsertDetalle(ventaID int, item models.ProductoItem) error {
	productoID := item.ProductID

	query := `
		INSERT INTO detalle_ventas (venta_id, producto_id, cantidad, precio_unitario, subtotal)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := DB.Exec(query, ventaID, productoID, item.Cantidad, item.Precio, item.Total)
	return err
}

// GetAllVentas retorna todas las ventas
func GetAllVentas(includeCanceladas bool) ([]models.VentaStats, error) {
	whereClause := ""
	if !includeCanceladas {
		whereClause = "WHERE v.estado != 'cancelada'"
	}

	query := `
		SELECT v.id, ve.nombre, COALESCE(c.nombre, 'Sin cliente'), 
		       v.total, v.payment_method, v.estado, v.tipo_entrega, v.created_at
		FROM ventas v
		JOIN vendedores ve ON v.vendedor_id = ve.id
		LEFT JOIN clientes c ON v.cliente_id = c.id
		` + whereClause + `
		ORDER BY v.created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ventas []models.VentaStats
	for rows.Next() {
		var v models.VentaStats

		if err := rows.Scan(&v.ID, &v.Vendedor, &v.Cliente, &v.Total, &v.PaymentMethod, &v.Estado, &v.TipoEntrega, &v.CreatedAt); err != nil {
			return nil, err
		}

		// Cargar items (detalles) para esta venta
		itemsQuery := `
			SELECT dv.id, dv.producto_id, dv.cantidad, p.tipo_pizza, p.precio
			FROM detalle_ventas dv
			JOIN productos p ON dv.producto_id = p.id
			WHERE dv.venta_id = ?
		`
		itemRows, err := DB.Query(itemsQuery, v.ID)
		if err == nil {
			for itemRows.Next() {
				var item models.ProductoItem
				var productoID int
				var tipo_pizza string
				var precio float64
				var cantidad int
				if err := itemRows.Scan(&item.DetalleID, &productoID, &cantidad, &tipo_pizza, &precio); err == nil {
					item.ProductID = productoID
					item.Cantidad = cantidad
					item.Tipo = tipo_pizza
					item.Precio = precio
					item.Total = float64(cantidad) * precio
					v.Items = append(v.Items, item)
				}
			}
			itemRows.Close()
		}

		if v.Items == nil {
			v.Items = []models.ProductoItem{} // Array vacío en lugar de null
		}

		ventas = append(ventas, v)
	}

	return ventas, nil
}

// GetResumen retorna el resumen de ventas
func GetResumen() (map[string]interface{}, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN (v.estado='pagada' OR v.estado='entregada') AND v.payment_method='efectivo' THEN v.total ELSE 0 END), 0) as efectivo,
			COALESCE(SUM(CASE WHEN (v.estado='pagada' OR v.estado='entregada') AND v.payment_method='transferencia' THEN v.total ELSE 0 END), 0) as transferencia,
			COALESCE(SUM(CASE WHEN v.estado='sin pagar' THEN v.total ELSE 0 END), 0) as pendiente,
			COALESCE(SUM(CASE WHEN v.estado='pagada' OR v.estado='entregada' THEN v.total ELSE 0 END), 0) as total_cobrado,
			COUNT(CASE WHEN v.estado='sin pagar' THEN 1 END) as ventas_sin_pagar,
			COUNT(CASE WHEN v.estado='pagada' OR v.estado='entregada' THEN 1 END) as ventas_pagadas,
			COUNT(CASE WHEN v.estado='entregada' THEN 1 END) as ventas_entregadas,
			COUNT(*) as ventas_totales
		FROM ventas v
		WHERE v.estado != 'cancelada'
	`

	var efectivo, transferencia, pendiente, total float64
	var sinPagar, pagadas, entregadas, totalVentas int

	err := DB.QueryRow(query).Scan(&efectivo, &transferencia, &pendiente, &total, &sinPagar, &pagadas, &entregadas, &totalVentas)
	if err != nil {
		log.Printf("Error en GetResumen: %v", err)
		return nil, err
	}

	// Ahora calcular items por separado
	itemsQuery := `
		SELECT 
			COALESCE(SUM(dv.cantidad), 0) as total_items,
			COALESCE(SUM(CASE WHEN v.tipo_entrega IN ('delivery', 'envio') OR (v.tipo_entrega IS NULL OR v.tipo_entrega = '') THEN dv.cantidad ELSE 0 END), 0) as total_delivery,
			COALESCE(SUM(CASE WHEN v.tipo_entrega='retiro' THEN dv.cantidad ELSE 0 END), 0) as total_retiro
		FROM detalle_ventas dv
		JOIN ventas v ON dv.venta_id = v.id
		WHERE v.estado != 'cancelada'
	`

	var totalItems, delivery, retiro int
	err = DB.QueryRow(itemsQuery).Scan(&totalItems, &delivery, &retiro)
	if err != nil {
		log.Printf("Error en GetResumen items: %v", err)
		totalItems, delivery, retiro = 0, 0, 0
	}

	return map[string]interface{}{
		"total_items":           totalItems,
		"total_delivery":        delivery,
		"total_retiro":          retiro,
		"efectivo_cobrado":      efectivo,
		"transferencia_cobrada": transferencia,
		"pendiente_cobro":       pendiente,
		"total_cobrado":         total,
		"ventas_sin_pagar":      sinPagar,
		"ventas_pagadas":        pagadas,
		"ventas_entregadas":     entregadas,
		"ventas_totales":        totalVentas,
	}, nil
}

// GetVendedoresConStats retorna vendedores con estadísticas
func GetVendedoresConStats() ([]map[string]interface{}, error) {
	vendedores, _ := GetVendedores()
	var result []map[string]interface{}

	for _, vendedor := range vendedores {
		// Query 1: Dinero sin JOIN (para evitar multiplicación)
		query := `
			SELECT 
				COUNT(DISTINCT v.id) as cantidad,
				COALESCE(SUM(CASE WHEN v.estado='sin pagar' THEN v.total ELSE 0 END), 0) as deuda,
				COALESCE(SUM(CASE WHEN v.estado='pagada' OR v.estado='entregada' THEN v.total ELSE 0 END), 0) as pagado,
				COALESCE(SUM(v.total), 0) as total
			FROM ventas v
			JOIN vendedores ve ON v.vendedor_id = ve.id
			WHERE ve.nombre = ? AND v.estado != 'cancelada'
		`

		var cantidad int
		var deuda, pagado, total float64

		err := DB.QueryRow(query, vendedor.Nombre).Scan(&cantidad, &deuda, &pagado, &total)
		if err != nil {
			log.Printf("Error consultando vendor %s: %v", vendedor.Nombre, err)
			continue
		}

		// Query 2: Items por separado
		itemsQuery := `
			SELECT COALESCE(SUM(dv.cantidad), 0) as total_items
			FROM detalle_ventas dv
			JOIN ventas v ON dv.venta_id = v.id
			JOIN vendedores ve ON v.vendedor_id = ve.id
			WHERE ve.nombre = ? AND v.estado != 'cancelada'
		`

		var totalItems int
		err = DB.QueryRow(itemsQuery, vendedor.Nombre).Scan(&totalItems)
		if err != nil {
			log.Printf("Error consultando items vendor %s: %v", vendedor.Nombre, err)
			totalItems = 0
		}

		result = append(result, map[string]interface{}{
			"nombre":      vendedor.Nombre,
			"cantidad":    cantidad,
			"total_items": totalItems,
			"deuda":       deuda,
			"pagado":      pagado,
			"total":       total,
		})
	}

	return result, nil
}

// UpdateVenta actualiza una venta (estado, pago, y productos)
func UpdateVenta(ventaID int, estado, paymentMethod, tipoEntrega string, productosEliminar []int, productos []map[string]interface{}) error {
	// 1. Actualizar estado, payment_method y tipo_entrega de la venta
	query := `UPDATE ventas SET estado = ?, payment_method = ?, tipo_entrega = ? WHERE id = ?`
	_, err := DB.Exec(query, estado, paymentMethod, tipoEntrega, ventaID)
	if err != nil {
		return fmt.Errorf("error actualizando venta: %w", err)
	}

	// 2. Eliminar productos si se proporcionan
	for _, detalleID := range productosEliminar {
		deleteQuery := `DELETE FROM detalle_ventas WHERE id = ?`
		_, err := DB.Exec(deleteQuery, detalleID)
		if err != nil {
			log.Printf("Error eliminando producto: %v", err)
			return fmt.Errorf("error eliminando producto: %w", err)
		}
	}

	// 3. Actualizar/insertar productos si se proporcionan
	for _, p := range productos {
		detalleID := p["detalle_id"]
		productoID := int(p["producto_id"].(float64))
		cantidad := int(p["cantidad"].(float64))

		if detalleID == nil {
			// Nuevo producto - insertar en detalle_ventas
			insertQuery := `INSERT INTO detalle_ventas (venta_id, producto_id, cantidad) VALUES (?, ?, ?)`
			_, err := DB.Exec(insertQuery, ventaID, productoID, cantidad)
			if err != nil {
				log.Printf("Error insertando producto: %v", err)
				return fmt.Errorf("error insertando producto: %w", err)
			}
		} else {
			// Actualizar cantidad existente
			detalleIDInt := int(detalleID.(float64))
			updateQuery := `UPDATE detalle_ventas SET cantidad = ? WHERE id = ?`
			_, err := DB.Exec(updateQuery, cantidad, detalleIDInt)
			if err != nil {
				log.Printf("Error actualizando producto: %v", err)
				return fmt.Errorf("error actualizando producto: %w", err)
			}
		}
	}

	// 4. Recalcular total de la venta
	var nuevoTotal float64
	totalQuery := `SELECT COALESCE(SUM(dv.cantidad * p.precio), 0) FROM detalle_ventas dv JOIN productos p ON dv.producto_id = p.id WHERE dv.venta_id = ?`
	err = DB.QueryRow(totalQuery, ventaID).Scan(&nuevoTotal)
	if err != nil {
		log.Printf("Error calculando nuevo total: %v", err)
	} else {
		_, err := DB.Exec(`UPDATE ventas SET total = ? WHERE id = ?`, nuevoTotal, ventaID)
		if err != nil {
			log.Printf("Error actualizando total: %v", err)
		}
	}

	return nil
}

// GetProductoByID obtiene un producto por ID
func GetProductoByID(id int) (*models.Producto, error) {
	var p models.Producto
	err := DB.QueryRow(`
		SELECT id, tipo_pizza, descripcion, precio, activo, created_at
		FROM productos WHERE id = ?
	`, id).Scan(&p.ID, &p.TipoPizza, &p.Descripcion, &p.Precio, &p.Activo, &p.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetUserByCredentials obtiene un usuario por credenciales
func GetUserByCredentials(username, plainPassword string) (*models.User, error) {
	var user models.User
	var storedHash string
	err := DB.QueryRow(
		"SELECT id, username, rol, password_hash FROM usuarios WHERE username = ?",
		username).Scan(&user.ID, &user.Username, &user.Rol, &storedHash)

	if err != nil {
		return nil, err
	}

	// Comparar la contraseña en texto plano con el hash almacenado
	if !VerifyPassword(storedHash, plainPassword) {
		return nil, fmt.Errorf("contraseña inválida")
	}

	return &user, nil
}

// CreateProducto crea un nuevo producto
func CreateProducto(tipoPizza, descripcion string, precio float64) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO productos (tipo_pizza, descripcion, precio, activo) VALUES (?, ?, ?, TRUE)",
		tipoPizza, descripcion, precio,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateProducto actualiza un producto
func UpdateProducto(id int, tipoPizza, descripcion string, precio float64, activo bool) error {
	_, err := DB.Exec(
		"UPDATE productos SET tipo_pizza = ?, precio = ?, descripcion = ?, activo = ? WHERE id = ?",
		tipoPizza, precio, descripcion, activo, id,
	)
	return err
}

// DeleteProducto desactiva un producto
func DeleteProducto(id int) error {
	result, err := DB.Exec("UPDATE productos SET activo = FALSE WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// CreateVendedor crea un nuevo vendedor
func CreateVendedor(nombre string) (int64, error) {
	result, err := DB.Exec(`INSERT INTO vendedores (nombre) VALUES (?)`, nombre)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateVendedor actualiza un vendedor
func UpdateVendedor(id int, nombre string) error {
	result, err := DB.Exec(`UPDATE vendedores SET nombre = ? WHERE id = ?`, nombre, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteVendedor elimina un vendedor
func DeleteVendedor(id int) error {
	result, err := DB.Exec(`DELETE FROM vendedores WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetAllUsers obtiene todos los usuarios sin contraseñas
func GetAllUsers() ([]models.User, error) {
	rows, err := DB.Query("SELECT id, username, rol FROM usuarios ORDER BY username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []models.User
	for rows.Next() {
		var usuario models.User
		if err := rows.Scan(&usuario.ID, &usuario.Username, &usuario.Rol); err != nil {
			return nil, err
		}
		usuarios = append(usuarios, usuario)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if usuarios == nil {
		usuarios = []models.User{}
	}

	return usuarios, nil
}

// UserExists verifica si un usuario existe
func UserExists(username string) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE username = ?)", username).Scan(&exists)
	return exists, err
}

// CreateUser crea un nuevo usuario con contraseña hasheada
func CreateUser(username, password, rol string) (int, error) {
	// Hash la contraseña
	hash, err := HashPassword(password)
	if err != nil {
		return 0, err
	}

	result, err := DB.Exec(
		"INSERT INTO usuarios (username, password_hash, rol) VALUES (?, ?, ?)",
		username, hash, rol,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	return int(id), err
}

// UpdateUser actualiza un usuario existente
func UpdateUser(id int, username, password, rol string) error {
	var query string
	var args []interface{}

	if password != "" {
		// Si se proporciona contraseña, actualizarla también
		hash, err := HashPassword(password)
		if err != nil {
			return err
		}
		query = "UPDATE usuarios SET username = ?, password_hash = ?, rol = ? WHERE id = ?"
		args = []interface{}{username, hash, rol, id}
	} else {
		// Si no se proporciona contraseña, solo actualizar username y rol
		query = "UPDATE usuarios SET username = ?, rol = ? WHERE id = ?"
		args = []interface{}{username, rol, id}
	}

	result, err := DB.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteUser elimina un usuario
func DeleteUser(id int) error {
	result, err := DB.Exec("DELETE FROM usuarios WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ClearDetalleVentas elimina todos los detalles de ventas
func ClearDetalleVentas() error {
	_, err := DB.Exec("DELETE FROM detalle_ventas")
	return err
}

// ClearVentas elimina todas las ventas
func ClearVentas() error {
	_, err := DB.Exec("DELETE FROM ventas")
	return err
}

// ClearClientes elimina todos los clientes
func ClearClientes() error {
	_, err := DB.Exec("DELETE FROM clientes")
	return err
}

// ClearVendedores elimina todos los vendedores
func ClearVendedores() error {
	_, err := DB.Exec("DELETE FROM vendedores")
	return err
}

// ClearProductos elimina todos los productos
func ClearProductos() error {
	_, err := DB.Exec("DELETE FROM productos")
	return err
}
