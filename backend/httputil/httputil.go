package httputil

import (
	"net/http"
)

// contextKey para almacenar parámetros
type contextKey string

const PathParamsKey contextKey = "path_params"

// PathParams contiene parámetros extraídos de la ruta
type PathParams map[string]string

// GetParam obtiene un parámetro de la ruta desde el request context
func GetParam(r *http.Request, key string) string {
	params, ok := r.Context().Value(PathParamsKey).(PathParams)
	if !ok {
		return ""
	}
	return params[key]
}
