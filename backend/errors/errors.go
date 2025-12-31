package errors

import (
	"encoding/json"
	"log"
	"net/http"
)

// ResponseError es la estructura estándar para errores API
type ResponseError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// ResponseSuccess es la estructura estándar para respuestas exitosas
type ResponseSuccess struct {
	Status  int         `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// CustomError es un error personalizado con código
type CustomError struct {
	Code     string
	Message  string
	HTTPCode int
	Details  string
}

// Errores predefinidos
var (
	ErrBadRequest = CustomError{
		Code:     "BAD_REQUEST",
		Message:  "Solicitud inválida",
		HTTPCode: http.StatusBadRequest,
	}
	ErrNotFound = CustomError{
		Code:     "NOT_FOUND",
		Message:  "Recurso no encontrado",
		HTTPCode: http.StatusNotFound,
	}
	ErrUnauthorized = CustomError{
		Code:     "UNAUTHORIZED",
		Message:  "No autorizado",
		HTTPCode: http.StatusUnauthorized,
	}
	ErrForbidden = CustomError{
		Code:     "FORBIDDEN",
		Message:  "Acceso denegado",
		HTTPCode: http.StatusForbidden,
	}
	ErrConflict = CustomError{
		Code:     "CONFLICT",
		Message:  "Recurso en conflicto",
		HTTPCode: http.StatusConflict,
	}
	ErrServerError = CustomError{
		Code:     "INTERNAL_SERVER_ERROR",
		Message:  "Error interno del servidor",
		HTTPCode: http.StatusInternalServerError,
	}
)

// WriteError escribe un error formateado a la respuesta HTTP
func WriteError(w http.ResponseWriter, err CustomError, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPCode)

	errResp := ResponseError{
		Status:  err.HTTPCode,
		Message: err.Message,
		Code:    err.Code,
		Details: details,
	}

	json.NewEncoder(w).Encode(errResp)
	if details != "" {
		log.Printf("⚠️ Error %s: %s - %s", err.Code, err.Message, details)
	}
}

// WriteSuccess escribe una respuesta exitosa formateada
func WriteSuccess(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ResponseSuccess{
		Status:  statusCode,
		Data:    data,
		Message: message,
	}

	json.NewEncoder(w).Encode(resp)
}

// WriteJSON escribe JSON directamente (para backward compatibility)
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
