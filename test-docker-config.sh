#!/bin/bash
# Docker Configuration Test Script
# Verifica que la configuración de Docker está correcta

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}╔═══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Pizza ECOS - Docker Configuration Test ${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════╝${NC}"
echo ""

# Test 1: Check docker installation
echo -e "${YELLOW}[1/8]${NC} Verificando Docker..."
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    echo -e "${GREEN}✓${NC} Docker instalado: $DOCKER_VERSION"
else
    echo -e "${RED}✗${NC} Docker no está instalado"
    exit 1
fi

# Test 2: Check docker-compose
echo -e "${YELLOW}[2/8]${NC} Verificando Docker Compose..."
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version)
    echo -e "${GREEN}✓${NC} Docker Compose instalado: $COMPOSE_VERSION"
else
    echo -e "${RED}✗${NC} Docker Compose no está instalado"
    exit 1
fi

# Test 3: Check docker-compose.yml
echo -e "${YELLOW}[3/8]${NC} Verificando docker-compose.yml..."
if [ -f "docker-compose.yml" ]; then
    echo -e "${GREEN}✓${NC} docker-compose.yml encontrado"
    docker-compose config > /dev/null 2>&1 && echo -e "${GREEN}✓${NC} Sintaxis válida" || echo -e "${RED}✗${NC} Sintaxis inválida"
else
    echo -e "${RED}✗${NC} docker-compose.yml no encontrado"
    exit 1
fi

# Test 4: Check .env file
echo -e "${YELLOW}[4/8]${NC} Verificando configuración de entorno..."
if [ -f ".env" ]; then
    echo -e "${GREEN}✓${NC} .env existe"
    
    # Check for required variables
    REQUIRED_VARS=("MYSQL_ROOT_PASSWORD" "DATABASE_URL" "JWT_SECRET")
    for var in "${REQUIRED_VARS[@]}"; do
        if grep -q "$var" .env; then
            echo -e "${GREEN}✓${NC} $var configurado"
        else
            echo -e "${YELLOW}⚠${NC} $var no encontrado en .env"
        fi
    done
else
    echo -e "${YELLOW}⚠${NC} .env no existe, crear desde .env.example:"
    echo "   cp .env.example .env"
fi

# Test 5: Check Dockerfile
echo -e "${YELLOW}[5/8]${NC} Verificando Dockerfile..."
if [ -f "backend/Dockerfile" ]; then
    echo -e "${GREEN}✓${NC} Dockerfile encontrado"
    
    # Check for required elements
    if grep -q "FROM golang" backend/Dockerfile; then
        echo -e "${GREEN}✓${NC} Base image (golang) configurada"
    fi
    
    if grep -q "EXPOSE 8080" backend/Dockerfile; then
        echo -e "${GREEN}✓${NC} Puerto 8080 expuesto"
    fi
    
    if grep -q "HEALTHCHECK" backend/Dockerfile; then
        echo -e "${GREEN}✓${NC} Health check configurado"
    fi
else
    echo -e "${RED}✗${NC} Dockerfile no encontrado"
    exit 1
fi

# Test 6: Check .dockerignore
echo -e "${YELLOW}[6/8]${NC} Verificando .dockerignore..."
if [ -f "backend/.dockerignore" ]; then
    echo -e "${GREEN}✓${NC} .dockerignore existe"
else
    echo -e "${YELLOW}⚠${NC} .dockerignore no existe (recomendado crearlo)"
fi

# Test 7: Check helper scripts
echo -e "${YELLOW}[7/8]${NC} Verificando scripts helper..."
if [ -f "docker-manage.sh" ]; then
    echo -e "${GREEN}✓${NC} docker-manage.sh encontrado"
    if [ -x "docker-manage.sh" ]; then
        echo -e "${GREEN}✓${NC} docker-manage.sh es ejecutable"
    else
        echo -e "${YELLOW}⚠${NC} docker-manage.sh no es ejecutable, hacerlo con: chmod +x docker-manage.sh"
    fi
else
    echo -e "${YELLOW}⚠${NC} docker-manage.sh no encontrado"
fi

if [ -f "docker-manage.ps1" ]; then
    echo -e "${GREEN}✓${NC} docker-manage.ps1 encontrado (para Windows)"
else
    echo -e "${YELLOW}⚠${NC} docker-manage.ps1 no encontrado"
fi

# Test 8: Check documentation
echo -e "${YELLOW}[8/8]${NC} Verificando documentación..."
DOCS=("DOCKER_SETUP.md" "DOCKER_CONFIG_SUMMARY.md" "DOCKER_COMMANDS.md")
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "${GREEN}✓${NC} $doc encontrado"
    else
        echo -e "${YELLOW}⚠${NC} $doc no encontrado"
    fi
done

# Summary
echo ""
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo -e "${GREEN}✓ Configuración de Docker verificada${NC}"
echo ""
echo -e "${BLUE}Próximos pasos:${NC}"
echo "  1. Editar .env con tus valores"
echo "  2. docker-compose up -d"
echo "  3. docker-compose logs -f backend"
echo "  4. Verificar: curl http://localhost:8080/api/v1/data"
echo ""
echo -e "${BLUE}Para más información:${NC}"
echo "  - DOCKER_SETUP.md (guía completa)"
echo "  - DOCKER_CONFIG_SUMMARY.md (resumen rápido)"
echo "  - DOCKER_COMMANDS.md (comandos frecuentes)"
echo ""
