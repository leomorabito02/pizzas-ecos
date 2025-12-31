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
	if string(JWTSecret) == "" {
		JWTSecret = []byte("ecos-auth-secret-key-change-in-production")
	}
}

// AuthMiddleware verifica que el request tenga un token JWT válido
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener token del header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Auth: Token no proporcionado", map[string]interface{}{
				"path": r.URL.Path,
				"ip":   r.RemoteAddr,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status":401,"message":"Token requerido","code":"UNAUTHORIZED"}`))
			return
		}

		// Extraer token del formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Auth: Formato de token inválido", map[string]interface{}{
				"path": r.URL.Path,
				"ip":   r.RemoteAddr,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status":401,"message":"Formato de token inválido","code":"UNAUTHORIZED"}`))
			return
		}

		tokenString := parts[1]

		// Validar token
		token, err := jwt.ParseWithClaims(tokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			logger.Warn("Auth: Token inválido o expirado", map[string]interface{}{
				"path":  r.URL.Path,
				"ip":    r.RemoteAddr,
				"error": err.Error(),
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status":401,"message":"Token inválido o expirado","code":"UNAUTHORIZED"}`))
			return
		}

		// Token válido, continuar
		logger.Debug("Auth: Token válido", map[string]interface{}{
			"path": r.URL.Path,
			"ip":   r.RemoteAddr,
		})
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware registra información de requests con structured logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Crear un response writer que capture el status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Procesar request
		next.ServeHTTP(wrapped, r)

		// Registrar con structured logging
		duration := time.Since(start)
		logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, wrapped.statusCode, duration, r.Header.Get("User-Agent"), "")
	})
}

// responseWriter es un wrapper para capturar status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware configura CORS
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Verificar si origen está permitido
			isAllowed := false
			if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
				isAllowed = true
			} else {
				for _, allowed := range allowedOrigins {
					if allowed == origin {
						isAllowed = true
						break
					}
				}
			}

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware recupera de panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("PANIC durante request", "PANIC", map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
					"error":  err,
					"ip":     r.RemoteAddr,
				})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status":500,"message":"Internal server error","code":"INTERNAL_SERVER_ERROR"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
