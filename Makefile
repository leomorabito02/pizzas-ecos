.PHONY: help build test clean docker-build docker-up docker-down docker-logs backend-build frontend-build dev ci

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Pizzas ECOS - Development Commands$(NC)"
	@echo ""
	@echo "$(GREEN)Docker:$(NC)"
	@echo "  make docker-build        Build Docker image"
	@echo "  make docker-up           Start services with docker-compose"
	@echo "  make docker-down         Stop services"
	@echo "  make docker-logs         View container logs"
	@echo "  make docker-clean        Remove containers and images"
	@echo ""
	@echo "$(GREEN)Backend:$(NC)"
	@echo "  make backend-build       Build backend binary"
	@echo "  make backend-run         Run backend locally"
	@echo "  make backend-test        Run backend tests"
	@echo "  make backend-test-cov    Run backend tests with coverage"
	@echo "  make backend-lint        Run Go linter"
	@echo "  make backend-fmt         Format Go code"
	@echo "  make backend-vet         Run Go vet"
	@echo ""
	@echo "$(GREEN)Frontend:$(NC)"
	@echo "  make frontend-build      Build frontend"
	@echo "  make frontend-test       Run frontend tests"
	@echo "  make frontend-install    Install dependencies"
	@echo ""
	@echo "$(GREEN)Development:$(NC)"
	@echo "  make dev                 Start full dev environment"
	@echo "  make ci                  Run full CI pipeline locally"
	@echo "  make clean               Clean build artifacts"
	@echo ""

# ===== Docker Commands =====

docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	cd backend && docker build -t pizzas-ecos:latest .
	@echo "$(GREEN)✓ Docker image built successfully$(NC)"

docker-up: ## Start services with docker-compose
	@echo "$(YELLOW)Starting services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo "Backend running at http://localhost:8080"

docker-down: ## Stop services
	@echo "$(YELLOW)Stopping services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

docker-logs: ## View logs
	docker-compose logs -f backend

docker-ps: ## Show running containers
	docker-compose ps

docker-exec: ## Execute shell in container
	docker-compose exec backend sh

docker-clean: ## Remove images and containers
	@echo "$(YELLOW)Cleaning Docker resources...$(NC)"
	docker-compose down -v
	docker rmi pizzas-ecos:latest 2>/dev/null || true
	@echo "$(GREEN)✓ Cleaned up Docker resources$(NC)"

docker-rebuild: docker-clean docker-build docker-up ## Clean rebuild and restart

# ===== Backend Commands =====

backend-build: ## Build backend binary
	@echo "$(YELLOW)Building backend...$(NC)"
	cd backend && go build -o pizzas-ecos .
	@echo "$(GREEN)✓ Backend built successfully$(NC)"

backend-run: backend-build ## Run backend locally
	@echo "$(YELLOW)Running backend...$(NC)"
	cd backend && ./pizzas-ecos

backend-test: ## Run backend tests
	@echo "$(YELLOW)Running backend tests...$(NC)"
	cd backend && go test ./... -v
	@echo "$(GREEN)✓ Tests completed$(NC)"

backend-test-cov: ## Run backend tests with coverage
	@echo "$(YELLOW)Running backend tests with coverage...$(NC)"
	cd backend && go test ./... -v -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out
	@echo "$(GREEN)✓ Coverage report generated$(NC)"

backend-lint: ## Run Go linter
	@echo "$(YELLOW)Linting backend...$(NC)"
	cd backend && golangci-lint run ./...
	@echo "$(GREEN)✓ Linting completed$(NC)"

backend-fmt: ## Format Go code
	@echo "$(YELLOW)Formatting backend code...$(NC)"
	cd backend && go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

backend-vet: ## Run Go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	cd backend && go vet ./...
	@echo "$(GREEN)✓ Vet check completed$(NC)"

backend-deps: ## Update backend dependencies
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	cd backend && go get -u ./...
	cd backend && go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

# ===== Frontend Commands =====

frontend-install: ## Install frontend dependencies
	@echo "$(YELLOW)Installing frontend dependencies...$(NC)"
	cd frontend && npm install
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

frontend-build: ## Build frontend
	@echo "$(YELLOW)Building frontend...$(NC)"
	cd frontend && npm run build
	@echo "$(GREEN)✓ Frontend built$(NC)"

frontend-test: ## Run frontend tests
	@echo "$(YELLOW)Running frontend tests...$(NC)"
	cd frontend && npm test
	@echo "$(GREEN)✓ Tests completed$(NC)"

frontend-dev: ## Start frontend dev server
	@echo "$(YELLOW)Starting frontend dev server...$(NC)"
	cd frontend && npm run dev

# ===== Development =====

dev: docker-up ## Start full development environment
	@echo "$(GREEN)✓ Development environment started$(NC)"
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:5000 (if running separately)"

ci: backend-lint backend-build backend-test docker-build ## Run full CI pipeline locally
	@echo "$(GREEN)✓ CI pipeline completed successfully$(NC)"

# ===== Utility Commands =====

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	cd backend && go clean
	rm -f backend/pizzas-ecos
	rm -f backend/coverage.out
	@echo "$(GREEN)✓ Cleaned$(NC)"

install-tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)✓ Tools installed$(NC)"

# ===== Information Commands =====

info-docker: ## Show Docker info
	@echo "$(BLUE)Docker Version:$(NC)"
	docker --version
	@echo "$(BLUE)Docker Compose Version:$(NC)"
	docker-compose --version
	@echo "$(BLUE)Go Version:$(NC)"
	go version

info-images: ## Show available images
	@echo "$(BLUE)Docker Images:$(NC)"
	docker images | grep pizzas-ecos || echo "No images found"

info-containers: ## Show running containers
	@echo "$(BLUE)Running Containers:$(NC)"
	docker ps | grep pizzas-ecos || echo "No containers running"

.DEFAULT_GOAL := help
