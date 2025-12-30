package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

// Base de datos global
var db *sql.DB

// JWT Secret key para firmar tokens
var jwtSecret = []byte("ecos-auth-secret-key-change-in-production")

// Estructuras de datos (idénticas al original para compatibilidad)
type ProductoItem struct {
	DetalleID int     `json:"detalle_id"` // id de la fila en detalle_ventas (para edición)
	Tipo      string  `json:"tipo"`       // "producto"
	ProductID int     `json:"product_id"` // producto_id
	Cantidad  int     `json:"cantidad"`   // cantidad
	Precio    float64 `json:"precio"`     // precio unitario
	Total     float64 `json:"total"`      // total (precio * cantidad)
}

type VentaRequest struct {
	Vendedor      string         `json:"vendedor"`
	Cliente       string         `json:"cliente"`
	Items         []ProductoItem `json:"items"` // array de items con producto_id
	PaymentMethod string         `json:"payment_method"`
	Estado        string         `json:"estado"`
	TipoEntrega   string         `json:"tipo_entrega"` // retiro o envio
}

type DataResponse struct {
	ClientesPorVendedor map[string][]string `json:"clientesPorVendedor"`
	Vendedores          []Vendedor          `json:"vendedores"`
	Productos           []Producto          `json:"productos"`
}

type Pizza struct {
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Precios     []float64 `json:"precios"`
}

type VentaStats struct {
	ID            int            `json:"id"`
	Vendedor      string         `json:"vendedor"`
	Cliente       string         `json:"cliente"`
	Total         float64        `json:"total"`
	PaymentMethod string         `json:"payment_method"`
	Estado        string         `json:"estado"`
	TipoEntrega   string         `json:"tipo_entrega"`
	CreatedAt     time.Time      `json:"created_at"`
	Items         []ProductoItem `json:"items"`
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

// Producto structs
type Producto struct {
	ID          int       `json:"id"`
	TipoPizza   string    `json:"tipo_pizza"`
	Descripcion string    `json:"descripcion"`
	Precio      float64   `json:"precio"`
	Activo      bool      `json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
}

type Vendedor struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type CrearProductoRequest struct {
	TipoPizza   string  `json:"tipo_pizza"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
}

type ActualizarProductoRequest struct {
	TipoPizza   string  `json:"tipo_pizza"`
	Precio      float64 `json:"precio"`
	Descripcion string  `json:"descripcion"`
	Activo      bool    `json:"activo"`
}

// inicDB inicializa la conexión a MySQL
func inicDB() error {
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	caCertPath := os.Getenv("DATABASE_CA_CERT")

	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL no configurada")
	}

	log.Printf("📍 DATABASE_URL configurada: %s", strings.Split(dbURL, "@")[0]+"@...")
	log.Printf("📍 DATABASE_CA_CERT: %v", caCertPath)

	// Registrar TLS config si hay certificado y existe el archivo
	hasCert := false
	if caCertPath != "" && caCertPath != "disabled" {
		caCert, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			log.Printf("⚠️  No se pudo leer certificado en %s: %v. Intentando sin TLS...", caCertPath, err)
		} else {
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				log.Printf("⚠️  No se pudo parsear certificado. Intentando sin TLS...")
			} else {
				tlsConfig := &tls.Config{
					RootCAs: caCertPool,
				}
				mysql.RegisterTLSConfig("custom", tlsConfig)
				hasCert = true
				log.Println("✅ Certificado TLS configurado")
			}
		}
	}

	// Convertir URL a DSN
	dsn := convertDSN(dbURL, hasCert)
	log.Printf("🔌 DSN preparado, intentando conexión...")

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error abriendo conexión: %w", err)
	}

	// Configurar pool de conexiones
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Probar conexión
	if err := db.Ping(); err != nil {
		return fmt.Errorf("error conectando a BD: %w", err)
	}

	log.Println("✅ Conectado a MySQL exitosamente")
	return nil
}

// convertDSN convierte URL MySQL a DSN
func convertDSN(url string, hasCert bool) string {
	url = strings.TrimPrefix(url, "mysql://")

	var credentials, rest string
	if idx := strings.IndexByte(url, '@'); idx != -1 {
		credentials = url[:idx]
		rest = url[idx+1:]
	}

	var host, dbPath string
	if idx := strings.IndexByte(rest, '/'); idx != -1 {
		host = rest[:idx]
		dbPath = rest[idx:]
	} else {
		host = rest
		dbPath = "/"
	}

	// Remover parámetros de la URL original
	if idx := strings.IndexByte(dbPath, '?'); idx != -1 {
		dbPath = dbPath[:idx]
	}

	suffix := "?parseTime=true"
	if hasCert {
		suffix = "?tls=custom&parseTime=true"
	}

	return fmt.Sprintf("%s@tcp(%s)%s%s", credentials, host, dbPath, suffix)
}

func main() {
	// Inicializar BD
	if err := inicDB(); err != nil {
		log.Fatalf("❌ Error inicializando BD: %v", err)
	}
	defer db.Close()

	// Middleware CORS con recover
	mux := http.NewServeMux()

	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("❌ PANIC en %s %s: %v", r.Method, r.URL.Path, err)
					w.Header().Set("Content-Type", "application/json")
					http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
				}
			}()

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	// Rutas API
	mux.HandleFunc("/api/login", handleLogin)                            // Login
	mux.HandleFunc("/api/verify-token", handleVerifyToken)               // Verificar token
	mux.HandleFunc("/api/submit", handleSubmit)                          // Guardar venta
	mux.HandleFunc("/api/data", handleData)                              // Obtener vendedores, clientes, pizzas
	mux.HandleFunc("/api/estadisticas", handleEstadisticas)              // Todas las ventas
	mux.HandleFunc("/api/estadisticas-sheet", handleEstadisticasSheet)   // Resumen
	mux.HandleFunc("/api/actualizar-venta", handleActualizarVenta)       // Actualizar venta
	mux.HandleFunc("/api/productos", handleProductos)                    // Listar productos
	mux.HandleFunc("/api/crear-producto", handleCrearProducto)           // Crear producto
	mux.HandleFunc("/api/actualizar-producto", handleActualizarProducto) // Actualizar producto
	mux.HandleFunc("/api/eliminar-producto", handleEliminarProducto)     // Eliminar producto
	mux.HandleFunc("/api/crear-vendedor", handleCrearVendedor)           // Crear vendedor
	mux.HandleFunc("/api/actualizar-vendedor", handleActualizarVendedor) // Actualizar vendedor
	mux.HandleFunc("/api/eliminar-vendedor", handleEliminarVendedor)     // Eliminar vendedor

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Backend en puerto %s", port)
	log.Printf("Iniciando servidor HTTP...")
	err := http.ListenAndServe(":"+port, corsHandler(mux))
	if err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}

// handleData retorna vendedores, clientes y productos
func handleData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Obtener vendedores
	vendedores, err := getVendedores()
	if err != nil {
		log.Printf("❌ Error en getVendedores: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error obteniendo vendedores: %v", err)})
		return
	}

	// Obtener clientes por vendedor
	clientesPorVendedor, err := getClientesPorVendedor()
	if err != nil {
		log.Printf("❌ Error en getClientesPorVendedor: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error obteniendo clientes: %v", err)})
		return
	}

	// Obtener productos
	productos, err := getProductos()
	if err != nil {
		log.Printf("❌ Error en getProductos: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error obteniendo productos: %v", err)})
		return
	}

	response := DataResponse{
		Vendedores:          vendedores,
		ClientesPorVendedor: clientesPorVendedor,
		Productos:           productos,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// handleSubmit guarda una nueva venta
func handleSubmit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req VentaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando JSON", http.StatusBadRequest)
		return
	}

	// Validar
	if req.Vendedor == "" || len(req.Items) == 0 {
		http.Error(w, "Vendedor e items requeridos", http.StatusBadRequest)
		return
	}

	// Obtener ID del vendedor
	vendedorID, err := getVendedorID(req.Vendedor)
	if err != nil {
		http.Error(w, fmt.Sprintf("Vendedor no encontrado: %v", err), http.StatusBadRequest)
		return
	}

	// Obtener o crear cliente
	var clienteID *int
	if req.Cliente != "" {
		id, err := getOrCreateCliente(req.Cliente)
		if err == nil {
			clienteID = &id
		}
	}

	// Calcular total
	total := 0.0
	for _, item := range req.Items {
		total += item.Total
	}

	// Insertar venta
	ventaID, err := insertVenta(clienteID, vendedorID, total, req.PaymentMethod, req.Estado, req.TipoEntrega)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error guardando venta: %v", err), http.StatusInternalServerError)
		return
	}

	// Insertar detalles de venta
	for _, item := range req.Items {
		if err := insertDetalle(ventaID, item); err != nil {
			log.Printf("Error insertando detalle: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "id": ventaID}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// handleEstadisticas retorna todas las ventas (incluyendo canceladas)
func handleEstadisticas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ventas, err := getAllVentas(true) // true = incluir canceladas
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ventas); err != nil {
		log.Printf("Error encoding ventas: %v", err)
	}
}

// handleEstadisticasSheet retorna resumen y vendedores
func handleEstadisticasSheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resumen, err := getResumen()
	if err != nil {
		log.Printf("Error en getResumen: %v", err)
		http.Error(w, fmt.Sprintf("Error obteniendo resumen: %v", err), http.StatusInternalServerError)
		return
	}

	vendedores, err := getVendedoresConStats()
	if err != nil {
		log.Printf("Error en getVendedoresConStats: %v", err)
		http.Error(w, fmt.Sprintf("Error obteniendo vendedores: %v", err), http.StatusInternalServerError)
		return
	}

	ventas, err := getAllVentas(false) // false = sin canceladas (solo para gráficos)
	if err != nil {
		log.Printf("Error en getAllVentas: %v", err)
		http.Error(w, fmt.Sprintf("Error obteniendo ventas: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"resumen":    resumen,
		"vendedores": vendedores,
		"ventas":     ventas,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding sheet response: %v", err)
	}
}

// handleActualizarVenta actualiza una venta (estado, pago, y productos)
func handleActualizarVenta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando JSON", http.StatusBadRequest)
		return
	}

	ventaID := int(req["id"].(float64))
	estado := req["estado"].(string)
	paymentMethod := req["payment_method"].(string)
	tipoEntrega := req["tipo_entrega"].(string)

	// 1. Actualizar estado, payment_method y tipo_entrega de la venta
	query := `UPDATE ventas SET estado = ?, payment_method = ?, tipo_entrega = ? WHERE id = ?`
	_, err := db.Exec(query, estado, paymentMethod, tipoEntrega, ventaID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error actualizando venta: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Eliminar productos si se proporcionan
	if eliminar, ok := req["productos_eliminar"].([]interface{}); ok && len(eliminar) > 0 {
		for _, id := range eliminar {
			detalleID := int(id.(float64))
			deleteQuery := `DELETE FROM detalle_ventas WHERE id = ?`
			_, err := db.Exec(deleteQuery, detalleID)
			if err != nil {
				log.Printf("Error eliminando producto: %v", err)
				http.Error(w, fmt.Sprintf("Error eliminando producto: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	// 3. Actualizar/insertar productos si se proporcionan
	if productos, ok := req["productos"].([]interface{}); ok {
		for _, p := range productos {
			item := p.(map[string]interface{})

			detalleID := item["detalle_id"]
			productoID := int(item["producto_id"].(float64))
			cantidad := int(item["cantidad"].(float64))

			if detalleID == nil {
				// Nuevo producto - insertar en detalle_ventas
				insertQuery := `INSERT INTO detalle_ventas (venta_id, producto_id, cantidad) VALUES (?, ?, ?)`
				_, err := db.Exec(insertQuery, ventaID, productoID, cantidad)
				if err != nil {
					log.Printf("Error insertando producto: %v", err)
					http.Error(w, fmt.Sprintf("Error insertando producto: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				// Actualizar cantidad existente
				detalleIDInt := int(detalleID.(float64))
				updateQuery := `UPDATE detalle_ventas SET cantidad = ? WHERE id = ?`
				_, err := db.Exec(updateQuery, cantidad, detalleIDInt)
				if err != nil {
					log.Printf("Error actualizando producto: %v", err)
					http.Error(w, fmt.Sprintf("Error actualizando producto: %v", err), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	// 4. Recalcular total de la venta
	var nuevoTotal float64
	totalQuery := `SELECT COALESCE(SUM(dv.cantidad * p.precio), 0) FROM detalle_ventas dv JOIN productos p ON dv.producto_id = p.id WHERE dv.venta_id = ?`
	err = db.QueryRow(totalQuery, ventaID).Scan(&nuevoTotal)
	if err != nil {
		log.Printf("Error calculando nuevo total: %v", err)
	} else {
		_, err := db.Exec(`UPDATE ventas SET total = ? WHERE id = ?`, nuevoTotal, ventaID)
		if err != nil {
			log.Printf("Error actualizando total: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"success": true}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Funciones auxiliares de BD

func getVendedores() ([]Vendedor, error) {
	rows, err := db.Query("SELECT id, nombre FROM vendedores ORDER BY nombre")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendedores []Vendedor
	for rows.Next() {
		var vendedor Vendedor
		if err := rows.Scan(&vendedor.ID, &vendedor.Nombre); err != nil {
			return nil, err
		}
		vendedores = append(vendedores, vendedor)
	}

	return vendedores, nil
}

func getVendedorID(nombre string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM vendedores WHERE nombre = ?", nombre).Scan(&id)
	return id, err
}

func getClientesPorVendedor() (map[string][]string, error) {
	result := make(map[string][]string)

	// Get clientes grouped by vendedor from actual sales
	query := `
		SELECT ve.nombre, c.nombre, c.apellido
		FROM ventas v
		JOIN vendedores ve ON v.vendedor_id = ve.id
		LEFT JOIN clientes c ON v.cliente_id = c.id
		WHERE c.id IS NOT NULL
		ORDER BY ve.nombre, c.nombre, c.apellido
	`

	rows, err := db.Query(query)
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

		// Avoid duplicates
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

func getOrCreateCliente(nombre string) (int, error) {
	// Buscar si existe
	var id int
	err := db.QueryRow("SELECT id FROM clientes WHERE nombre = ?", nombre).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Crear si no existe
	res, err := db.Exec("INSERT INTO clientes (nombre) VALUES (?)", nombre)
	if err != nil {
		return 0, err
	}

	idInt, _ := res.LastInsertId()
	return int(idInt), nil
}

// getProductos retorna lista de productos activos
func getProductos() ([]Producto, error) {
	var productos []Producto

	rows, err := db.Query(`
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
		var p Producto
		if err := rows.Scan(&p.ID, &p.TipoPizza, &p.Descripcion, &p.Precio, &p.Activo, &p.CreatedAt); err != nil {
			return nil, err
		}
		productos = append(productos, p)
	}

	return productos, nil
}

func insertVenta(clienteID *int, vendedorID int, total float64, payment, estado, tipoEntrega string) (int, error) {
	query := `
		INSERT INTO ventas (cliente_id, vendedor_id, total, payment_method, estado, tipo_entrega)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	res, err := db.Exec(query, clienteID, vendedorID, total, payment, estado, tipoEntrega)
	if err != nil {
		return 0, err
	}

	id, _ := res.LastInsertId()
	return int(id), nil
}

func insertDetalle(ventaID int, item ProductoItem) error {
	// item.ProductID es el producto_id
	productoID := item.ProductID

	query := `
		INSERT INTO detalle_ventas (venta_id, producto_id, cantidad, precio_unitario, subtotal)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, ventaID, productoID, item.Cantidad, item.Precio, item.Total)
	return err
}

func getAllVentas(includeCanceladas bool) ([]VentaStats, error) {
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

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ventas []VentaStats
	for rows.Next() {
		var v VentaStats

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
		itemRows, err := db.Query(itemsQuery, v.ID)
		if err == nil {
			for itemRows.Next() {
				var item ProductoItem
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
			v.Items = []ProductoItem{} // Array vacío en lugar de null
		}

		ventas = append(ventas, v)
	}

	return ventas, nil
}

func getResumen() (map[string]interface{}, error) {
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

	err := db.QueryRow(query).Scan(&efectivo, &transferencia, &pendiente, &total, &sinPagar, &pagadas, &entregadas, &totalVentas)
	if err != nil {
		log.Printf("Error en getResumen: %v", err)
		return nil, err
	}

	// Ahora calcular items por separado
	itemsQuery := `
		SELECT 
			COALESCE(SUM(dv.cantidad), 0) as total_items,
			COALESCE(SUM(CASE WHEN v.tipo_entrega IN ('delivery', 'envio') THEN dv.cantidad ELSE 0 END), 0) as total_delivery,
			COALESCE(SUM(CASE WHEN v.tipo_entrega='retiro' THEN dv.cantidad ELSE 0 END), 0) as total_retiro
		FROM detalle_ventas dv
		JOIN ventas v ON dv.venta_id = v.id
		WHERE v.estado != 'cancelada'
	`

	var totalItems, delivery, retiro int
	err = db.QueryRow(itemsQuery).Scan(&totalItems, &delivery, &retiro)
	if err != nil {
		log.Printf("Error en getResumen items: %v", err)
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

func getVendedoresConStats() ([]map[string]interface{}, error) {
	vendedores, _ := getVendedores()
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

		err := db.QueryRow(query, vendedor.Nombre).Scan(&cantidad, &deuda, &pagado, &total)
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
		err = db.QueryRow(itemsQuery, vendedor.Nombre).Scan(&totalItems)
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

// handleLogin autentica usuario y devuelve JWT token
func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "Usuario y contraseña requeridos", http.StatusBadRequest)
		return
	}

	// Hashear la contraseña ingresada
	hash := sha256.Sum256([]byte(loginReq.Password))
	passwordHash := hex.EncodeToString(hash[:])

	// Verificar credenciales en base de datos
	var user User
	err := db.QueryRow(
		"SELECT id, username, rol FROM usuarios WHERE username = ? AND password_hash = ?",
		loginReq.Username, passwordHash).Scan(&user.ID, &user.Username, &user.Rol)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		} else {
			http.Error(w, fmt.Sprintf("Error en base de datos: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Generar JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		Username: user.Username,
		Rol:      user.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generando token: %v", err), http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Token: tokenString,
		User:  user,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding login response: %v", err)
	}
}

// handleVerifyToken verifica y decodifica un JWT token
func handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	type TokenRequest struct {
		Token string `json:"token"`
	}

	var tokenReq TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&tokenReq); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if tokenReq.Token == "" {
		http.Error(w, "Token requerido", http.StatusBadRequest)
		return
	}

	// Parsear y validar token
	token, err := jwt.ParseWithClaims(tokenReq.Token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(*TokenClaims)

	response := map[string]interface{}{
		"valid":    true,
		"username": claims.Username,
		"rol":      claims.Rol,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding verify response: %v", err)
	}
}

// handleProductos obtiene la lista de productos
func handleProductos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`
		SELECT id, tipo_pizza, descripcion, precio, activo, created_at
		FROM productos
		WHERE activo = TRUE
		ORDER BY tipo_pizza
	`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error consultando productos: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var productos []Producto

	for rows.Next() {
		var p Producto
		if err := rows.Scan(&p.ID, &p.TipoPizza, &p.Descripcion, &p.Precio, &p.Activo, &p.CreatedAt); err != nil {
			continue
		}
		productos = append(productos, p)
	}

	if len(productos) == 0 {
		productos = []Producto{}
	}

	if err := json.NewEncoder(w).Encode(productos); err != nil {
		log.Printf("Error encoding productos: %v", err)
	}
}

// handleCrearProducto crea un nuevo producto
func handleCrearProducto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req CrearProductoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.TipoPizza == "" || req.Precio <= 0 {
		http.Error(w, "Tipo de pizza y precio son requeridos", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"INSERT INTO productos (tipo_pizza, descripcion, precio, activo) VALUES (?, ?, ?, TRUE)",
		req.TipoPizza, req.Descripcion, req.Precio,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			http.Error(w, "Este tipo de pizza ya existe", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Error creando producto: %v", err), http.StatusInternalServerError)
		}
		return
	}

	id, _ := result.LastInsertId()
	response := map[string]interface{}{
		"id":          id,
		"tipo_pizza":  req.TipoPizza,
		"descripcion": req.Descripcion,
		"precio":      req.Precio,
		"message":     "Producto creado exitosamente",
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding crear-producto response: %v", err)
	}
}

// handleActualizarProducto actualiza precio y descripción de un producto
func handleActualizarProducto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	productoID := r.URL.Query().Get("id")
	if productoID == "" {
		http.Error(w, "ID de producto requerido", http.StatusBadRequest)
		return
	}

	var req ActualizarProductoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		"UPDATE productos SET tipo_pizza = ?, precio = ?, descripcion = ?, activo = ? WHERE id = ?",
		req.TipoPizza, req.Precio, req.Descripcion, req.Activo, productoID,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error actualizando producto: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":      productoID,
		"message": "Producto actualizado exitosamente",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding actualizar-producto response: %v", err)
	}
}

// handleEliminarProducto elimina (desactiva) un producto
func handleEliminarProducto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	productoID := r.URL.Query().Get("id")
	if productoID == "" {
		http.Error(w, "ID de producto requerido", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE productos SET activo = FALSE WHERE id = ?", productoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error eliminando producto: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Producto no encontrado", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message": "Producto eliminado exitosamente",
		"id":      productoID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding eliminar-producto response: %v", err)
	}
}

// handleCrearVendedor crea un nuevo vendedor
func handleCrearVendedor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	nombre, ok := req["nombre"].(string)
	if !ok || nombre == "" {
		http.Error(w, "El nombre del vendedor es requerido", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`INSERT INTO vendedores (nombre) VALUES (?)`, nombre)
	if err != nil {
		log.Printf("Error creando vendedor: %v", err)
		http.Error(w, "Error creando vendedor", http.StatusInternalServerError)
		return
	}

	vendedorID, _ := result.LastInsertId()

	response := map[string]interface{}{
		"message": "Vendedor creado exitosamente",
		"id":      vendedorID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding crear-vendedor response: %v", err)
	}
}

// handleActualizarVendedor actualiza el nombre de un vendedor
func handleActualizarVendedor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vendedorID := r.URL.Query().Get("id")
	if vendedorID == "" {
		http.Error(w, "ID de vendedor requerido", http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	nombre, ok := req["nombre"].(string)
	if !ok || nombre == "" {
		http.Error(w, "El nombre del vendedor es requerido", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`UPDATE vendedores SET nombre = ? WHERE id = ?`, nombre, vendedorID)
	if err != nil {
		log.Printf("Error actualizando vendedor: %v", err)
		http.Error(w, "Error actualizando vendedor", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Vendedor no encontrado", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message": "Vendedor actualizado exitosamente",
		"id":      vendedorID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding actualizar-vendedor response: %v", err)
	}
}

// handleEliminarVendedor elimina un vendedor
func handleEliminarVendedor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	vendedorID := r.URL.Query().Get("id")
	if vendedorID == "" {
		http.Error(w, "ID de vendedor requerido", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`DELETE FROM vendedores WHERE id = ?`, vendedorID)
	if err != nil {
		log.Printf("Error eliminando vendedor: %v", err)
		http.Error(w, "Error eliminando vendedor", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Vendedor no encontrado", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message": "Vendedor eliminado exitosamente",
		"id":      vendedorID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding eliminar-vendedor response: %v", err)
	}
}
