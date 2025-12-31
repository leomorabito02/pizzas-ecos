# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar módulos
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux go build -o pizzas-ecos .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Copiar binary desde builder
COPY --from=builder /app/pizzas-ecos .

# Copiar .env si existe (opcional - usar variables de entorno en Docker)
COPY --from=builder /app/.env* ./

# Exponer puerto
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/api/v1/data || exit 1

# Comando de inicio
CMD ["./pizzas-ecos"]
