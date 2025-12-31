
#!/bin/bash
# Health check script for Pizzas ECOS

echo "🔍 Checking Pizzas ECOS Setup..."
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ISSUES=0

# Check Docker
echo "📦 Checking Docker..."
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    echo -e "${GREEN}✓${NC} Docker installed: $DOCKER_VERSION"
else
    echo -e "${RED}✗${NC} Docker not found"
    ISSUES=$((ISSUES + 1))
fi

# Check Docker Compose
echo "🐳 Checking Docker Compose..."
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version)
    echo -e "${GREEN}✓${NC} Docker Compose installed: $COMPOSE_VERSION"
else
    echo -e "${RED}✗${NC} Docker Compose not found"
    ISSUES=$((ISSUES + 1))
fi

# Check Go
echo "🐹 Checking Go..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo -e "${GREEN}✓${NC} Go installed: $GO_VERSION"
else
    echo -e "${RED}✗${NC} Go not found"
    ISSUES=$((ISSUES + 1))
fi

# Check backend .env
echo "⚙️  Checking backend configuration..."
if [ -f "backend/.env" ]; then
    echo -e "${GREEN}✓${NC} backend/.env exists"
else
    if [ -f "backend/.env.example" ]; then
        echo -e "${YELLOW}⚠${NC} backend/.env missing (using .env.example as template)"
        echo "   Run: cp backend/.env.example backend/.env"
        ISSUES=$((ISSUES + 1))
    else
        echo -e "${RED}✗${NC} Neither backend/.env nor backend/.env.example found"
        ISSUES=$((ISSUES + 1))
    fi
fi

# Check Docker image
echo "🖼️  Checking Docker image..."
if docker images | grep -q pizzas-ecos; then
    IMAGE_SIZE=$(docker images | grep pizzas-ecos | head -1 | awk '{print $NF}')
    echo -e "${GREEN}✓${NC} pizzas-ecos image exists (size: $IMAGE_SIZE)"
else
    echo -e "${YELLOW}⚠${NC} pizzas-ecos image not built"
    echo "   Run: docker build -t pizzas-ecos:latest ./backend"
    ISSUES=$((ISSUES + 1))
fi

# Check backend files
echo "📄 Checking backend files..."
BACKEND_FILES=("backend/main.go" "backend/go.mod" "backend/Dockerfile")
for file in "${BACKEND_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $file exists"
    else
        echo -e "${RED}✗${NC} $file missing"
        ISSUES=$((ISSUES + 1))
    fi
done

# Check frontend files
echo "🎨 Checking frontend files..."
FRONTEND_FILES=("frontend/index.html" "frontend/admin.html" "frontend/estadisticas.html")
for file in "${FRONTEND_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $file exists"
    else
        echo -e "${RED}✗${NC} $file missing"
        ISSUES=$((ISSUES + 1))
    fi
done

# Check docker-compose.yml
echo "🐳 Checking docker-compose.yml..."
if [ -f "docker-compose.yml" ]; then
    echo -e "${GREEN}✓${NC} docker-compose.yml exists"
else
    echo -e "${RED}✗${NC} docker-compose.yml missing"
    ISSUES=$((ISSUES + 1))
fi

# Check Makefile
echo "🛠️  Checking Makefile..."
if [ -f "Makefile" ]; then
    echo -e "${GREEN}✓${NC} Makefile exists"
else
    echo -e "${YELLOW}⚠${NC} Makefile not found"
fi

# Check documentation
echo "📚 Checking documentation..."
DOCS=("QUICK_START.md" "DOCKER.md" "README.md" "STATUS.md")
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "${GREEN}✓${NC} $doc exists"
    else
        echo -e "${YELLOW}⚠${NC} $doc missing"
    fi
done

# Check git
echo "📋 Checking git..."
if [ -d ".git" ]; then
    echo -e "${GREEN}✓${NC} Git repository initialized"
else
    echo -e "${YELLOW}⚠${NC} Not a git repository"
fi

# Check running containers
echo "🚀 Checking running containers..."
if docker ps 2>/dev/null | grep -q pizzas-ecos; then
    echo -e "${GREEN}✓${NC} pizzas-ecos container is running"
else
    echo -e "${YELLOW}⚠${NC} pizzas-ecos container not running"
    echo "   Run: docker-compose up -d"
fi

# Summary
echo ""
echo "═══════════════════════════════════════════════════════════════"
if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed! Ready to go!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Configure backend/.env with your credentials"
    echo "  2. Run: docker-compose up -d"
    echo "  3. Visit: http://localhost:8080/api/v1/data"
    echo ""
else
    echo -e "${YELLOW}⚠ Found $ISSUES issue(s) to fix${NC}"
    echo ""
    echo "Common fixes:"
    echo "  • Install Docker: https://www.docker.com/products/docker-desktop"
    echo "  • Copy .env template: cp backend/.env.example backend/.env"
    echo "  • Build image: docker build -t pizzas-ecos:latest ./backend"
    echo "  • Start services: docker-compose up -d"
    echo ""
fi

exit $ISSUES
