package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implementa limitación de tasa por IP
type RateLimiter struct {
	requestsPerSecond int
	ips               map[string]*ipData
	mu                sync.RWMutex
	cleanupInterval   time.Duration
}

// ipData contiene datos de rate limit por IP
type ipData struct {
	count     int
	timestamp time.Time
	lastSeen  time.Time
}

// NewRateLimiter crea un nuevo rate limiter
// requestsPerSecond: máximo de requests permitidos por segundo por IP
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	rl := &RateLimiter{
		requestsPerSecond: requestsPerSecond,
		ips:               make(map[string]*ipData),
		cleanupInterval:   5 * time.Minute,
	}

	// Limpieza periódica de IPs antiguas
	go rl.cleanup()

	return rl
}

// Allow verifica si la IP puede hacer una request
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	data, exists := rl.ips[ip]

	if !exists {
		// Primera vez que vemos esta IP
		rl.ips[ip] = &ipData{
			count:     1,
			timestamp: now,
			lastSeen:  now,
		}
		return true
	}

	// Chequear si pasó 1 segundo desde el primer request
	if now.Sub(data.timestamp) >= time.Second {
		// Reset del contador
		data.count = 1
		data.timestamp = now
		data.lastSeen = now
		return true
	}

	// Dentro del mismo segundo, verificar límite
	data.count++
	data.lastSeen = now
	return data.count <= rl.requestsPerSecond
}

// cleanup limpia IPs que no han sido vistas recientemente
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, data := range rl.ips {
			if now.Sub(data.lastSeen) > rl.cleanupInterval {
				delete(rl.ips, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware retorna un middleware de rate limiting
func Middleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener IP del cliente
			ip := getClientIP(r)

			// Verificar límite
			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"status":429,"message":"Too many requests","code":"RATE_LIMIT_EXCEEDED"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extrae la IP del cliente del request
func getClientIP(r *http.Request) string {
	// Intentar obtener de X-Forwarded-For (proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Intentar X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Usar RemoteAddr
	return r.RemoteAddr
}
