//go:build ignore

// migrate_encrypt.go — script de migración one-shot para encriptar datos existentes.
// Ejecutar UNA SOLA VEZ después de aplicar el ALTER TABLE en Neon.
//
// Uso:
//
//	go run migrate_encrypt.go
//
// Requiere las variables de entorno DATABASE_URL, ENCRYPTION_KEY y HMAC_KEY
// (se cargan desde .env automáticamente).
package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"pizzas-ecos/security"
)

func main() {
	godotenv.Load()

	if err := security.InitCrypto(); err != nil {
		log.Fatalf("❌ InitCrypto: %v", err)
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("❌ Abriendo DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Ping DB: %v", err)
	}

	log.Println("✅ Conectado a la base de datos")

	migrateVendedores(db)
	migrateClientes(db)
	migrateUsuarios(db)

	log.Println("✅ Migración completada exitosamente")
}

// migrateVendedores encripta nombres de vendedores en texto plano.
// Detecta si ya están encriptados (Decrypt exitoso) y los omite.
func migrateVendedores(db *sql.DB) {
	log.Println("→ Migrando vendedores...")
	rows, err := db.Query("SELECT id, nombre FROM vendedores")
	if err != nil {
		log.Fatalf("❌ Query vendedores: %v", err)
	}
	defer rows.Close()

	type row struct {
		id     int
		nombre string
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.nombre); err != nil {
			log.Fatalf("❌ Scan vendedor: %v", err)
		}
		// Si ya está encriptado, omitir
		if _, err := security.Decrypt(r.nombre); err == nil {
			log.Printf("  [skip] vendedor %d ya encriptado", r.id)
			continue
		}
		pending = append(pending, r)
	}
	rows.Close()

	for _, r := range pending {
		enc, err := security.Encrypt(r.nombre)
		if err != nil {
			log.Fatalf("❌ Encrypt vendedor %d: %v", r.id, err)
		}
		hash := security.HMACField(r.nombre)
		if _, err := db.Exec("UPDATE vendedores SET nombre = $1, nombre_hash = $2 WHERE id = $3", enc, hash, r.id); err != nil {
			log.Fatalf("❌ Update vendedor %d: %v", r.id, err)
		}
		log.Printf("  [ok] vendedor %d (%s)", r.id, r.nombre)
	}
	log.Printf("  Vendedores migrados: %d", len(pending))
}

// migrateClientes encripta nombres y teléfonos de clientes.
func migrateClientes(db *sql.DB) {
	log.Println("→ Migrando clientes...")
	rows, err := db.Query("SELECT id, nombre, COALESCE(telefono, '') FROM clientes")
	if err != nil {
		log.Fatalf("❌ Query clientes: %v", err)
	}
	defer rows.Close()

	type row struct {
		id       int
		nombre   string
		telefono string
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.nombre, &r.telefono); err != nil {
			log.Fatalf("❌ Scan cliente: %v", err)
		}
		if _, err := security.Decrypt(r.nombre); err == nil {
			log.Printf("  [skip] cliente %d ya encriptado", r.id)
			continue
		}
		pending = append(pending, r)
	}
	rows.Close()

	for _, r := range pending {
		encNombre, err := security.Encrypt(r.nombre)
		if err != nil {
			log.Fatalf("❌ Encrypt nombre cliente %d: %v", r.id, err)
		}
		hash := security.HMACField(r.nombre)

		trimTel := strings.TrimSpace(r.telefono)
		if trimTel != "" && trimTel != "0" {
			encTel, err := security.Encrypt(trimTel)
			if err != nil {
				log.Fatalf("❌ Encrypt telefono cliente %d: %v", r.id, err)
			}
			if _, err := db.Exec(
				"UPDATE clientes SET nombre = $1, nombre_hash = $2, telefono = $3 WHERE id = $4",
				encNombre, hash, encTel, r.id,
			); err != nil {
				log.Fatalf("❌ Update cliente %d: %v", r.id, err)
			}
		} else {
			if _, err := db.Exec(
				"UPDATE clientes SET nombre = $1, nombre_hash = $2 WHERE id = $3",
				encNombre, hash, r.id,
			); err != nil {
				log.Fatalf("❌ Update cliente %d: %v", r.id, err)
			}
		}
		log.Printf("  [ok] cliente %d (%s)", r.id, r.nombre)
	}
	log.Printf("  Clientes migrados: %d", len(pending))
}

// migrateUsuarios encripta usernames de usuarios.
func migrateUsuarios(db *sql.DB) {
	log.Println("→ Migrando usuarios...")
	rows, err := db.Query("SELECT id, username FROM usuarios")
	if err != nil {
		log.Fatalf("❌ Query usuarios: %v", err)
	}
	defer rows.Close()

	type row struct {
		id       int
		username string
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.username); err != nil {
			log.Fatalf("❌ Scan usuario: %v", err)
		}
		if _, err := security.Decrypt(r.username); err == nil {
			log.Printf("  [skip] usuario %d ya encriptado", r.id)
			continue
		}
		pending = append(pending, r)
	}
	rows.Close()

	for _, r := range pending {
		enc, err := security.Encrypt(r.username)
		if err != nil {
			log.Fatalf("❌ Encrypt username %d: %v", r.id, err)
		}
		hash := security.HMACField(r.username)
		if _, err := db.Exec(
			"UPDATE usuarios SET username = $1, username_hash = $2 WHERE id = $3",
			enc, hash, r.id,
		); err != nil {
			log.Fatalf("❌ Update usuario %d: %v", r.id, err)
		}
		log.Printf("  [ok] usuario %d (%s)", r.id, r.username)
	}
	log.Printf("  Usuarios migrados: %d", len(pending))
}
