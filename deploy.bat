@echo off
REM Script para ayudar con el despliegue de Pizzas ECOS en Windows

echo.
echo 🍕 Pizzas ECOS - Deployment Helper
echo ====================================
echo.

REM Menú principal
echo ¿Qué deseas hacer?
echo 1. Verificar instalación local
echo 2. Iniciar desarrollo local
echo 3. Preparar para despliegue
echo 4. Ver instrucciones de despliegue
echo.
set /p option="Selecciona opción (1-4): "

if "%option%"=="1" goto check_install
if "%option%"=="2" goto start_dev
if "%option%"=="3" goto prepare_deploy
if "%option%"=="4" goto show_docs
goto end

:check_install
echo.
echo 🔍 Verificando instalación...
echo.

REM Verificar Go
where go >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    echo ✅ Go instalado
    go version
) else (
    echo ❌ Go no instalado
)

REM Verificar .env en backend
if exist "backend\.env" (
    echo ✅ backend\.env existe
) else (
    echo ⚠️  backend\.env no encontrado
    echo    Debes crear uno con tus credenciales
)

REM Verificar credenciales
if exist "venta-pizzas-ecos.json" (
    echo ✅ Credenciales de Google existen
) else (
    echo ⚠️  venta-pizzas-ecos.json no encontrado
)

REM Verificar archivos del frontend
if exist "frontend\index.html" (
    echo ✅ Archivos del frontend existen
) else (
    echo ❌ Archivos del frontend no encontrados
)

echo.
echo Instalación verificada ✓
goto end

:start_dev
echo.
echo 🚀 Iniciando desarrollo local...
echo.

REM Crear dos ventanas CMD - una para backend, otra para frontend
start cmd /k "cd backend && go run main.go"
timeout /t 2 /nobreak

echo Iniciando frontend en puerto 5000...
start cmd /k "cd frontend && python -m http.server 5000"

echo.
echo ✅ Servicios iniciados:
echo    Frontend: http://localhost:5000
echo    Backend:  http://localhost:8080
echo.
goto end

:prepare_deploy
echo.
echo 📦 Preparando para despliegue...
echo.

REM Verificar .env
if exist ".env" (
    echo ⚠️  Archivo .env encontrado en raíz
    echo    Asegúrate que esté en .gitignore
)

REM Verificar credenciales
if exist "venta-pizzas-ecos.json" (
    echo ⚠️  Credenciales JSON encontradas
    echo    Asegúrate que estén en .gitignore
)

echo.
echo 🔍 Verificando Git...
git status >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    echo ✅ Repositorio Git válido
    
    echo.
    echo 📝 Cambios pendientes:
    git status --short
    
    echo.
    set /p commit="¿Deseas hacer commit? (s/n): "
    if /i "%commit%"=="s" (
        git add .
        set /p msg="Mensaje de commit: "
        git commit -m "%msg%"
        git push
        echo ✅ Cambios subidos a GitHub
    )
) else (
    echo ❌ No es un repositorio Git
)

echo.
echo Próximo paso: Visita https://vercel.com/import
goto end

:show_docs
echo.
echo 📖 Instrucciones de despliegue
echo ================================
echo.
type DEPLOYMENT.md
goto end

:end
echo.
pause
