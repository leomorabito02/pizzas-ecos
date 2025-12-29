#!/bin/bash

# Script para ayudar con el despliegue de Pizzas ECOS

echo "🍕 Pizzas ECOS - Deployment Helper"
echo "===================================="
echo ""

# Detectar el sistema operativo
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    echo "⚠️  Detectado Windows. Usa el archivo deploy.bat en su lugar"
    exit 1
fi

# Menú principal
echo "¿Qué deseas hacer?"
echo "1. Verificar instalación local"
echo "2. Iniciar desarrollo local"
echo "3. Preparar para despliegue"
echo "4. Ver instrucciones de despliegue"
echo ""
read -p "Selecciona opción (1-4): " option

case $option in
    1)
        echo ""
        echo "🔍 Verificando instalación..."
        echo ""
        
        # Verificar Go
        if command -v go &> /dev/null; then
            echo "✅ Go instalado: $(go version)"
        else
            echo "❌ Go no instalado"
        fi
        
        # Verificar archivo .env en backend
        if [ -f "backend/.env" ]; then
            echo "✅ backend/.env existe"
        else
            echo "⚠️  backend/.env no encontrado"
            echo "   Copiar: cp backend/.env.example backend/.env"
        fi
        
        # Verificar credenciales
        if [ -f "venta-pizzas-ecos.json" ]; then
            echo "✅ Credenciales de Google existen"
        else
            echo "⚠️  venta-pizzas-ecos.json no encontrado"
        fi
        
        # Verificar archivos del frontend
        if [ -f "frontend/index.html" ] && [ -f "frontend/form.js" ]; then
            echo "✅ Archivos del frontend existen"
        else
            echo "❌ Archivos del frontend no encontrados"
        fi
        
        echo ""
        echo "Instalación verificada ✓"
        ;;
        
    2)
        echo ""
        echo "🚀 Iniciando desarrollo local..."
        echo ""
        
        # Verificar que no esté ya corriendo
        if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
            echo "⚠️  Puerto 8080 ya está en uso"
        else
            echo "Iniciando backend en puerto 8080..."
            cd backend && go run main.go &
            BACKEND_PID=$!
        fi
        
        sleep 2
        
        # Frontend
        if lsof -Pi :5000 -sTCP:LISTEN -t >/dev/null ; then
            echo "⚠️  Puerto 5000 ya está en uso"
        else
            echo "Iniciando frontend en puerto 5000..."
            cd frontend && python -m http.server 5000 &
        fi
        
        echo ""
        echo "✅ Servicios iniciados:"
        echo "   Frontend: http://localhost:5000"
        echo "   Backend:  http://localhost:8080"
        echo ""
        echo "Presiona Ctrl+C para detener"
        wait
        ;;
        
    3)
        echo ""
        echo "📦 Preparando para despliegue..."
        echo ""
        
        # Verificar .env
        if [ -f ".env" ]; then
            echo "⚠️  Archivo .env encontrado en raíz"
            echo "   Asegúrate que NO esté commiteado (.gitignore)"
        fi
        
        # Verificar credenciales
        if [ -f "venta-pizzas-ecos.json" ]; then
            echo "⚠️  Credenciales JSON encontradas"
            echo "   Asegúrate que NO estén commiteadas (.gitignore)"
        fi
        
        # Check git
        echo ""
        echo "🔍 Verificando Git..."
        if git status &> /dev/null; then
            echo "✅ Repositorio Git válido"
            
            # Ver cambios
            echo ""
            echo "📝 Cambios pendientes:"
            git status --short
            
            echo ""
            echo "¿Deseas hacer commit? (y/n)"
            read -p "> " commit
            if [ "$commit" == "y" ]; then
                git add .
                read -p "Mensaje de commit: " msg
                git commit -m "$msg"
                git push
                echo "✅ Cambios subidos a GitHub"
            fi
        else
            echo "❌ No es un repositorio Git"
        fi
        
        echo ""
        echo "Próximo paso: Visita https://vercel.com/import"
        ;;
        
    4)
        echo ""
        echo "📖 Instrucciones de despliegue"
        echo "================================"
        echo ""
        cat DEPLOYMENT.md
        ;;
        
    *)
        echo "Opción no válida"
        ;;
esac
