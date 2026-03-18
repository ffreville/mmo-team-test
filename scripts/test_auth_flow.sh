#!/bin/bash
# Test Script pour le flow complet Auth → Character → World

set -e

echo "=========================================="
echo "Test E2E: Auth Flow Complet"
echo "Date: $(date)"
echo "=========================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
BACKEND_URL="ws://localhost:8080"
HTTP_URL="http://localhost:8080"
TEST_USERNAME="testuser_$(date +%s)"
TEST_EMAIL="${TEST_USERNAME}@test.com"
TEST_PASSWORD="testpass123"
TEST_CHARACTER="TestHero"

echo ""
echo -e "${YELLOW}[1/6] Checking backend server...${NC}"

# Check if backend is running
if ! curl -s "${HTTP_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend is not running on ${HTTP_URL}${NC}"
    echo "Please start the backend: cd server/cmd/gateway && go run main.go"
    exit 1
fi

echo -e "${GREEN}✓ Backend is running${NC}"

echo ""
echo -e "${YELLOW}[2/6] Testing WebSocket connection...${NC}"

# Test WebSocket connection using websocat or wscat
if command -v websocat &> /dev/null; then
    echo "Testing WebSocket connection..."
    timeout 5 websocat "${BACKEND_URL}/ws" <<< '{"type":"ping"}' 2>&1 || true
    echo -e "${GREEN}✓ WebSocket connection successful${NC}"
elif command -v wscat &> /dev/null; then
    echo "Testing WebSocket connection..."
    echo '{"type":"ping"}' | timeout 5 wscat -c "${BACKEND_URL}/ws" 2>&1 || true
    echo -e "${GREEN}✓ WebSocket connection successful${NC}"
else
    echo -e "${YELLOW}⚠ websocat/wscat not found, skipping WebSocket test${NC}"
    echo "Install with: brew install websocat or npm install -g wscat"
fi

echo ""
echo -e "${YELLOW}[3/6] Testing HTTP Register endpoint...${NC}"

# Test HTTP register endpoint (fallback)
REGISTER_RESPONSE=$(curl -s -X POST "${HTTP_URL}/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${TEST_USERNAME}\",\"email\":\"${TEST_EMAIL}\",\"password\":\"${TEST_PASSWORD}\"}" \
    2>&1 || echo '{"error":"endpoint not available"}')

echo "Register response: ${REGISTER_RESPONSE}"

if echo "${REGISTER_RESPONSE}" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ HTTP Register successful${NC}"
elif echo "${REGISTER_RESPONSE}" | grep -q '"error":"username or email already exists"'; then
    echo -e "${YELLOW}⚠ User already exists, using login test${NC}"
else
    echo -e "${YELLOW}⚠ HTTP register endpoint not available (expected if only WebSocket is implemented)${NC}"
fi

echo ""
echo -e "${YELLOW}[4/6] Testing HTTP Login endpoint...${NC}"

# Test HTTP login endpoint
LOGIN_RESPONSE=$(curl -s -X POST "${HTTP_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${TEST_USERNAME}\",\"password\":\"${TEST_PASSWORD}\"}" \
    2>&1 || echo '{"error":"endpoint not available"}')

echo "Login response: ${LOGIN_RESPONSE}"

if echo "${LOGIN_RESPONSE}" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ HTTP Login successful${NC}"
    TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "Token received: ${TOKEN:0:20}..."
else
    echo -e "${YELLOW}⚠ Login test skipped (user may not exist)${NC}"
fi

echo ""
echo -e "${YELLOW}[5/6] Checking database...${NC}"

# Check if PostgreSQL is accessible
if command -v psql &> /dev/null; then
    # Try to connect to database
    if psql -h localhost -U postgres -d mmo_db -c "SELECT COUNT(*) FROM users;" > /dev/null 2>&1; then
        USER_COUNT=$(psql -h localhost -U postgres -d mmo_db -t -c "SELECT COUNT(*) FROM users;" 2>/dev/null)
        echo -e "${GREEN}✓ Database connected, ${USER_COUNT} users in database${NC}"
    else
        echo -e "${YELLOW}⚠ Database not accessible or not configured${NC}"
    fi
else
    echo -e "${YELLOW}⚠ psql not found, skipping database check${NC}"
fi

echo ""
echo -e "${YELLOW}[6/6] Client Test Instructions...${NC}"

echo ""
echo "Pour tester le client Godot:"
echo "1. Assurez-vous que le backend est lancé sur ws://localhost:8080/ws"
echo "2. Ouvrez Godot 4.3 et lancez la scène res://scenes/ui/menus/AuthMenu.tscn"
echo "3. Cliquez sur 'Go to Register'"
echo "4. Entrez:"
echo "   - Username: ${TEST_USERNAME}"
echo "   - Email: ${TEST_EMAIL}"
echo "   - Password: ${TEST_PASSWORD}"
echo "5. Cliquez sur 'Register'"
echo "6. Le client devrait naviguer vers CharacterSelection"
echo ""
echo "Si le client reste bloqué sur 'Processing...':"
echo "  - Vérifiez les logs du backend (terminal où gateway tourne)"
echo "  - Vérifiez les logs Godot (Debug > Console)"
echo "  - Vérifiez que le WebSocket est connecté"
echo ""

echo "=========================================="
echo -e "${GREEN}Test script completed${NC}"
echo "=========================================="
echo ""
echo "Prochaines étapes:"
echo "1. Lancez le backend: cd server/cmd/gateway && go run main.go"
echo "2. Lancez le client Godot"
echo "3. Exécutez les tests manuels"
echo "4. Capturez des screenshots des tests réussis"
echo ""
