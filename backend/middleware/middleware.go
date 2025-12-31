package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"pizzas-ecos/logger"
	"pizzas-ecos/models"
)

var JWTSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(JWTSecret) == 0 {
		JWTSecret = []byte("uytrewghbvcxbvnbvnz")
	}
}

/* =========================
   AUTH MIDDLEWARE
========================= */

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 🔴 CLAVE: ignorar preflight OPTIONS
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path
		method := r.Method

		// 🔐 RUTAS COMPLETAMENTE PROTEGIDAS
		protectedPaths := map[string]bool{
			"/api/v1/admin": true, // Admin dashboard
		}

		if protectedPaths[path] {
			goto requireAuth
		}

		// 🔐 OPERACIONES PROTEGIDAS (POST/PUT/DELETE en productos y vendedores)
		// POST crear productos (solo admin)
		if method == http.MethodPost && (path == "/api/v1/productos" || path == "/api/v1/crear-producto") {
			goto requireAuth
		}

		// PUT/DELETE editar/eliminar productos (solo admin)
		if (method == http.MethodPut || method == http.MethodDelete) && path == "/api/v1/productos" {
			goto requireAuth
		}
		if method == http.MethodPut && (path == "/api/v1/actualizar-producto" || path == "/api/v1/productos/:id") {
			goto requireAuth
		}
		if method == http.MethodDelete && (path == "/api/v1/eliminar-producto" || path == "/api/v1/productos/:id") {
			goto requireAuth
		}

		// POST crear vendedores (solo admin)
		if method == http.MethodPost && (path == "/api/v1/vendedores" || path == "/api/v1/crear-vendedor") {
			goto requireAuth
		}

		// PUT/DELETE editar/eliminar vendedores (solo admin)
		if (method == http.MethodPut || method == http.MethodDelete) && path == "/api/v1/vendedores" {
			goto requireAuth
		}
		if method == http.MethodPut && (path == "/api/v1/actualizar-vendedor" || path == "/api/v1/vendedores/:id") {
			goto requireAuth
		}
		if method == http.MethodDelete && (path == "/api/v1/eliminar-vendedor" || path == "/api/v1/vendedores/:id") {
			goto requireAuth
		}

		// POST crear usuarios (solo admin)
		if method == http.MethodPost && (path == "/api/v1/usuarios" || path == "/api/v1/crear-usuario") {
			goto requireAuth
		}

		// PUT/DELETE usuarios (solo admin)
		if (method == http.MethodPut || method == http.MethodDelete) && (path == "/api/v1/usuarios" || path == "/api/v1/actualizar-usuario" || path == "/api/v1/eliminar-usuario") {
			goto requireAuth
		}

		// ✅ TODO LO DEMÁS ES PÚBLICO
		// - Todos pueden autenticarse (/login, /auth/login)
		// - Todos pueden ver datos iniciales (/data)
		// - Todos pueden ver productos (GET)
		// - Todos pueden ver vendedores (GET)
		// - Todos pueden crear/editar ventas
		// - Todos pueden ver estadísticas
		next.ServeHTTP(w, r)
		return

	requireAuth:
		// Validar token JWT
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			unauthorized(w, "Token requerido para esta acción")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			unauthorized(w, "Formato de token inválido")
			return
		}

		token, err := jwt.ParseWithClaims(
			parts[1],
			&models.TokenClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return JWTSecret, nil
			},
		)

		if err != nil || !token.Valid {
			unauthorized(w, "Token inválido o expirado")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"status":401,"message":"` + msg + `","code":"UNAUTHORIZED"}`))
}

/* =========================
   LOGGING
========================= */

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		logger.LogHTTPRequest(
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			wrapped.statusCode,
			time.Since(start),
			r.Header.Get("User-Agent"),
			"",
		)
	})
}

/* =========================
   CORS
========================= */

func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Permitir cualquier origen
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Responder a preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	// Si el wildcard "*" está en la lista, permitir cualquier origen
	for _, o := range allowed {
		if o == "*" {
			return true
		}
		if o == origin {
			return true
		}
	}
	return false
}

/* =========================
   RECOVERY
========================= */

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("PANIC", "panic", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status":500,"message":"Error interno","code":"INTERNAL_SERVER_ERROR"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

/* =========================
   RESPONSE WRAPPER
========================= */

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
