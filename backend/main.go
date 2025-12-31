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

func initDB() error {
	return config.InitDB()
}

func main() {

	// 1. Base de datos
	if err := initDB(); err != nil {
		log.Fatalf("❌ Error inicializando BD: %v", err)
	}

	// 2. Router
	mux := http.NewServeMux()
	apiRouter := routes.SetupRoutes()
	apiRouter.Register(mux)

	logger.Info("Rutas registradas", map[string]interface{}{
		"count": len(apiRouter.GetRoutes()),
	})

	// 3. Seguridad y control
	limiter := ratelimit.NewRateLimiter(50)
	ddosDetector := security.NewDDoSDetector(500, 10*time.Second)

	// 4. CORS origins
	corsOrigins := []string{
		"http://localhost:5000",
		"https://ecos-ventas-pizzas.netlify.app",
	}

	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		corsOrigins = strings.Split(env, ",")
	}

	// 5. Middleware chain (orden correcto)
	handler := http.Handler(mux)

	handler = middleware.RecoveryMiddleware(handler)
	handler = middleware.CORSMiddleware(corsOrigins)(handler)
	handler = security.Middleware(ddosDetector)(handler)
	handler = ratelimit.Middleware(limiter)(handler)
	handler = middleware.LoggingMiddleware(handler)

	// 6. Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Servidor iniciado", map[string]interface{}{
		"port": port,
		"env":  os.Getenv("GO_ENV"),
	})

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Error crítico del servidor", "SERVER_ERROR", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
