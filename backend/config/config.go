package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"pizzas-ecos/database"
	"pizzas-ecos/security"
)

// InitDB inicializa la conexión a PostgreSQL
func InitDB() error {
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL no configurada")
	}

	log.Printf("📍 DATABASE_URL configurada: %s", strings.Split(dbURL, "@")[0]+"@...")

	db, err := sql.Open("postgres", dbURL)
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

	// Inicializar encriptación de campos sensibles
	if err := security.InitCrypto(); err != nil {
		log.Fatalf("❌ Error inicializando crypto: %v", err)
	}

	log.Println("✅ Conectado a PostgreSQL exitosamente")
	return nil
}
