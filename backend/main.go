package main

import (
	"context"
	//"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	//"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Configuración global
// Extrae el ID de la URL del Google Sheets
const SheetURL = "https://docs.google.com/spreadsheets/d/1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0/edit?gid=865373207"
const SpreadsheetID = "1E8bLD1DKp3ZrsmLb05O7cAJ-Qn929yBSTrZ18BSeVk0" // ID extraído de la URL

var srvSheets *sheets.Service

func main() {
	// 1. Autenticación con Google Sheets
	// Lee el archivo credentials.json en local o usa variable de entorno en producción
	var credsData []byte
	var err error

	// En Render/producción usa variable de entorno
	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credsJSON != "" {
		credsData = []byte(credsJSON)
	} else {
		// En local, lee del archivo (intenta múltiples rutas)
		credsPaths := []string{
			"venta-pizzas-ecos.json",
			"../venta-pizzas-ecos.json",
		}

		for _, path := range credsPaths {
			credsData, err = os.ReadFile(path)
			if err == nil {
				break
			}
		}

		if err != nil {
			log.Fatalf("No se pudo leer credentials desde ninguna ruta: %v", err)
		}
	}

	ctx := context.Background()
	srvSheets, err = sheets.NewService(ctx, option.WithCredentialsJSON(credsData))
	if err != nil {
		log.Fatalf("No se pudo iniciar el cliente de Sheets: %v", err)
	}

	// 2. Middleware CORS
	mux := http.NewServeMux()

	// Envolver todas las rutas con CORS
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Responder a preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	// Rutas API
	mux.HandleFunc("/api/submit", handleSubmit)                        // Recibe el POST del form
	mux.HandleFunc("/api/data", handleData)                            // Devuelve JSON para los selects
	mux.HandleFunc("/api/estadisticas", handleEstadisticas)            // Estadísticas generales (desde Ventas)
	mux.HandleFunc("/api/estadisticas-sheet", handleEstadisticasSheet) // Estadísticas desde sheet "estadisticas"
	mux.HandleFunc("/api/actualizar-venta", handleActualizarVenta)     // Actualizar estado de venta

	// 3. Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Backend escuchando en puerto %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(mux)))
}

// Estructura para recibir datos del frontend
type ComboItem struct {
	Tipo     string  `json:"tipo"`     // muzza o jamon
	Combo    int     `json:"combo"`    // 0, 1 o 2 (índices)
	Cantidad int     `json:"cantidad"` // cantidad de combos
	Precio   float64 `json:"precio"`   // precio unitario
	Total    float64 `json:"total"`    // total (precio * cantidad)
}

type VentaRequest struct {
	Vendedor      string      `json:"vendedor"`
	Cliente       string      `json:"cliente"`
	Combos        []ComboItem `json:"combos"` // array de combos
	PaymentMethod string      `json:"payment_method"`
	Estado        string      `json:"estado"`
	TipoEntrega   string      `json:"tipo_entrega"` // retiro o envio
}

// Estructura para responder con datos procesados
type DataResponse struct {
	ClientesPorVendedor map[string][]string `json:"clientesPorVendedor"`
	Vendedores          []string            `json:"vendedores"`
	Pizzas              map[string]Pizza    `json:"pizzas"`
}

type Pizza struct {
	Nombre  string    `json:"nombre"`
	Combos  []string  `json:"combos"`
	Precios []float64 `json:"precios"`
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parsear JSON del body
	var venta VentaRequest
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&venta); err != nil {
		http.Error(w, "Error al parsear JSON", http.StatusBadRequest)
		log.Printf("Error decodificando JSON: %v", err)
		return
	}

	// Validar datos
	if venta.Vendedor == "" || venta.Cliente == "" || len(venta.Combos) == 0 {
		http.Error(w, "Faltan datos requeridos", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Leer desde B4 hacia abajo para encontrar la primera fila libre
	respLastRow, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "Ventas!B4:B").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo última fila: %v", err)
		http.Error(w, "Error leyendo datos", http.StatusInternalServerError)
		return
	}

	// Calcular próximo ID: buscar la primera fila vacía en B4:B
	proximoID := 1
	filaVaciaIndex := 0 // índice relativo a B4 (0 = fila 4, 1 = fila 5, etc)

	if respLastRow.Values != nil && len(respLastRow.Values) > 0 {
		// Iterar sobre las filas para encontrar la primera vacía
		for i, row := range respLastRow.Values {
			if len(row) == 0 || row[0] == nil || row[0] == "" {
				// Encontramos una fila vacía
				proximoID = i + 1
				filaVaciaIndex = i
				break
			}
		}
		// Si no encontramos vacía, el próximo ID es después de la última
		if filaVaciaIndex == 0 && len(respLastRow.Values) > 0 {
			proximoID = len(respLastRow.Values) + 1
			filaVaciaIndex = len(respLastRow.Values)
		}
	}

	// La fila real en Google Sheets es 4 + filaVaciaIndex
	filaReal := 4 + filaVaciaIndex

	// Preparar contadores para cada tipo de combo
	// Columnas: B=ID, C=Vendedor, D=Cliente, E=Muzza-C1, F=Muzza-C2, G=Muzza-C3, H=(ignorar), I=Jamón-C1, J=Jamón-C2, K=Jamón-C3, L=(ignorar), M=MetodoPago, N=Estado, O=Entrega
	contadores := map[string]int{
		"muzza-c1": 0,
		"muzza-c2": 0,
		"muzza-c3": 0,
		"jamon-c1": 0,
		"jamon-c2": 0,
		"jamon-c3": 0,
	}

	// Contar combos
	for _, combo := range venta.Combos {
		key := fmt.Sprintf("%s-c%d", combo.Tipo, combo.Combo+1)
		contadores[key] += combo.Cantidad
	}

	// 1. Leer la fila vacía actual para copiar su formato
	filaReal = 4 + filaVaciaIndex
	//filaSiguiente := filaReal + 1
	//rangoLectura := fmt.Sprintf("Ventas!A%d:O%d", filaReal, filaReal)

	//respFilaVacia, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, rangoLectura).
	//	Context(ctx).
	//	Do()
	if err != nil {
		log.Printf("Error leyendo fila vacía: %v", err)
		// Continuamos de todas formas
	}

	// 2. Insertar una nueva fila en la posición filaReal (desplaza hacia abajo)
	sheetID := getSheetID(ctx, "Ventas")
	if sheetID == 0 {
		log.Printf("Error: No se pudo obtener el ID de la hoja Ventas")
		http.Error(w, "Error obteniendo ID de hoja", http.StatusInternalServerError)
		return
	}

	insertRequest := &sheets.InsertDimensionRequest{
		Range: &sheets.DimensionRange{
			SheetId:    sheetID,
			Dimension:  "ROWS",
			StartIndex: int64(filaReal - 1), // startIndex es 0-basado, así que fila 4 = índice 3
			EndIndex:   int64(filaReal),     // endIndex es exclusivo
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				InsertDimension: insertRequest,
			},
		},
	}

	_, err = srvSheets.Spreadsheets.BatchUpdate(SpreadsheetID, batchUpdateRequest).
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error insertando fila: %v", err)
		http.Error(w, "Error insertando fila", http.StatusInternalServerError)
		return
	}

	// 3. La fila insertada heredará automáticamente el formato de la fila anterior
	// No necesitamos copiar nada - Google Sheets lo hace automáticamente

	// 3.5 Arrastrar fórmulas de columnas H, L, P, Q, R, S de la fila anterior a la nueva fila
	// Estas columnas contienen funciones automatizadas que deben copiarse y actualizarse
	filasParaCopiarFormulas := []string{"H", "L", "P", "Q", "R", "S"}
	for _, col := range filasParaCopiarFormulas {
		// Leer fórmula de la fila anterior (filaReal - 1)
		rangoFuente := fmt.Sprintf("Ventas!%s%d", col, filaReal-1)
		respFormula, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, rangoFuente).
			ValueRenderOption("FORMULA"). // Leer la FÓRMULA, no el valor calculado
			Context(ctx).
			Do()
		if err != nil {
			log.Printf("Error leyendo fórmula de %s%d: %v", col, filaReal-1, err)
			continue
		}

		// Si hay contenido (fórmula), copiarla a la nueva fila
		if respFormula.Values != nil && len(respFormula.Values) > 0 && len(respFormula.Values[0]) > 0 {
			formulaValue := fmt.Sprintf("%v", respFormula.Values[0][0])

			// Actualizar la fórmula: reemplazar números de fila SOLO en referencias relativas
			// No tocar referencias absolutas ($E$9, $F$9, etc)
			formulaActualizada := actualizarFormulaRelativa(formulaValue, filaReal-1, filaReal)

			rangoDestino := fmt.Sprintf("Ventas!%s%d", col, filaReal)
			rbFormula := &sheets.ValueRange{
				Values: [][]interface{}{{formulaActualizada}},
			}
			_, err = srvSheets.Spreadsheets.Values.Update(SpreadsheetID, rangoDestino, rbFormula).
				ValueInputOption("USER_ENTERED").
				Context(ctx).
				Do()
			if err != nil {
				log.Printf("Error copiando fórmula a %s%d: %v", col, filaReal, err)
			}
		}
	}

	// 4. Preparar datos para escribir en secciones (sin H y L)

	// Escribir B:G (ID, Vendedor, Cliente, Muzza combos)
	valuesB_G := []interface{}{
		proximoID,              // B
		venta.Vendedor,         // C
		venta.Cliente,          // D
		contadores["muzza-c1"], // E
		contadores["muzza-c2"], // F
		contadores["muzza-c3"], // G
	}
	rbB_G := &sheets.ValueRange{
		Values: [][]interface{}{valuesB_G},
	}
	_, err = srvSheets.Spreadsheets.Values.Update(SpreadsheetID, fmt.Sprintf("Ventas!B%d:G%d", filaReal, filaReal), rbB_G).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error escribiendo datos principales: %v", err)
		http.Error(w, "Error escribiendo en Sheets", http.StatusInternalServerError)
		return
	}

	// Escribir I:K (Jamón combos)
	valuesI_K := []interface{}{
		contadores["jamon-c1"], // I
		contadores["jamon-c2"], // J
		contadores["jamon-c3"], // K
	}
	rbI_K := &sheets.ValueRange{
		Values: [][]interface{}{valuesI_K},
	}
	_, err = srvSheets.Spreadsheets.Values.Update(SpreadsheetID, fmt.Sprintf("Ventas!I%d:K%d", filaReal, filaReal), rbI_K).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error escribiendo Jamón: %v", err)
		http.Error(w, "Error escribiendo en Sheets", http.StatusInternalServerError)
		return
	}

	// Escribir M:O (Pago, Estado, Entrega)
	valuesM_O := []interface{}{
		venta.PaymentMethod, // M
		venta.Estado,        // N
		venta.TipoEntrega,   // O
	}
	rbM_O := &sheets.ValueRange{
		Values: [][]interface{}{valuesM_O},
	}
	_, err = srvSheets.Spreadsheets.Values.Update(SpreadsheetID, fmt.Sprintf("Ventas!M%d:O%d", filaReal, filaReal), rbM_O).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error escribiendo metadata: %v", err)
		http.Error(w, "Error escribiendo en Sheets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Venta guardada"}`)
}

func handleData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// Leer vendedores desde la hoja "datos" columna C a partir de C9
	respVendedores, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "datos!C9:C").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo vendedores: %v", err)
		http.Error(w, "Error leyendo vendedores", http.StatusInternalServerError)
		return
	}

	// Procesar vendedores (detente en la primera celda vacía)
	vendedores := []string{}
	for _, row := range respVendedores.Values {
		if len(row) == 0 || fmt.Sprintf("%v", row[0]) == "" {
			break // Detener al encontrar una celda vacía
		}
		vendedor := fmt.Sprintf("%v", row[0])
		vendedores = append(vendedores, vendedor)
	}

	// Leer precios de Muzza desde H9:J9
	respMuzza, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "datos!H9:J9").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo precios Muzza: %v", err)
		http.Error(w, "Error leyendo precios Muzza", http.StatusInternalServerError)
		return
	}

	// Leer precios de Jamón desde H11:J11
	respJamon, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "datos!H11:J11").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo precios Jamón: %v", err)
		http.Error(w, "Error leyendo precios Jamón", http.StatusInternalServerError)
		return
	}

	// Procesar precios Muzza
	preciosMuzza := []float64{}
	if len(respMuzza.Values) > 0 && len(respMuzza.Values[0]) >= 3 {
		for i := 0; i < 3; i++ {
			precio := parseFloat(respMuzza.Values[0][i])
			preciosMuzza = append(preciosMuzza, precio)
		}
	}

	// Procesar precios Jamón
	preciosJamon := []float64{}
	if len(respJamon.Values) > 0 && len(respJamon.Values[0]) >= 3 {
		for i := 0; i < 3; i++ {
			precio := parseFloat(respJamon.Values[0][i])
			preciosJamon = append(preciosJamon, precio)
		}
	}

	// Leer datos históricos de ventas para construir clientes por vendedor
	// Leer desde D4 en adelante (columna C=Vendedor, columna D=Cliente)
	respVentas, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "Ventas!C4:D").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo ventas: %v", err)
		http.Error(w, "Error leyendo datos de ventas", http.StatusInternalServerError)
		return
	}

	clientesPorVendedor := make(map[string][]string)

	// Procesar filas de ventas (comenzar desde fila 4 - sin encabezado)
	for _, row := range respVentas.Values {
		// Validar que la fila tenga al menos 2 columnas
		if len(row) < 2 {
			continue
		}

		vendedor := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		cliente := strings.TrimSpace(fmt.Sprintf("%v", row[1]))

		// Evitar vacíos y valores "nil" o "<nil>"
		if vendedor == "" || cliente == "" || vendedor == "<nil>" || cliente == "<nil>" {
			continue
		}

		// Evitar duplicados
		if !contains(clientesPorVendedor[vendedor], cliente) {
			clientesPorVendedor[vendedor] = append(clientesPorVendedor[vendedor], cliente)
		}
	}

	// Construir mapa de pizzas
	pizzas := make(map[string]Pizza)
	pizzas["muzza"] = Pizza{
		Nombre:  "Muzza",
		Combos:  []string{"Combo 1", "Combo 2", "Combo 3"},
		Precios: preciosMuzza,
	}
	pizzas["jamon"] = Pizza{
		Nombre:  "Jamón",
		Combos:  []string{"Combo 1", "Combo 2", "Combo 3"},
		Precios: preciosJamon,
	}

	// Responder con JSON
	response := DataResponse{
		ClientesPorVendedor: clientesPorVendedor,
		Vendedores:          vendedores,
		Pizzas:              pizzas,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Funciones auxiliares
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func parseFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case string:
		// Limpiar espacios
		s := strings.TrimSpace(v)

		// Remover $ si existe
		s = strings.ReplaceAll(s, "$", "")

		// Convertir formato argentino (1.000,50) a formato Go (1000.50)
		// Reemplazar . con vacío (son separadores de miles)
		s = strings.ReplaceAll(s, ".", "")
		// Reemplazar , con . (es el separador decimal)
		s = strings.ReplaceAll(s, ",", ".")

		// Usar strconv para convertir string a float64
		result, err := strconv.ParseFloat(s, 64)
		if err != nil {
			log.Printf("Error parseando float: %s -> %v", v, err)
			return 0
		}
		return result
	default:
		return 0
	}
}

// Obtener el ID de la hoja por su nombre
func getSheetID(ctx context.Context, sheetName string) int64 {
	spreadsheet, err := srvSheets.Spreadsheets.Get(SpreadsheetID).
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error obteniendo información de la hoja: %v", err)
		return 0
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet.Properties.SheetId
		}
	}

	log.Printf("No se encontró la hoja: %s", sheetName)
	return 0
}

// actualizarFormulaRelativa actualiza solo referencias relativas (sin $) en una fórmula
// Ejemplo: =SUM(E4:G4) → =SUM(E5:G5) pero =SUM($E$9:$G$9) se mantiene igual
func actualizarFormulaRelativa(formula string, filaAnterior, filaReal int) string {
	filaAnteriorStr := strconv.Itoa(filaAnterior)
	filaRealStr := strconv.Itoa(filaReal)

	var resultado strings.Builder
	i := 0

	for i < len(formula) {
		// Buscar si hay un $ antes del número
		if i > 0 && formula[i-1] == '$' {
			// Es una referencia absoluta, mantener igual
			resultado.WriteByte(formula[i])
			i++
		} else if i < len(formula) && isDigit(formula[i]) {
			// Encontramos un número, verificar si es el número de fila anterior
			numStr := ""
			j := i
			for j < len(formula) && isDigit(formula[j]) {
				numStr += string(formula[j])
				j++
			}

			// Reemplazar si es el número de fila anterior
			if numStr == filaAnteriorStr {
				resultado.WriteString(filaRealStr)
			} else {
				resultado.WriteString(numStr)
			}
			i = j
		} else {
			resultado.WriteByte(formula[i])
			i++
		}
	}

	return resultado.String()
}

// isDigit verifica si un byte es un dígito
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// handleEstadisticas devuelve todas las estadísticas de ventas
func handleEstadisticas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// Leer precios de Muzza desde H9:J9
	respMuzza, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "datos!H9:J9").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo precios Muzza: %v", err)
		http.Error(w, "Error leyendo precios Muzza", http.StatusInternalServerError)
		return
	}

	// Leer precios de Jamón desde H11:J11
	respJamon, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "datos!H11:J11").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo precios Jamón: %v", err)
		http.Error(w, "Error leyendo precios Jamón", http.StatusInternalServerError)
		return
	}

	// Procesar precios Muzza
	preciosMuzza := []float64{}
	if len(respMuzza.Values) > 0 && len(respMuzza.Values[0]) >= 3 {
		for i := 0; i < 3; i++ {
			precio := parseFloat(respMuzza.Values[0][i])
			preciosMuzza = append(preciosMuzza, precio)
		}
	}

	// Procesar precios Jamón
	preciosJamon := []float64{}
	if len(respJamon.Values) > 0 && len(respJamon.Values[0]) >= 3 {
		for i := 0; i < 3; i++ {
			precio := parseFloat(respJamon.Values[0][i])
			preciosJamon = append(preciosJamon, precio)
		}
	}

	// Leer todas las ventas desde B4 hasta P (última fila)
	respVentas, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "Ventas!B4:P").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo ventas: %v", err)
		http.Error(w, "Error leyendo ventas", http.StatusInternalServerError)
		return
	}

	// Procesar ventas
	var ventas []map[string]interface{}

	if respVentas.Values != nil {
		for _, row := range respVentas.Values {
			if len(row) < 4 {
				continue // Saltar filas incompletas
			}

			id := fmt.Sprintf("%v", row[0])
			if id == "" {
				continue // Saltar si no hay ID
			}

			venta := map[string]interface{}{
				"id":             parseFloat(row[0]),
				"vendedor":       fmt.Sprintf("%v", row[1]),
				"cliente":        fmt.Sprintf("%v", row[2]),
				"combos":         []map[string]interface{}{}, // Se llenarán después
				"total":          0.0,
				"estado":         "",
				"payment_method": "",
				"tipo_entrega":   "",
			}

			// Combos Muzza (E, F, G = columnas 3, 4, 5)
			if len(row) > 5 {
				for i := 0; i < 3; i++ {
					cantidad := int(parseFloat(row[3+i]))
					if cantidad > 0 {
						combo := map[string]interface{}{
							"tipo":     "muzza",
							"combo":    i,
							"cantidad": cantidad,
						}
						venta["combos"] = append(venta["combos"].([]map[string]interface{}), combo)
					}
				}
			}

			// Total (P = columna 14)
			if len(row) > 14 {
				venta["total"] = parseFloat(row[14])
			}

			// Combos Jamón (I, J, K = columnas 7, 8, 9)
			if len(row) > 9 {
				for i := 0; i < 3; i++ {
					cantidad := int(parseFloat(row[7+i]))
					if cantidad > 0 {
						combo := map[string]interface{}{
							"tipo":     "jamon",
							"combo":    i,
							"cantidad": cantidad,
						}
						venta["combos"] = append(venta["combos"].([]map[string]interface{}), combo)
					}
				}
			}

			// Payment Method (M = columna 11)
			if len(row) > 11 {
				venta["payment_method"] = fmt.Sprintf("%v", row[11])
			}

			// Estado (N = columna 12)
			if len(row) > 12 {
				venta["estado"] = fmt.Sprintf("%v", row[12])
			}

			// Tipo Entrega (O = columna 13)
			if len(row) > 13 {
				venta["tipo_entrega"] = fmt.Sprintf("%v", row[13])
			}

			ventas = append(ventas, venta)
		}
	}

	// Respuesta
	response := map[string]interface{}{
		"ventas":       ventas,
		"preciosMuzza": preciosMuzza,
		"preciosJamon": preciosJamon,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleEstadisticasSheet trae los datos directamente del sheet "estadisticas"
func handleEstadisticasSheet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// Leer estadísticas generales (C5:C6, C10:C11, G5:G9, G12:G15)
	ranges := []string{
		"estadisticas!C5:C6",   // Total Muzzas, Total Jamones
		"estadisticas!C10:C11", // Delivery, Retiro
		"estadisticas!G5:G9",   // Efectivo, Transferencia, Total, Total con Sin Cobrar
		"estadisticas!G12:G15", // Sin Pagar, Pagadas, Entregadas, Totales
		"estadisticas!B24:I",   // Vendedores y sus datos (B24 hasta fin)
	}

	resps, err := srvSheets.Spreadsheets.Values.BatchGet(SpreadsheetID).
		Ranges(ranges...).
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error leyendo estadísticas: %v", err)
		http.Error(w, "Error leyendo estadísticas", http.StatusInternalServerError)
		return
	}

	// Parsear los valores
	var totalMuzzas, totalJamones, totalDelivery, totalRetiro float64
	var pendienteCobro, efectivoCobrado, transferenciaCobrada, totalCobrado, totalConSinCobrar float64
	var ventasSinPagar, ventasPagadas, ventasEntregadas, ventasTotales float64

	// C5:C6 - Total Muzzas, Total Jamones
	if len(resps.ValueRanges) > 0 && len(resps.ValueRanges[0].Values) > 0 {
		totalMuzzas = parseFloat(resps.ValueRanges[0].Values[0][0])
		if len(resps.ValueRanges[0].Values) > 1 {
			totalJamones = parseFloat(resps.ValueRanges[0].Values[1][0])
		}
	}

	// C10:C11 - Delivery, Retiro
	if len(resps.ValueRanges) > 1 && len(resps.ValueRanges[1].Values) > 0 {
		totalDelivery = parseFloat(resps.ValueRanges[1].Values[0][0])
		if len(resps.ValueRanges[1].Values) > 1 {
			totalRetiro = parseFloat(resps.ValueRanges[1].Values[1][0])
		}
	}

	// G5:G9 - Pendiente, Efectivo, Transferencia, Total, Total con Sin Cobrar
	if len(resps.ValueRanges) > 2 && len(resps.ValueRanges[2].Values) > 0 {
		pendienteCobro = parseFloat(resps.ValueRanges[2].Values[0][0])
		if len(resps.ValueRanges[2].Values) > 1 {
			efectivoCobrado = parseFloat(resps.ValueRanges[2].Values[1][0])
		}
		if len(resps.ValueRanges[2].Values) > 2 {
			transferenciaCobrada = parseFloat(resps.ValueRanges[2].Values[2][0])
		}
		if len(resps.ValueRanges[2].Values) > 3 {
			totalCobrado = parseFloat(resps.ValueRanges[2].Values[3][0])
		}
		if len(resps.ValueRanges[2].Values) > 4 {
			totalConSinCobrar = parseFloat(resps.ValueRanges[2].Values[4][0])
		}
	}

	// G12:G15 - Sin Pagar, Pagadas, Entregadas, Totales
	if len(resps.ValueRanges) > 3 && len(resps.ValueRanges[3].Values) > 0 {
		ventasSinPagar = parseFloat(resps.ValueRanges[3].Values[0][0])
		if len(resps.ValueRanges[3].Values) > 1 {
			ventasPagadas = parseFloat(resps.ValueRanges[3].Values[1][0])
		}
		if len(resps.ValueRanges[3].Values) > 2 {
			ventasEntregadas = parseFloat(resps.ValueRanges[3].Values[2][0])
		}
		if len(resps.ValueRanges[3].Values) > 3 {
			ventasTotales = parseFloat(resps.ValueRanges[3].Values[3][0])
		}
	}

	// B24:I - Vendedores y sus datos
	var vendedores []map[string]interface{}
	if len(resps.ValueRanges) > 4 && resps.ValueRanges[4].Values != nil {
		for _, row := range resps.ValueRanges[4].Values {
			if len(row) < 2 {
				continue // Saltar filas incompletas
			}

			nombre := fmt.Sprintf("%v", row[0])
			if nombre == "" || nombre == "<nil>" {
				break // Detener al encontrar primera fila vacía
			}

			vendedor := map[string]interface{}{
				"nombre":          nombre,
				"cantidad_ventas": 0.0,
				"muzzas":          0.0,
				"jamones":         0.0,
				"sin_pagar":       0.0,
				"pagado":          0.0,
				"total":           0.0,
			}

			// B24 - Nombre (ya está)
			// C24 - Cantidad de ventas
			if len(row) > 1 {
				vendedor["cantidad_ventas"] = parseFloat(row[1])
			}
			// D24 - Muzzas
			if len(row) > 2 {
				vendedor["muzzas"] = parseFloat(row[2])
			}
			// E24 - Jamones
			if len(row) > 3 {
				vendedor["jamones"] = parseFloat(row[3])
			}
			// G24 - Sin Pagar
			if len(row) > 5 {
				vendedor["sin_pagar"] = parseFloat(row[5])
			}
			// H24 - Pagado
			if len(row) > 6 {
				vendedor["pagado"] = parseFloat(row[6])
			}
			// I24 - Total
			if len(row) > 7 {
				vendedor["total"] = parseFloat(row[7])
			}

			vendedores = append(vendedores, vendedor)
		}
	}

	// Armar respuesta
	response := map[string]interface{}{
		"resumen": map[string]interface{}{
			"total_muzzas":          totalMuzzas,
			"total_jamones":         totalJamones,
			"total_delivery":        totalDelivery,
			"total_retiro":          totalRetiro,
			"pendiente_cobro":       pendienteCobro,
			"efectivo_cobrado":      efectivoCobrado,
			"transferencia_cobrada": transferenciaCobrada,
			"total_cobrado":         totalCobrado,
			"total_con_sin_cobrar":  totalConSinCobrar,
			"ventas_sin_pagar":      ventasSinPagar,
			"ventas_pagadas":        ventasPagadas,
			"ventas_entregadas":     ventasEntregadas,
			"ventas_totales":        ventasTotales,
		},
		"vendedores": vendedores,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleActualizarVenta actualiza el estado, pago y combos de una venta
func handleActualizarVenta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	type Combo struct {
		Tipo     string  `json:"tipo"`
		Combo    int     `json:"combo"`
		Cantidad float64 `json:"cantidad"`
	}

	var req struct {
		ID            float64 `json:"id"`
		Estado        string  `json:"estado"`
		PaymentMethod string  `json:"payment_method"`
		Combos        []Combo `json:"combos"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Encontrar la fila de la venta por ID
	respVentas, err := srvSheets.Spreadsheets.Values.Get(SpreadsheetID, "Ventas!B4:B").
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error buscando venta: %v", err)
		http.Error(w, "Error buscando venta", http.StatusInternalServerError)
		return
	}

	var filaVenta int
	if respVentas.Values != nil {
		for i, row := range respVentas.Values {
			if len(row) > 0 && int(parseFloat(row[0])) == int(req.ID) {
				filaVenta = 4 + i // Fila real (4 + índice)
				break
			}
		}
	}

	if filaVenta == 0 {
		http.Error(w, "Venta no encontrada", http.StatusNotFound)
		return
	}

	// Preparar batch update para combos (E:K) y estado/pago (M:N)
	data := []*sheets.ValueRange{}

	// Actualizar combos si existen
	if len(req.Combos) > 0 {
		// Inicializar array con valores vacíos para E:K (7 columnas)
		comboValues := []interface{}{0, 0, 0, 0, 0, 0, 0}

		// Mapear combos al array: E=0, F=1, G=2, H(skip), I=4, J=5, K=6
		for _, combo := range req.Combos {
			if combo.Tipo == "muzza" {
				// Muzzas: E(0), F(1), G(2)
				if combo.Combo >= 0 && combo.Combo < 3 {
					comboValues[combo.Combo] = int(combo.Cantidad)
				}
			} else if combo.Tipo == "jamon" {
				// Jamones: I(4), J(5), K(6)
				if combo.Combo >= 0 && combo.Combo < 3 {
					comboValues[combo.Combo+4] = int(combo.Cantidad)
				}
			}
		}

		data = append(data, &sheets.ValueRange{
			Range:  fmt.Sprintf("Ventas!E%d:K%d", filaVenta, filaVenta),
			Values: [][]interface{}{comboValues},
		})
	}

	// Actualizar M (Payment Method) y N (Estado)
	data = append(data, &sheets.ValueRange{
		Range:  fmt.Sprintf("Ventas!M%d:N%d", filaVenta, filaVenta),
		Values: [][]interface{}{{req.PaymentMethod, req.Estado}},
	})

	// BatchUpdate
	batchReq := &sheets.BatchUpdateValuesRequest{
		Data:             data,
		ValueInputOption: "USER_ENTERED",
	}
	_, err = srvSheets.Spreadsheets.Values.BatchUpdate(SpreadsheetID, batchReq).
		Context(ctx).
		Do()
	if err != nil {
		log.Printf("Error actualizando venta: %v", err)
		http.Error(w, "Error actualizando venta", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Venta actualizada"}`)
}
