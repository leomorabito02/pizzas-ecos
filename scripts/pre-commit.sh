#!/bin/bash
# Pre-commit hook para validar código antes de hacer commit
# Instalar: cp scripts/pre-commit .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit

set -e

echo "🔍 Running pre-commit checks..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track if any check failed
FAILED=0

# ===== BACKEND CHECKS =====
echo -e "\n${YELLOW}→ Checking backend...${NC}"

if [ -d "backend" ]; then
    cd backend
    
    # Check Go formatting
    if ! go fmt ./... >/dev/null 2>&1; then
        echo -e "${RED}✗ Go formatting failed${NC}"
        FAILED=1
    else
        echo -e "${GREEN}✓ Go formatting OK${NC}"
    fi
    
    # Check Go vet
    if ! go vet ./... >/dev/null 2>&1; then
        echo -e "${RED}✗ Go vet failed${NC}"
        FAILED=1
    else
        echo -e "${GREEN}✓ Go vet OK${NC}"
    fi
    
    # Check if code compiles
    if ! go build -v -o /tmp/test-build ./. >/dev/null 2>&1; then
        echo -e "${RED}✗ Backend build failed${NC}"
        FAILED=1
    else
        echo -e "${GREEN}✓ Backend builds successfully${NC}"
        rm -f /tmp/test-build
    fi
    
    cd ..
fi

# ===== FRONTEND CHECKS =====
echo -e "\n${YELLOW}→ Checking frontend...${NC}"

if [ -d "frontend" ]; then
    # Check HTML files
    html_files=$(find frontend -maxdepth 1 -name "*.html" 2>/dev/null || true)
    if [ -n "$html_files" ]; then
        for file in $html_files; do
            if ! grep -q "<!DOCTYPE html>" "$file"; then
                echo -e "${RED}✗ $file: Missing DOCTYPE${NC}"
                FAILED=1
            fi
        done
        if [ $FAILED -eq 0 ]; then
            echo -e "${GREEN}✓ HTML files OK${NC}"
        fi
    fi
    
    # Check for common mistakes in JS
    js_files=$(find frontend/js -name "*.js" 2>/dev/null || true)
    if [ -n "$js_files" ]; then
        for file in $js_files; do
            # Check for console.log left in code
            if grep -q "console\\.log" "$file" && [ ! "$file" = "frontend/js/logger.js" ]; then
                echo -e "${YELLOW}⚠ $file: Contains console.log (check if intentional)${NC}"
            fi
            # Check for TODO/FIXME
            if grep -q "TODO\|FIXME" "$file"; then
                echo -e "${YELLOW}⚠ $file: Contains TODO/FIXME${NC}"
            fi
        done
        echo -e "${GREEN}✓ JavaScript files checked${NC}"
    fi
fi

# ===== GENERAL CHECKS =====
echo -e "\n${YELLOW}→ Running general checks...${NC}"

# Check for large files
large_files=$(find . -type f -size +10M ! -path "./.git/*" ! -path "./node_modules/*" ! -path "./vendor/*" 2>/dev/null || true)
if [ -n "$large_files" ]; then
    echo -e "${YELLOW}⚠ Large files detected:${NC}"
    echo "$large_files"
fi

# Check for secrets
if git rev-parse --git-dir >/dev/null 2>&1; then
    staged_files=$(git diff --cached --name-only 2>/dev/null || true)
    if [ -n "$staged_files" ]; then
        # Check for .env files
        if echo "$staged_files" | grep -q "\.env" && ! echo "$staged_files" | grep -q "\.env\.example"; then
            echo -e "${RED}✗ .env file should not be committed${NC}"
            echo "Use: git rm --cached .env"
            FAILED=1
        fi
        
        # Check for API keys
        if echo "$staged_files" | xargs grep -l "AKIA\|aws_secret\|password.*=\|api_key.*=" 2>/dev/null | grep -v "\.example\|sample" | head -1; then
            echo -e "${RED}✗ Possible secrets detected in staged files${NC}"
            FAILED=1
        fi
    fi
fi

# Check if required files exist
echo -e "\n${YELLOW}→ Checking required files...${NC}"
required_files=("README.md" "go.mod" "docker-compose.yml" ".gitignore")
for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo -e "${YELLOW}⚠ Missing: $file${NC}"
    else
        echo -e "${GREEN}✓ $file present${NC}"
    fi
done

# ===== FINAL RESULT =====
echo ""
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All pre-commit checks passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Pre-commit checks failed!${NC}"
    echo "Fix the issues above or use: git commit --no-verify (not recommended)"
    exit 1
fi
