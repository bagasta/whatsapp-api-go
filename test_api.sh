#!/bin/bash

# WhatsApp API - Testing Script
# Comprehensive API testing dengan contoh lengkap

# Configuration
BASE_URL="http://localhost:3000"
USER_ID="test_user_001"
AGENT_ID="test_bot_001"
API_KEY="test_key_12345"
PHONE="6281234567890"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
print_header() {
    echo ""
    echo -e "${BLUE}==================================="
    echo -e "$1"
    echo -e "===================================${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local headers=$5
    
    print_info "Testing: $name"
    echo "Request: $method $endpoint"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi
    
    local response
    if [ "$method" = "GET" ]; then
        response=$(curl -s $headers "$BASE_URL$endpoint")
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -s -X DELETE $headers "$BASE_URL$endpoint")
    else
        response=$(curl -s -X $method $headers -H "Content-Type: application/json" -d "$data" "$BASE_URL$endpoint")
    fi
    
    echo "Response:"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo ""
    
    # Check if response contains error
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        print_error "Request failed"
        return 1
    else
        print_success "Request successful"
        return 0
    fi
}

# Main Testing Flow
print_header "WhatsApp API - Comprehensive Testing"

# 1. Health Check
print_header "1. Health Check"
test_endpoint "Health Check" "GET" "/health"

# 2. Metrics
print_header "2. Metrics"
test_endpoint "Metrics" "GET" "/metrics"

# 3. Create Session
print_header "3. Create Session"
SESSION_DATA='{
  "userId": "'"$USER_ID"'",
  "agentId": "'"$AGENT_ID"'",
  "agentName": "Test Support Bot"
}'
test_endpoint "Create Session" "POST" "/sessions" "$SESSION_DATA"

# Wait for session to initialize
print_info "Waiting 3 seconds for session initialization..."
sleep 3

# 4. Get Session Status
print_header "4. Get Session Status"
test_endpoint "Get Session" "GET" "/sessions/$AGENT_ID"

# 5. Get QR Code
print_header "5. Get QR Code"
print_info "Getting QR code for WhatsApp authentication..."
QR_RESPONSE=$(curl -s -X POST "$BASE_URL/sessions/$AGENT_ID/qr")
echo "$QR_RESPONSE" | jq '.'

if echo "$QR_RESPONSE" | jq -e '.qr.base64' > /dev/null 2>&1; then
    print_success "QR code received"
    print_info "Saving QR code to qr_code.png..."
    echo "$QR_RESPONSE" | jq -r '.qr.base64' | base64 -d > qr_code.png 2>/dev/null
    if [ -f qr_code.png ]; then
        print_success "QR code saved to qr_code.png"
        print_info "Scan this QR code with WhatsApp to authenticate"
    fi
else
    print_info "No QR code (session might be already authenticated or not ready)"
fi
echo ""

# 6. Send Message (will fail if not authenticated)
print_header "6. Send Message"
MESSAGE_DATA='{
  "to": "'"$PHONE"'",
  "message": "Hello from WhatsApp API Test!"
}'
test_endpoint "Send Message" "POST" "/agents/$AGENT_ID/messages" "$MESSAGE_DATA" "-H 'Authorization: Bearer $API_KEY'"

# 7. Send Message with Quote
print_header "7. Send Message with Quote"
QUOTE_DATA='{
  "to": "'"$PHONE"'",
  "message": "This is a quoted reply",
  "quotedMessageId": "true_6281234567890@c.us_EXAMPLE123"
}'
test_endpoint "Send Quoted Message" "POST" "/agents/$AGENT_ID/messages" "$QUOTE_DATA" "-H 'Authorization: Bearer $API_KEY'"

# 8. Send Media (Base64)
print_header "8. Send Media (Base64)"
# Tiny 1x1 PNG image
TINY_PNG="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
MEDIA_DATA='{
  "to": "'"$PHONE"'",
  "data": "'"$TINY_PNG"'",
  "caption": "Test image from API",
  "filename": "test.png",
  "mimeType": "image/png"
}'
test_endpoint "Send Media (Base64)" "POST" "/agents/$AGENT_ID/media" "$MEDIA_DATA" "-H 'Authorization: Bearer $API_KEY'"

# 9. Send Media (URL)
print_header "9. Send Media (URL)"
URL_MEDIA_DATA='{
  "to": "'"$PHONE"'",
  "url": "https://picsum.photos/200/300",
  "caption": "Random image from URL"
}'
test_endpoint "Send Media (URL)" "POST" "/agents/$AGENT_ID/media" "$URL_MEDIA_DATA" "-H 'Authorization: Bearer $API_KEY'"

# 10. Execute AI Agent
print_header "10. Execute AI Agent"
AI_DATA='{
  "input": "Hello AI, how are you?",
  "session_id": "'"$PHONE"'",
  "parameters": {
    "max_steps": 5
  }
}'
test_endpoint "Execute AI" "POST" "/agents/$AGENT_ID/run" "$AI_DATA" "-H 'Authorization: Bearer $API_KEY'"

# 11. Test Error Cases
print_header "11. Testing Error Cases"

print_info "Test 1: Invalid payload (missing required fields)"
test_endpoint "Invalid Session Creation" "POST" "/sessions" '{"userId": "test"}'

print_info "Test 2: Unauthorized (no Bearer token)"
test_endpoint "Unauthorized Message" "POST" "/agents/$AGENT_ID/messages" "$MESSAGE_DATA"

print_info "Test 3: Invalid agent ID"
test_endpoint "Invalid Agent" "GET" "/sessions/invalid_agent_id_xyz"

# 12. Reconnect Session
print_header "12. Reconnect Session"
test_endpoint "Reconnect Session" "POST" "/sessions/$AGENT_ID/reconnect"

# 13. Multiple Agents Test
print_header "13. Multiple Agents Test"

print_info "Creating Agent 2..."
AGENT2_DATA='{
  "userId": "'"$USER_ID"'",
  "agentId": "test_bot_002",
  "agentName": "Second Test Bot"
}'
test_endpoint "Create Agent 2" "POST" "/sessions" "$AGENT2_DATA"

print_info "Creating Agent 3..."
AGENT3_DATA='{
  "userId": "'"$USER_ID"'",
  "agentId": "test_bot_003",
  "agentName": "Third Test Bot"
}'
test_endpoint "Create Agent 3" "POST" "/sessions" "$AGENT3_DATA"

print_info "Getting all agent statuses..."
test_endpoint "Get Agent 1 Status" "GET" "/sessions/$AGENT_ID"
test_endpoint "Get Agent 2 Status" "GET" "/sessions/test_bot_002"
test_endpoint "Get Agent 3 Status" "GET" "/sessions/test_bot_003"

# 14. Cleanup (optional)
print_header "14. Cleanup (Optional)"
read -p "Do you want to delete test sessions? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    test_endpoint "Delete Agent 1" "DELETE" "/sessions/$AGENT_ID"
    test_endpoint "Delete Agent 2" "DELETE" "/sessions/test_bot_002"
    test_endpoint "Delete Agent 3" "DELETE" "/sessions/test_bot_003"
    print_success "Cleanup completed"
else
    print_info "Cleanup skipped"
fi

# Summary
print_header "Testing Summary"
echo "Base URL: $BASE_URL"
echo "User ID: $USER_ID"
echo "Agent ID: $AGENT_ID"
echo "Phone: $PHONE"
echo ""
print_success "All tests completed!"
echo ""
echo "Next steps:"
echo "1. Check qr_code.png and scan with WhatsApp"
echo "2. Monitor logs: tail -f /var/log/whatsapp-api.log"
echo "3. Check database: psql -h localhost -U whatsapp_user -d whatsapp"
echo ""
