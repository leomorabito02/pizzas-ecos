package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"pizzas-ecos/config"
	"pizzas-ecos/logger"
	"pizzas-ecos/middleware"
	"pizzas-ecos/ratelimit"
	"pizzas-ecos/routes"
	"pizzas-ecos/security"
)

// inicDB inicializa la conexión a la base de datos
// CORS middleware is properly configured
func inicDB() error {
	return config.InitDB()
}

func main() {
	// 1. Inicializar Base de Datos
	if err := inicDB(); err != nil {
		log.Fatalf("❌ Error inicializando BD: %v", err)
	}

	// 2. Configuración del Router (Mux)
	mux := http.NewServeMux()
	apiRouter := routes.SetupRoutes()
	apiRouter.Register(mux)

	// Logging de rutas para auditoría técnica
	logger.Info("Rutas registradas", map[string]interface{}{
		"count": len(apiRouter.GetRoutes()),
	})

	// 3. Configuración de Seguridad y Control de Tráfico
	limiter := ratelimit.NewRateLimiter(50)                    // 50 req/s
	ddosDetector := security.NewDDoSDetector(500, 10*time.Second) // 500 req en 10s

	// 4. Configuración de Orígenes CORS
	// Priorizamos variable de entorno para flexibilidad en el deploy
	corsOrigins := []string{"http://localhost:5000", "https://ecos-ventas-pizzas.netlify.app"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		corsOrigins = strings.Split(envOrigins, ",")
	}

	// 5. Cadena de Middlewares (Arquitectura de Cebolla)
	// Mi opinión: El orden aquí es vital para que el CORS no sea bloqueado por seguridad previa.
	// La ejecución es de AFUERA hacia ADENTRO.
	
	// Paso 1: Mux básico con Logging (lo más interno)
	handler := middleware.LoggingMiddleware(mux)

	// Paso 2: Aplicar Rate Limit
	handler = ratelimit.Middleware(limiter)(handler)

	// Paso 3: Aplicar Protección DDoS
	handler = security.Middleware(ddosDetector)(handler)

	// Paso 4: Aplicar CORS (Debe estar afuera para responder OPTIONS rápidamente)
	handler = middleware.CORSMiddleware(corsOrigins)(handler)

	// Paso 5: Aplicar Recovery (El escudo más externo que atrapa errores de todos los anteriores)
	handler = middleware.RecoveryMiddleware(handler)

	// 6. Lanzamiento del Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Servidor iniciado exitosamente", map[string]interface{}{
		"port": port,
		"env":  os.Getenv("GO_ENV"),
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Error crítico en el servidor", "SERVER_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
	}
}