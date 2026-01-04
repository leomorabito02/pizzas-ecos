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

// GetClientesPorVendedor obtiene clientes agrupados por vendedor (solo clientes que han tenido ventas con ese vendedor)
func GetClientesPorVendedor() (map[string][]models.Cliente, error) {
	result := make(map[string][]models.Cliente)

	// Query para obtener clientes por vendedor basándose en ventas
	query := `
		SELECT DISTINCT
			v.nombre as vendedor,
			c.id,
			c.nombre,
			COALESCE(c.telefono, 0) as telefono
		FROM ventas vt
		JOIN vendedores v ON vt.vendedor_id = v.id
		JOIN clientes c ON vt.cliente_id = c.id
		ORDER BY v.nombre, c.nombre
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Procesar resultados y agrupar por vendedor
	for rows.Next() {
		var vendedor string
		var cliente models.Cliente
		if err := rows.Scan(&vendedor, &cliente.ID, &cliente.Nombre, &cliente.Telefono); err != nil {
			return nil, err
		}
		cliente.Nombre = strings.TrimSpace(cliente.Nombre)
		result[vendedor] = append(result[vendedor], cliente)
	}

	return result, nil
}

// GetOrCreateCliente obtiene o crea un cliente
func GetOrCreateCliente(nombre string) (int, error) {
	// Limpieza básica
	nombre = strings.TrimSpace(nombre)

	// Usamos transacción para asegurar lectura consistente o escritura
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // Rollback seguro si no hay commit

	var id int
	// Intentamos buscar primero
	err = tx.QueryRow("SELECT id FROM clientes WHERE nombre = ?", nombre).Scan(&id)
	if err == nil {
		return id, nil // Ya existe, no hacemos commit porque fue solo lectura, rollback limpia el contexto
	}

	// Si no existe, creamos
	res, err := tx.Exec("INSERT INTO clientes (nombre) VALUES (?)", nombre)
	if err != nil {
		return 0, err
	}

	idInt, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return int(idInt), nil
}

// GetClienteByNombre devuelve id y telefono (0 si null) y si existe
func GetClienteByNombre(nombre string) (int, int, bool, error) {
	var id int
	var telefono sql.NullInt64
	err := DB.QueryRow("SELECT id, telefono FROM clientes WHERE nombre = ?", nombre).Scan(&id, &telefono)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	tel := 0
	if telefono.Valid {
		tel = int(telefono.Int64)
	}
	return id, tel, true, nil
}

// CreateClienteWithTelefono crea un cliente con telefono opcional
func CreateClienteWithTelefono(nombre string, telefono *int) (int, error) {
	if telefono != nil {
		res, err := DB.Exec("INSERT INTO clientes (nombre, telefono) VALUES (?, ?)", nombre, *telefono)
		if err != nil {
			return 0, err
		}
		id64, _ := res.LastInsertId()
		return int(id64), nil
	}
	res, err := DB.Exec("INSERT INTO clientes (nombre) VALUES (?)", nombre)
	if err != nil {
		return 0, err
	}
	id64, _ := res.LastInsertId()
	return int(id64), nil
}

// UpdateClienteTelefono actualiza el telefono de un cliente
func UpdateClienteTelefono(id int, telefono *int) error {
	if telefono == nil {
		_, err := DB.Exec("UPDATE clientes SET telefono = NULL WHERE id = ?", id)
		return err
	}
	_, err := DB.Exec("UPDATE clientes SET telefono = ? WHERE id = ?", *telefono, id)
	return err
}

// UpdateVentaClienteID asocia una venta a un cliente
func UpdateVentaClienteID(ventaID int, clienteID int) error {
	_, err := DB.Exec("UPDATE ventas SET cliente_id = ? WHERE id = ?", clienteID, ventaID)
	return err
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

	// 1. Obtener solo las ventas (sin detalles)
	ventasQuery := `
		SELECT v.id, ve.nombre, COALESCE(c.nombre, 'Sin cliente'), 
		       COALESCE(c.telefono, 0), v.total, v.payment_method, v.estado, v.tipo_entrega, v.created_at
		FROM ventas v
		JOIN vendedores ve ON v.vendedor_id = ve.id
		LEFT JOIN clientes c ON v.cliente_id = c.id
		` + whereClause + `
		ORDER BY v.created_at DESC
	`

	rows, err := DB.Query(ventasQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Construir un mapa de punteros y una lista de IDs para preservar orden
	ventaOrder := []int{}
	ventaIDs := []int{}
	ventasMap := make(map[int]*models.VentaStats)

	for rows.Next() {
		v := &models.VentaStats{}
		if err := rows.Scan(&v.ID, &v.Vendedor, &v.Cliente, &v.TelefonoCliente, &v.Total, &v.PaymentMethod, &v.Estado, &v.TipoEntrega, &v.CreatedAt); err != nil {
			return nil, err
		}
		v.Items = []models.ProductoItem{}
		ventaOrder = append(ventaOrder, v.ID)
		ventaIDs = append(ventaIDs, v.ID)
		ventasMap[v.ID] = v
	}

	// 2. Si hay ventas, obtener todos los items de una sola vez
	if len(ventaIDs) > 0 {
		// Construir placeholders para la query
		placeholders := ""
		args := make([]interface{}, len(ventaIDs))
		for i, id := range ventaIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args[i] = id
		}

		itemsQuery := `
			SELECT dv.venta_id, dv.id, dv.producto_id, dv.cantidad, p.tipo_pizza, p.precio
			FROM detalle_ventas dv
			JOIN productos p ON dv.producto_id = p.id
			WHERE dv.venta_id IN (` + placeholders + `)
			ORDER BY dv.venta_id, dv.id
		`

		itemRows, err := DB.Query(itemsQuery, args...)
		if err == nil {
			for itemRows.Next() {
				var ventaID int
				var item models.ProductoItem
				var productoID int
				var tipo_pizza string
				var precio float64
				var cantidad int

				if err := itemRows.Scan(&ventaID, &item.DetalleID, &productoID, &cantidad, &tipo_pizza, &precio); err == nil {
					item.ProductID = productoID
					item.Cantidad = cantidad
					item.Tipo = tipo_pizza
					item.Precio = precio
					item.Total = float64(cantidad) * precio

					if venta, ok := ventasMap[ventaID]; ok {
						venta.Items = append(venta.Items, item)
					}
				}
			}
			itemRows.Close()
		}
	}

	// Reconstruir slice en el mismo orden en que se obtuvieron las ventas
	ventas := make([]models.VentaStats, 0, len(ventaOrder))
	for _, id := range ventaOrder {
		if vptr, ok := ventasMap[id]; ok {
			ventas = append(ventas, *vptr)
		}
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

	// Ahora calcular cantidad de ventas por tipo de entrega
	itemsQuery := `
		SELECT 
			COALESCE(COUNT(DISTINCT CASE WHEN v.tipo_entrega IN ('delivery', 'envio') OR (v.tipo_entrega IS NULL OR v.tipo_entrega = '') THEN v.id END), 0) as total_ventas_delivery,
			COALESCE(COUNT(DISTINCT CASE WHEN v.tipo_entrega='retiro' THEN v.id END), 0) as total_ventas_retiro
		FROM ventas v
		WHERE v.estado != 'cancelada'
	`

	var delivery, retiro int
	err = DB.QueryRow(itemsQuery).Scan(&delivery, &retiro)
	if err != nil {
		log.Printf("Error en GetResumen items: %v", err)
		delivery, retiro = 0, 0
	}

	return map[string]interface{}{
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

// UpdateVenta actualiza una venta de forma atómica usando transacciones
func UpdateVenta(ventaID int, estado, paymentMethod, tipoEntrega string, productosEliminar []int, productos []map[string]interface{}) error {
	// 1. Iniciar Transacción
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}

	// Defer para Rollback en caso de pánico o error no manejado explícitamente
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-lanzar pánico después de rollback
		} else if err != nil {
			tx.Rollback() // Rollback si hay error retornado
		}
	}()

	// 2. Actualizar cabecera de venta
	query := `UPDATE ventas SET estado = ?, payment_method = ?, tipo_entrega = ? WHERE id = ?`
	if _, err = tx.Exec(query, estado, paymentMethod, tipoEntrega, ventaID); err != nil {
		return fmt.Errorf("error actualizando cabecera venta: %w", err)
	}

	// 3. Eliminar productos (Batch)
	if len(productosEliminar) > 0 {
		// Nota de eficiencia: Podríamos usar IN (?) dinámico, pero un loop simple dentro de Tx es aceptable para pocos items
		deleteQuery := `DELETE FROM detalle_ventas WHERE id = ?`
		stmt, err := tx.Prepare(deleteQuery)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, detalleID := range productosEliminar {
			if _, err = stmt.Exec(detalleID); err != nil {
				return fmt.Errorf("error eliminando producto %d: %w", detalleID, err)
			}
		}
	}

	// 4. Upsert (Insertar o Actualizar) productos
	// Preparamos los statements fuera del loop para eficiencia
	insertStmt, err := tx.Prepare(`INSERT INTO detalle_ventas (venta_id, producto_id, cantidad, precio_unitario, subtotal) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	updateStmt, err := tx.Prepare(`UPDATE detalle_ventas SET cantidad = ?, subtotal = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	for _, p := range productos {
		detalleID := p["detalle_id"]
		productoID := int(p["producto_id"].(float64))
		cantidad := int(p["cantidad"].(float64))

		// Necesitamos el precio actual del producto para consistencia
		var precio float64
		err = tx.QueryRow("SELECT precio FROM productos WHERE id = ?", productoID).Scan(&precio)
		if err != nil {
			return fmt.Errorf("producto %d no encontrado o inactivo", productoID)
		}

		subtotal := float64(cantidad) * precio

		if detalleID == nil {
			if _, err = insertStmt.Exec(ventaID, productoID, cantidad, precio, subtotal); err != nil {
				return err
			}
		} else {
			detalleIDInt := int(detalleID.(float64))
			if _, err = updateStmt.Exec(cantidad, subtotal, detalleIDInt); err != nil {
				return err
			}
		}
	}

	// 5. Recalcular total usando la misma transacción (ve los cambios no confirmados)
	var nuevoTotal float64
	// Sumamos directamente de detalle_ventas que ya tiene el subtotal actualizado
	totalQuery := `SELECT COALESCE(SUM(subtotal), 0) FROM detalle_ventas WHERE venta_id = ?`
	if err = tx.QueryRow(totalQuery, ventaID).Scan(&nuevoTotal); err != nil {
		return fmt.Errorf("error recalculando total: %w", err)
	}

	if _, err = tx.Exec(`UPDATE ventas SET total = ? WHERE id = ?`, nuevoTotal, ventaID); err != nil {
		return fmt.Errorf("error actualizando total final: %w", err)
	}

	// 6. Commit final
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error en commit: %w", err)
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
