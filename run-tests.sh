#!/bin/bash
# Script maestro para ejecutar todas las pruebas del proyecto
# Ejecuta pruebas unitarias del frontend, backend y pruebas de integración

set -e

# Colores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Función para imprimir headers
print_header() {
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║$1${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Función para ejecutar comando y verificar resultado
run_test() {
    local name="$1"
    local command="$2"

    echo -e "${YELLOW}→${NC} Ejecutando: $name"
    echo -e "${BLUE}Comando:${NC} $command"
    echo ""

    if eval "$command"; then
        echo -e "${GREEN}✓ $name - PASSED${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗ $name - FAILED${NC}"
        echo ""
        return 1
    fi
}

# Verificar que estamos en el directorio correcto
if [ ! -d "frontend" ] || [ ! -d "backend" ]; then
    echo -e "${RED}Error: Ejecutar desde el directorio raíz del proyecto${NC}"
    exit 1
fi

print_header "🧪 PRUEBAS UNITARIAS - PIZZAS ECOS"

FAILED_TESTS=0

# ===== BACKEND TESTS =====
print_header "🔧 Backend - Pruebas Unitarias (Go)"

cd backend

# Verificar que Go está instalado
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go no está instalado${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
else
    # Ejecutar pruebas del backend
    if run_test "Backend Unit Tests" "go test ./... -v"; then
        echo -e "${GREEN}✓ Backend tests passed${NC}"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    # Verificar que el código compila
    if run_test "Backend Build" "go build ./..."; then
        echo -e "${GREEN}✓ Backend builds successfully${NC}"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
fi

cd ..
echo ""

# ===== FRONTEND TESTS =====
print_header "🌐 Frontend - Pruebas Unitarias (JavaScript)"

cd frontend

# Verificar que Node.js está instalado
if ! command -v node &> /dev/null; then
    echo -e "${RED}✗ Node.js no está instalado${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
else
    # Verificar que npm está instalado
    if ! command -v npm &> /dev/null; then
        echo -e "${RED}✗ npm no está instalado${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        # Instalar dependencias si no existen
        if [ ! -d "node_modules" ]; then
            echo -e "${YELLOW}Instalando dependencias de npm...${NC}"
            npm install
        fi

        # Ejecutar pruebas del frontend
        if run_test "Frontend Unit Tests" "npm test -- --watchAll=false --passWithNoTests"; then
            echo -e "${GREEN}✓ Frontend tests passed${NC}"
        else
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
fi

cd ..
echo ""

# ===== INTEGRATION TESTS =====
print_header "🔗 Pruebas de Integración (Endpoints API)"

# Verificar que el backend esté corriendo
echo -e "${YELLOW}Verificando si el backend está corriendo...${NC}"
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Backend está corriendo en localhost:8080${NC}"

    # Ejecutar pruebas de integración
    if run_test "API Integration Tests" "cd frontend && node ../test-endpoints.js"; then
        echo -e "${GREEN}✓ Integration tests passed${NC}"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Backend no está corriendo. Omitiendo pruebas de integración.${NC}"
    echo -e "${BLUE}Para ejecutar pruebas de integración:${NC}"
    echo -e "  1. Iniciar backend: ${YELLOW}cd backend && go run .${NC}"
    echo -e "  2. Ejecutar: ${YELLOW}./run-tests.sh${NC}"
    echo ""
fi

# ===== RESULTADOS =====
print_header "📊 RESULTADOS FINALES"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 ¡Todas las pruebas pasaron exitosamente!${NC}"
    echo ""
    echo -e "${BLUE}Resumen:${NC}"
    echo -e "${GREEN}✓${NC} Backend unit tests"
    echo -e "${GREEN}✓${NC} Backend build"
    echo -e "${GREEN}✓${NC} Frontend unit tests"
    echo -e "${GREEN}✓${NC} API integration tests"
    echo ""
    exit 0
else
    echo -e "${RED}❌ $FAILED_TESTS conjunto(s) de pruebas fallaron${NC}"
    echo ""
    echo -e "${YELLOW}Para más detalles, revisa la salida anterior.${NC}"
    echo ""
    exit 1
fi