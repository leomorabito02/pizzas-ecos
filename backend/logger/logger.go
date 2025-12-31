package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// LogLevel define niveles de log
type LogLevel string

const (
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
	DebugLevel LogLevel = "DEBUG"
)

// LogEntry es una entrada de log estructurada
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     LogLevel    `json:"level"`
	Message   string      `json:"message"`
	Code      string      `json:"code,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Trace     string      `json:"trace,omitempty"`
}

var logger = log.New(os.Stdout, "", 0)

// logJSON escribe un log en formato JSON
func logJSON(entry LogEntry) {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	data, _ := json.Marshal(entry)
	logger.Println(string(data))
}

// Info registra un mensaje de información
func Info(message string, details ...interface{}) {
	logJSON(LogEntry{
		Level:   InfoLevel,
		Message: message,
		Details: details,
	})
}

// Warn registra una advertencia
func Warn(message string, details ...interface{}) {
	logJSON(LogEntry{
		Level:   WarnLevel,
		Message: message,
		Details: details,
	})
}

// Error registra un error
func Error(message string, code string, details ...interface{}) {
	logJSON(LogEntry{
		Level:   ErrorLevel,
		Message: message,
		Code:    code,
		Details: details,
	})
}

// Debug registra un debug (solo si DEBUG=true)
func Debug(message string, details ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		logJSON(LogEntry{
			Level:   DebugLevel,
			Message: message,
			Details: details,
		})
	}
}

// HTTPRequest registra una request HTTP
type HTTPRequestLog struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	IP         string `json:"ip"`
	StatusCode int    `json:"status_code"`
	Duration   string `json:"duration_ms"`
	UserAgent  string `json:"user_agent,omitempty"`
	Error      string `json:"error,omitempty"`
}

// LogHTTPRequest registra un request HTTP procesado
func LogHTTPRequest(method, path, ip string, statusCode int, duration time.Duration, userAgent, errMsg string) {
	entry := LogEntry{
		Level:   InfoLevel,
		Message: "HTTP Request",
		Details: HTTPRequestLog{
			Method:     method,
			Path:       path,
			IP:         ip,
			StatusCode: statusCode,
			Duration:   duration.String(),
			UserAgent:  userAgent,
			Error:      errMsg,
		},
	}
	logJSON(entry)
}

// DatabaseLog registra operaciones de BD
type DatabaseLog struct {
	Operation    string `json:"operation"`
	Table        string `json:"table"`
	Duration     string `json:"duration_ms"`
	Error        string `json:"error,omitempty"`
	RowsAffected int    `json:"rows_affected,omitempty"`
}

// LogDatabase registra operación de BD
func LogDatabase(operation, table string, duration time.Duration, rowsAffected int, errMsg string) {
	entry := LogEntry{
		Level:   InfoLevel,
		Message: "Database Operation",
		Details: DatabaseLog{
			Operation:    operation,
			Table:        table,
			Duration:     duration.String(),
			Error:        errMsg,
			RowsAffected: rowsAffected,
		},
	}
	logJSON(entry)
}
