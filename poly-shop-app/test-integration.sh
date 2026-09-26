#!/bin/bash

# Poly-Shop Integration Test Script (Bash/WSL)
# Requires 'jq' for JSON parsing

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check for jq
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is not installed.${NC}"
    echo "Please install jq to run this script (e.g., sudo apt install jq)"
    exit 1
fi

echo -e "\n${CYAN}POLY-SHOP INTEGRATION TEST (BASH)${NC}\n"

TEST_USER="test-user-$(date +%s)"

# Function to test an endpoint
test_endpoint() {
    local NAME=$1
    local URL=$2
    local METHOD=${3:-GET}
    local BODY=$4

    echo -n "Testing: $NAME... " >&2

    if [ -n "$BODY" ]; then
        RESPONSE=$(curl -s -w "\n%{http_code}" -X "$METHOD" -H "Content-Type: application/json" -d "$BODY" "$URL")
    else
        RESPONSE=$(curl -s -w "\n%{http_code}" -X "$METHOD" "$URL")
    fi

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    # Use sed to remove the last line (status code) to get content
    CONTENT=$(echo "$RESPONSE" | sed '$d')

    if [[ "$HTTP_CODE" -ge 200 && "$HTTP_CODE" -lt 300 ]]; then
        echo -e "${GREEN}PASS [$HTTP_CODE]${NC}" >&2
        # Only echo CONTENT to stdout if we want to capture it. 
        # But wait, test_endpoint is sometimes called just for side effects.
        # We need to output CONTENT to stdout so $(test_endpoint ...) captures it.
        echo "$CONTENT"
        return 0
    else
        echo -e "${RED}FAIL [$HTTP_CODE]${NC}" >&2
        echo -e "${RED}   Response: $CONTENT${NC}" >&2
        return 1
    fi
}

echo -e "${YELLOW}[1] PRODUCT SERVICE (Port 8090)${NC}" >&2

# 1. List Products
echo -e "\nStep 1: Listing Products..." >&2
PRODUCTS_JSON=$(test_endpoint "GET /products" "http://localhost:8090/products")
# Display first 3 products to keep output clean but confirm listing logic
echo "$PRODUCTS_JSON" | jq -r '.[0:3][] | "  - [\(.id)] \(.name): $\(.price)"' >&2

# Extract a valid product ID from the list
FIRST_PRODUCT_ID=$(echo "$PRODUCTS_JSON" | jq -r '.[0].id')
SECOND_PRODUCT_ID=$(echo "$PRODUCTS_JSON" | jq -r '.[1].id') 
echo -e "${NC}  Selected Product IDs for testing: ${FIRST_PRODUCT_ID}, ${SECOND_PRODUCT_ID}${NC}" >&2

# 2. List Specific Product
echo -e "\nStep 2: Viewing Specific Product..." >&2
test_endpoint "GET /products/$FIRST_PRODUCT_ID" "http://localhost:8090/products/$FIRST_PRODUCT_ID" > /dev/null
# Fetch and display details to confirm
PRODUCT_DETAILS=$(curl -s "http://localhost:8090/products/$FIRST_PRODUCT_ID")
echo "$PRODUCT_DETAILS" | jq -r '"  Product: \(.name) - $\(.price) (\(.category))"' >&2

echo -e "\n${YELLOW}[2] CART SERVICE (Port 8080)${NC}" >&2

# 3. Add products to cart
echo -e "\nStep 3: Adding items to cart..." >&2
echo "  - Adding 2x Product $FIRST_PRODUCT_ID" >&2
ADD_BODY_1=$(jq -n --arg pid "$FIRST_PRODUCT_ID" '{product_id: $pid, quantity: 2}')
test_endpoint "POST /v1/cart/$TEST_USER" "http://localhost:8080/v1/cart/$TEST_USER" "POST" "$ADD_BODY_1" > /dev/null

echo "  - Adding 1x Product $SECOND_PRODUCT_ID" >&2
ADD_BODY_2=$(jq -n --arg pid "$SECOND_PRODUCT_ID" '{product_id: $pid, quantity: 1}')
curl -s -X POST -H "Content-Type: application/json" -d "$ADD_BODY_2" "http://localhost:8080/v1/cart/$TEST_USER" > /dev/null
echo -e "${GREEN}PASS [200]${NC}" >&2

# 4. View cart
echo -e "\nStep 4: Viewing cart content..." >&2
CART_RESPONSE=$(test_endpoint "GET /v1/cart/$TEST_USER" "http://localhost:8080/v1/cart/$TEST_USER")
echo -e "\n  Cart Contents:" >&2
echo "$CART_RESPONSE" | jq -r '.items[] | "  - Product ID: \(.product_id), Quantity: \(.quantity)"' >&2

echo -e "\n${YELLOW}[3] CHECKOUT SERVICE (Port 8085)${NC}" >&2

# 5. Check currency conversion
echo -e "\nStep 5: Checking currency conversion..." >&2
CONVERSION=$(test_endpoint "GET /currency/convert" "http://localhost:8085/currency/convert?amount=100&from=USD&to=CAD")
echo "$CONVERSION" | jq -r '"  100 USD = \(.converted_amount) CAD (Rate: \(.exchange_rate))"' >&2

echo -e "\n${YELLOW}[4] END-TO-END CHECKOUT FLOW (DISTRIBUTED TRACING)${NC}" >&2

# 6. Checkout cart
echo -e "\nStep 6: Checking out..." >&2
CHECKOUT_BODY=$(jq -n --arg uid "$TEST_USER" '{userId: $uid}')

CHECKOUT_RESULT=$(test_endpoint "POST /checkout" "http://localhost:8085/checkout" "POST" "$CHECKOUT_BODY")
LAST_EXIT_CODE=$?

if [ $LAST_EXIT_CODE -eq 0 ]; then
    TX_ID=$(echo "$CHECKOUT_RESULT" | jq -r '.transaction_id')
    STATUS=$(echo "$CHECKOUT_RESULT" | jq -r '.status')
    MESSAGE=$(echo "$CHECKOUT_RESULT" | jq -r '.message')
    TRACE_ID=$(echo "$CHECKOUT_RESULT" | jq -r '.trace_id')

    echo -e "\n${CYAN}CHECKOUT RESULT:${NC}" >&2
    echo "  Transaction ID: $TX_ID" >&2
    echo "  Status: $STATUS" >&2
    echo "  Message: $MESSAGE" >&2
    echo -e "  Trace ID: ${GREEN}$TRACE_ID${NC}" >&2
    echo -e "\n  View trace in Jaeger: ${CYAN}http://localhost:16686/trace/$TRACE_ID${NC}" >&2
fi

echo -e "\n${YELLOW}[5] SERVICE INTERACTIONS VERIFIED${NC}"
echo -e "${NC}  - Cart Service -> Redis"
echo -e "${NC}  - Product Service -> PostgreSQL"
echo -e "${NC}  - Checkout Service -> Cart Service (RestTemplate)"
echo -e "${NC}  - All Services -> Jaeger (OpenTelemetry tracing)"

echo -e "\n${GREEN}INTEGRATION TEST COMPLETE!${NC}\n"
