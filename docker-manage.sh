#!/bin/bash
# Docker management script for Pizza ECOS project
# Usage: ./docker-manage.sh [command]

set -e

COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="pizzas-ecos"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Commands
cmd_help() {
    cat << EOF
${BLUE}Pizza ECOS - Docker Management${NC}

Usage: ./docker-manage.sh [command]

Commands:
    up              Levantar contenedores en background
    down            Detener y remover contenedores
    restart         Reiniciar todos los contenedores
    logs [service]  Ver logs (backend|mysql|all)
    shell [service] Acceder a shell del contenedor (backend|mysql)
    ps              Ver estado de contenedores
    clean           Limpiar contenedores detenidos
    rebuild         Reconstruir imagen del backend
    test            Correr tests
    db-init         Inicializar base de datos
    db-backup       Hacer backup de MySQL
    db-restore      Restaurar backup de MySQL
    health          Ver health checks
    stats           Ver estadísticas de recursos
    help            Mostrar esta ayuda

Examples:
    ./docker-manage.sh up
    ./docker-manage.sh logs backend
    ./docker-manage.sh shell mysql
    ./docker-manage.sh db-backup

EOF
}

cmd_up() {
    print_header "Levantando contenedores"
    docker-compose -f "$COMPOSE_FILE" up -d
    print_success "Contenedores levantados"
    sleep 2
    cmd_ps
    echo ""
    print_header "Próximos pasos"
    echo "Backend:     http://localhost:8080"
    echo "MySQL:       localhost:3306"
    echo "Ver logs:    ./docker-manage.sh logs"
}

cmd_down() {
    print_header "Deteniendo contenedores"
    docker-compose -f "$COMPOSE_FILE" down
    print_success "Contenedores detenidos"
}

cmd_restart() {
    print_header "Reiniciando contenedores"
    docker-compose -f "$COMPOSE_FILE" restart
    print_success "Contenedores reiniciados"
    sleep 2
    cmd_ps
}

cmd_logs() {
    local service=${1:-all}
    print_header "Logs de $service"
    
    if [ "$service" = "all" ] || [ -z "$service" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f
    else
        docker-compose -f "$COMPOSE_FILE" logs -f "$service"
    fi
}

cmd_shell() {
    local service=${1:-backend}
    print_header "Abriendo shell en $service"
    docker-compose -f "$COMPOSE_FILE" exec "$service" sh
}

cmd_ps() {
    print_header "Estado de contenedores"
    docker-compose -f "$COMPOSE_FILE" ps
}

cmd_clean() {
    print_header "Limpiando Docker"
    docker container prune -f
    docker image prune -f
    print_success "Limpieza completada"
}

cmd_rebuild() {
    print_header "Reconstruyendo imagen del backend"
    docker-compose -f "$COMPOSE_FILE" build --no-cache backend
    print_success "Imagen reconstruida"
}

cmd_test() {
    print_header "Ejecutando tests"
    docker-compose -f "$COMPOSE_FILE" exec backend go test -v ./...
}

cmd_db_init() {
    print_header "Inicializando base de datos"
    
    if ! docker-compose -f "$COMPOSE_FILE" ps mysql | grep -q "Up"; then
        print_error "MySQL no está corriendo"
        print_warning "Levantando MySQL..."
        docker-compose -f "$COMPOSE_FILE" up -d mysql
        sleep 5
    fi
    
    docker-compose -f "$COMPOSE_FILE" exec mysql mysql -u root -p"${MYSQL_ROOT_PASSWORD:-root}" < backend/create_user.sql
    print_success "Base de datos inicializada"
}

cmd_db_backup() {
    print_header "Haciendo backup de MySQL"
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="backups/mysql_backup_${timestamp}.sql"
    
    mkdir -p backups
    
    docker-compose -f "$COMPOSE_FILE" exec mysql mysqldump -u root -p"${MYSQL_ROOT_PASSWORD:-root}" --all-databases > "$backup_file"
    
    print_success "Backup creado: $backup_file"
    ls -lh "$backup_file"
}

cmd_db_restore() {
    local backup_file=$1
    
    if [ -z "$backup_file" ]; then
        print_error "Debes especificar archivo de backup"
        echo "Backups disponibles:"
        ls -lh backups/ 2>/dev/null || echo "No hay backups"
        exit 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        print_error "Archivo no existe: $backup_file"
        exit 1
    fi
    
    print_header "Restaurando backup: $backup_file"
    
    if ! docker-compose -f "$COMPOSE_FILE" ps mysql | grep -q "Up"; then
        print_warning "Levantando MySQL..."
        docker-compose -f "$COMPOSE_FILE" up -d mysql
        sleep 5
    fi
    
    docker-compose -f "$COMPOSE_FILE" exec -T mysql mysql -u root -p"${MYSQL_ROOT_PASSWORD:-root}" < "$backup_file"
    print_success "Backup restaurado"
}

cmd_health() {
    print_header "Health Checks"
    
    echo ""
    echo "Backend:"
    docker-compose -f "$COMPOSE_FILE" exec backend wget --quiet --tries=1 --spider http://localhost:8080/api/v1/data && print_success "Backend OK" || print_error "Backend FAILED"
    
    echo ""
    echo "MySQL:"
    docker-compose -f "$COMPOSE_FILE" exec mysql mysqladmin ping -h localhost -u root -p"${MYSQL_ROOT_PASSWORD:-root}" && print_success "MySQL OK" || print_error "MySQL FAILED"
}

cmd_stats() {
    print_header "Estadísticas de Recursos"
    docker stats --no-stream
}

# Main
main() {
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_error "No se encontró $COMPOSE_FILE"
        exit 1
    fi
    
    case "${1:-help}" in
        up)         cmd_up ;;
        down)       cmd_down ;;
        restart)    cmd_restart ;;
        logs)       cmd_logs "$2" ;;
        shell)      cmd_shell "$2" ;;
        ps)         cmd_ps ;;
        clean)      cmd_clean ;;
        rebuild)    cmd_rebuild ;;
        test)       cmd_test ;;
        db-init)    cmd_db_init ;;
        db-backup)  cmd_db_backup ;;
        db-restore) cmd_db_restore "$2" ;;
        health)     cmd_health ;;
        stats)      cmd_stats ;;
        *)          cmd_help ;;
    esac
}

main "$@"
