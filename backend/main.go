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

// inicDB inicializa la conexión a MySQL
func inicDB() error {
	return config.InitDB()
}

func main() {
	// Inicializar BD
	if err := inicDB(); err != nil {
		log.Fatalf("❌ Error inicializando BD: %v", err)
	}

	// Crear mux
	mux := http.NewServeMux()

	// Registrar rutas API v1 + legacy (todo en un solo router)
	apiRouter := routes.SetupRoutes()
	// Las rutas legacy se agregan al mismo router, no crear dos routers
	apiRouter.Register(mux)

	// Imprimir rutas (útil para debugging)
	logger.Info("Rutas registradas", map[string]interface{}{
		"count": len(apiRouter.GetRoutes()),
	})
	for _, route := range apiRouter.GetRoutes() {
		logger.Debug("Route", map[string]interface{}{
			"method": route.Method,
			"path":   route.Path,
			"name":   route.Name,
		})
	}

	// Crear rate limiter (50 requests por segundo por IP)
	limiter := ratelimit.NewRateLimiter(50)

	// Crear DDoS detector (más de 500 requests en 10 segundos por IP)
	ddosDetector := security.NewDDoSDetector(500, 10*time.Second)

	// Aplicar middlewares globales en orden
	// 1. DDoS Protection
	// 2. CORS
	// 3. Rate Limiting
	// 4. Logging
	// 5. Recovery
	ddosMiddleware := security.Middleware(ddosDetector)

	// Configurar orígenes CORS desde variable de entorno o usar valores por defecto
	corsOrigins := []string{"http://localhost:5000", "https://ecos-ventas-pizzas.netlify.app"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		corsOrigins = strings.Split(envOrigins, ",")
	}
	corsMiddleware := middleware.CORSMiddleware(corsOrigins)
	rateLimitMiddleware := ratelimit.Middleware(limiter)
	loggingMiddleware := middleware.LoggingMiddleware
	recoveryMiddleware := middleware.RecoveryMiddleware

	handler := ddosMiddleware(corsMiddleware(rateLimitMiddleware(loggingMiddleware(recoveryMiddleware(mux)))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Servidor iniciado", map[string]interface{}{
		"port": port,
	})
	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		logger.Error("Error iniciando servidor", "SERVER_ERROR", map[string]interface{}{"error": err.Error()})
	}
}
