#!/bin/bash
# Script para verificar que el backend funciona

echo "🚀 Iniciando backend..."
go run main.go &
BACKEND_PID=$!

# Esperar a que inicie
sleep 2

echo ""
echo "✅ Backend iniciado (PID: $BACKEND_PID)"
echo ""
echo "📡 Probando endpoint /api/data..."
curl -s http://localhost:8080/api/data | jq '.' | head -20

echo ""
echo "📝 Backend está corriendo en http://localhost:8080"
echo "Presiona Ctrl+C para detener"
echo ""

wait $BACKEND_PID
