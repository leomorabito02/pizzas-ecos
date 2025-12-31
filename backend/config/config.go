package config

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"pizzas-ecos/database"
)

// InitDB inicializa la conexión a MySQL
func InitDB() error {
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

	db, err := sql.Open("mysql", dsn)
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

	// Guardar en el package database
	database.DB = db

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
