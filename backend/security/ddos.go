package security

import (
	"net/http"
	"sync"
	"time"
)

// DDoSDetector detecta intentos de DDoS
type DDoSDetector struct {
	ipThresholds     map[string]int       // Requests por IP en ventana actual
	ipBlocklist      map[string]time.Time // IPs bloqueadas con timestamp
	windowStart      time.Time
	windowDuration   time.Duration
	maxRequestsPerIP int // Máximo de requests en ventana
	blockDuration    time.Duration
	mu               sync.RWMutex
}

// NewDDoSDetector crea un detector de DDoS
func NewDDoSDetector(maxRequestsPerIP int, windowDuration time.Duration) *DDoSDetector {
	d := &DDoSDetector{
		ipThresholds:     make(map[string]int),
		ipBlocklist:      make(map[string]time.Time),
		windowStart:      time.Now(),
		windowDuration:   windowDuration,
		maxRequestsPerIP: maxRequestsPerIP,
		blockDuration:    5 * time.Minute,
	}

	// Cleanup periódico
	go d.cleanup()

	return d
}

// IsBlocked verifica si una IP está bloqueada
func (d *DDoSDetector) IsBlocked(ip string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	blockTime, exists := d.ipBlocklist[ip]
	if !exists {
		return false
	}

	// Desbloquear si pasó el tiempo
	if time.Since(blockTime) > d.blockDuration {
		return false
	}

	return true
}

// RecordRequest registra un request de una IP
func (d *DDoSDetector) RecordRequest(ip string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Resetear ventana si pasó el tiempo
	if time.Since(d.windowStart) > d.windowDuration {
		d.ipThresholds = make(map[string]int)
		d.windowStart = time.Now()
	}

	// Incrementar contador
	d.ipThresholds[ip]++

	// Detectar DDoS
	if d.ipThresholds[ip] > d.maxRequestsPerIP {
		d.ipBlocklist[ip] = time.Now()
		return false // Bloqueado
	}

	return true // Permitido
}

// GetBlockedCount retorna cantidad de IPs bloqueadas
func (d *DDoSDetector) GetBlockedCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.ipBlocklist)
}

// cleanup limpia IPs desbloqueadas
func (d *DDoSDetector) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		d.mu.Lock()
		now := time.Now()
		for ip, blockTime := range d.ipBlocklist {
			if now.Sub(blockTime) > d.blockDuration {
				delete(d.ipBlocklist, ip)
			}
		}
		d.mu.Unlock()
	}
}

// Middleware retorna middleware de DDoS protection
func Middleware(detector *DDoSDetector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			// Verificar si IP está bloqueada
			if detector.IsBlocked(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"status":403,"message":"IP blocked due to suspicious activity","code":"IP_BLOCKED"}`))
				return
			}

			// Registrar request
			if !detector.RecordRequest(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"status":403,"message":"Too many requests from this IP","code":"DDOS_DETECTED"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extrae IP del cliente
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
