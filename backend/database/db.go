package database

import (
	"database/sql"
	"fmt"
	"log"
	"pizzas-ecos/models"
	"pizzas-ecos/security"
	"strings"
)

var DB *sql.DB

// GetVendedores retorna lista de vendedores (nombres desencriptados)
func GetVendedores() ([]models.Vendedor, error) {
	rows, err := DB.Query("SELECT id, nombre FROM vendedores ORDER BY nombre")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendedores []models.Vendedor
	for rows.Next() {
		var vendedor models.Vendedor
		var encNombre string
		if err := rows.Scan(&vendedor.ID, &encNombre); err != nil {
			return nil, err
		}
		nombre, err := security.Decrypt(encNombre)
		if err != nil {
			return nil, fmt.Errorf("error desencriptando vendedor %d: %w", vendedor.ID, err)
		}
		vendedor.Nombre = nombre
		vendedores = append(vendedores, vendedor)
	}

	return vendedores, nil
}

// GetVendedorID obtiene el ID de un vendedor buscando por HMAC del nombre
func GetVendedorID(nombre string) (int, error) {
	var id int
	hash := security.HMACField(nombre)
	err := DB.QueryRow("SELECT id FROM vendedores WHERE nombre_hash = $1", hash).Scan(&id)
	return id, err
}

// GetClientesPorVendedor obtiene clientes agrupados por vendedor (nombres desencriptados)
func GetClientesPorVendedor() (map[string][]models.Cliente, error) {
	result := make(map[string][]models.Cliente)

	query := `
		SELECT DISTINCT
			v.nombre as vendedor,
			c.id,
			c.nombre,
			COALESCE(c.telefono, '') as telefono
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

	for rows.Next() {
		var encVendedor, encNombre, encTelefono string
		var cliente models.Cliente
		if err := rows.Scan(&encVendedor, &cliente.ID, &encNombre, &encTelefono); err != nil {
			return nil, err
		}
		vendedor, err := security.Decrypt(encVendedor)
		if err != nil {
			return nil, fmt.Errorf("error desencriptando vendedor en clientes: %w", err)
		}
		nombre, err := security.Decrypt(encNombre)
		if err != nil {
			return nil, fmt.Errorf("error desencriptando cliente %d: %w", cliente.ID, err)
		}
		cliente.Nombre = strings.TrimSpace(nombre)
		if encTelefono != "" {
			decTel, err := security.Decrypt(encTelefono)
			if err == nil {
				var tel int
				fmt.Sscanf(decTel, "%d", &tel)
				cliente.Telefono = tel
			}
		}
		result[vendedor] = append(result[vendedor], cliente)
	}

	return result, nil
}

// GetOrCreateCliente obtiene o crea un cliente usando HMAC para búsqueda
func GetOrCreateCliente(nombre string) (int, error) {
	nombre = strings.TrimSpace(nombre)

	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int
	hash := security.HMACField(nombre)
	err = tx.QueryRow("SELECT id FROM clientes WHERE nombre_hash = $1", hash).Scan(&id)
	if err == nil {
		return id, nil
	}

	encNombre, err := security.Encrypt(nombre)
	if err != nil {
		return 0, fmt.Errorf("error encriptando nombre cliente: %w", err)
	}

	var idInt int64
	err = tx.QueryRow("INSERT INTO clientes (nombre, nombre_hash) VALUES ($1, $2) RETURNING id", encNombre, hash).Scan(&idInt)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return int(idInt), nil
}

// GetClienteByNombre devuelve id y telefono (0 si null) y si existe; busca por HMAC
func GetClienteByNombre(nombre string) (int, int, bool, error) {
	var id int
	var encTelefono sql.NullString
	hash := security.HMACField(nombre)
	err := DB.QueryRow("SELECT id, telefono FROM clientes WHERE nombre_hash = $1", hash).Scan(&id, &encTelefono)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	tel := 0
	if encTelefono.Valid && encTelefono.String != "" {
		decTel, err := security.Decrypt(encTelefono.String)
		if err == nil {
			fmt.Sscanf(decTel, "%d", &tel)
		}
	}
	return id, tel, true, nil
}

// CreateClienteWithTelefono crea un cliente con telefono opcional (ambos encriptados)
func CreateClienteWithTelefono(nombre string, telefono *int) (int, error) {
	encNombre, err := security.Encrypt(nombre)
	if err != nil {
		return 0, fmt.Errorf("error encriptando nombre: %w", err)
	}
	nombreHash := security.HMACField(nombre)

	var id64 int64
	if telefono != nil {
		encTel, err := security.Encrypt(fmt.Sprintf("%d", *telefono))
		if err != nil {
			return 0, fmt.Errorf("error encriptando telefono: %w", err)
		}
		err = DB.QueryRow(
			"INSERT INTO clientes (nombre, nombre_hash, telefono) VALUES ($1, $2, $3) RETURNING id",
			encNombre, nombreHash, encTel,
		).Scan(&id64)
		if err != nil {
			return 0, err
		}
		return int(id64), nil
	}
	err = DB.QueryRow(
		"INSERT INTO clientes (nombre, nombre_hash) VALUES ($1, $2) RETURNING id",
		encNombre, nombreHash,
	).Scan(&id64)
	if err != nil {
		return 0, err
	}
	return int(id64), nil
}

// UpdateClienteTelefono actualiza el telefono de un cliente (encriptado)
func UpdateClienteTelefono(id int, telefono *int) error {
	if telefono == nil {
		_, err := DB.Exec("UPDATE clientes SET telefono = NULL WHERE id = $1", id)
		return err
	}
	encTel, err := security.Encrypt(fmt.Sprintf("%d", *telefono))
	if err != nil {
		return fmt.Errorf("error encriptando telefono: %w", err)
	}
	_, err = DB.Exec("UPDATE clientes SET telefono = $1 WHERE id = $2", encTel, id)
	return err
}

// UpdateVentaClienteID asocia una venta a un cliente
func UpdateVentaClienteID(ventaID int, clienteID int) error {
	_, err := DB.Exec("UPDATE ventas SET cliente_id = $1 WHERE id = $2", clienteID, ventaID)
	return err
}

// GetProductos retorna lista de productos activos
func GetProductos() ([]models.Producto, error) {
	var productos []models.Producto

	rows, err := DB.Query(`
		SELECT id, tipo_pizza, descripcion, precio, activo, es_combo, created_at
		FROM productos
		WHERE activo = TRUE
		ORDER BY tipo_pizza
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comboIDs []int
	prodMap := make(map[int]*models.Producto)

	for rows.Next() {
		var p models.Producto
		if err := rows.Scan(&p.ID, &p.TipoPizza, &p.Descripcion, &p.Precio, &p.Activo, &p.EsCombo, &p.CreatedAt); err != nil {
			return nil, err
		}
		productos = append(productos, p)
	}
	rows.Close()

	for i := range productos {
		prodMap[productos[i].ID] = &productos[i]
		if productos[i].EsCombo {
			comboIDs = append(comboIDs, productos[i].ID)
		}
	}

	if len(comboIDs) > 0 {
		placeholders := ""
		args := make([]interface{}, len(comboIDs))
		for i, id := range comboIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += fmt.Sprintf("$%d", i+1)
			args[i] = id
		}
		compRows, err := DB.Query(`SELECT combo_id, producto_id, cantidad FROM combo_productos WHERE combo_id IN (`+placeholders+`)`, args...)
		if err == nil {
			for compRows.Next() {
				var comp models.ComboProducto
				if err := compRows.Scan(&comp.ComboID, &comp.ProductoID, &comp.Cantidad); err == nil {
					if p, ok := prodMap[comp.ComboID]; ok {
						p.Componentes = append(p.Componentes, comp)
					}
				}
			}
			compRows.Close()
		}
	}

	return productos, nil
}

// InsertVenta inserta una nueva venta
func InsertVenta(clienteID *int, vendedorID int, total float64, payment, estado, tipoEntrega string) (int, error) {
	query := `
		INSERT INTO ventas (cliente_id, vendedor_id, total, payment_method, estado, tipo_entrega)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int64
	err := DB.QueryRow(query, clienteID, vendedorID, total, payment, estado, tipoEntrega).Scan(&id)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertDetalle inserta un detalle de venta
func InsertDetalle(ventaID int, item models.ProductoItem) error {
	productoID := item.ProductID

	query := `
		INSERT INTO detalle_ventas (venta_id, producto_id, cantidad, precio_unitario, subtotal)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := DB.Exec(query, ventaID, productoID, item.Cantidad, item.Precio, item.Total)
	return err
}

// GetAllVentas retorna todas las ventas (paginadas)
func GetAllVentas(includeCanceladas bool, limit, offset int) ([]models.VentaStats, error) {
	whereClause := ""
	if !includeCanceladas {
		whereClause = "WHERE v.estado != 'cancelada'"
	}

	// 1. Obtener solo las ventas (sin detalles) con paginación
	ventasQuery := `
		SELECT v.id, ve.nombre, COALESCE(c.nombre, 'Sin cliente'), 
		       c.telefono, v.total, v.payment_method, v.estado, v.tipo_entrega, v.created_at
		FROM ventas v
		JOIN vendedores ve ON v.vendedor_id = ve.id
		LEFT JOIN clientes c ON v.cliente_id = c.id
		` + whereClause + `
		ORDER BY v.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := DB.Query(ventasQuery, limit, offset)
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
		var encVendedor, encCliente string
		var encTelefono sql.NullString
		if err := rows.Scan(&v.ID, &encVendedor, &encCliente, &encTelefono, &v.Total, &v.PaymentMethod, &v.Estado, &v.TipoEntrega, &v.CreatedAt); err != nil {
			return nil, err
		}
		vendedor, err := security.Decrypt(encVendedor)
		if err != nil {
			return nil, fmt.Errorf("error desencriptando vendedor en venta %d: %w", v.ID, err)
		}
		v.Vendedor = vendedor
		// El cliente puede ser 'Sin cliente' (literal no encriptado) o un nombre encriptado
		if decCliente, err := security.Decrypt(encCliente); err == nil {
			v.Cliente = decCliente
		} else {
			v.Cliente = encCliente
		}
		if encTelefono.Valid && encTelefono.String != "" {
			if decTel, err := security.Decrypt(encTelefono.String); err == nil {
				var tel int
				fmt.Sscanf(decTel, "%d", &tel)
				v.TelefonoCliente = &tel
			}
		} else {
			v.TelefonoCliente = nil
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
			placeholders += fmt.Sprintf("$%d", i+1)
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
			COALESCE(SUM(CASE WHEN v.estado='sin_pagar' THEN v.total ELSE 0 END), 0) as pendiente,
			COALESCE(SUM(CASE WHEN v.estado='pagada' OR v.estado='entregada' THEN v.total ELSE 0 END), 0) as total_cobrado,
			COUNT(CASE WHEN v.estado='sin_pagar' THEN 1 END) as ventas_sin_pagar,
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

	// Calcular cantidad de ventas por tipo de entrega
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

	// Calcular productos vendidos (tal cual se vendieron, incluyendo Combos)
	productosQuery := `
		SELECT p.tipo_pizza as nombre, COALESCE(SUM(dv.cantidad), 0) as cantidad, p.precio, p.es_combo
		FROM detalle_ventas dv
		JOIN ventas v ON dv.venta_id = v.id
		JOIN productos p ON dv.producto_id = p.id
		WHERE v.estado != 'cancelada'
		GROUP BY p.id, p.tipo_pizza, p.precio, p.es_combo
		ORDER BY cantidad DESC
	`
	rows, err := DB.Query(productosQuery)
	productosVendidos := []map[string]interface{}{}
	if err == nil {
		for rows.Next() {
			var nombre string
			var cantidad int
			var precio float64
			var esCombo bool
			if err := rows.Scan(&nombre, &cantidad, &precio, &esCombo); err == nil {
				productosVendidos = append(productosVendidos, map[string]interface{}{
					"nombre":   nombre,
					"cantidad": cantidad,
					"precio":   precio,
					"es_combo": esCombo,
				})
			}
		}
		rows.Close()
	}

	// Calcular desglose de componentes individuales de los combos
	desgloseQuery := `
		SELECT nombre, SUM(cantidad) as cantidad FROM (
			SELECT p.tipo_pizza as nombre, SUM(dv.cantidad) as cantidad
			FROM detalle_ventas dv
			JOIN productos p ON dv.producto_id = p.id
			JOIN ventas v ON dv.venta_id = v.id
			WHERE v.estado != 'cancelada' AND p.es_combo = FALSE
			GROUP BY p.id, p.tipo_pizza
			UNION ALL
			SELECT pc.tipo_pizza as nombre, SUM(dv.cantidad * cp.cantidad) as cantidad
			FROM detalle_ventas dv
			JOIN productos p ON dv.producto_id = p.id
			JOIN ventas v ON dv.venta_id = v.id
			JOIN combo_productos cp ON p.id = cp.combo_id
			JOIN productos pc ON cp.producto_id = pc.id
			WHERE v.estado != 'cancelada' AND p.es_combo = TRUE
			GROUP BY pc.id, pc.tipo_pizza
		) desglose
		GROUP BY nombre
		ORDER BY cantidad DESC
	`
	rowsDesglose, err := DB.Query(desgloseQuery)
	productosDesglosados := []map[string]interface{}{}
	if err == nil {
		for rowsDesglose.Next() {
			var nombre string
			var cantidad int
			if err := rowsDesglose.Scan(&nombre, &cantidad); err == nil {
				productosDesglosados = append(productosDesglosados, map[string]interface{}{
					"nombre":   nombre,
					"cantidad": cantidad,
				})
			}
		}
		rowsDesglose.Close()
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
		"productos_vendidos":    productosVendidos,
		"productos_desglosados": productosDesglosados,
	}, nil
}

// GetVendedoresConStats retorna vendedores con estadísticas detalladas sin bucles N+1
func GetVendedoresConStats() ([]map[string]interface{}, error) {
	vendedores, err := GetVendedores()
	if err != nil {
		return nil, err
	}
	
	// Crear mapa para acceso rápido
	vendedoresMap := make(map[int]map[string]interface{})
	for _, v := range vendedores {
		vendedoresMap[v.ID] = map[string]interface{}{
			"nombre":                v.Nombre,
			"cantidad":              0,
			"total_items":           0,
			"deuda":                 0.0,
			"deuda_efectivo":        0.0,
			"deuda_transferencia":   0.0,
			"pagado":                0.0,
			"pagado_efectivo":       0.0,
			"pagado_transferencia":  0.0,
			"total":                 0.0,
			"productos_vendidos":    []map[string]interface{}{},
			"deudores":              []map[string]interface{}{},
		}
	}

	// 1. Obtener métricas agregadas por vendedor (ventas, pagos, deudas)
	statsQuery := `
		SELECT 
			ve.id as vendedor_id,
			COUNT(DISTINCT v.id) as cantidad,
			COALESCE(SUM(CASE WHEN v.estado='sin_pagar' THEN v.total ELSE 0 END), 0) as deuda,
			COALESCE(SUM(CASE WHEN v.estado='sin_pagar' AND v.payment_method='efectivo' THEN v.total ELSE 0 END), 0) as deuda_efectivo,
			COALESCE(SUM(CASE WHEN v.estado='sin_pagar' AND v.payment_method='transferencia' THEN v.total ELSE 0 END), 0) as deuda_transferencia,
			COALESCE(SUM(CASE WHEN v.estado IN ('pagada', 'entregada') THEN v.total ELSE 0 END), 0) as pagado,
			COALESCE(SUM(CASE WHEN v.estado IN ('pagada', 'entregada') AND v.payment_method='efectivo' THEN v.total ELSE 0 END), 0) as pagado_efectivo,
			COALESCE(SUM(CASE WHEN v.estado IN ('pagada', 'entregada') AND v.payment_method='transferencia' THEN v.total ELSE 0 END), 0) as pagado_transferencia,
			COALESCE(SUM(v.total), 0) as total
		FROM vendedores ve
		LEFT JOIN ventas v ON v.vendedor_id = ve.id AND v.estado != 'cancelada'
		GROUP BY ve.id
	`
	rows, err := DB.Query(statsQuery)
	if err == nil {
		for rows.Next() {
			var vId, cantidad int
			var deuda, deudaEf, deudaTr, pagado, pagadoEf, pagadoTr, total float64
			if err := rows.Scan(&vId, &cantidad, &deuda, &deudaEf, &deudaTr, &pagado, &pagadoEf, &pagadoTr, &total); err == nil {
				if vm, ok := vendedoresMap[vId]; ok {
					vm["cantidad"] = cantidad
					vm["deuda"] = deuda
					vm["deuda_efectivo"] = deudaEf
					vm["deuda_transferencia"] = deudaTr
					vm["pagado"] = pagado
					vm["pagado_efectivo"] = pagadoEf
					vm["pagado_transferencia"] = pagadoTr
					vm["total"] = total
				}
			}
		}
		rows.Close()
	}

	// 2. Obtener productos vendidos por vendedor
	productosQuery := `
		SELECT v.vendedor_id, p.tipo_pizza as nombre, SUM(dv.cantidad) as cantidad
		FROM detalle_ventas dv
		JOIN ventas v ON dv.venta_id = v.id
		JOIN productos p ON dv.producto_id = p.id
		WHERE v.estado != 'cancelada'
		GROUP BY v.vendedor_id, p.id, p.tipo_pizza
	`
	pRows, err := DB.Query(productosQuery)
	if err == nil {
		for pRows.Next() {
			var vId, cantidad int
			var nombre string
			if err := pRows.Scan(&vId, &nombre, &cantidad); err == nil {
				if vm, ok := vendedoresMap[vId]; ok {
					productos := vm["productos_vendidos"].([]map[string]interface{})
					productos = append(productos, map[string]interface{}{
						"nombre":   nombre,
						"cantidad": cantidad,
					})
					vm["productos_vendidos"] = productos
					vm["total_items"] = vm["total_items"].(int) + cantidad
				}
			}
		}
		pRows.Close()
	}

	// 3. Obtener lista de deudores por vendedor (nombres desencriptados)
	deudoresQuery := `
		SELECT v.vendedor_id, COALESCE(c.nombre, '') as cliente, v.payment_method, v.total
		FROM ventas v
		LEFT JOIN clientes c ON v.cliente_id = c.id
		WHERE v.estado = 'sin_pagar'
	`
	dRows, err := DB.Query(deudoresQuery)
	if err == nil {
		for dRows.Next() {
			var vId int
			var encCliente, paymentMethod string
			var total float64
			if err := dRows.Scan(&vId, &encCliente, &paymentMethod, &total); err == nil {
				cliente := encCliente
				if encCliente != "" {
					if dec, err := security.Decrypt(encCliente); err == nil {
						cliente = dec
					}
				} else {
					cliente = "Sin cliente"
				}
				if vm, ok := vendedoresMap[vId]; ok {
					deudores := vm["deudores"].([]map[string]interface{})
					deudores = append(deudores, map[string]interface{}{
						"cliente":        cliente,
						"payment_method": paymentMethod,
						"total":          total,
					})
					vm["deudores"] = deudores
				}
			}
		}
		dRows.Close()
	}

	// Reconstruir array resultante en el orden original de vendedores
	var result []map[string]interface{}
	for _, v := range vendedores {
		result = append(result, vendedoresMap[v.ID])
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
	query := `UPDATE ventas SET estado = $1, payment_method = $2, tipo_entrega = $3 WHERE id = $4`
	if _, err = tx.Exec(query, estado, paymentMethod, tipoEntrega, ventaID); err != nil {
		return fmt.Errorf("error actualizando cabecera venta: %w", err)
	}

	// 3. Eliminar productos (Batch)
	if len(productosEliminar) > 0 {
		// Nota de eficiencia: Podríamos usar IN (?) dinámico, pero un loop simple dentro de Tx es aceptable para pocos items
		deleteQuery := `DELETE FROM detalle_ventas WHERE id = $1`
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
	insertStmt, err := tx.Prepare(`INSERT INTO detalle_ventas (venta_id, producto_id, cantidad, precio_unitario, subtotal) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	updateStmt, err := tx.Prepare(`UPDATE detalle_ventas SET cantidad = $1, subtotal = $2 WHERE id = $3`)
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
		err = tx.QueryRow("SELECT precio FROM productos WHERE id = $1", productoID).Scan(&precio)
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
	totalQuery := `SELECT COALESCE(SUM(subtotal), 0) FROM detalle_ventas WHERE venta_id = $1`
	if err = tx.QueryRow(totalQuery, ventaID).Scan(&nuevoTotal); err != nil {
		return fmt.Errorf("error recalculando total: %w", err)
	}

	if _, err = tx.Exec(`UPDATE ventas SET total = $1 WHERE id = $2`, nuevoTotal, ventaID); err != nil {
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
		SELECT id, tipo_pizza, descripcion, precio, activo, es_combo, created_at
		FROM productos WHERE id = $1
	`, id).Scan(&p.ID, &p.TipoPizza, &p.Descripcion, &p.Precio, &p.Activo, &p.EsCombo, &p.CreatedAt)

	if err != nil {
		return nil, err
	}

	if p.EsCombo {
		compRows, err := DB.Query(`SELECT combo_id, producto_id, cantidad FROM combo_productos WHERE combo_id = $1`, p.ID)
		if err == nil {
			for compRows.Next() {
				var comp models.ComboProducto
				if err := compRows.Scan(&comp.ComboID, &comp.ProductoID, &comp.Cantidad); err == nil {
					p.Componentes = append(p.Componentes, comp)
				}
			}
			compRows.Close()
		}
	}

	return &p, nil
}

// GetUserByCredentials obtiene un usuario por credenciales; busca por HMAC del username
func GetUserByCredentials(username, plainPassword string) (*models.User, error) {
	var user models.User
	var encUsername, storedHash string
	hash := security.HMACField(username)
	err := DB.QueryRow(
		"SELECT id, username, rol, password_hash FROM usuarios WHERE username_hash = $1",
		hash).Scan(&user.ID, &encUsername, &user.Rol, &storedHash)

	if err != nil {
		return nil, err
	}

	decUsername, err := security.Decrypt(encUsername)
	if err != nil {
		return nil, fmt.Errorf("error desencriptando username: %w", err)
	}
	user.Username = decUsername

	if !VerifyPassword(storedHash, plainPassword) {
		return nil, fmt.Errorf("contraseña inválida")
	}

	return &user, nil
}

// CreateProducto crea un nuevo producto
func CreateProducto(req *models.CrearProductoRequest) (int64, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("error iniciando transacción: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	var id int64
	err = tx.QueryRow(
		"INSERT INTO productos (tipo_pizza, descripcion, precio, activo, es_combo) VALUES ($1, $2, $3, TRUE, $4) RETURNING id",
		req.TipoPizza, req.Descripcion, req.Precio, req.EsCombo,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	if req.EsCombo && len(req.Componentes) > 0 {
		stmt, err := tx.Prepare("INSERT INTO combo_productos (combo_id, producto_id, cantidad) VALUES ($1, $2, $3)")
		if err != nil {
			return 0, err
		}
		defer stmt.Close()
		for _, comp := range req.Componentes {
			if _, err = stmt.Exec(id, comp.ProductoID, comp.Cantidad); err != nil {
				return 0, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateProducto actualiza un producto
func UpdateProducto(id int, req *models.ActualizarProductoRequest) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(
		"UPDATE productos SET tipo_pizza = $1, precio = $2, descripcion = $3, activo = $4, es_combo = $5 WHERE id = $6",
		req.TipoPizza, req.Precio, req.Descripcion, req.Activo, req.EsCombo, id,
	)
	if err != nil {
		return err
	}

	if req.EsCombo {
		_, err = tx.Exec("DELETE FROM combo_productos WHERE combo_id = $1", id)
		if err != nil {
			return err
		}
		if len(req.Componentes) > 0 {
			stmt, err := tx.Prepare("INSERT INTO combo_productos (combo_id, producto_id, cantidad) VALUES ($1, $2, $3)")
			if err != nil {
				return err
			}
			defer stmt.Close()
			for _, comp := range req.Componentes {
				if _, err = stmt.Exec(id, comp.ProductoID, comp.Cantidad); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

// DeleteProducto desactiva un producto
func DeleteProducto(id int) error {
	result, err := DB.Exec("UPDATE productos SET activo = FALSE WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// CreateVendedor crea un nuevo vendedor (nombre encriptado)
func CreateVendedor(nombre string) (int64, error) {
	encNombre, err := security.Encrypt(nombre)
	if err != nil {
		return 0, fmt.Errorf("error encriptando nombre vendedor: %w", err)
	}
	nombreHash := security.HMACField(nombre)
	var id int64
	err = DB.QueryRow(
		`INSERT INTO vendedores (nombre, nombre_hash) VALUES ($1, $2) RETURNING id`,
		encNombre, nombreHash,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateVendedor actualiza un vendedor (nombre encriptado)
func UpdateVendedor(id int, nombre string) error {
	encNombre, err := security.Encrypt(nombre)
	if err != nil {
		return fmt.Errorf("error encriptando nombre vendedor: %w", err)
	}
	nombreHash := security.HMACField(nombre)
	result, err := DB.Exec(
		`UPDATE vendedores SET nombre = $1, nombre_hash = $2 WHERE id = $3`,
		encNombre, nombreHash, id,
	)
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
	result, err := DB.Exec(`DELETE FROM vendedores WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetAllUsers obtiene todos los usuarios sin contraseñas (usernames desencriptados)
func GetAllUsers() ([]models.User, error) {
	rows, err := DB.Query("SELECT id, username, rol FROM usuarios ORDER BY username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []models.User
	for rows.Next() {
		var usuario models.User
		var encUsername string
		if err := rows.Scan(&usuario.ID, &encUsername, &usuario.Rol); err != nil {
			return nil, err
		}
		decUsername, err := security.Decrypt(encUsername)
		if err != nil {
			return nil, fmt.Errorf("error desencriptando username %d: %w", usuario.ID, err)
		}
		usuario.Username = decUsername
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

// UserExists verifica si un usuario existe buscando por HMAC del username
func UserExists(username string) (bool, error) {
	var exists bool
	hash := security.HMACField(username)
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE username_hash = $1)", hash).Scan(&exists)
	return exists, err
}

// CreateUser crea un nuevo usuario con contraseña hasheada y username encriptado
func CreateUser(username, password, rol string) (int, error) {
	passHash, err := HashPassword(password)
	if err != nil {
		return 0, err
	}

	encUsername, err := security.Encrypt(username)
	if err != nil {
		return 0, fmt.Errorf("error encriptando username: %w", err)
	}
	usernameHash := security.HMACField(username)

	var id int64
	err = DB.QueryRow(
		"INSERT INTO usuarios (username, username_hash, password_hash, rol) VALUES ($1, $2, $3, $4) RETURNING id",
		encUsername, usernameHash, passHash, rol,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// UpdateUser actualiza un usuario existente (username encriptado)
func UpdateUser(id int, username, password, rol string) error {
	encUsername, err := security.Encrypt(username)
	if err != nil {
		return fmt.Errorf("error encriptando username: %w", err)
	}
	usernameHash := security.HMACField(username)

	var result sql.Result
	if password != "" {
		passHash, err := HashPassword(password)
		if err != nil {
			return err
		}
		result, err = DB.Exec(
			"UPDATE usuarios SET username = $1, username_hash = $2, password_hash = $3, rol = $4 WHERE id = $5",
			encUsername, usernameHash, passHash, rol, id,
		)
		if err != nil {
			return err
		}
	} else {
		result, err = DB.Exec(
			"UPDATE usuarios SET username = $1, username_hash = $2, rol = $3 WHERE id = $4",
			encUsername, usernameHash, rol, id,
		)
		if err != nil {
			return err
		}
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteUser elimina un usuario
func DeleteUser(id int) error {
	result, err := DB.Exec("DELETE FROM usuarios WHERE id = $1", id)
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
