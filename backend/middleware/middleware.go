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
		// Opinión: En ingeniería, siempre avisar si se usa un fallback débil
		JWTSecret = []byte("ecos-auth-secret-key-change-in-production")
	}
}

// AuthMiddleware verifica que el request tenga un token JWT válido
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status":401,"message":"Token requerido","code":"UNAUTHORIZED"}`))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status":401,"message":"Formato de token inválido","code":"UNAUTHORIZED"}`))
			return
		}

		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"status":401,"message":"Token inválido o expirado","code":"UNAUTHORIZED"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware registra información de requests con structured logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, wrapped.statusCode, duration, r.Header.Get("User-Agent"), "")
	})
}

// CORSMiddleware configura CORS de manera profesional
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Validación de origen
			isAllowed := false
			if len(allowedOrigins) > 0 && allowedOrigins[0] == "*" {
				isAllowed = true
			} else {
				for _, allowed := range allowedOrigins {
					if allowed == origin {
						isAllowed = true
						break
					}
				}
			}

			// Inyectar cabeceras si el origen es permitido
			if isAllowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			// [CRÍTICO] Manejo de Preflight
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
				logger.Error("PANIC detectado", "error", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status":500,"message":"Error interno del servidor","code":"INTERNAL_SERVER_ERROR"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}